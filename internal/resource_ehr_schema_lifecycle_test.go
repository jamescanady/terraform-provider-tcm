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

type mockEHRSchema struct {
	ID             string   `json:"id"`
	TenantID       *string  `json:"tenantId,omitempty"`
	EndpointID     *string  `json:"endpointId,omitempty"`
	RequiredFields []string `json:"requiredFields,omitempty"`
	IsDisabled     bool     `json:"isDisabled"`
}

func newEHRSchemaMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	var (
		mu      sync.Mutex
		counter atomic.Int64
		store   = map[string]*mockEHRSchema{}
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/ehr/schema", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p struct {
			TenantID       *string  `json:"tenantId"`
			EndpointID     *string  `json:"endpointId"`
			RequiredFields []string `json:"requiredFields"`
			IsDisabled     bool     `json:"isDisabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id := fmt.Sprintf("es-%d", counter.Add(1))
		es := &mockEHRSchema{
			ID:             id,
			TenantID:       p.TenantID,
			EndpointID:     p.EndpointID,
			RequiredFields: p.RequiredFields,
			IsDisabled:     p.IsDisabled,
		}
		mu.Lock()
		store[id] = es
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(es) //nolint:errcheck
	})

	mux.HandleFunc("/v1/ehr/schema/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/ehr/schema/")
		switch r.Method {
		case http.MethodGet:
			mu.Lock()
			es, ok := store[id]
			var snap mockEHRSchema
			if ok {
				snap = *es
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
				TenantID       *string  `json:"tenantId"`
				EndpointID     *string  `json:"endpointId"`
				RequiredFields []string `json:"requiredFields"`
				IsDisabled     bool     `json:"isDisabled"`
			}
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mu.Lock()
			es, ok := store[id]
			var updated mockEHRSchema
			if ok {
				es.TenantID = p.TenantID
				es.EndpointID = p.EndpointID
				es.RequiredFields = p.RequiredFields
				es.IsDisabled = p.IsDisabled
				updated = *es
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

func TestAccEHRSchemaResource_lifecycle(t *testing.T) {
	srv := newEHRSchemaMockServer(t)
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
resource "tcm_ehr_schema" "test" {
  required_fields = ["field1", "field2"]
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tcm_ehr_schema.test", "id"),
					resource.TestCheckResourceAttr("tcm_ehr_schema.test", "required_fields.#", "2"),
					resource.TestCheckResourceAttr("tcm_ehr_schema.test", "required_fields.0", "field1"),
					resource.TestCheckResourceAttr("tcm_ehr_schema.test", "required_fields.1", "field2"),
					resource.TestCheckResourceAttr("tcm_ehr_schema.test", "is_disabled", "false"),
				),
			},
			{
				Config: providerConfig + `
resource "tcm_ehr_schema" "test" {
  tenant_id       = "tenant-001"
  endpoint_id     = "ep-001"
  required_fields = ["field1", "field2", "field3"]
  is_disabled     = true
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tcm_ehr_schema.test", "tenant_id", "tenant-001"),
					resource.TestCheckResourceAttr("tcm_ehr_schema.test", "endpoint_id", "ep-001"),
					resource.TestCheckResourceAttr("tcm_ehr_schema.test", "required_fields.#", "3"),
					resource.TestCheckResourceAttr("tcm_ehr_schema.test", "required_fields.2", "field3"),
					resource.TestCheckResourceAttr("tcm_ehr_schema.test", "is_disabled", "true"),
				),
			},
			{
				ResourceName:      "tcm_ehr_schema.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEHRSchemaDataSource_read(t *testing.T) {
	srv := newEHRSchemaMockServer(t)
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
resource "tcm_ehr_schema" "seed" {
  tenant_id       = "tenant-ds"
  required_fields = ["fieldA"]
}

data "tcm_ehr_schema" "test" {
  id = tcm_ehr_schema.seed.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tcm_ehr_schema.test", "id", "tcm_ehr_schema.seed", "id"),
					resource.TestCheckResourceAttr("data.tcm_ehr_schema.test", "tenant_id", "tenant-ds"),
					resource.TestCheckResourceAttr("data.tcm_ehr_schema.test", "required_fields.#", "1"),
					resource.TestCheckResourceAttr("data.tcm_ehr_schema.test", "required_fields.0", "fieldA"),
					resource.TestCheckResourceAttr("data.tcm_ehr_schema.test", "is_disabled", "false"),
				),
			},
		},
	})
}
