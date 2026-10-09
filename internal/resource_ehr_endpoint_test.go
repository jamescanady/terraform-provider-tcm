package internal

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEHREndpoint_toPayload_required(t *testing.T) {
	r := &EHREndpointResource{}
	m := &EHREndpointResourceModel{
		ID:         types.StringNull(),
		Name:       types.StringValue("test-endpoint"),
		IsDisabled: types.BoolValue(false),
	}
	p := r.toPayload(m)
	if p.Name != "test-endpoint" {
		t.Errorf("expected Name test-endpoint, got %q", p.Name)
	}
	if p.IsDisabled {
		t.Errorf("expected IsDisabled false")
	}
	if p.ID != nil {
		t.Errorf("expected ID nil when model ID is null")
	}
	if p.Description != nil {
		t.Errorf("expected Description nil when not set")
	}
}

func TestEHREndpoint_toPayload_withIDAndDescription(t *testing.T) {
	r := &EHREndpointResource{}
	m := &EHREndpointResourceModel{
		ID:          types.StringValue("ep-001"),
		Name:        types.StringValue("my-endpoint"),
		Description: types.StringValue("endpoint desc"),
		IsDisabled:  types.BoolValue(true),
	}
	p := r.toPayload(m)
	if p.ID == nil || *p.ID != "ep-001" {
		t.Errorf("expected ID ep-001")
	}
	if p.Description == nil || *p.Description != "endpoint desc" {
		t.Errorf("expected Description 'endpoint desc'")
	}
	if !p.IsDisabled {
		t.Errorf("expected IsDisabled true")
	}
}

func TestEHREndpoint_applyResult(t *testing.T) {
	id := "ep-result"
	desc := "result desc"
	p := &ehrEndpointPayload{
		ID:          &id,
		Name:        "result-endpoint",
		Description: &desc,
		IsDisabled:  true,
	}
	m := &EHREndpointResourceModel{}
	applyEHREndpointResult(m, p)

	if m.ID.ValueString() != "ep-result" {
		t.Errorf("expected ID ep-result, got %q", m.ID.ValueString())
	}
	if m.Name.ValueString() != "result-endpoint" {
		t.Errorf("expected Name result-endpoint, got %q", m.Name.ValueString())
	}
	if m.Description.ValueString() != "result desc" {
		t.Errorf("expected Description 'result desc', got %q", m.Description.ValueString())
	}
	if !m.IsDisabled.ValueBool() {
		t.Errorf("expected IsDisabled true")
	}
}
