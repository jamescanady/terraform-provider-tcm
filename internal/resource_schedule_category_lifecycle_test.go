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

type mockScheduleCategory struct {
	ID          string  `json:"id"`
	Code        *string `json:"code,omitempty"`
	Description *string `json:"description,omitempty"`
	IsDisabled  bool    `json:"isDisabled"`
}

func newScheduleCategoryMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	var (
		mu      sync.Mutex
		counter atomic.Int64
		store   = map[string]*mockScheduleCategory{}
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/ScheduleCategory", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p struct {
			Code        *string `json:"code"`
			Description *string `json:"description"`
			IsDisabled  bool    `json:"isDisabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id := fmt.Sprintf("sc-%d", counter.Add(1))
		item := &mockScheduleCategory{
			ID:          id,
			Code:        p.Code,
			Description: p.Description,
			IsDisabled:  p.IsDisabled,
		}
		mu.Lock()
		store[id] = item
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(item) //nolint:errcheck
	})

	mux.HandleFunc("/v1/ScheduleCategory/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/ScheduleCategory/")
		switch r.Method {
		case http.MethodGet:
			mu.Lock()
			item, ok := store[id]
			var snap mockScheduleCategory
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
				Code        *string `json:"code"`
				Description *string `json:"description"`
				IsDisabled  bool    `json:"isDisabled"`
			}
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mu.Lock()
			item, ok := store[id]
			var updated mockScheduleCategory
			if ok {
				item.Code = p.Code
				item.Description = p.Description
				item.IsDisabled = p.IsDisabled
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

func TestAccScheduleCategoryResource_lifecycle(t *testing.T) {
	srv := newScheduleCategoryMockServer(t)
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
resource "tcm_schedule_category" "test" {
  code = "CAT001"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tcm_schedule_category.test", "id"),
					resource.TestCheckResourceAttr("tcm_schedule_category.test", "code", "CAT001"),
					resource.TestCheckResourceAttr("tcm_schedule_category.test", "is_disabled", "false"),
				),
			},
			{
				Config: providerConfig + `
resource "tcm_schedule_category" "test" {
  code        = "CAT001"
  description = "Updated category"
  is_disabled = true
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tcm_schedule_category.test", "code", "CAT001"),
					resource.TestCheckResourceAttr("tcm_schedule_category.test", "description", "Updated category"),
					resource.TestCheckResourceAttr("tcm_schedule_category.test", "is_disabled", "true"),
				),
			},
			{
				ResourceName:      "tcm_schedule_category.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccScheduleCategoryDataSource_read(t *testing.T) {
	srv := newScheduleCategoryMockServer(t)
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
resource "tcm_schedule_category" "seed" {
  code        = "DS-CAT"
  description = "DataSource Category"
}

data "tcm_schedule_category" "test" {
  id = tcm_schedule_category.seed.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tcm_schedule_category.test", "id", "tcm_schedule_category.seed", "id"),
					resource.TestCheckResourceAttr("data.tcm_schedule_category.test", "code", "DS-CAT"),
					resource.TestCheckResourceAttr("data.tcm_schedule_category.test", "description", "DataSource Category"),
					resource.TestCheckResourceAttr("data.tcm_schedule_category.test", "is_disabled", "false"),
				),
			},
		},
	})
}
