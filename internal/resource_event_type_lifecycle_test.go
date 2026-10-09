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

type mockEventType struct {
	ID          string  `json:"id"`
	ProductID   string  `json:"productId"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	IsDisabled  bool    `json:"isDisabled"`
}

func newEventTypeMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	var (
		mu      sync.Mutex
		counter atomic.Int64
		store   = map[string]*mockEventType{}
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/EventType", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p struct {
			ProductID   string  `json:"productId"`
			Name        string  `json:"name"`
			Description *string `json:"description"`
			IsDisabled  bool    `json:"isDisabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id := fmt.Sprintf("et-%d", counter.Add(1))
		et := &mockEventType{
			ID:          id,
			ProductID:   p.ProductID,
			Name:        p.Name,
			Description: p.Description,
			IsDisabled:  p.IsDisabled,
		}
		mu.Lock()
		store[id] = et
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(et) //nolint:errcheck
	})

	mux.HandleFunc("/v1/EventType/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/EventType/")
		switch r.Method {
		case http.MethodGet:
			mu.Lock()
			et, ok := store[id]
			var snap mockEventType
			if ok {
				snap = *et
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
				ProductID   string  `json:"productId"`
				Name        string  `json:"name"`
				Description *string `json:"description"`
				IsDisabled  bool    `json:"isDisabled"`
			}
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mu.Lock()
			et, ok := store[id]
			var updated mockEventType
			if ok {
				et.ProductID = p.ProductID
				et.Name = p.Name
				et.Description = p.Description
				et.IsDisabled = p.IsDisabled
				updated = *et
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

func TestAccEventTypeResource_lifecycle(t *testing.T) {
	srv := newEventTypeMockServer(t)
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
resource "tcm_event_type" "test" {
  product_id = "product-001"
  name       = "TestEvent"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tcm_event_type.test", "id"),
					resource.TestCheckResourceAttr("tcm_event_type.test", "product_id", "product-001"),
					resource.TestCheckResourceAttr("tcm_event_type.test", "name", "TestEvent"),
					resource.TestCheckResourceAttr("tcm_event_type.test", "is_disabled", "false"),
				),
			},
			{
				Config: providerConfig + `
resource "tcm_event_type" "test" {
  product_id  = "product-001"
  name        = "TestEventUpdated"
  description = "An updated event type"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tcm_event_type.test", "name", "TestEventUpdated"),
					resource.TestCheckResourceAttr("tcm_event_type.test", "description", "An updated event type"),
					resource.TestCheckResourceAttr("tcm_event_type.test", "is_disabled", "false"),
				),
			},
			{
				ResourceName:      "tcm_event_type.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEventTypeDataSource_read(t *testing.T) {
	srv := newEventTypeMockServer(t)
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
resource "tcm_event_type" "seed" {
  product_id  = "product-ds"
  name        = "SeedEvent"
  description = "Seed event type"
}

data "tcm_event_type" "test" {
  id = tcm_event_type.seed.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tcm_event_type.test", "id", "tcm_event_type.seed", "id"),
					resource.TestCheckResourceAttr("data.tcm_event_type.test", "product_id", "product-ds"),
					resource.TestCheckResourceAttr("data.tcm_event_type.test", "name", "SeedEvent"),
					resource.TestCheckResourceAttr("data.tcm_event_type.test", "description", "Seed event type"),
					resource.TestCheckResourceAttr("data.tcm_event_type.test", "is_disabled", "false"),
				),
			},
		},
	})
}
