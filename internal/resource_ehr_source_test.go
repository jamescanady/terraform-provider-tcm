package internal

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEHRSource_toPayload_defaults(t *testing.T) {
	r := &EHRSourceResource{}
	m := &EHRSourceResourceModel{
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
	if p.TenantID != nil {
		t.Errorf("expected TenantID nil when not set")
	}
	if p.SourceID != nil {
		t.Errorf("expected SourceID nil when not set")
	}
}

func TestEHRSource_toPayload_withFields(t *testing.T) {
	r := &EHRSourceResource{}
	m := &EHRSourceResourceModel{
		ID:          types.StringValue("src-001"),
		TenantID:    types.StringValue("tenant-001"),
		SourceID:    types.StringValue("source-001"),
		Name:        types.StringValue("my-source"),
		Description: types.StringValue("source desc"),
		IsDisabled:  types.BoolValue(true),
	}
	p := r.toPayload(m)
	if p.ID == nil || *p.ID != "src-001" {
		t.Errorf("expected ID src-001")
	}
	if p.TenantID == nil || *p.TenantID != "tenant-001" {
		t.Errorf("expected TenantID tenant-001")
	}
	if p.SourceID == nil || *p.SourceID != "source-001" {
		t.Errorf("expected SourceID source-001")
	}
	if p.Name == nil || *p.Name != "my-source" {
		t.Errorf("expected Name my-source")
	}
	if p.Description == nil || *p.Description != "source desc" {
		t.Errorf("expected Description 'source desc'")
	}
	if !p.IsDisabled {
		t.Errorf("expected IsDisabled true")
	}
}

func TestEHRSource_applyResult(t *testing.T) {
	id := "src-result"
	tenantID := "tenant-result"
	sourceID := "source-result"
	name := "result-source"
	desc := "result desc"
	p := &ehrSourcePayload{
		ID:          &id,
		TenantID:    &tenantID,
		SourceID:    &sourceID,
		Name:        &name,
		Description: &desc,
		IsDisabled:  true,
	}
	m := &EHRSourceResourceModel{}
	applyEHRSourceResult(m, p)

	if m.ID.ValueString() != "src-result" {
		t.Errorf("expected ID src-result, got %q", m.ID.ValueString())
	}
	if m.TenantID.ValueString() != "tenant-result" {
		t.Errorf("expected TenantID tenant-result, got %q", m.TenantID.ValueString())
	}
	if m.SourceID.ValueString() != "source-result" {
		t.Errorf("expected SourceID source-result, got %q", m.SourceID.ValueString())
	}
	if m.Name.ValueString() != "result-source" {
		t.Errorf("expected Name result-source, got %q", m.Name.ValueString())
	}
	if m.Description.ValueString() != "result desc" {
		t.Errorf("expected Description 'result desc', got %q", m.Description.ValueString())
	}
	if !m.IsDisabled.ValueBool() {
		t.Errorf("expected IsDisabled true")
	}
}
