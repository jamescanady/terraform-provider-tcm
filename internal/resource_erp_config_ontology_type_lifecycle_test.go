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

type mockERPConfigOntologyType struct {
	ID             string  `json:"id"`
	ERPConfigID    *string `json:"erpConfigId,omitempty"`
	OntologyTypeID *string `json:"ontologyTypeId,omitempty"`
}

func newERPConfigOntologyTypeMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	var (
		mu      sync.Mutex
		counter atomic.Int64
		store   = map[string]*mockERPConfigOntologyType{}
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/ERPConfigOntologyType", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p struct {
			ERPConfigID    *string `json:"erpConfigId"`
			OntologyTypeID *string `json:"ontologyTypeId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id := fmt.Sprintf("ecot-%d", counter.Add(1))
		item := &mockERPConfigOntologyType{
			ID:             id,
			ERPConfigID:    p.ERPConfigID,
			OntologyTypeID: p.OntologyTypeID,
		}
		mu.Lock()
		store[id] = item
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(item) //nolint:errcheck
	})

	mux.HandleFunc("/v1/ERPConfigOntologyType/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/ERPConfigOntologyType/")
		switch r.Method {
		case http.MethodGet:
			mu.Lock()
			item, ok := store[id]
			var snap mockERPConfigOntologyType
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
				ERPConfigID    *string `json:"erpConfigId"`
				OntologyTypeID *string `json:"ontologyTypeId"`
			}
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mu.Lock()
			item, ok := store[id]
			var updated mockERPConfigOntologyType
			if ok {
				item.ERPConfigID = p.ERPConfigID
				item.OntologyTypeID = p.OntologyTypeID
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

func TestAccERPConfigOntologyTypeResource_lifecycle(t *testing.T) {
	srv := newERPConfigOntologyTypeMockServer(t)
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
resource "tcm_erp_config_ontology_type" "test" {
  erp_config_id    = "erp-001"
  ontology_type_id = "ot-001"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tcm_erp_config_ontology_type.test", "id"),
					resource.TestCheckResourceAttr("tcm_erp_config_ontology_type.test", "erp_config_id", "erp-001"),
					resource.TestCheckResourceAttr("tcm_erp_config_ontology_type.test", "ontology_type_id", "ot-001"),
				),
			},
			{
				Config: providerConfig + `
resource "tcm_erp_config_ontology_type" "test" {
  erp_config_id    = "erp-002"
  ontology_type_id = "ot-002"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tcm_erp_config_ontology_type.test", "erp_config_id", "erp-002"),
					resource.TestCheckResourceAttr("tcm_erp_config_ontology_type.test", "ontology_type_id", "ot-002"),
				),
			},
			{
				ResourceName:      "tcm_erp_config_ontology_type.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccERPConfigOntologyTypeDataSource_read(t *testing.T) {
	srv := newERPConfigOntologyTypeMockServer(t)
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
resource "tcm_erp_config_ontology_type" "seed" {
  erp_config_id    = "erp-ds"
  ontology_type_id = "ot-ds"
}

data "tcm_erp_config_ontology_type" "test" {
  id = tcm_erp_config_ontology_type.seed.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tcm_erp_config_ontology_type.test", "id", "tcm_erp_config_ontology_type.seed", "id"),
					resource.TestCheckResourceAttr("data.tcm_erp_config_ontology_type.test", "erp_config_id", "erp-ds"),
					resource.TestCheckResourceAttr("data.tcm_erp_config_ontology_type.test", "ontology_type_id", "ot-ds"),
				),
			},
		},
	})
}
