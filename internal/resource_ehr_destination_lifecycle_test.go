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

type mockEHRDestination struct {
	ID             string  `json:"id"`
	JsonMetadataID *string `json:"jsonMetadataId,omitempty"`
	DestinationID  *string `json:"destinationId,omitempty"`
	Name           *string `json:"name,omitempty"`
	Description    *string `json:"description,omitempty"`
	IsDisabled     bool    `json:"isDisabled"`
}

func newEHRDestinationMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	var (
		mu      sync.Mutex
		counter atomic.Int64
		store   = map[string]*mockEHRDestination{}
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/ehr/destination", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p struct {
			JsonMetadataID *string `json:"jsonMetadataId"`
			DestinationID  *string `json:"destinationId"`
			Name           *string `json:"name"`
			Description    *string `json:"description"`
			IsDisabled     bool    `json:"isDisabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id := fmt.Sprintf("dst-%d", counter.Add(1))
		dst := &mockEHRDestination{
			ID:             id,
			JsonMetadataID: p.JsonMetadataID,
			DestinationID:  p.DestinationID,
			Name:           p.Name,
			Description:    p.Description,
			IsDisabled:     p.IsDisabled,
		}
		mu.Lock()
		store[id] = dst
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(dst) //nolint:errcheck
	})

	mux.HandleFunc("/v1/ehr/destination/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/ehr/destination/")
		switch r.Method {
		case http.MethodGet:
			mu.Lock()
			dst, ok := store[id]
			var snap mockEHRDestination
			if ok {
				snap = *dst
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
				JsonMetadataID *string `json:"jsonMetadataId"`
				DestinationID  *string `json:"destinationId"`
				Name           *string `json:"name"`
				Description    *string `json:"description"`
				IsDisabled     bool    `json:"isDisabled"`
			}
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mu.Lock()
			dst, ok := store[id]
			var updated mockEHRDestination
			if ok {
				dst.JsonMetadataID = p.JsonMetadataID
				dst.DestinationID = p.DestinationID
				dst.Name = p.Name
				dst.Description = p.Description
				dst.IsDisabled = p.IsDisabled
				updated = *dst
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

func TestAccEHRDestinationResource_lifecycle(t *testing.T) {
	srv := newEHRDestinationMockServer(t)
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
resource "tcm_ehr_destination" "test" {
  name = "test-destination"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tcm_ehr_destination.test", "id"),
					resource.TestCheckResourceAttr("tcm_ehr_destination.test", "name", "test-destination"),
					resource.TestCheckResourceAttr("tcm_ehr_destination.test", "is_disabled", "false"),
				),
			},
			{
				Config: providerConfig + `
resource "tcm_ehr_destination" "test" {
  name        = "test-destination-updated"
  description = "Updated description"
  is_disabled = true
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tcm_ehr_destination.test", "name", "test-destination-updated"),
					resource.TestCheckResourceAttr("tcm_ehr_destination.test", "description", "Updated description"),
					resource.TestCheckResourceAttr("tcm_ehr_destination.test", "is_disabled", "true"),
				),
			},
			{
				ResourceName:      "tcm_ehr_destination.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEHRDestinationDataSource_read(t *testing.T) {
	srv := newEHRDestinationMockServer(t)
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
resource "tcm_ehr_destination" "seed" {
  name        = "ds-destination"
  description = "Datasource test"
}

data "tcm_ehr_destination" "test" {
  id = tcm_ehr_destination.seed.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tcm_ehr_destination.test", "id", "tcm_ehr_destination.seed", "id"),
					resource.TestCheckResourceAttr("data.tcm_ehr_destination.test", "name", "ds-destination"),
					resource.TestCheckResourceAttr("data.tcm_ehr_destination.test", "description", "Datasource test"),
					resource.TestCheckResourceAttr("data.tcm_ehr_destination.test", "is_disabled", "false"),
				),
			},
		},
	})
}
