package internal

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEventConsumer_toPayload_defaults(t *testing.T) {
	r := &EventConsumerResource{}
	m := &EventConsumerResourceModel{
		ID:         types.StringNull(),
		Name:       types.StringValue("MyConsumer"),
		Endpoint:   types.StringValue("https://example.com/events"),
		IsDisabled: types.BoolValue(false),
	}
	p := r.toPayload(m)
	if p.Name != "MyConsumer" {
		t.Errorf("expected Name MyConsumer, got %q", p.Name)
	}
	if p.Endpoint != "https://example.com/events" {
		t.Errorf("expected Endpoint https://example.com/events, got %q", p.Endpoint)
	}
	if p.IsDisabled {
		t.Errorf("expected IsDisabled false")
	}
	if p.ID != nil {
		t.Errorf("expected ID nil when model ID is null")
	}
	if p.AuthorizationParameters != nil {
		t.Errorf("expected AuthorizationParameters nil when not set")
	}
}

func TestEventConsumer_toPayload_authorizationParameters(t *testing.T) {
	r := &EventConsumerResource{}
	jsonStr := `{"key":"val"}`
	m := &EventConsumerResourceModel{
		ID:                      types.StringNull(),
		Name:                    types.StringValue("Consumer"),
		Endpoint:                types.StringValue("https://example.com"),
		IsDisabled:              types.BoolValue(false),
		AuthorizationParameters: types.StringValue(jsonStr),
	}
	p := r.toPayload(m)
	if p.AuthorizationParameters == nil {
		t.Fatal("expected AuthorizationParameters to be non-nil")
	}
	if string(*p.AuthorizationParameters) != jsonStr {
		t.Errorf("expected AuthorizationParameters %q, got %q", jsonStr, string(*p.AuthorizationParameters))
	}
}

func TestEventConsumer_applyResult(t *testing.T) {
	idStr := "ec-result"
	tenantStr := "tenant-001"
	descStr := "consumer desc"
	authTypeStr := "API_KEY"
	jsonStr := `{"apiKey":"secret"}`
	version := int64(5)

	rawMsg := json.RawMessage(jsonStr)

	p := &eventConsumerPayload{
		ID:                      &idStr,
		TenantID:                &tenantStr,
		Name:                    "ResultConsumer",
		Description:             &descStr,
		Endpoint:                "https://result.example.com",
		AuthorizationType:       &authTypeStr,
		AuthorizationParameters: &rawMsg,
		IsDisabled:              true,
		Version:                 &version,
	}

	m := &EventConsumerResourceModel{}
	applyEventConsumerResult(m, p)

	if m.ID.ValueString() != "ec-result" {
		t.Errorf("expected ID ec-result, got %q", m.ID.ValueString())
	}
	if m.TenantID.ValueString() != "tenant-001" {
		t.Errorf("expected TenantID tenant-001, got %q", m.TenantID.ValueString())
	}
	if m.Name.ValueString() != "ResultConsumer" {
		t.Errorf("expected Name ResultConsumer, got %q", m.Name.ValueString())
	}
	if m.Description.ValueString() != "consumer desc" {
		t.Errorf("expected Description 'consumer desc', got %q", m.Description.ValueString())
	}
	if m.Endpoint.ValueString() != "https://result.example.com" {
		t.Errorf("expected Endpoint https://result.example.com, got %q", m.Endpoint.ValueString())
	}
	if m.AuthorizationType.ValueString() != "API_KEY" {
		t.Errorf("expected AuthorizationType API_KEY, got %q", m.AuthorizationType.ValueString())
	}
	if m.AuthorizationParameters.ValueString() != jsonStr {
		t.Errorf("expected AuthorizationParameters %q, got %q", jsonStr, m.AuthorizationParameters.ValueString())
	}
	if !m.IsDisabled.ValueBool() {
		t.Errorf("expected IsDisabled true")
	}
	if m.Version.ValueInt64() != 5 {
		t.Errorf("expected Version 5, got %d", m.Version.ValueInt64())
	}
}

func TestEventConsumer_applyResult_nilAuthParams(t *testing.T) {
	p := &eventConsumerPayload{
		Name:     "Minimal",
		Endpoint: "https://minimal.example.com",
	}
	m := &EventConsumerResourceModel{}
	applyEventConsumerResult(m, p)
	if !m.AuthorizationParameters.IsNull() {
		t.Errorf("expected AuthorizationParameters to be null when payload field is nil")
	}
	if !m.Version.IsNull() {
		t.Errorf("expected Version to be null when payload field is nil")
	}
}
