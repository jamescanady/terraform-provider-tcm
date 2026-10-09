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

type mockEHREndpoint struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	IsDisabled  bool    `json:"isDisabled"`
}

func newEHREndpointMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	var (
		mu      sync.Mutex
		counter atomic.Int64
		store   = map[string]*mockEHREndpoint{}
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/ehr/endpoint", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p struct {
			Name        string  `json:"name"`
			Description *string `json:"description"`
			IsDisabled  bool    `json:"isDisabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id := fmt.Sprintf("ep-%d", counter.Add(1))
		ep := &mockEHREndpoint{
			ID:          id,
			Name:        p.Name,
			Description: p.Description,
			IsDisabled:  p.IsDisabled,
		}
		mu.Lock()
		store[id] = ep
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ep) //nolint:errcheck
	})

	mux.HandleFunc("/v1/ehr/endpoint/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/ehr/endpoint/")
		switch r.Method {
		case http.MethodGet:
			mu.Lock()
			ep, ok := store[id]
			var snap mockEHREndpoint
			if ok {
				snap = *ep
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
				Name        string  `json:"name"`
				Description *string `json:"description"`
				IsDisabled  bool    `json:"isDisabled"`
			}
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mu.Lock()
			ep, ok := store[id]
			var updated mockEHREndpoint
			if ok {
				ep.Name = p.Name
				ep.Description = p.Description
				ep.IsDisabled = p.IsDisabled
				updated = *ep
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

func TestAccEHREndpointResource_lifecycle(t *testing.T) {
	srv := newEHREndpointMockServer(t)
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
resource "tcm_ehr_endpoint" "test" {
  name = "test-endpoint"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tcm_ehr_endpoint.test", "id"),
					resource.TestCheckResourceAttr("tcm_ehr_endpoint.test", "name", "test-endpoint"),
					resource.TestCheckResourceAttr("tcm_ehr_endpoint.test", "is_disabled", "false"),
				),
			},
			{
				Config: providerConfig + `
resource "tcm_ehr_endpoint" "test" {
  name        = "test-endpoint-updated"
  description = "Updated description"
  is_disabled = true
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tcm_ehr_endpoint.test", "name", "test-endpoint-updated"),
					resource.TestCheckResourceAttr("tcm_ehr_endpoint.test", "description", "Updated description"),
					resource.TestCheckResourceAttr("tcm_ehr_endpoint.test", "is_disabled", "true"),
				),
			},
			{
				ResourceName:      "tcm_ehr_endpoint.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEHREndpointDataSource_read(t *testing.T) {
	srv := newEHREndpointMockServer(t)
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
resource "tcm_ehr_endpoint" "seed" {
  name        = "ds-endpoint"
  description = "Datasource test"
}

data "tcm_ehr_endpoint" "test" {
  id = tcm_ehr_endpoint.seed.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tcm_ehr_endpoint.test", "id", "tcm_ehr_endpoint.seed", "id"),
					resource.TestCheckResourceAttr("data.tcm_ehr_endpoint.test", "name", "ds-endpoint"),
					resource.TestCheckResourceAttr("data.tcm_ehr_endpoint.test", "description", "Datasource test"),
					resource.TestCheckResourceAttr("data.tcm_ehr_endpoint.test", "is_disabled", "false"),
				),
			},
		},
	})
}
