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

type mockERP struct {
	ID               string  `json:"id"`
	Name             *string `json:"name,omitempty"`
	Description      *string `json:"description,omitempty"`
	Version          *string `json:"version,omitempty"`
	IsDeleted        *bool   `json:"isDeleted,omitempty"`
	CreatedBy        *string `json:"createdBy,omitempty"`
	CreatedDate      *string `json:"createdDate,omitempty"`
	LastModifiedBy   *string `json:"lastModifiedBy,omitempty"`
	LastModifiedDate *string `json:"lastModifiedDate,omitempty"`
}

func newERPMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	var (
		mu      sync.Mutex
		counter atomic.Int64
		store   = map[string]*mockERP{}
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/ERP", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p struct {
			Name        *string `json:"name"`
			Description *string `json:"description"`
			Version     *string `json:"version"`
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id := fmt.Sprintf("erp-%d", counter.Add(1))
		erp := &mockERP{
			ID:          id,
			Name:        p.Name,
			Description: p.Description,
			Version:     p.Version,
		}
		mu.Lock()
		store[id] = erp
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(erp) //nolint:errcheck
	})

	mux.HandleFunc("/v1/ERP/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/ERP/")
		switch r.Method {
		case http.MethodGet:
			mu.Lock()
			erp, ok := store[id]
			var snap mockERP
			if ok {
				snap = *erp
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
				Version     *string `json:"version"`
			}
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mu.Lock()
			erp, ok := store[id]
			var updated mockERP
			if ok {
				erp.Name = p.Name
				erp.Description = p.Description
				erp.Version = p.Version
				updated = *erp
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

func TestAccERPResource_lifecycle(t *testing.T) {
	srv := newERPMockServer(t)
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
resource "tcm_erp" "test" {
  name        = "TestERP"
  description = "A test ERP"
  version     = "1.0"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tcm_erp.test", "id"),
					resource.TestCheckResourceAttr("tcm_erp.test", "name", "TestERP"),
					resource.TestCheckResourceAttr("tcm_erp.test", "description", "A test ERP"),
					resource.TestCheckResourceAttr("tcm_erp.test", "version", "1.0"),
					resource.TestCheckResourceAttr("tcm_erp.test", "is_deleted", "false"),
				),
			},
			{
				Config: providerConfig + `
resource "tcm_erp" "test" {
  name        = "UpdatedERP"
  description = "Updated description"
  version     = "2.0"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tcm_erp.test", "name", "UpdatedERP"),
					resource.TestCheckResourceAttr("tcm_erp.test", "description", "Updated description"),
					resource.TestCheckResourceAttr("tcm_erp.test", "version", "2.0"),
					resource.TestCheckResourceAttr("tcm_erp.test", "is_deleted", "false"),
				),
			},
			{
				ResourceName:      "tcm_erp.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccERPDataSource_read(t *testing.T) {
	srv := newERPMockServer(t)
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
resource "tcm_erp" "seed" {
  name        = "SeedERP"
  description = "A seed ERP"
  version     = "1.0"
}

data "tcm_erp" "test" {
  id = tcm_erp.seed.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tcm_erp.test", "id", "tcm_erp.seed", "id"),
					resource.TestCheckResourceAttr("data.tcm_erp.test", "name", "SeedERP"),
					resource.TestCheckResourceAttr("data.tcm_erp.test", "description", "A seed ERP"),
					resource.TestCheckResourceAttr("data.tcm_erp.test", "version", "1.0"),
					resource.TestCheckResourceAttr("data.tcm_erp.test", "is_deleted", "false"),
				),
			},
		},
	})
}
