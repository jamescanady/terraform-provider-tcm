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

type mockTenantProductEnvironment struct {
	ID                   string  `json:"id"`
	TenantID             string  `json:"tenantId"`
	ProductEnvironmentID string  `json:"productEnvironmentId"`
	NamespaceID          *string `json:"namespaceId,omitempty"`
	ProductTenantCode    *string `json:"productTenantCode,omitempty"`
	ProductAlias         *string `json:"productAlias,omitempty"`
	IsDisabled           bool    `json:"isDisabled"`
}

func newTenantProductEnvironmentMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	var (
		mu      sync.Mutex
		counter atomic.Int64
		store   = map[string]*mockTenantProductEnvironment{}
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/TenantProductEnvironment", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p struct {
			TenantID             string  `json:"tenantId"`
			ProductEnvironmentID string  `json:"productEnvironmentId"`
			NamespaceID          *string `json:"namespaceId"`
			ProductTenantCode    *string `json:"productTenantCode"`
			ProductAlias         *string `json:"productAlias"`
			IsDisabled           bool    `json:"isDisabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id := fmt.Sprintf("tpe-%d", counter.Add(1))
		tpe := &mockTenantProductEnvironment{
			ID:                   id,
			TenantID:             p.TenantID,
			ProductEnvironmentID: p.ProductEnvironmentID,
			NamespaceID:          p.NamespaceID,
			ProductTenantCode:    p.ProductTenantCode,
			ProductAlias:         p.ProductAlias,
			IsDisabled:           p.IsDisabled,
		}
		mu.Lock()
		store[id] = tpe
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tpe) //nolint:errcheck
	})

	mux.HandleFunc("/v1/TenantProductEnvironment/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/TenantProductEnvironment/")
		switch r.Method {
		case http.MethodGet:
			mu.Lock()
			tpe, ok := store[id]
			var snap mockTenantProductEnvironment
			if ok {
				snap = *tpe
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
				TenantID             string  `json:"tenantId"`
				ProductEnvironmentID string  `json:"productEnvironmentId"`
				NamespaceID          *string `json:"namespaceId"`
				ProductTenantCode    *string `json:"productTenantCode"`
				ProductAlias         *string `json:"productAlias"`
				IsDisabled           bool    `json:"isDisabled"`
			}
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mu.Lock()
			tpe, ok := store[id]
			var updated mockTenantProductEnvironment
			if ok {
				tpe.NamespaceID = p.NamespaceID
				tpe.ProductTenantCode = p.ProductTenantCode
				tpe.ProductAlias = p.ProductAlias
				tpe.IsDisabled = p.IsDisabled
				updated = *tpe
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

func TestAccTenantProductEnvironmentResource_lifecycle(t *testing.T) {
	srv := newTenantProductEnvironmentMockServer(t)
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
resource "tcm_tenant_product_environment" "test" {
  tenant_id              = "tenant-abc"
  product_environment_id = "penv-xyz"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tcm_tenant_product_environment.test", "id"),
					resource.TestCheckResourceAttr("tcm_tenant_product_environment.test", "tenant_id", "tenant-abc"),
					resource.TestCheckResourceAttr("tcm_tenant_product_environment.test", "product_environment_id", "penv-xyz"),
					resource.TestCheckResourceAttr("tcm_tenant_product_environment.test", "is_disabled", "false"),
				),
			},
			{
				Config: providerConfig + `
resource "tcm_tenant_product_environment" "test" {
  tenant_id              = "tenant-abc"
  product_environment_id = "penv-xyz"
  namespace_id           = "ns-001"
  product_tenant_code    = "PTC001"
  product_alias          = "alias-a"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tcm_tenant_product_environment.test", "tenant_id", "tenant-abc"),
					resource.TestCheckResourceAttr("tcm_tenant_product_environment.test", "product_environment_id", "penv-xyz"),
					resource.TestCheckResourceAttr("tcm_tenant_product_environment.test", "namespace_id", "ns-001"),
					resource.TestCheckResourceAttr("tcm_tenant_product_environment.test", "product_tenant_code", "PTC001"),
					resource.TestCheckResourceAttr("tcm_tenant_product_environment.test", "product_alias", "alias-a"),
					resource.TestCheckResourceAttr("tcm_tenant_product_environment.test", "is_disabled", "false"),
				),
			},
			{
				ResourceName:      "tcm_tenant_product_environment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTenantProductEnvironmentDataSource_read(t *testing.T) {
	srv := newTenantProductEnvironmentMockServer(t)
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
resource "tcm_tenant_product_environment" "seed" {
  tenant_id              = "tenant-ds"
  product_environment_id = "penv-ds"
  namespace_id           = "ns-ds"
  product_tenant_code    = "PTCDS"
}

data "tcm_tenant_product_environment" "test" {
  id = tcm_tenant_product_environment.seed.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tcm_tenant_product_environment.test", "id", "tcm_tenant_product_environment.seed", "id"),
					resource.TestCheckResourceAttr("data.tcm_tenant_product_environment.test", "tenant_id", "tenant-ds"),
					resource.TestCheckResourceAttr("data.tcm_tenant_product_environment.test", "product_environment_id", "penv-ds"),
					resource.TestCheckResourceAttr("data.tcm_tenant_product_environment.test", "namespace_id", "ns-ds"),
					resource.TestCheckResourceAttr("data.tcm_tenant_product_environment.test", "product_tenant_code", "PTCDS"),
					resource.TestCheckResourceAttr("data.tcm_tenant_product_environment.test", "is_disabled", "false"),
				),
			},
		},
	})
}
