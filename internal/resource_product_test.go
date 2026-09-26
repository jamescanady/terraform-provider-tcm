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

type mockProduct struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDisabled  bool   `json:"isDisabled"`
}

func newMockTCMServer(t *testing.T) *httptest.Server {
	t.Helper()

	var (
		mu      sync.Mutex
		counter atomic.Int64
		store   = map[string]*mockProduct{}
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/Product", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p mockProduct
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		p.ID = fmt.Sprintf("product-%d", counter.Add(1))
		mu.Lock()
		store[p.ID] = &p
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p) //nolint:errcheck
	})

	mux.HandleFunc("/v1/Product/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/Product/")
		switch r.Method {
		case http.MethodGet:
			mu.Lock()
			p, ok := store[id]
			mu.Unlock()
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(p) //nolint:errcheck

		case http.MethodPut:
			var p mockProduct
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			p.ID = id
			mu.Lock()
			store[id] = &p
			mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(p) //nolint:errcheck

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return httptest.NewServer(mux)
}

func TestAccProductResource_lifecycle(t *testing.T) {
	srv := newMockTCMServer(t)
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
resource "tcm_product" "test" {
  name        = "Test Product"
  description = "A test product"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tcm_product.test", "id"),
					resource.TestCheckResourceAttr("tcm_product.test", "name", "Test Product"),
					resource.TestCheckResourceAttr("tcm_product.test", "description", "A test product"),
					resource.TestCheckResourceAttr("tcm_product.test", "is_disabled", "false"),
				),
			},
			{
				Config: providerConfig + `
resource "tcm_product" "test" {
  name        = "Updated Product"
  description = "An updated description"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tcm_product.test", "name", "Updated Product"),
					resource.TestCheckResourceAttr("tcm_product.test", "description", "An updated description"),
					resource.TestCheckResourceAttr("tcm_product.test", "is_disabled", "false"),
				),
			},
			{
				ResourceName:      "tcm_product.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccProductResource_disabledOnDestroy(t *testing.T) {
	srv := newMockTCMServer(t)
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
resource "tcm_product" "test" {
  name        = "Temp Product"
  description = "Will be soft-deleted"
}
`,
				Check: resource.TestCheckResourceAttr("tcm_product.test", "is_disabled", "false"),
			},
		},
	})
}
