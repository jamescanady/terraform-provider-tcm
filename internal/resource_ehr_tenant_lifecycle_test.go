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

type mockEHRTenant struct {
	ID                 string  `json:"id"`
	TenantID           *string `json:"tenantId,omitempty"`
	Name               string  `json:"name"`
	Description        *string `json:"description,omitempty"`
	RedoxEHRIdentifier string  `json:"redoxEHRIdentifier"`
	SymplrURLSlug      string  `json:"symplrUrlSlug"`
	RedoxEnvironment   string  `json:"redoxEnvironment"`
	IsDisabled         bool    `json:"isDisabled"`
}

func newEHRTenantMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	var (
		mu      sync.Mutex
		counter atomic.Int64
		store   = map[string]*mockEHRTenant{}
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/ehr/tenant", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p struct {
			TenantID           *string `json:"tenantId"`
			Name               string  `json:"name"`
			Description        *string `json:"description"`
			RedoxEHRIdentifier string  `json:"redoxEHRIdentifier"`
			SymplrURLSlug      string  `json:"symplrUrlSlug"`
			RedoxEnvironment   string  `json:"redoxEnvironment"`
			IsDisabled         bool    `json:"isDisabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id := fmt.Sprintf("tn-%d", counter.Add(1))
		tn := &mockEHRTenant{
			ID:                 id,
			TenantID:           p.TenantID,
			Name:               p.Name,
			Description:        p.Description,
			RedoxEHRIdentifier: p.RedoxEHRIdentifier,
			SymplrURLSlug:      p.SymplrURLSlug,
			RedoxEnvironment:   p.RedoxEnvironment,
			IsDisabled:         p.IsDisabled,
		}
		mu.Lock()
		store[id] = tn
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tn) //nolint:errcheck
	})

	mux.HandleFunc("/v1/ehr/tenant/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/ehr/tenant/")
		switch r.Method {
		case http.MethodGet:
			mu.Lock()
			tn, ok := store[id]
			var snap mockEHRTenant
			if ok {
				snap = *tn
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
				TenantID           *string `json:"tenantId"`
				Name               string  `json:"name"`
				Description        *string `json:"description"`
				RedoxEHRIdentifier string  `json:"redoxEHRIdentifier"`
				SymplrURLSlug      string  `json:"symplrUrlSlug"`
				RedoxEnvironment   string  `json:"redoxEnvironment"`
				IsDisabled         bool    `json:"isDisabled"`
			}
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mu.Lock()
			tn, ok := store[id]
			var updated mockEHRTenant
			if ok {
				tn.TenantID = p.TenantID
				tn.Name = p.Name
				tn.Description = p.Description
				tn.RedoxEHRIdentifier = p.RedoxEHRIdentifier
				tn.SymplrURLSlug = p.SymplrURLSlug
				tn.RedoxEnvironment = p.RedoxEnvironment
				tn.IsDisabled = p.IsDisabled
				updated = *tn
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

func TestAccEHRTenantResource_lifecycle(t *testing.T) {
	srv := newEHRTenantMockServer(t)
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
resource "tcm_ehr_tenant" "test" {
  name                 = "test-tenant"
  redox_ehr_identifier = "ehr-001"
  symplr_url_slug      = "test-slug"
  redox_environment    = "production"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tcm_ehr_tenant.test", "id"),
					resource.TestCheckResourceAttr("tcm_ehr_tenant.test", "name", "test-tenant"),
					resource.TestCheckResourceAttr("tcm_ehr_tenant.test", "redox_ehr_identifier", "ehr-001"),
					resource.TestCheckResourceAttr("tcm_ehr_tenant.test", "symplr_url_slug", "test-slug"),
					resource.TestCheckResourceAttr("tcm_ehr_tenant.test", "redox_environment", "production"),
					resource.TestCheckResourceAttr("tcm_ehr_tenant.test", "is_disabled", "false"),
				),
			},
			{
				Config: providerConfig + `
resource "tcm_ehr_tenant" "test" {
  name                 = "test-tenant-updated"
  description          = "Updated description"
  redox_ehr_identifier = "ehr-002"
  symplr_url_slug      = "test-slug-v2"
  redox_environment    = "staging"
  is_disabled          = true
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tcm_ehr_tenant.test", "name", "test-tenant-updated"),
					resource.TestCheckResourceAttr("tcm_ehr_tenant.test", "description", "Updated description"),
					resource.TestCheckResourceAttr("tcm_ehr_tenant.test", "redox_ehr_identifier", "ehr-002"),
					resource.TestCheckResourceAttr("tcm_ehr_tenant.test", "symplr_url_slug", "test-slug-v2"),
					resource.TestCheckResourceAttr("tcm_ehr_tenant.test", "redox_environment", "staging"),
					resource.TestCheckResourceAttr("tcm_ehr_tenant.test", "is_disabled", "true"),
				),
			},
			{
				ResourceName:      "tcm_ehr_tenant.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEHRTenantDataSource_read(t *testing.T) {
	srv := newEHRTenantMockServer(t)
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
resource "tcm_ehr_tenant" "seed" {
  name                 = "ds-tenant"
  redox_ehr_identifier = "ehr-ds"
  symplr_url_slug      = "ds-slug"
  redox_environment    = "production"
}

data "tcm_ehr_tenant" "test" {
  id = tcm_ehr_tenant.seed.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tcm_ehr_tenant.test", "id", "tcm_ehr_tenant.seed", "id"),
					resource.TestCheckResourceAttr("data.tcm_ehr_tenant.test", "name", "ds-tenant"),
					resource.TestCheckResourceAttr("data.tcm_ehr_tenant.test", "redox_ehr_identifier", "ehr-ds"),
					resource.TestCheckResourceAttr("data.tcm_ehr_tenant.test", "symplr_url_slug", "ds-slug"),
					resource.TestCheckResourceAttr("data.tcm_ehr_tenant.test", "redox_environment", "production"),
					resource.TestCheckResourceAttr("data.tcm_ehr_tenant.test", "is_disabled", "false"),
				),
			},
		},
	})
}
