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

type mockOntologyType struct {
	ID          string  `json:"id"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	IsDisabled  bool    `json:"isDisabled"`
}

func newOntologyTypeMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	var (
		mu      sync.Mutex
		counter atomic.Int64
		store   = map[string]*mockOntologyType{}
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/OntologyType", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p struct {
			Name        *string `json:"name"`
			Description *string `json:"description"`
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id := fmt.Sprintf("ot-%d", counter.Add(1))
		item := &mockOntologyType{
			ID:          id,
			Name:        p.Name,
			Description: p.Description,
			IsDisabled:  false,
		}
		mu.Lock()
		store[id] = item
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(item) //nolint:errcheck
	})

	mux.HandleFunc("/v1/OntologyType/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/OntologyType/")
		switch r.Method {
		case http.MethodGet:
			mu.Lock()
			item, ok := store[id]
			var snap mockOntologyType
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
				Name        *string `json:"name"`
				Description *string `json:"description"`
			}
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mu.Lock()
			item, ok := store[id]
			var updated mockOntologyType
			if ok {
				item.Name = p.Name
				item.Description = p.Description
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

func TestAccOntologyTypeResource_lifecycle(t *testing.T) {
	srv := newOntologyTypeMockServer(t)
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
resource "tcm_ontology_type" "test" {
  name = "My Ontology"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tcm_ontology_type.test", "id"),
					resource.TestCheckResourceAttr("tcm_ontology_type.test", "name", "My Ontology"),
					resource.TestCheckResourceAttr("tcm_ontology_type.test", "is_disabled", "false"),
				),
			},
			{
				Config: providerConfig + `
resource "tcm_ontology_type" "test" {
  name        = "Updated Ontology"
  description = "An ontology description"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tcm_ontology_type.test", "name", "Updated Ontology"),
					resource.TestCheckResourceAttr("tcm_ontology_type.test", "description", "An ontology description"),
					resource.TestCheckResourceAttr("tcm_ontology_type.test", "is_disabled", "false"),
				),
			},
			{
				ResourceName:      "tcm_ontology_type.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOntologyTypeDataSource_read(t *testing.T) {
	srv := newOntologyTypeMockServer(t)
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
resource "tcm_ontology_type" "seed" {
  name        = "DS Ontology"
  description = "DataSource Ontology"
}

data "tcm_ontology_type" "test" {
  id = tcm_ontology_type.seed.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tcm_ontology_type.test", "id", "tcm_ontology_type.seed", "id"),
					resource.TestCheckResourceAttr("data.tcm_ontology_type.test", "name", "DS Ontology"),
					resource.TestCheckResourceAttr("data.tcm_ontology_type.test", "description", "DataSource Ontology"),
					resource.TestCheckResourceAttr("data.tcm_ontology_type.test", "is_disabled", "false"),
				),
			},
		},
	})
}
