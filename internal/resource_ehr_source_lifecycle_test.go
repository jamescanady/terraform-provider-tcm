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

type mockEHRSource struct {
	ID          string  `json:"id"`
	TenantID    *string `json:"tenantId,omitempty"`
	SourceID    *string `json:"sourceId,omitempty"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	IsDisabled  bool    `json:"isDisabled"`
}

func newEHRSourceMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	var (
		mu      sync.Mutex
		counter atomic.Int64
		store   = map[string]*mockEHRSource{}
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/ehr/source", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p struct {
			TenantID    *string `json:"tenantId"`
			SourceID    *string `json:"sourceId"`
			Name        *string `json:"name"`
			Description *string `json:"description"`
			IsDisabled  bool    `json:"isDisabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id := fmt.Sprintf("src-%d", counter.Add(1))
		src := &mockEHRSource{
			ID:          id,
			TenantID:    p.TenantID,
			SourceID:    p.SourceID,
			Name:        p.Name,
			Description: p.Description,
			IsDisabled:  p.IsDisabled,
		}
		mu.Lock()
		store[id] = src
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(src) //nolint:errcheck
	})

	mux.HandleFunc("/v1/ehr/source/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/ehr/source/")
		switch r.Method {
		case http.MethodGet:
			mu.Lock()
			src, ok := store[id]
			var snap mockEHRSource
			if ok {
				snap = *src
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
				TenantID    *string `json:"tenantId"`
				SourceID    *string `json:"sourceId"`
				Name        *string `json:"name"`
				Description *string `json:"description"`
				IsDisabled  bool    `json:"isDisabled"`
			}
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mu.Lock()
			src, ok := store[id]
			var updated mockEHRSource
			if ok {
				src.TenantID = p.TenantID
				src.SourceID = p.SourceID
				src.Name = p.Name
				src.Description = p.Description
				src.IsDisabled = p.IsDisabled
				updated = *src
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

func TestAccEHRSourceResource_lifecycle(t *testing.T) {
	srv := newEHRSourceMockServer(t)
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
resource "tcm_ehr_source" "test" {
  name = "test-source"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tcm_ehr_source.test", "id"),
					resource.TestCheckResourceAttr("tcm_ehr_source.test", "name", "test-source"),
					resource.TestCheckResourceAttr("tcm_ehr_source.test", "is_disabled", "false"),
				),
			},
			{
				Config: providerConfig + `
resource "tcm_ehr_source" "test" {
  name        = "test-source-updated"
  description = "Updated description"
  is_disabled = true
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tcm_ehr_source.test", "name", "test-source-updated"),
					resource.TestCheckResourceAttr("tcm_ehr_source.test", "description", "Updated description"),
					resource.TestCheckResourceAttr("tcm_ehr_source.test", "is_disabled", "true"),
				),
			},
			{
				ResourceName:      "tcm_ehr_source.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEHRSourceDataSource_read(t *testing.T) {
	srv := newEHRSourceMockServer(t)
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
resource "tcm_ehr_source" "seed" {
  name        = "ds-source"
  description = "Datasource test"
}

data "tcm_ehr_source" "test" {
  id = tcm_ehr_source.seed.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tcm_ehr_source.test", "id", "tcm_ehr_source.seed", "id"),
					resource.TestCheckResourceAttr("data.tcm_ehr_source.test", "name", "ds-source"),
					resource.TestCheckResourceAttr("data.tcm_ehr_source.test", "description", "Datasource test"),
					resource.TestCheckResourceAttr("data.tcm_ehr_source.test", "is_disabled", "false"),
				),
			},
		},
	})
}
