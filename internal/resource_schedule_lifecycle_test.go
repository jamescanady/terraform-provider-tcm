//go:build integration

package internal_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type mockSchedule struct {
	ID                         string           `json:"id"`
	Name                       *string          `json:"name,omitempty"`
	Description                *string          `json:"description,omitempty"`
	TenantProductEnvironmentID string           `json:"tenantProductEnvironmentId"`
	ScheduleCategoryID         string           `json:"scheduleCategoryId"`
	Enabled                    bool             `json:"enabled"`
	ScheduleDetails            *json.RawMessage `json:"scheduleDetails,omitempty"`
}

func newScheduleMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	var (
		mu      sync.Mutex
		counter atomic.Int64
		store   = map[string]*mockSchedule{}
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/Schedule", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p struct {
			Name                       *string          `json:"name"`
			Description                *string          `json:"description"`
			TenantProductEnvironmentID string           `json:"tenantProductEnvironmentId"`
			ScheduleCategoryID         string           `json:"scheduleCategoryId"`
			Enabled                    bool             `json:"enabled"`
			ScheduleDetails            *json.RawMessage `json:"scheduleDetails"`
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id := fmt.Sprintf("sched-%d", counter.Add(1))
		item := &mockSchedule{
			ID:                         id,
			Name:                       p.Name,
			Description:                p.Description,
			TenantProductEnvironmentID: p.TenantProductEnvironmentID,
			ScheduleCategoryID:         p.ScheduleCategoryID,
			Enabled:                    p.Enabled,
			ScheduleDetails:            p.ScheduleDetails,
		}
		mu.Lock()
		store[id] = item
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(item) //nolint:errcheck
	})

	mux.HandleFunc("/v1/Schedule/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/Schedule/")
		switch r.Method {
		case http.MethodGet:
			mu.Lock()
			item, ok := store[id]
			var snap mockSchedule
			if ok {
				snap = *item
			}
			mu.Unlock()
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(&snap) //nolint:errcheck

		case http.MethodPut:
			var p struct {
				Name                       *string          `json:"name"`
				Description                *string          `json:"description"`
				TenantProductEnvironmentID string           `json:"tenantProductEnvironmentId"`
				ScheduleCategoryID         string           `json:"scheduleCategoryId"`
				Enabled                    bool             `json:"enabled"`
				ScheduleDetails            *json.RawMessage `json:"scheduleDetails"`
			}
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mu.Lock()
			item, ok := store[id]
			var updated mockSchedule
			if ok {
				item.Name = p.Name
				item.Description = p.Description
				item.TenantProductEnvironmentID = p.TenantProductEnvironmentID
				item.ScheduleCategoryID = p.ScheduleCategoryID
				item.Enabled = p.Enabled
				item.ScheduleDetails = p.ScheduleDetails
				updated = *item
			}
			mu.Unlock()
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(&updated) //nolint:errcheck

		case http.MethodDelete:
			mu.Lock()
			_, ok := store[id]
			if ok {
				delete(store, id)
			}
			mu.Unlock()
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return httptest.NewServer(mux)
}

func TestAccScheduleResource_lifecycle(t *testing.T) {
	srv := newScheduleMockServer(t)
	t.Cleanup(srv.Close)

	providerConfig := fmt.Sprintf(`
provider "tcm" {
  base_url = %q
  token    = "test-token"
}
`, srv.URL)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories(),
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "tcm_schedule" "test" {
  tenant_product_environment_id = "tpe-001"
  schedule_category_id          = "cat-001"
  name                          = "My Schedule"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tcm_schedule.test", "id"),
					resource.TestCheckResourceAttr("tcm_schedule.test", "tenant_product_environment_id", "tpe-001"),
					resource.TestCheckResourceAttr("tcm_schedule.test", "schedule_category_id", "cat-001"),
					resource.TestCheckResourceAttr("tcm_schedule.test", "name", "My Schedule"),
					resource.TestCheckResourceAttr("tcm_schedule.test", "enabled", "false"),
				),
			},
			{
				Config: providerConfig + `
resource "tcm_schedule" "test" {
  tenant_product_environment_id = "tpe-001"
  schedule_category_id          = "cat-001"
  name                          = "Updated Schedule"
  description                   = "A schedule description"
  enabled                       = true
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tcm_schedule.test", "name", "Updated Schedule"),
					resource.TestCheckResourceAttr("tcm_schedule.test", "description", "A schedule description"),
					resource.TestCheckResourceAttr("tcm_schedule.test", "enabled", "true"),
				),
			},
			{
				ResourceName:      "tcm_schedule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccScheduleDataSource_read(t *testing.T) {
	srv := newScheduleMockServer(t)
	t.Cleanup(srv.Close)

	providerConfig := fmt.Sprintf(`
provider "tcm" {
  base_url = %q
  token    = "test-token"
}
`, srv.URL)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories(),
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "tcm_schedule" "seed" {
  tenant_product_environment_id = "tpe-ds"
  schedule_category_id          = "cat-ds"
  name                          = "DataSource Schedule"
}

data "tcm_schedule" "test" {
  id = tcm_schedule.seed.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tcm_schedule.test", "id", "tcm_schedule.seed", "id"),
					resource.TestCheckResourceAttr("data.tcm_schedule.test", "name", "DataSource Schedule"),
					resource.TestCheckResourceAttr("data.tcm_schedule.test", "tenant_product_environment_id", "tpe-ds"),
					resource.TestCheckResourceAttr("data.tcm_schedule.test", "schedule_category_id", "cat-ds"),
					resource.TestCheckResourceAttr("data.tcm_schedule.test", "enabled", "false"),
				),
			},
		},
	})
}
