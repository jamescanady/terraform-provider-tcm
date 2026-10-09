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

type mockEHRJsonMetadata struct {
	ID           string  `json:"id"`
	TenantID     *string `json:"tenantId,omitempty"`
	EndpointID   *string `json:"endpointId,omitempty"`
	FacilityCode *string `json:"facilityCode,omitempty"`
	IsDisabled   bool    `json:"isDisabled"`
}

func newEHRJsonMetadataMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	var (
		mu      sync.Mutex
		counter atomic.Int64
		store   = map[string]*mockEHRJsonMetadata{}
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/ehr/jsonMetadata", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p struct {
			TenantID     *string `json:"tenantId"`
			EndpointID   *string `json:"endpointId"`
			FacilityCode *string `json:"facilityCode"`
			IsDisabled   bool    `json:"isDisabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id := fmt.Sprintf("jm-%d", counter.Add(1))
		jm := &mockEHRJsonMetadata{
			ID:           id,
			TenantID:     p.TenantID,
			EndpointID:   p.EndpointID,
			FacilityCode: p.FacilityCode,
			IsDisabled:   p.IsDisabled,
		}
		mu.Lock()
		store[id] = jm
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jm) //nolint:errcheck
	})

	mux.HandleFunc("/v1/ehr/jsonMetadata/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/ehr/jsonMetadata/")
		switch r.Method {
		case http.MethodGet:
			mu.Lock()
			jm, ok := store[id]
			var snap mockEHRJsonMetadata
			if ok {
				snap = *jm
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
				TenantID     *string `json:"tenantId"`
				EndpointID   *string `json:"endpointId"`
				FacilityCode *string `json:"facilityCode"`
				IsDisabled   bool    `json:"isDisabled"`
			}
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mu.Lock()
			jm, ok := store[id]
			var updated mockEHRJsonMetadata
			if ok {
				jm.TenantID = p.TenantID
				jm.EndpointID = p.EndpointID
				jm.FacilityCode = p.FacilityCode
				jm.IsDisabled = p.IsDisabled
				updated = *jm
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

func TestAccEHRJsonMetadataResource_lifecycle(t *testing.T) {
	srv := newEHRJsonMetadataMockServer(t)
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
resource "tcm_ehr_json_metadata" "test" {
  facility_code = "FAC001"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tcm_ehr_json_metadata.test", "id"),
					resource.TestCheckResourceAttr("tcm_ehr_json_metadata.test", "facility_code", "FAC001"),
					resource.TestCheckResourceAttr("tcm_ehr_json_metadata.test", "is_disabled", "false"),
				),
			},
			{
				Config: providerConfig + `
resource "tcm_ehr_json_metadata" "test" {
  tenant_id     = "tenant-001"
  endpoint_id   = "ep-001"
  facility_code = "FAC002"
  is_disabled   = true
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tcm_ehr_json_metadata.test", "tenant_id", "tenant-001"),
					resource.TestCheckResourceAttr("tcm_ehr_json_metadata.test", "endpoint_id", "ep-001"),
					resource.TestCheckResourceAttr("tcm_ehr_json_metadata.test", "facility_code", "FAC002"),
					resource.TestCheckResourceAttr("tcm_ehr_json_metadata.test", "is_disabled", "true"),
				),
			},
			{
				ResourceName:      "tcm_ehr_json_metadata.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEHRJsonMetadataDataSource_read(t *testing.T) {
	srv := newEHRJsonMetadataMockServer(t)
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
resource "tcm_ehr_json_metadata" "seed" {
  tenant_id     = "tenant-ds"
  facility_code = "FAC-DS"
}

data "tcm_ehr_json_metadata" "test" {
  id = tcm_ehr_json_metadata.seed.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tcm_ehr_json_metadata.test", "id", "tcm_ehr_json_metadata.seed", "id"),
					resource.TestCheckResourceAttr("data.tcm_ehr_json_metadata.test", "tenant_id", "tenant-ds"),
					resource.TestCheckResourceAttr("data.tcm_ehr_json_metadata.test", "facility_code", "FAC-DS"),
					resource.TestCheckResourceAttr("data.tcm_ehr_json_metadata.test", "is_disabled", "false"),
				),
			},
		},
	})
}
