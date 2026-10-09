package internal

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEHRSchema_toPayload_requiredFields(t *testing.T) {
	r := &EHRSchemaResource{}
	fields := types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("field1"),
		types.StringValue("field2"),
	})
	m := &EHRSchemaResourceModel{
		ID:             types.StringNull(),
		RequiredFields: fields,
		IsDisabled:     types.BoolValue(false),
	}
	p := r.toPayload(m)
	if len(p.RequiredFields) != 2 {
		t.Fatalf("expected 2 RequiredFields, got %d", len(p.RequiredFields))
	}
	if p.RequiredFields[0] != "field1" {
		t.Errorf("expected RequiredFields[0] = field1, got %q", p.RequiredFields[0])
	}
	if p.RequiredFields[1] != "field2" {
		t.Errorf("expected RequiredFields[1] = field2, got %q", p.RequiredFields[1])
	}
	if p.IsDisabled {
		t.Errorf("expected IsDisabled false")
	}
	if p.ID != nil {
		t.Errorf("expected ID nil when model ID is null")
	}
}

func TestEHRSchema_toPayload_nullRequiredFieldsOmitted(t *testing.T) {
	r := &EHRSchemaResource{}
	m := &EHRSchemaResourceModel{
		ID:             types.StringNull(),
		RequiredFields: types.ListNull(types.StringType),
		IsDisabled:     types.BoolValue(false),
	}
	p := r.toPayload(m)
	if len(p.RequiredFields) != 0 {
		t.Errorf("expected empty RequiredFields when model list is null, got %v", p.RequiredFields)
	}
}

func TestEHRSchema_toPayload_withOptionalFields(t *testing.T) {
	r := &EHRSchemaResource{}
	m := &EHRSchemaResourceModel{
		ID:             types.StringValue("es-001"),
		TenantID:       types.StringValue("tenant-001"),
		EndpointID:     types.StringValue("ep-001"),
		RequiredFields: types.ListNull(types.StringType),
		IsDisabled:     types.BoolValue(true),
	}
	p := r.toPayload(m)
	if p.ID == nil || *p.ID != "es-001" {
		t.Errorf("expected ID es-001")
	}
	if p.TenantID == nil || *p.TenantID != "tenant-001" {
		t.Errorf("expected TenantID tenant-001")
	}
	if p.EndpointID == nil || *p.EndpointID != "ep-001" {
		t.Errorf("expected EndpointID ep-001")
	}
	if !p.IsDisabled {
		t.Errorf("expected IsDisabled true")
	}
}

func TestEHRSchema_applyResult(t *testing.T) {
	ctx := context.Background()
	id := "es-result"
	tenantID := "tenant-result"
	endpointID := "ep-result"
	p := &ehrSchemaPayload{
		ID:             &id,
		TenantID:       &tenantID,
		EndpointID:     &endpointID,
		RequiredFields: []string{"fieldA", "fieldB"},
		IsDisabled:     true,
	}
	m := &EHRSchemaResourceModel{}
	applyEHRSchemaResult(ctx, m, p)

	if m.ID.ValueString() != "es-result" {
		t.Errorf("expected ID es-result, got %q", m.ID.ValueString())
	}
	if m.TenantID.ValueString() != "tenant-result" {
		t.Errorf("expected TenantID tenant-result, got %q", m.TenantID.ValueString())
	}
	if m.EndpointID.ValueString() != "ep-result" {
		t.Errorf("expected EndpointID ep-result, got %q", m.EndpointID.ValueString())
	}
	if !m.IsDisabled.ValueBool() {
		t.Errorf("expected IsDisabled true")
	}
	elems := m.RequiredFields.Elements()
	if len(elems) != 2 {
		t.Fatalf("expected 2 RequiredFields, got %d", len(elems))
	}
	if elems[0].(types.String).ValueString() != "fieldA" {
		t.Errorf("expected RequiredFields[0] = fieldA, got %q", elems[0].(types.String).ValueString())
	}
	if elems[1].(types.String).ValueString() != "fieldB" {
		t.Errorf("expected RequiredFields[1] = fieldB, got %q", elems[1].(types.String).ValueString())
	}
}

func TestEHRSchema_applyResult_nilRequiredFields(t *testing.T) {
	ctx := context.Background()
	p := &ehrSchemaPayload{
		IsDisabled: false,
	}
	m := &EHRSchemaResourceModel{}
	applyEHRSchemaResult(ctx, m, p)
	if !m.RequiredFields.IsNull() {
		t.Errorf("expected RequiredFields null when payload field is nil")
	}
}
