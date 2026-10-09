package internal

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEventTypeConsumer_toPayload_defaults(t *testing.T) {
	r := &EventTypeConsumerResource{}
	m := &EventTypeConsumerResourceModel{
		ID:         types.StringNull(),
		IsDisabled: types.BoolValue(false),
	}
	p := r.toPayload(m)
	if p.IsDisabled {
		t.Errorf("expected IsDisabled false")
	}
	if p.ID != nil {
		t.Errorf("expected ID nil when model ID is null")
	}
	if p.EventTypeID != nil {
		t.Errorf("expected EventTypeID nil when not set")
	}
	if p.EventConsumerID != nil {
		t.Errorf("expected EventConsumerID nil when not set")
	}
}

func TestEventTypeConsumer_toPayload_withFields(t *testing.T) {
	r := &EventTypeConsumerResource{}
	m := &EventTypeConsumerResourceModel{
		ID:                         types.StringValue("etc-001"),
		EventTypeID:                types.StringValue("et-001"),
		EventConsumerID:            types.StringValue("ec-001"),
		TenantProductEnvironmentID: types.StringValue("tpe-001"),
		IsDisabled:                 types.BoolValue(true),
	}
	p := r.toPayload(m)
	if p.ID == nil || *p.ID != "etc-001" {
		t.Errorf("expected ID etc-001")
	}
	if p.EventTypeID == nil || *p.EventTypeID != "et-001" {
		t.Errorf("expected EventTypeID et-001")
	}
	if p.EventConsumerID == nil || *p.EventConsumerID != "ec-001" {
		t.Errorf("expected EventConsumerID ec-001")
	}
	if p.TenantProductEnvironmentID == nil || *p.TenantProductEnvironmentID != "tpe-001" {
		t.Errorf("expected TenantProductEnvironmentID tpe-001")
	}
	if !p.IsDisabled {
		t.Errorf("expected IsDisabled true")
	}
}

func TestEventTypeConsumer_applyResult_consumerIDMapping(t *testing.T) {
	// consumerId (response field) must map to m.EventConsumerID
	consumerID := "ec-from-response"
	etID := "et-from-response"
	id := "etc-result"
	p := &eventTypeConsumerPayload{
		ID:          &id,
		ConsumerID:  &consumerID, // response field
		EventTypeID: &etID,
		IsDisabled:  false,
	}
	m := &EventTypeConsumerResourceModel{}
	applyEventTypeConsumerResult(m, p)
	if m.ID.ValueString() != "etc-result" {
		t.Errorf("expected ID etc-result, got %q", m.ID.ValueString())
	}
	if m.EventConsumerID.ValueString() != "ec-from-response" {
		t.Errorf("expected EventConsumerID ec-from-response (mapped from ConsumerID), got %q", m.EventConsumerID.ValueString())
	}
	if m.EventTypeID.ValueString() != "et-from-response" {
		t.Errorf("expected EventTypeID et-from-response, got %q", m.EventTypeID.ValueString())
	}
}

func TestEventTypeConsumer_applyResult_computedFields(t *testing.T) {
	consumerName := "My Consumer"
	endpoint := "https://consumer.example.com"
	tenantName := "TenantA"
	productName := "ProductB"
	p := &eventTypeConsumerPayload{
		ConsumerName:     &consumerName,
		ConsumerEndpoint: &endpoint,
		TenantName:       &tenantName,
		ProductName:      &productName,
	}
	m := &EventTypeConsumerResourceModel{}
	applyEventTypeConsumerResult(m, p)
	if m.ConsumerName.ValueString() != "My Consumer" {
		t.Errorf("expected ConsumerName 'My Consumer', got %q", m.ConsumerName.ValueString())
	}
	if m.ConsumerEndpoint.ValueString() != "https://consumer.example.com" {
		t.Errorf("expected ConsumerEndpoint, got %q", m.ConsumerEndpoint.ValueString())
	}
	if m.TenantName.ValueString() != "TenantA" {
		t.Errorf("expected TenantName TenantA, got %q", m.TenantName.ValueString())
	}
	if m.ProductName.ValueString() != "ProductB" {
		t.Errorf("expected ProductName ProductB, got %q", m.ProductName.ValueString())
	}
}
