package internal

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEventType_toPayload_defaults(t *testing.T) {
	r := &EventTypeResource{}
	m := &EventTypeResourceModel{
		ID:         types.StringNull(),
		ProductID:  types.StringValue("prod-001"),
		Name:       types.StringValue("MyEvent"),
		IsDisabled: types.BoolValue(false),
	}
	p := r.toPayload(m)
	if p.ProductID != "prod-001" {
		t.Errorf("expected ProductID prod-001, got %q", p.ProductID)
	}
	if p.Name != "MyEvent" {
		t.Errorf("expected Name MyEvent, got %q", p.Name)
	}
	if p.IsDisabled != false {
		t.Errorf("expected IsDisabled false")
	}
	if p.ID != nil {
		t.Errorf("expected ID nil when model ID is null")
	}
	if p.Description != nil {
		t.Errorf("expected Description nil when not set")
	}
}

func TestEventType_toPayload_withIDAndDescription(t *testing.T) {
	r := &EventTypeResource{}
	m := &EventTypeResourceModel{
		ID:          types.StringValue("uuid-001"),
		ProductID:   types.StringValue("prod-002"),
		Name:        types.StringValue("AnotherEvent"),
		Description: types.StringValue("desc here"),
		IsDisabled:  types.BoolValue(true),
	}
	p := r.toPayload(m)
	if p.ID == nil || *p.ID != "uuid-001" {
		t.Errorf("expected ID uuid-001")
	}
	if p.Description == nil || *p.Description != "desc here" {
		t.Errorf("expected Description 'desc here'")
	}
	if !p.IsDisabled {
		t.Errorf("expected IsDisabled true")
	}
}

func TestEventType_applyResult(t *testing.T) {
	m := &EventTypeResourceModel{}
	id := "result-id"
	desc := "result desc"
	creator := "alice"
	created := "2024-01-01"
	modifier := "bob"
	modified := "2024-01-02"
	p := &eventTypePayload{
		ID:             &id,
		ProductID:      "prod-999",
		Name:           "ResultEvent",
		Description:    &desc,
		IsDisabled:     true,
		CreatedBy:      &creator,
		Created:        &created,
		LastModifiedBy: &modifier,
		LastModified:   &modified,
	}
	applyEventTypeResult(m, p)
	if m.ID.ValueString() != "result-id" {
		t.Errorf("expected ID result-id, got %q", m.ID.ValueString())
	}
	if m.ProductID.ValueString() != "prod-999" {
		t.Errorf("expected ProductID prod-999, got %q", m.ProductID.ValueString())
	}
	if m.Name.ValueString() != "ResultEvent" {
		t.Errorf("expected Name ResultEvent, got %q", m.Name.ValueString())
	}
	if m.Description.ValueString() != "result desc" {
		t.Errorf("expected Description 'result desc', got %q", m.Description.ValueString())
	}
	if !m.IsDisabled.ValueBool() {
		t.Errorf("expected IsDisabled true")
	}
	if m.CreatedBy.ValueString() != "alice" {
		t.Errorf("expected CreatedBy alice")
	}
	if m.LastModifiedBy.ValueString() != "bob" {
		t.Errorf("expected LastModifiedBy bob")
	}
}
