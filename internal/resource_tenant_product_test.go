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

type mockTenantProduct struct {
	ID                string  `json:"id"`
	TenantID          string  `json:"tenantId"`
	ProductID         string  `json:"productId"`
	TenantProductCode *string `json:"tenantProductCode,omitempty"`
	IsDisabled        bool    `json:"isDisabled"`
}

func newTenantProductMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	var (
		mu      sync.Mutex
		counter atomic.Int64
		store   = map[string]*mockTenantProduct{}
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/TenantProduct", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p struct {
			TenantID          string  `json:"tenantId"`
			ProductID         string  `json:"productId"`
			TenantProductCode *string `json:"tenantProductCode"`
			IsDisabled        bool    `json:"isDisabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id := fmt.Sprintf("tp-%d", counter.Add(1))
		tp := &mockTenantProduct{
			ID:                id,
			TenantID:          p.TenantID,
			ProductID:         p.ProductID,
			TenantProductCode: p.TenantProductCode,
			IsDisabled:        p.IsDisabled,
		}
		mu.Lock()
		store[id] = tp
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tp) //nolint:errcheck
	})

	mux.HandleFunc("/v1/TenantProduct/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/TenantProduct/")
		switch r.Method {
		case http.MethodGet:
			mu.Lock()
			tp, ok := store[id]
			var snap mockTenantProduct
			if ok {
				snap = *tp
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
				TenantProductCode *string `json:"tenantProductCode"`
				IsDisabled        bool    `json:"isDisabled"`
			}
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mu.Lock()
			tp, ok := store[id]
			var updated mockTenantProduct
			if ok {
				tp.TenantProductCode = p.TenantProductCode
				tp.IsDisabled = p.IsDisabled
				updated = *tp
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

func TestAccTenantProductResource_lifecycle(t *testing.T) {
	srv := newTenantProductMockServer(t)
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
resource "tcm_tenant_product" "test" {
  tenant_id  = "tenant-abc"
  product_id = "product-xyz"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tcm_tenant_product.test", "id"),
					resource.TestCheckResourceAttr("tcm_tenant_product.test", "tenant_id", "tenant-abc"),
					resource.TestCheckResourceAttr("tcm_tenant_product.test", "product_id", "product-xyz"),
					resource.TestCheckResourceAttr("tcm_tenant_product.test", "is_disabled", "false"),
				),
			},
			{
				Config: providerConfig + `
resource "tcm_tenant_product" "test" {
  tenant_id           = "tenant-abc"
  product_id          = "product-xyz"
  tenant_product_code = "TPC001"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tcm_tenant_product.test", "tenant_id", "tenant-abc"),
					resource.TestCheckResourceAttr("tcm_tenant_product.test", "product_id", "product-xyz"),
					resource.TestCheckResourceAttr("tcm_tenant_product.test", "tenant_product_code", "TPC001"),
					resource.TestCheckResourceAttr("tcm_tenant_product.test", "is_disabled", "false"),
				),
			},
			{
				ResourceName:      "tcm_tenant_product.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTenantProductDataSource_read(t *testing.T) {
	srv := newTenantProductMockServer(t)
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
resource "tcm_tenant_product" "seed" {
  tenant_id           = "tenant-ds"
  product_id          = "product-ds"
  tenant_product_code = "TPCDS"
}

data "tcm_tenant_product" "test" {
  id = tcm_tenant_product.seed.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tcm_tenant_product.test", "id", "tcm_tenant_product.seed", "id"),
					resource.TestCheckResourceAttr("data.tcm_tenant_product.test", "tenant_id", "tenant-ds"),
					resource.TestCheckResourceAttr("data.tcm_tenant_product.test", "product_id", "product-ds"),
					resource.TestCheckResourceAttr("data.tcm_tenant_product.test", "tenant_product_code", "TPCDS"),
					resource.TestCheckResourceAttr("data.tcm_tenant_product.test", "is_disabled", "false"),
				),
			},
		},
	})
}
