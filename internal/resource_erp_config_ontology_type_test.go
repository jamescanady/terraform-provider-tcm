package internal

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestERPConfigOntologyType_toPayload_noID(t *testing.T) {
	r := &ERPConfigOntologyTypeResource{}
	m := &ERPConfigOntologyTypeResourceModel{
		ID:             types.StringValue("should-not-appear"),
		ERPConfigID:    types.StringValue("erp-cfg-001"),
		OntologyTypeID: types.StringValue("ot-001"),
	}
	p := r.toPayload(m)
	// toPayload for ERPConfigOntologyType does NOT set ID
	if p.ID != nil {
		t.Errorf("expected p.ID nil (ERPConfigOntologyType toPayload never sets ID), got %v", *p.ID)
	}
	if p.ERPConfigID == nil || *p.ERPConfigID != "erp-cfg-001" {
		t.Errorf("expected ERPConfigID erp-cfg-001")
	}
	if p.OntologyTypeID == nil || *p.OntologyTypeID != "ot-001" {
		t.Errorf("expected OntologyTypeID ot-001")
	}
}

func TestERPConfigOntologyType_toPayload_nullFields(t *testing.T) {
	r := &ERPConfigOntologyTypeResource{}
	m := &ERPConfigOntologyTypeResourceModel{
		ID: types.StringNull(),
	}
	p := r.toPayload(m)
	if p.ERPConfigID != nil {
		t.Errorf("expected ERPConfigID nil when model field is null")
	}
	if p.OntologyTypeID != nil {
		t.Errorf("expected OntologyTypeID nil when model field is null")
	}
}

func TestERPConfigOntologyType_applyResult(t *testing.T) {
	id := "ecot-result"
	erpCfgID := "erp-cfg-result"
	otID := "ot-result"
	p := &erpConfigOntologyTypePayload{
		ID:             &id,
		ERPConfigID:    &erpCfgID,
		OntologyTypeID: &otID,
	}
	m := &ERPConfigOntologyTypeResourceModel{}
	applyERPConfigOntologyTypeResult(m, p)

	if m.ID.ValueString() != "ecot-result" {
		t.Errorf("expected ID ecot-result, got %q", m.ID.ValueString())
	}
	if m.ERPConfigID.ValueString() != "erp-cfg-result" {
		t.Errorf("expected ERPConfigID erp-cfg-result, got %q", m.ERPConfigID.ValueString())
	}
	if m.OntologyTypeID.ValueString() != "ot-result" {
		t.Errorf("expected OntologyTypeID ot-result, got %q", m.OntologyTypeID.ValueString())
	}
}
