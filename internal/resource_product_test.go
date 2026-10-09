package internal

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestProduct_applyResult_withID(t *testing.T) {
	id := "prod-123"
	p := &productPayload{
		ID:          &id,
		Name:        "My Product",
		Description: "A description",
		IsDisabled:  false,
	}
	var m ProductResourceModel
	applyResult(&m, p)

	if m.ID.ValueString() != "prod-123" {
		t.Errorf("expected ID prod-123, got %s", m.ID.ValueString())
	}
	if m.Name.ValueString() != "My Product" {
		t.Errorf("expected Name My Product, got %s", m.Name.ValueString())
	}
	if m.Description.ValueString() != "A description" {
		t.Errorf("expected Description A description, got %s", m.Description.ValueString())
	}
	if m.IsDisabled.ValueBool() != false {
		t.Error("expected IsDisabled false")
	}
}

func TestProduct_applyResult_nilID(t *testing.T) {
	p := &productPayload{
		ID:          nil,
		Name:        "Product B",
		Description: "Desc B",
		IsDisabled:  true,
	}
	m := ProductResourceModel{
		ID: types.StringValue("existing-id"),
	}
	applyResult(&m, p)

	// ID should remain unchanged when payload ID is nil
	if m.ID.ValueString() != "existing-id" {
		t.Errorf("expected ID to remain existing-id, got %s", m.ID.ValueString())
	}
	if !m.IsDisabled.ValueBool() {
		t.Error("expected IsDisabled true")
	}
	if m.Name.ValueString() != "Product B" {
		t.Errorf("expected Name Product B, got %s", m.Name.ValueString())
	}
}

func TestProduct_applyResult_disabled(t *testing.T) {
	id := "prod-disabled"
	p := &productPayload{
		ID:          &id,
		Name:        "Disabled Product",
		Description: "Will be soft-deleted",
		IsDisabled:  true,
	}
	var m ProductResourceModel
	applyResult(&m, p)

	if !m.IsDisabled.ValueBool() {
		t.Error("expected IsDisabled to be true")
	}
	if m.ID.ValueString() != "prod-disabled" {
		t.Errorf("expected ID prod-disabled, got %s", m.ID.ValueString())
	}
}

func TestProduct_payload_fields(t *testing.T) {
	// Verify productPayload struct can hold all required fields.
	id := "p-1"
	p := productPayload{
		ID:          &id,
		Name:        "name",
		Description: "desc",
		IsDisabled:  false,
	}
	if *p.ID != "p-1" {
		t.Errorf("unexpected ID: %s", *p.ID)
	}
	if p.Name != "name" {
		t.Errorf("unexpected Name: %s", p.Name)
	}
}
