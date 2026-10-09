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

type mockNamespace struct {
	ID             string  `json:"id"`
	Name           *string `json:"name,omitempty"`
	Description    *string `json:"description,omitempty"`
	IsDefault      bool    `json:"isDefault"`
	IsDisabled     bool    `json:"isDisabled"`
	CreatedDate    string  `json:"createdDate"`
	CreatedBy      *string `json:"createdBy,omitempty"`
	LastModified   string  `json:"lastModified"`
	LastModifiedBy *string `json:"lastModifiedBy,omitempty"`
}

func newNamespaceMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	var (
		mu      sync.Mutex
		counter atomic.Int64
		store   = map[string]*mockNamespace{}
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/Namespace", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p struct {
			Name        *string `json:"name"`
			Description *string `json:"description"`
			IsDisabled  bool    `json:"isDisabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id := fmt.Sprintf("ns-%d", counter.Add(1))
		author := "test-user"
		ns := &mockNamespace{
			ID:             id,
			Name:           p.Name,
			Description:    p.Description,
			IsDefault:      false,
			IsDisabled:     p.IsDisabled,
			CreatedDate:    "2024-01-01T00:00:00Z",
			CreatedBy:      &author,
			LastModified:   "2024-01-01T00:00:00Z",
			LastModifiedBy: &author,
		}
		mu.Lock()
		store[id] = ns
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ns) //nolint:errcheck
	})

	mux.HandleFunc("/v1/Namespace/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/Namespace/")
		switch r.Method {
		case http.MethodGet:
			mu.Lock()
			ns, ok := store[id]
			var snap mockNamespace
			if ok {
				snap = *ns
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
				IsDisabled  bool    `json:"isDisabled"`
			}
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mu.Lock()
			ns, ok := store[id]
			var updated mockNamespace
			if ok {
				ns.Name = p.Name
				ns.Description = p.Description
				ns.IsDisabled = p.IsDisabled
				ns.LastModified = "2024-01-02T00:00:00Z"
				updated = *ns
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

func TestAccNamespaceResource_lifecycle(t *testing.T) {
	srv := newNamespaceMockServer(t)
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
resource "tcm_namespace" "test" {
  name        = "Test Namespace"
  description = "A test namespace"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tcm_namespace.test", "id"),
					resource.TestCheckResourceAttr("tcm_namespace.test", "name", "Test Namespace"),
					resource.TestCheckResourceAttr("tcm_namespace.test", "description", "A test namespace"),
					resource.TestCheckResourceAttr("tcm_namespace.test", "is_disabled", "false"),
					resource.TestCheckResourceAttr("tcm_namespace.test", "is_default", "false"),
					resource.TestCheckResourceAttrSet("tcm_namespace.test", "created_date"),
					resource.TestCheckResourceAttrSet("tcm_namespace.test", "last_modified"),
				),
			},
			{
				Config: providerConfig + `
resource "tcm_namespace" "test" {
  name        = "Updated Namespace"
  description = "Updated description"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tcm_namespace.test", "name", "Updated Namespace"),
					resource.TestCheckResourceAttr("tcm_namespace.test", "description", "Updated description"),
					resource.TestCheckResourceAttr("tcm_namespace.test", "is_disabled", "false"),
				),
			},
			{
				ResourceName:      "tcm_namespace.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNamespaceDataSource_read(t *testing.T) {
	srv := newNamespaceMockServer(t)
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
resource "tcm_namespace" "seed" {
  name        = "Seed Namespace"
  description = "For datasource test"
}

data "tcm_namespace" "test" {
  id = tcm_namespace.seed.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tcm_namespace.test", "id", "tcm_namespace.seed", "id"),
					resource.TestCheckResourceAttr("data.tcm_namespace.test", "name", "Seed Namespace"),
					resource.TestCheckResourceAttr("data.tcm_namespace.test", "description", "For datasource test"),
					resource.TestCheckResourceAttr("data.tcm_namespace.test", "is_disabled", "false"),
					resource.TestCheckResourceAttr("data.tcm_namespace.test", "is_default", "false"),
					resource.TestCheckResourceAttrSet("data.tcm_namespace.test", "created_date"),
					resource.TestCheckResourceAttrSet("data.tcm_namespace.test", "last_modified"),
				),
			},
		},
	})
}
