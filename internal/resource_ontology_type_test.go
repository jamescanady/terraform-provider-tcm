package internal

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestOntologyType_toPayload_noID(t *testing.T) {
	r := &OntologyTypeResource{}
	m := &OntologyTypeResourceModel{
		ID:   types.StringValue("ot-should-not-appear"),
		Name: types.StringValue("My Ontology"),
	}
	p := r.toPayload(m)
	// toPayload for OntologyType does NOT set ID
	if p.ID != nil {
		t.Errorf("expected p.ID nil (OntologyType toPayload never sets ID), got %v", *p.ID)
	}
	if p.Name == nil || *p.Name != "My Ontology" {
		t.Errorf("expected Name 'My Ontology'")
	}
}

func TestOntologyType_toPayload_optional(t *testing.T) {
	r := &OntologyTypeResource{}
	m := &OntologyTypeResourceModel{
		ID:          types.StringNull(),
		Name:        types.StringValue("OT"),
		Description: types.StringValue("desc"),
	}
	p := r.toPayload(m)
	if p.Description == nil || *p.Description != "desc" {
		t.Errorf("expected Description 'desc'")
	}
}

func TestOntologyType_toPayload_nullFieldsOmitted(t *testing.T) {
	r := &OntologyTypeResource{}
	m := &OntologyTypeResourceModel{
		ID: types.StringNull(),
	}
	p := r.toPayload(m)
	if p.Name != nil {
		t.Errorf("expected Name nil when model Name is null")
	}
	if p.Description != nil {
		t.Errorf("expected Description nil when model Description is null")
	}
}

func TestOntologyType_applyResult(t *testing.T) {
	id := "ot-result"
	name := "Result OT"
	desc := "result desc"
	disabled := true
	p := &ontologyTypePayload{
		ID:          &id,
		Name:        &name,
		Description: &desc,
		IsDisabled:  &disabled,
	}
	m := &OntologyTypeResourceModel{}
	applyOntologyTypeResult(m, p)

	if m.ID.ValueString() != "ot-result" {
		t.Errorf("expected ID ot-result, got %q", m.ID.ValueString())
	}
	if m.Name.ValueString() != "Result OT" {
		t.Errorf("expected Name 'Result OT', got %q", m.Name.ValueString())
	}
	if m.Description.ValueString() != "result desc" {
		t.Errorf("expected Description 'result desc', got %q", m.Description.ValueString())
	}
	if !m.IsDisabled.ValueBool() {
		t.Errorf("expected IsDisabled true")
	}
}

func TestOntologyType_applyResult_nilIsDisabledDefaultsFalse(t *testing.T) {
	p := &ontologyTypePayload{}
	m := &OntologyTypeResourceModel{}
	applyOntologyTypeResult(m, p)
	if m.IsDisabled.ValueBool() {
		t.Errorf("expected IsDisabled false when payload IsDisabled is nil")
	}
}
