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

type mockSystemInfo struct {
	ID                   string  `json:"id"`
	ProductID            string  `json:"productId"`
	TenantID             *string `json:"tenantId,omitempty"`
	NamespaceID          *string `json:"namespaceId,omitempty"`
	ProductEnvironmentID *string `json:"productEnvironmentId,omitempty"`
	FlowID               *string `json:"flowId,omitempty"`
	ConnectionType       *string `json:"connectionType,omitempty"`
	Description          *string `json:"description,omitempty"`
	Host                 *string `json:"host,omitempty"`
	FlowVersion          *string `json:"flowVersion,omitempty"`
	BaseApiPath          *string `json:"baseApiPath,omitempty"`
	Endpoint             *string `json:"endpoint,omitempty"`
	OAuthScope           *string `json:"oAuthScope,omitempty"`
	IsDisabled           bool    `json:"isDisabled"`
	NamespaceName        *string `json:"namespaceName,omitempty"`
	TenantName           *string `json:"tenantName,omitempty"`
	ProductName          *string `json:"productName,omitempty"`
}

func newSystemInfoMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	var (
		mu      sync.Mutex
		counter atomic.Int64
		store   = map[string]*mockSystemInfo{}
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/SystemInfo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p struct {
			ProductID            string  `json:"productId"`
			TenantID             *string `json:"tenantId"`
			NamespaceID          *string `json:"namespaceId"`
			ProductEnvironmentID *string `json:"productEnvironmentId"`
			FlowID               *string `json:"flowId"`
			ConnectionType       *string `json:"connectionType"`
			Description          *string `json:"description"`
			Host                 *string `json:"host"`
			FlowVersion          *string `json:"flowVersion"`
			BaseApiPath          *string `json:"baseApiPath"`
			Endpoint             *string `json:"endpoint"`
			OAuthScope           *string `json:"oAuthScope"`
			IsDisabled           bool    `json:"isDisabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id := fmt.Sprintf("si-%d", counter.Add(1))
		si := &mockSystemInfo{
			ID:                   id,
			ProductID:            p.ProductID,
			TenantID:             p.TenantID,
			NamespaceID:          p.NamespaceID,
			ProductEnvironmentID: p.ProductEnvironmentID,
			FlowID:               p.FlowID,
			ConnectionType:       p.ConnectionType,
			Description:          p.Description,
			Host:                 p.Host,
			FlowVersion:          p.FlowVersion,
			BaseApiPath:          p.BaseApiPath,
			Endpoint:             p.Endpoint,
			OAuthScope:           p.OAuthScope,
			IsDisabled:           p.IsDisabled,
		}
		mu.Lock()
		store[id] = si
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(si) //nolint:errcheck
	})

	mux.HandleFunc("/v1/SystemInfo/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/SystemInfo/")
		switch r.Method {
		case http.MethodGet:
			mu.Lock()
			si, ok := store[id]
			var snap mockSystemInfo
			if ok {
				snap = *si
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
				ProductID            string  `json:"productId"`
				TenantID             *string `json:"tenantId"`
				NamespaceID          *string `json:"namespaceId"`
				ProductEnvironmentID *string `json:"productEnvironmentId"`
				FlowID               *string `json:"flowId"`
				ConnectionType       *string `json:"connectionType"`
				Description          *string `json:"description"`
				Host                 *string `json:"host"`
				FlowVersion          *string `json:"flowVersion"`
				BaseApiPath          *string `json:"baseApiPath"`
				Endpoint             *string `json:"endpoint"`
				OAuthScope           *string `json:"oAuthScope"`
				IsDisabled           bool    `json:"isDisabled"`
			}
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mu.Lock()
			si, ok := store[id]
			var updated mockSystemInfo
			if ok {
				si.ProductID = p.ProductID
				si.TenantID = p.TenantID
				si.NamespaceID = p.NamespaceID
				si.ProductEnvironmentID = p.ProductEnvironmentID
				si.FlowID = p.FlowID
				si.ConnectionType = p.ConnectionType
				si.Description = p.Description
				si.Host = p.Host
				si.FlowVersion = p.FlowVersion
				si.BaseApiPath = p.BaseApiPath
				si.Endpoint = p.Endpoint
				si.OAuthScope = p.OAuthScope
				si.IsDisabled = p.IsDisabled
				updated = *si
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

func TestAccSystemInfoResource_lifecycle(t *testing.T) {
	srv := newSystemInfoMockServer(t)
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
resource "tcm_system_info" "test" {
  product_id    = "product-001"
  host          = "https://api.example.com"
  flow_version  = "1.0"
  base_api_path = "/api/v1"
  o_auth_scope  = "openid profile"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tcm_system_info.test", "id"),
					resource.TestCheckResourceAttr("tcm_system_info.test", "product_id", "product-001"),
					resource.TestCheckResourceAttr("tcm_system_info.test", "host", "https://api.example.com"),
					resource.TestCheckResourceAttr("tcm_system_info.test", "flow_version", "1.0"),
					resource.TestCheckResourceAttr("tcm_system_info.test", "base_api_path", "/api/v1"),
					resource.TestCheckResourceAttr("tcm_system_info.test", "o_auth_scope", "openid profile"),
					resource.TestCheckResourceAttr("tcm_system_info.test", "is_disabled", "false"),
				),
			},
			{
				Config: providerConfig + `
resource "tcm_system_info" "test" {
  product_id      = "product-001"
  host            = "https://api-v2.example.com"
  flow_version    = "2.0"
  base_api_path   = "/api/v2"
  o_auth_scope    = "openid profile email"
  connection_type = "Http"
  description     = "Updated system info"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tcm_system_info.test", "host", "https://api-v2.example.com"),
					resource.TestCheckResourceAttr("tcm_system_info.test", "flow_version", "2.0"),
					resource.TestCheckResourceAttr("tcm_system_info.test", "base_api_path", "/api/v2"),
					resource.TestCheckResourceAttr("tcm_system_info.test", "o_auth_scope", "openid profile email"),
					resource.TestCheckResourceAttr("tcm_system_info.test", "connection_type", "Http"),
					resource.TestCheckResourceAttr("tcm_system_info.test", "description", "Updated system info"),
					resource.TestCheckResourceAttr("tcm_system_info.test", "is_disabled", "false"),
				),
			},
			{
				ResourceName:      "tcm_system_info.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSystemInfoDataSource_read(t *testing.T) {
	srv := newSystemInfoMockServer(t)
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
resource "tcm_system_info" "seed" {
  product_id      = "product-ds"
  host            = "https://ds.example.com"
  flow_version    = "1.0"
  base_api_path   = "/api"
  o_auth_scope    = "openid"
  connection_type = "Http"
}

data "tcm_system_info" "test" {
  id = tcm_system_info.seed.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tcm_system_info.test", "id", "tcm_system_info.seed", "id"),
					resource.TestCheckResourceAttr("data.tcm_system_info.test", "product_id", "product-ds"),
					resource.TestCheckResourceAttr("data.tcm_system_info.test", "host", "https://ds.example.com"),
					resource.TestCheckResourceAttr("data.tcm_system_info.test", "flow_version", "1.0"),
					resource.TestCheckResourceAttr("data.tcm_system_info.test", "base_api_path", "/api"),
					resource.TestCheckResourceAttr("data.tcm_system_info.test", "o_auth_scope", "openid"),
					resource.TestCheckResourceAttr("data.tcm_system_info.test", "connection_type", "Http"),
					resource.TestCheckResourceAttr("data.tcm_system_info.test", "is_disabled", "false"),
				),
			},
		},
	})
}
