package internal

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestTenantProduct_applyTenantProductResult_withID(t *testing.T) {
	id := "tp-001"
	code := "TPC001"
	p := &tenantProductPayload{
		ID:                &id,
		TenantID:          "tenant-abc",
		ProductID:         "product-xyz",
		TenantProductCode: &code,
		IsDisabled:        false,
	}
	var m TenantProductResourceModel
	applyTenantProductResult(&m, p)

	if m.ID.ValueString() != "tp-001" {
		t.Errorf("expected ID tp-001, got %s", m.ID.ValueString())
	}
	if m.TenantID.ValueString() != "tenant-abc" {
		t.Errorf("expected TenantID tenant-abc, got %s", m.TenantID.ValueString())
	}
	if m.ProductID.ValueString() != "product-xyz" {
		t.Errorf("expected ProductID product-xyz, got %s", m.ProductID.ValueString())
	}
	if m.TenantProductCode.ValueString() != "TPC001" {
		t.Errorf("expected TenantProductCode TPC001, got %s", m.TenantProductCode.ValueString())
	}
	if m.IsDisabled.ValueBool() {
		t.Error("expected IsDisabled false")
	}
}

func TestTenantProduct_applyTenantProductResult_nilTenantProductCode(t *testing.T) {
	id := "tp-002"
	p := &tenantProductPayload{
		ID:                &id,
		TenantID:          "tenant-def",
		ProductID:         "product-ghi",
		TenantProductCode: nil,
		IsDisabled:        false,
	}
	var m TenantProductResourceModel
	applyTenantProductResult(&m, p)

	if !m.TenantProductCode.IsNull() {
		t.Error("expected TenantProductCode to be null when payload TenantProductCode is nil")
	}
}

func TestTenantProduct_applyTenantProductResult_nilID(t *testing.T) {
	p := &tenantProductPayload{
		ID:                nil,
		TenantID:          "tenant-xyz",
		ProductID:         "product-abc",
		TenantProductCode: nil,
		IsDisabled:        false,
	}
	m := TenantProductResourceModel{
		ID: types.StringValue("existing-tp-id"),
	}
	applyTenantProductResult(&m, p)

	// ID should remain unchanged when payload ID is nil
	if m.ID.ValueString() != "existing-tp-id" {
		t.Errorf("expected ID existing-tp-id, got %s", m.ID.ValueString())
	}
}

func TestTenantProduct_applyTenantProductResult_disabled(t *testing.T) {
	id := "tp-dis"
	p := &tenantProductPayload{
		ID:                &id,
		TenantID:          "tenant-1",
		ProductID:         "product-1",
		TenantProductCode: nil,
		IsDisabled:        true,
	}
	var m TenantProductResourceModel
	applyTenantProductResult(&m, p)

	if !m.IsDisabled.ValueBool() {
		t.Error("expected IsDisabled true")
	}
}
