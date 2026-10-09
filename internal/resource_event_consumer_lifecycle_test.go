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

type mockEventConsumer struct {
	ID                      string           `json:"id"`
	TenantID                *string          `json:"tenantId,omitempty"`
	Name                    string           `json:"name"`
	Description             *string          `json:"description,omitempty"`
	Endpoint                string           `json:"endpoint"`
	AuthorizationType       *string          `json:"authorizationType,omitempty"`
	AuthorizationParameters *json.RawMessage `json:"authorizationParameters,omitempty"`
	IsDisabled              bool             `json:"isDisabled"`
	Version                 *int64           `json:"version,omitempty"`
}

func newEventConsumerMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	var (
		mu      sync.Mutex
		counter atomic.Int64
		store   = map[string]*mockEventConsumer{}
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/EventConsumer", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var p struct {
			TenantID                *string          `json:"tenantId"`
			Name                    string           `json:"name"`
			Description             *string          `json:"description"`
			Endpoint                string           `json:"endpoint"`
			AuthorizationType       *string          `json:"authorizationType"`
			AuthorizationParameters *json.RawMessage `json:"authorizationParameters"`
			IsDisabled              bool             `json:"isDisabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id := fmt.Sprintf("ec-%d", counter.Add(1))
		version := int64(1)
		ec := &mockEventConsumer{
			ID:                      id,
			TenantID:                p.TenantID,
			Name:                    p.Name,
			Description:             p.Description,
			Endpoint:                p.Endpoint,
			AuthorizationType:       p.AuthorizationType,
			AuthorizationParameters: p.AuthorizationParameters,
			IsDisabled:              p.IsDisabled,
			Version:                 &version,
		}
		mu.Lock()
		store[id] = ec
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ec) //nolint:errcheck
	})

	mux.HandleFunc("/v1/EventConsumer/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/EventConsumer/")
		switch r.Method {
		case http.MethodGet:
			mu.Lock()
			ec, ok := store[id]
			var snap mockEventConsumer
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
				TenantID                *string          `json:"tenantId"`
				Name                    string           `json:"name"`
				Description             *string          `json:"description"`
				Endpoint                string           `json:"endpoint"`
				AuthorizationType       *string          `json:"authorizationType"`
				AuthorizationParameters *json.RawMessage `json:"authorizationParameters"`
				IsDisabled              bool             `json:"isDisabled"`
			}
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mu.Lock()
			ec, ok := store[id]
			var updated mockEventConsumer
			if ok {
				ec.TenantID = p.TenantID
				ec.Name = p.Name
				ec.Description = p.Description
				ec.Endpoint = p.Endpoint
				ec.AuthorizationType = p.AuthorizationType
				ec.AuthorizationParameters = p.AuthorizationParameters
				ec.IsDisabled = p.IsDisabled
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

func TestAccEventConsumerResource_lifecycle(t *testing.T) {
	srv := newEventConsumerMockServer(t)
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
resource "tcm_event_consumer" "test" {
  name     = "TestConsumer"
  endpoint = "https://consumer.example.com/events"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tcm_event_consumer.test", "id"),
					resource.TestCheckResourceAttr("tcm_event_consumer.test", "name", "TestConsumer"),
					resource.TestCheckResourceAttr("tcm_event_consumer.test", "endpoint", "https://consumer.example.com/events"),
					resource.TestCheckResourceAttr("tcm_event_consumer.test", "is_disabled", "false"),
				),
			},
			{
				Config: providerConfig + `
resource "tcm_event_consumer" "test" {
  name               = "TestConsumerUpdated"
  endpoint           = "https://consumer-v2.example.com/events"
  description        = "Updated consumer"
  authorization_type = "API_KEY"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tcm_event_consumer.test", "name", "TestConsumerUpdated"),
					resource.TestCheckResourceAttr("tcm_event_consumer.test", "endpoint", "https://consumer-v2.example.com/events"),
					resource.TestCheckResourceAttr("tcm_event_consumer.test", "description", "Updated consumer"),
					resource.TestCheckResourceAttr("tcm_event_consumer.test", "authorization_type", "API_KEY"),
					resource.TestCheckResourceAttr("tcm_event_consumer.test", "is_disabled", "false"),
				),
			},
			{
				ResourceName:      "tcm_event_consumer.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEventConsumerDataSource_read(t *testing.T) {
	srv := newEventConsumerMockServer(t)
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
resource "tcm_event_consumer" "seed" {
  name        = "SeedConsumer"
  endpoint    = "https://seed.example.com/events"
  description = "Seed consumer"
}

data "tcm_event_consumer" "test" {
  id = tcm_event_consumer.seed.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.tcm_event_consumer.test", "id", "tcm_event_consumer.seed", "id"),
					resource.TestCheckResourceAttr("data.tcm_event_consumer.test", "name", "SeedConsumer"),
					resource.TestCheckResourceAttr("data.tcm_event_consumer.test", "endpoint", "https://seed.example.com/events"),
					resource.TestCheckResourceAttr("data.tcm_event_consumer.test", "description", "Seed consumer"),
					resource.TestCheckResourceAttr("data.tcm_event_consumer.test", "is_disabled", "false"),
				),
			},
		},
	})
}
