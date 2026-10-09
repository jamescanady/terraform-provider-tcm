package internal

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEHRJsonMetadata_toPayload_defaults(t *testing.T) {
	r := &EHRJsonMetadataResource{}
	m := &EHRJsonMetadataResourceModel{
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
	if p.EndpointID != nil {
		t.Errorf("expected EndpointID nil when not set")
	}
	if p.FacilityCode != nil {
		t.Errorf("expected FacilityCode nil when not set")
	}
}

func TestEHRJsonMetadata_toPayload_withFields(t *testing.T) {
	r := &EHRJsonMetadataResource{}
	m := &EHRJsonMetadataResourceModel{
		ID:           types.StringValue("jm-001"),
		TenantID:     types.StringValue("tenant-001"),
		EndpointID:   types.StringValue("ep-001"),
		FacilityCode: types.StringValue("FAC001"),
		IsDisabled:   types.BoolValue(true),
	}
	p := r.toPayload(m)
	if p.ID == nil || *p.ID != "jm-001" {
		t.Errorf("expected ID jm-001")
	}
	if p.TenantID == nil || *p.TenantID != "tenant-001" {
		t.Errorf("expected TenantID tenant-001")
	}
	if p.EndpointID == nil || *p.EndpointID != "ep-001" {
		t.Errorf("expected EndpointID ep-001")
	}
	if p.FacilityCode == nil || *p.FacilityCode != "FAC001" {
		t.Errorf("expected FacilityCode FAC001")
	}
	if !p.IsDisabled {
		t.Errorf("expected IsDisabled true")
	}
}

func TestEHRJsonMetadata_applyResult(t *testing.T) {
	id := "jm-result"
	tenantID := "tenant-result"
	endpointID := "ep-result"
	facilityCode := "FAC-RESULT"
	p := &ehrJsonMetadataPayload{
		ID:           &id,
		TenantID:     &tenantID,
		EndpointID:   &endpointID,
		FacilityCode: &facilityCode,
		IsDisabled:   true,
	}
	m := &EHRJsonMetadataResourceModel{}
	applyEHRJsonMetadataResult(m, p)

	if m.ID.ValueString() != "jm-result" {
		t.Errorf("expected ID jm-result, got %q", m.ID.ValueString())
	}
	if m.TenantID.ValueString() != "tenant-result" {
		t.Errorf("expected TenantID tenant-result, got %q", m.TenantID.ValueString())
	}
	if m.EndpointID.ValueString() != "ep-result" {
		t.Errorf("expected EndpointID ep-result, got %q", m.EndpointID.ValueString())
	}
	if m.FacilityCode.ValueString() != "FAC-RESULT" {
		t.Errorf("expected FacilityCode FAC-RESULT, got %q", m.FacilityCode.ValueString())
	}
	if !m.IsDisabled.ValueBool() {
		t.Errorf("expected IsDisabled true")
	}
}

func TestEHRJsonMetadata_applyResult_nilOptionals(t *testing.T) {
	p := &ehrJsonMetadataPayload{
		IsDisabled: false,
	}
	m := &EHRJsonMetadataResourceModel{}
	applyEHRJsonMetadataResult(m, p)
	if !m.TenantID.IsNull() {
		t.Errorf("expected TenantID null when payload field is nil")
	}
	if !m.EndpointID.IsNull() {
		t.Errorf("expected EndpointID null when payload field is nil")
	}
	if !m.FacilityCode.IsNull() {
		t.Errorf("expected FacilityCode null when payload field is nil")
	}
	if m.IsDisabled.ValueBool() {
		t.Errorf("expected IsDisabled false")
	}
}
