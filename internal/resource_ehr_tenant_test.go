package internal

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEHRTenant_toPayload_required(t *testing.T) {
	r := &EHRTenantResource{}
	m := &EHRTenantResourceModel{
		ID:                 types.StringNull(),
		Name:               types.StringValue("test-tenant"),
		RedoxEHRIdentifier: types.StringValue("ehr-001"),
		SymplrURLSlug:      types.StringValue("test-slug"),
		RedoxEnvironment:   types.StringValue("production"),
		IsDisabled:         types.BoolValue(false),
	}
	p := r.toPayload(m)
	if p.Name != "test-tenant" {
		t.Errorf("expected Name test-tenant, got %q", p.Name)
	}
	if p.RedoxEHRIdentifier != "ehr-001" {
		t.Errorf("expected RedoxEHRIdentifier ehr-001, got %q", p.RedoxEHRIdentifier)
	}
	if p.SymplrURLSlug != "test-slug" {
		t.Errorf("expected SymplrURLSlug test-slug, got %q", p.SymplrURLSlug)
	}
	if p.RedoxEnvironment != "production" {
		t.Errorf("expected RedoxEnvironment production, got %q", p.RedoxEnvironment)
	}
	if p.IsDisabled {
		t.Errorf("expected IsDisabled false")
	}
	if p.ID != nil {
		t.Errorf("expected ID nil when model ID is null")
	}
	if p.TenantID != nil {
		t.Errorf("expected TenantID nil when not set")
	}
	if p.Description != nil {
		t.Errorf("expected Description nil when not set")
	}
}

func TestEHRTenant_toPayload_withOptionalFields(t *testing.T) {
	r := &EHRTenantResource{}
	m := &EHRTenantResourceModel{
		ID:                 types.StringValue("tn-001"),
		TenantID:           types.StringValue("tenant-001"),
		Name:               types.StringValue("my-tenant"),
		Description:        types.StringValue("tenant desc"),
		RedoxEHRIdentifier: types.StringValue("ehr-002"),
		SymplrURLSlug:      types.StringValue("slug-v2"),
		RedoxEnvironment:   types.StringValue("staging"),
		IsDisabled:         types.BoolValue(true),
	}
	p := r.toPayload(m)
	if p.ID == nil || *p.ID != "tn-001" {
		t.Errorf("expected ID tn-001")
	}
	if p.TenantID == nil || *p.TenantID != "tenant-001" {
		t.Errorf("expected TenantID tenant-001")
	}
	if p.Description == nil || *p.Description != "tenant desc" {
		t.Errorf("expected Description 'tenant desc'")
	}
	if !p.IsDisabled {
		t.Errorf("expected IsDisabled true")
	}
}

func TestEHRTenant_applyResult(t *testing.T) {
	id := "tn-result"
	tenantID := "tenant-result"
	desc := "result desc"
	p := &ehrTenantPayload{
		ID:                 &id,
		TenantID:           &tenantID,
		Name:               "result-tenant",
		Description:        &desc,
		RedoxEHRIdentifier: "ehr-result",
		SymplrURLSlug:      "result-slug",
		RedoxEnvironment:   "production",
		IsDisabled:         true,
	}
	m := &EHRTenantResourceModel{}
	applyEHRTenantResult(m, p)

	if m.ID.ValueString() != "tn-result" {
		t.Errorf("expected ID tn-result, got %q", m.ID.ValueString())
	}
	if m.TenantID.ValueString() != "tenant-result" {
		t.Errorf("expected TenantID tenant-result, got %q", m.TenantID.ValueString())
	}
	if m.Name.ValueString() != "result-tenant" {
		t.Errorf("expected Name result-tenant, got %q", m.Name.ValueString())
	}
	if m.Description.ValueString() != "result desc" {
		t.Errorf("expected Description 'result desc', got %q", m.Description.ValueString())
	}
	if m.RedoxEHRIdentifier.ValueString() != "ehr-result" {
		t.Errorf("expected RedoxEHRIdentifier ehr-result, got %q", m.RedoxEHRIdentifier.ValueString())
	}
	if m.SymplrURLSlug.ValueString() != "result-slug" {
		t.Errorf("expected SymplrURLSlug result-slug, got %q", m.SymplrURLSlug.ValueString())
	}
	if m.RedoxEnvironment.ValueString() != "production" {
		t.Errorf("expected RedoxEnvironment production, got %q", m.RedoxEnvironment.ValueString())
	}
	if !m.IsDisabled.ValueBool() {
		t.Errorf("expected IsDisabled true")
	}
}
