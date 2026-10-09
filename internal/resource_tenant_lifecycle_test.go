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

type mockTenant struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	GlobalTenantCode string `json:"globalTenantCode"`
	TenantShortCode  string `json:"tenantShortCode,omitempty"`
	IsDisabled       bool   `json:"isDisabled"`
}

func newTenantMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	var (
		mu      sync.Mutex
		counter atomic.Int64
		store   = map[string]*mockTenant{}
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/Tenant", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p struct {
			Name             string `json:"name"`
			Description      string `json:"description"`
			GlobalTenantCode string `json:"globalTenantCode"`
			TenantShortCode  string `json:"tenantShortCode"`
			IsDisabled       bool   `json:"isDisabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id := fmt.Sprintf("tenant-%d", counter.Add(1))
		tenant := &mockTenant{
			ID:               id,
			Name:             p.Name,
			Description:      p.Description,
			GlobalTenantCode: p.GlobalTenantCode,
			TenantShortCode:  p.TenantShortCode,
			IsDisabled:       p.IsDisabled,
		}
		mu.Lock()
		store[id] = tenant
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tenant) //nolint:errcheck
	})

	mux.HandleFunc("/v1/Tenant/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/Tenant/")
		switch r.Method {
		case http.MethodGet:
			mu.Lock()
			tenant, ok := store[id]
			var snap mockTenant
			if ok {
				snap = *tenant
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
				Name             string `json:"name"`
				Description      string `json:"description"`
				GlobalTenantCode string `json:"globalTenantCode"`
				IsDisabled       bool   `json:"isDisabled"`
			}
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mu.Lock()
			tenant, ok := store[id]
			var updated mockTenant
			if ok {
				tenant.Name = p.Name
				tenant.Description = p.Description
				tenant.GlobalTenantCode = p.GlobalTenantCode
				tenant.IsDisabled = p.IsDisabled
				updated = *tenant
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

func TestAccTenantResource_lifecycle(t *testing.T) {
	srv := newTenantMockServer(t)
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
resource "tcm_tenant" "test" {
  name               = "Test Tenant"
  description        = "A test tenant"
  global_tenant_code = "GTC001"
  tenant_short_code  = "tsc001"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tcm_tenant.test", "id"),
					resource.TestCheckResourceAttr("tcm_tenant.test", "name", "Test Tenant"),
					resource.TestCheckResourceAttr("tcm_tenant.test", "description", "A test tenant"),
					resource.TestCheckResourceAttr("tcm_tenant.test", "global_tenant_code", "GTC001"),
					resource.TestCheckResourceAttr("tcm_tenant.test", "tenant_short_code", "tsc001"),
					resource.TestCheckResourceAttr("tcm_tenant.test", "is_disabled", "false"),
				),
			},
			{
				Config: providerConfig + `
resource "tcm_tenant" "test" {
  name               = "Updated Tenant"
  description        = "Updated description"
  global_tenant_code = "GTC002"
  tenant_short_code  = "tsc001"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tcm_tenant.test", "name", "Updated Tenant"),
					resource.TestCheckResourceAttr("tcm_tenant.test", "description", "Updated description"),
					resource.TestCheckResourceAttr("tcm_tenant.test", "global_tenant_code", "GTC002"),
					resource.TestCheckResourceAttr("tcm_tenant.test", "tenant_short_code", "tsc001"),
					resource.TestCheckResourceAttr("tcm_tenant.test", "is_disabled", "false"),
				),
			},
			{
				ResourceName:      "tcm_tenant.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTenantResource_shortCodeRequiresReplace(t *testing.T) {
	srv := newTenantMockServer(t)
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
resource "tcm_tenant" "test" {
  name               = "Tenant A"
  description        = "First tenant"
  global_tenant_code = "GTC001"
  tenant_short_code  = "tsca"
}
`,
				Check: resource.TestCheckResourceAttr("tcm_tenant.test", "tenant_short_code", "tsca"),
			},
			{
				Config: providerConfig + `
resource "tcm_tenant" "test" {
  name               = "Tenant B"
  description        = "Second tenant"
  global_tenant_code = "GTC001"
  tenant_short_code  = "tscb"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tcm_tenant.test", "tenant_short_code", "tscb"),
					resource.TestCheckResourceAttr("tcm_tenant.test", "name", "Tenant B"),
				),
			},
		},
	})
}

func TestAccTenantDataSource_read(t *testing.T) {
	srv := newTenantMockServer(t)
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
resource "tcm_tenant" "seed" {
  name               = "Seed Tenant"
  description        = "For datasource test"
  global_tenant_code = "GTC_DS"
  tenant_short_code  = "tscds"
}

data "tcm_tenant" "test" {
  id = tcm_tenant.seed.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tcm_tenant.test", "id", "tcm_tenant.seed", "id"),
					resource.TestCheckResourceAttr("data.tcm_tenant.test", "name", "Seed Tenant"),
					resource.TestCheckResourceAttr("data.tcm_tenant.test", "description", "For datasource test"),
					resource.TestCheckResourceAttr("data.tcm_tenant.test", "global_tenant_code", "GTC_DS"),
					resource.TestCheckResourceAttr("data.tcm_tenant.test", "tenant_short_code", "tscds"),
					resource.TestCheckResourceAttr("data.tcm_tenant.test", "is_disabled", "false"),
				),
			},
		},
	})
}
