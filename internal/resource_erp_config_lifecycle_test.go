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

type mockERPConfig struct {
	ID                         string   `json:"id"`
	ERPID                      *string  `json:"erpId,omitempty"`
	TenantProductEnvironmentID *string  `json:"tenantProductEnvironmentId,omitempty"`
	URL                        *string  `json:"url,omitempty"`
	ClientID                   *string  `json:"clientId,omitempty"`
	ClientSecret               *string  `json:"clientSecret,omitempty"`
	LookupID                   *string  `json:"lookupId,omitempty"`
	HostName                   *string  `json:"hostName,omitempty"`
	TenantSlug                 *string  `json:"tenantSlug,omitempty"`
	APIVersion                 *string  `json:"apiVersion,omitempty"`
	OntologyTypeIDs            []string `json:"ontologyTypeIds,omitempty"`
	GlobalTenantCode           *string  `json:"globalTenantCode,omitempty"`
	ERPName                    *string  `json:"erpName,omitempty"`
	IsDeleted                  *bool    `json:"isDeleted,omitempty"`
}

func newERPConfigMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	var (
		mu      sync.Mutex
		counter atomic.Int64
		store   = map[string]*mockERPConfig{}
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/ERPConfig", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p struct {
			ERPID                      *string  `json:"erpId"`
			TenantProductEnvironmentID *string  `json:"tenantProductEnvironmentId"`
			URL                        *string  `json:"url"`
			ClientID                   *string  `json:"clientId"`
			ClientSecret               *string  `json:"clientSecret"`
			LookupID                   *string  `json:"lookupId"`
			HostName                   *string  `json:"hostName"`
			TenantSlug                 *string  `json:"tenantSlug"`
			APIVersion                 *string  `json:"apiVersion"`
			OntologyTypeIDs            []string `json:"ontologyTypeIds"`
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id := fmt.Sprintf("erpconfig-%d", counter.Add(1))
		ec := &mockERPConfig{
			ID:                         id,
			ERPID:                      p.ERPID,
			TenantProductEnvironmentID: p.TenantProductEnvironmentID,
			URL:                        p.URL,
			ClientID:                   p.ClientID,
			ClientSecret:               p.ClientSecret,
			LookupID:                   p.LookupID,
			HostName:                   p.HostName,
			TenantSlug:                 p.TenantSlug,
			APIVersion:                 p.APIVersion,
			OntologyTypeIDs:            p.OntologyTypeIDs,
		}
		mu.Lock()
		store[id] = ec
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ec) //nolint:errcheck
	})

	mux.HandleFunc("/v1/ERPConfig/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/ERPConfig/")
		switch r.Method {
		case http.MethodGet:
			mu.Lock()
			ec, ok := store[id]
			var snap mockERPConfig
			if ok {
				snap = *ec
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
				ERPID                      *string  `json:"erpId"`
				TenantProductEnvironmentID *string  `json:"tenantProductEnvironmentId"`
				URL                        *string  `json:"url"`
				ClientID                   *string  `json:"clientId"`
				ClientSecret               *string  `json:"clientSecret"`
				LookupID                   *string  `json:"lookupId"`
				HostName                   *string  `json:"hostName"`
				TenantSlug                 *string  `json:"tenantSlug"`
				APIVersion                 *string  `json:"apiVersion"`
				OntologyTypeIDs            []string `json:"ontologyTypeIds"`
			}
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mu.Lock()
			ec, ok := store[id]
			var updated mockERPConfig
			if ok {
				ec.ERPID = p.ERPID
				ec.TenantProductEnvironmentID = p.TenantProductEnvironmentID
				ec.URL = p.URL
				ec.ClientID = p.ClientID
				ec.ClientSecret = p.ClientSecret
				ec.LookupID = p.LookupID
				ec.HostName = p.HostName
				ec.TenantSlug = p.TenantSlug
				ec.APIVersion = p.APIVersion
				ec.OntologyTypeIDs = p.OntologyTypeIDs
				updated = *ec
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

func TestAccERPConfigResource_lifecycle(t *testing.T) {
	srv := newERPConfigMockServer(t)
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
resource "tcm_erp_config" "test" {
  erp_id                        = "erp-001"
  tenant_product_environment_id = "tpe-001"
  url                           = "https://erp.example.com"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tcm_erp_config.test", "id"),
					resource.TestCheckResourceAttr("tcm_erp_config.test", "erp_id", "erp-001"),
					resource.TestCheckResourceAttr("tcm_erp_config.test", "tenant_product_environment_id", "tpe-001"),
					resource.TestCheckResourceAttr("tcm_erp_config.test", "url", "https://erp.example.com"),
					resource.TestCheckResourceAttr("tcm_erp_config.test", "is_deleted", "false"),
				),
			},
			{
				Config: providerConfig + `
resource "tcm_erp_config" "test" {
  erp_id                        = "erp-001"
  tenant_product_environment_id = "tpe-002"
  url                           = "https://erp-v2.example.com"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tcm_erp_config.test", "tenant_product_environment_id", "tpe-002"),
					resource.TestCheckResourceAttr("tcm_erp_config.test", "url", "https://erp-v2.example.com"),
					resource.TestCheckResourceAttr("tcm_erp_config.test", "is_deleted", "false"),
				),
			},
			{
				ResourceName:            "tcm_erp_config.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ontology_type_ids"},
			},
		},
	})
}

func TestAccERPConfigDataSource_read(t *testing.T) {
	srv := newERPConfigMockServer(t)
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
resource "tcm_erp_config" "seed" {
  erp_id                        = "erp-ds"
  tenant_product_environment_id = "tpe-ds"
  url                           = "https://ds.erp.example.com"
}

data "tcm_erp_config" "test" {
  id = tcm_erp_config.seed.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tcm_erp_config.test", "id", "tcm_erp_config.seed", "id"),
					resource.TestCheckResourceAttr("data.tcm_erp_config.test", "erp_id", "erp-ds"),
					resource.TestCheckResourceAttr("data.tcm_erp_config.test", "tenant_product_environment_id", "tpe-ds"),
					resource.TestCheckResourceAttr("data.tcm_erp_config.test", "url", "https://ds.erp.example.com"),
					resource.TestCheckResourceAttr("data.tcm_erp_config.test", "is_deleted", "false"),
				),
			},
		},
	})
}
