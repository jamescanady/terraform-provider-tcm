package internal

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestERP_toPayload_noID(t *testing.T) {
	r := &ERPResource{}
	m := &ERPResourceModel{
		ID:          types.StringValue("should-not-appear"),
		Name:        types.StringValue("MyERP"),
		Description: types.StringValue("desc"),
		Version:     types.StringValue("1.0"),
	}
	p := r.toPayload(m)
	// toPayload for ERP does NOT set ID
	if p.ID != nil {
		t.Errorf("expected p.ID nil (ERP toPayload never sets ID), got %v", *p.ID)
	}
	if p.Name == nil || *p.Name != "MyERP" {
		t.Errorf("expected Name MyERP")
	}
	if p.Description == nil || *p.Description != "desc" {
		t.Errorf("expected Description 'desc'")
	}
	if p.Version == nil || *p.Version != "1.0" {
		t.Errorf("expected Version 1.0")
	}
}

func TestERP_toPayload_nullFieldsOmitted(t *testing.T) {
	r := &ERPResource{}
	m := &ERPResourceModel{
		ID: types.StringNull(),
	}
	p := r.toPayload(m)
	if p.Name != nil {
		t.Errorf("expected Name nil when model Name is null")
	}
	if p.Description != nil {
		t.Errorf("expected Description nil when model Description is null")
	}
	if p.Version != nil {
		t.Errorf("expected Version nil when model Version is null")
	}
}

func TestERP_applyResult(t *testing.T) {
	id := "erp-result"
	name := "Result ERP"
	desc := "result desc"
	version := "2.0"
	deleted := true
	createdBy := "alice"
	createdDate := "2024-01-01"
	modifiedBy := "bob"
	modifiedDate := "2024-01-02"
	p := &erpPayload{
		ID:               &id,
		Name:             &name,
		Description:      &desc,
		Version:          &version,
		IsDeleted:        &deleted,
		CreatedBy:        &createdBy,
		CreatedDate:      &createdDate,
		LastModifiedBy:   &modifiedBy,
		LastModifiedDate: &modifiedDate,
	}
	m := &ERPResourceModel{}
	applyERPResult(m, p)

	if m.ID.ValueString() != "erp-result" {
		t.Errorf("expected ID erp-result, got %q", m.ID.ValueString())
	}
	if m.Name.ValueString() != "Result ERP" {
		t.Errorf("expected Name 'Result ERP', got %q", m.Name.ValueString())
	}
	if m.Description.ValueString() != "result desc" {
		t.Errorf("expected Description 'result desc', got %q", m.Description.ValueString())
	}
	if m.Version.ValueString() != "2.0" {
		t.Errorf("expected Version 2.0, got %q", m.Version.ValueString())
	}
	if !m.IsDeleted.ValueBool() {
		t.Errorf("expected IsDeleted true")
	}
	if m.CreatedBy.ValueString() != "alice" {
		t.Errorf("expected CreatedBy alice, got %q", m.CreatedBy.ValueString())
	}
	if m.LastModifiedBy.ValueString() != "bob" {
		t.Errorf("expected LastModifiedBy bob, got %q", m.LastModifiedBy.ValueString())
	}
}

func TestERP_applyResult_nilIsDeletedDefaultsFalse(t *testing.T) {
	p := &erpPayload{}
	m := &ERPResourceModel{}
	applyERPResult(m, p)
	if m.IsDeleted.ValueBool() {
		t.Errorf("expected IsDeleted false when payload IsDeleted is nil")
	}
}
