package internal

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestERPConfig_toPayload_ontologyTypeIDs(t *testing.T) {
	r := &ERPConfigResource{}
	ids := types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("ot-001"),
		types.StringValue("ot-002"),
	})
	m := &ERPConfigResourceModel{
		OntologyTypeIDs: ids,
	}
	p := r.toPayload(m)
	if len(p.OntologyTypeIDs) != 2 {
		t.Fatalf("expected 2 OntologyTypeIDs, got %d", len(p.OntologyTypeIDs))
	}
	if p.OntologyTypeIDs[0] != "ot-001" {
		t.Errorf("expected OntologyTypeIDs[0] = ot-001, got %q", p.OntologyTypeIDs[0])
	}
	if p.OntologyTypeIDs[1] != "ot-002" {
		t.Errorf("expected OntologyTypeIDs[1] = ot-002, got %q", p.OntologyTypeIDs[1])
	}
}

func TestERPConfig_toPayload_noID(t *testing.T) {
	r := &ERPConfigResource{}
	m := &ERPConfigResourceModel{
		ERPID: types.StringValue("erp-001"),
		URL:   types.StringValue("https://example.com"),
	}
	p := r.toPayload(m)
	// toPayload for ERPConfig does NOT set ID
	if p.ID != nil {
		t.Errorf("expected p.ID nil (ERPConfig toPayload never sets ID), got %v", *p.ID)
	}
	if p.ERPID == nil || *p.ERPID != "erp-001" {
		t.Errorf("expected ERPID erp-001")
	}
	if p.URL == nil || *p.URL != "https://example.com" {
		t.Errorf("expected URL https://example.com")
	}
}

func TestERPConfig_toPayload_nullOntologyTypeIDsOmitted(t *testing.T) {
	r := &ERPConfigResource{}
	m := &ERPConfigResourceModel{
		OntologyTypeIDs: types.ListNull(types.StringType),
	}
	p := r.toPayload(m)
	if len(p.OntologyTypeIDs) != 0 {
		t.Errorf("expected empty OntologyTypeIDs when model list is null, got %v", p.OntologyTypeIDs)
	}
}

func TestERPConfig_applyResult_doesNotOverwriteOntologyTypeIDs(t *testing.T) {
	// applyERPConfigResult must NOT touch OntologyTypeIDs (server doesn't return it).
	initialIDs := types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("ot-preserved"),
	})
	m := &ERPConfigResourceModel{
		OntologyTypeIDs: initialIDs,
	}
	id := "cfg-result"
	erpID := "erp-result"
	// Server response does NOT include OntologyTypeIDs
	p := &erpConfigPayload{
		ID:    &id,
		ERPID: &erpID,
		// OntologyTypeIDs is nil / empty in response
	}
	applyERPConfigResult(m, p)

	// OntologyTypeIDs must be preserved from state
	if m.OntologyTypeIDs.IsNull() {
		t.Errorf("expected OntologyTypeIDs to be preserved, got null")
	}
	elems := m.OntologyTypeIDs.Elements()
	if len(elems) != 1 {
		t.Fatalf("expected 1 element preserved, got %d", len(elems))
	}
	if elems[0].(types.String).ValueString() != "ot-preserved" {
		t.Errorf("expected preserved value ot-preserved, got %q", elems[0].(types.String).ValueString())
	}
}

func TestERPConfig_applyResult_fields(t *testing.T) {
	id := "cfg-001"
	erpID := "erp-001"
	tpeID := "tpe-001"
	url := "https://erp.example.com"
	deleted := false
	m := &ERPConfigResourceModel{
		OntologyTypeIDs: types.ListNull(types.StringType),
	}
	p := &erpConfigPayload{
		ID:                         &id,
		ERPID:                      &erpID,
		TenantProductEnvironmentID: &tpeID,
		URL:                        &url,
		IsDeleted:                  &deleted,
	}
	applyERPConfigResult(m, p)
	if m.ID.ValueString() != "cfg-001" {
		t.Errorf("expected ID cfg-001, got %q", m.ID.ValueString())
	}
	if m.ERPID.ValueString() != "erp-001" {
		t.Errorf("expected ERPID erp-001, got %q", m.ERPID.ValueString())
	}
	if m.TenantProductEnvironmentID.ValueString() != "tpe-001" {
		t.Errorf("expected TenantProductEnvironmentID tpe-001, got %q", m.TenantProductEnvironmentID.ValueString())
	}
	if m.URL.ValueString() != "https://erp.example.com" {
		t.Errorf("expected URL, got %q", m.URL.ValueString())
	}
	if m.IsDeleted.ValueBool() {
		t.Errorf("expected IsDeleted false")
	}
}
