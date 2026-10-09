package internal

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEHRDestination_toPayload_defaults(t *testing.T) {
	r := &EHRDestinationResource{}
	m := &EHRDestinationResourceModel{
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
	if p.Name != nil {
		t.Errorf("expected Name nil when not set")
	}
	if p.JsonMetadataID != nil {
		t.Errorf("expected JsonMetadataID nil when not set")
	}
	if p.DestinationID != nil {
		t.Errorf("expected DestinationID nil when not set")
	}
}

func TestEHRDestination_toPayload_withFields(t *testing.T) {
	r := &EHRDestinationResource{}
	m := &EHRDestinationResourceModel{
		ID:             types.StringValue("dst-001"),
		JsonMetadataID: types.StringValue("jm-001"),
		DestinationID:  types.StringValue("dest-001"),
		Name:           types.StringValue("my-destination"),
		Description:    types.StringValue("dest desc"),
		IsDisabled:     types.BoolValue(true),
	}
	p := r.toPayload(m)
	if p.ID == nil || *p.ID != "dst-001" {
		t.Errorf("expected ID dst-001")
	}
	if p.JsonMetadataID == nil || *p.JsonMetadataID != "jm-001" {
		t.Errorf("expected JsonMetadataID jm-001")
	}
	if p.DestinationID == nil || *p.DestinationID != "dest-001" {
		t.Errorf("expected DestinationID dest-001")
	}
	if p.Name == nil || *p.Name != "my-destination" {
		t.Errorf("expected Name my-destination")
	}
	if p.Description == nil || *p.Description != "dest desc" {
		t.Errorf("expected Description 'dest desc'")
	}
	if !p.IsDisabled {
		t.Errorf("expected IsDisabled true")
	}
}

func TestEHRDestination_applyResult(t *testing.T) {
	id := "dst-result"
	jmID := "jm-result"
	destID := "dest-result"
	name := "result-destination"
	desc := "result desc"
	p := &ehrDestinationPayload{
		ID:             &id,
		JsonMetadataID: &jmID,
		DestinationID:  &destID,
		Name:           &name,
		Description:    &desc,
		IsDisabled:     true,
	}
	m := &EHRDestinationResourceModel{}
	applyEHRDestinationResult(m, p)

	if m.ID.ValueString() != "dst-result" {
		t.Errorf("expected ID dst-result, got %q", m.ID.ValueString())
	}
	if m.JsonMetadataID.ValueString() != "jm-result" {
		t.Errorf("expected JsonMetadataID jm-result, got %q", m.JsonMetadataID.ValueString())
	}
	if m.DestinationID.ValueString() != "dest-result" {
		t.Errorf("expected DestinationID dest-result, got %q", m.DestinationID.ValueString())
	}
	if m.Name.ValueString() != "result-destination" {
		t.Errorf("expected Name result-destination, got %q", m.Name.ValueString())
	}
	if m.Description.ValueString() != "result desc" {
		t.Errorf("expected Description 'result desc', got %q", m.Description.ValueString())
	}
	if !m.IsDisabled.ValueBool() {
		t.Errorf("expected IsDisabled true")
	}
}
