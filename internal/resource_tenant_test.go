package internal

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestTenant_applyTenantResult_withID(t *testing.T) {
	id := "tenant-001"
	p := &tenantPayload{
		ID:               &id,
		Name:             "Test Tenant",
		Description:      "A test tenant",
		GlobalTenantCode: "GTC001",
		TenantShortCode:  "tsc001",
		IsDisabled:       false,
	}
	var m TenantResourceModel
	applyTenantResult(&m, p)

	if m.ID.ValueString() != "tenant-001" {
		t.Errorf("expected ID tenant-001, got %s", m.ID.ValueString())
	}
	if m.Name.ValueString() != "Test Tenant" {
		t.Errorf("expected Name Test Tenant, got %s", m.Name.ValueString())
	}
	if m.Description.ValueString() != "A test tenant" {
		t.Errorf("expected Description A test tenant, got %s", m.Description.ValueString())
	}
	if m.GlobalTenantCode.ValueString() != "GTC001" {
		t.Errorf("expected GlobalTenantCode GTC001, got %s", m.GlobalTenantCode.ValueString())
	}
	if m.TenantShortCode.ValueString() != "tsc001" {
		t.Errorf("expected TenantShortCode tsc001, got %s", m.TenantShortCode.ValueString())
	}
	if m.IsDisabled.ValueBool() {
		t.Error("expected IsDisabled false")
	}
}

func TestTenant_applyTenantResult_nilID(t *testing.T) {
	p := &tenantPayload{
		ID:               nil,
		Name:             "Tenant B",
		Description:      "Desc B",
		GlobalTenantCode: "GTC002",
		TenantShortCode:  "",
		IsDisabled:       false,
	}
	m := TenantResourceModel{
		ID: types.StringValue("pre-existing-id"),
	}
	applyTenantResult(&m, p)

	// ID should be unchanged when payload ID is nil
	if m.ID.ValueString() != "pre-existing-id" {
		t.Errorf("expected ID pre-existing-id, got %s", m.ID.ValueString())
	}
}

func TestTenant_applyTenantResult_emptyShortCode(t *testing.T) {
	id := "tenant-sc"
	p := &tenantPayload{
		ID:               &id,
		Name:             "Tenant SC",
		Description:      "Desc SC",
		GlobalTenantCode: "GTCSC",
		TenantShortCode:  "", // empty — should not update model field
		IsDisabled:       false,
	}
	m := TenantResourceModel{
		TenantShortCode: types.StringValue("original-code"),
	}
	applyTenantResult(&m, p)

	// TenantShortCode should remain unchanged when payload value is empty
	if m.TenantShortCode.ValueString() != "original-code" {
		t.Errorf("expected TenantShortCode to remain original-code, got %s", m.TenantShortCode.ValueString())
	}
}

func TestTenant_applyTenantResult_disabled(t *testing.T) {
	id := "tenant-dis"
	p := &tenantPayload{
		ID:               &id,
		Name:             "Disabled Tenant",
		Description:      "This tenant is disabled",
		GlobalTenantCode: "GTC-DIS",
		TenantShortCode:  "tdis",
		IsDisabled:       true,
	}
	var m TenantResourceModel
	applyTenantResult(&m, p)

	if !m.IsDisabled.ValueBool() {
		t.Error("expected IsDisabled true")
	}
}
