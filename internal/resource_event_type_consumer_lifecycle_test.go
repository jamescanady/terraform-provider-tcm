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

// mockEventTypeConsumer stores data using consumerId (the response field name).
type mockEventTypeConsumer struct {
	ID                         string  `json:"id"`
	EventTypeID                *string `json:"eventTypeId,omitempty"`
	ConsumerID                 *string `json:"consumerId,omitempty"` // response field
	TenantProductEnvironmentID *string `json:"tenantProductEnvironmentId,omitempty"`
	IsDisabled                 bool    `json:"isDisabled"`
}

func newEventTypeConsumerMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	var (
		mu      sync.Mutex
		counter atomic.Int64
		store   = map[string]*mockEventTypeConsumer{}
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/EventTypeConsumer", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Request uses eventConsumerId; response uses consumerId.
		var p struct {
			EventTypeID                *string `json:"eventTypeId"`
			EventConsumerID            *string `json:"eventConsumerId"` // write field
			TenantProductEnvironmentID *string `json:"tenantProductEnvironmentId"`
			IsDisabled                 bool    `json:"isDisabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id := fmt.Sprintf("etc-%d", counter.Add(1))
		etc := &mockEventTypeConsumer{
			ID:                         id,
			EventTypeID:                p.EventTypeID,
			ConsumerID:                 p.EventConsumerID, // map write→read field
			TenantProductEnvironmentID: p.TenantProductEnvironmentID,
			IsDisabled:                 p.IsDisabled,
		}
		mu.Lock()
		store[id] = etc
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(etc) //nolint:errcheck
	})

	mux.HandleFunc("/v1/EventTypeConsumer/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/EventTypeConsumer/")
		switch r.Method {
		case http.MethodGet:
			mu.Lock()
			etc, ok := store[id]
			var snap mockEventTypeConsumer
			if ok {
				snap = *etc
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
				EventTypeID                *string `json:"eventTypeId"`
				EventConsumerID            *string `json:"eventConsumerId"` // write field
				TenantProductEnvironmentID *string `json:"tenantProductEnvironmentId"`
				IsDisabled                 bool    `json:"isDisabled"`
			}
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mu.Lock()
			etc, ok := store[id]
			var updated mockEventTypeConsumer
			if ok {
				etc.EventTypeID = p.EventTypeID
				etc.ConsumerID = p.EventConsumerID // map write→read field
				etc.TenantProductEnvironmentID = p.TenantProductEnvironmentID
				etc.IsDisabled = p.IsDisabled
				updated = *etc
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

func TestAccEventTypeConsumerResource_lifecycle(t *testing.T) {
	srv := newEventTypeConsumerMockServer(t)
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
resource "tcm_event_type_consumer" "test" {
  event_type_id     = "et-001"
  event_consumer_id = "ec-001"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tcm_event_type_consumer.test", "id"),
					resource.TestCheckResourceAttr("tcm_event_type_consumer.test", "event_type_id", "et-001"),
					resource.TestCheckResourceAttr("tcm_event_type_consumer.test", "event_consumer_id", "ec-001"),
					resource.TestCheckResourceAttr("tcm_event_type_consumer.test", "is_disabled", "false"),
				),
			},
			{
				Config: providerConfig + `
resource "tcm_event_type_consumer" "test" {
  event_type_id     = "et-001"
  event_consumer_id = "ec-002"
  is_disabled       = true
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tcm_event_type_consumer.test", "event_type_id", "et-001"),
					resource.TestCheckResourceAttr("tcm_event_type_consumer.test", "event_consumer_id", "ec-002"),
					resource.TestCheckResourceAttr("tcm_event_type_consumer.test", "is_disabled", "true"),
				),
			},
			{
				ResourceName:      "tcm_event_type_consumer.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEventTypeConsumerDataSource_read(t *testing.T) {
	srv := newEventTypeConsumerMockServer(t)
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
resource "tcm_event_type_consumer" "seed" {
  event_type_id     = "et-ds-001"
  event_consumer_id = "ec-ds-001"
}

data "tcm_event_type_consumer" "test" {
  id = tcm_event_type_consumer.seed.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tcm_event_type_consumer.test", "id", "tcm_event_type_consumer.seed", "id"),
					resource.TestCheckResourceAttr("data.tcm_event_type_consumer.test", "event_type_id", "et-ds-001"),
					resource.TestCheckResourceAttr("data.tcm_event_type_consumer.test", "event_consumer_id", "ec-ds-001"),
					resource.TestCheckResourceAttr("data.tcm_event_type_consumer.test", "is_disabled", "false"),
				),
			},
		},
	})
}
