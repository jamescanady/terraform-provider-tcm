package internal

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestTenantProductEnvironment_toPayload_defaults(t *testing.T) {
	r := &TenantProductEnvironmentResource{}
	m := &TenantProductEnvironmentResourceModel{
		ID:                   types.StringNull(),
		TenantID:             types.StringValue("tenant-abc"),
		ProductEnvironmentID: types.StringValue("penv-xyz"),
		NamespaceID:          types.StringNull(),
		ProductTenantCode:    types.StringNull(),
		ProductAlias:         types.StringNull(),
		IsDisabled:           types.BoolValue(false),
	}

	p := r.toPayload(m)

	if p.TenantID != "tenant-abc" {
		t.Errorf("expected TenantID tenant-abc, got %s", p.TenantID)
	}
	if p.ProductEnvironmentID != "penv-xyz" {
		t.Errorf("expected ProductEnvironmentID penv-xyz, got %s", p.ProductEnvironmentID)
	}
	if p.ID != nil {
		t.Errorf("expected ID nil when model ID is null, got %v", p.ID)
	}
	if p.NamespaceID != nil {
		t.Error("expected NamespaceID nil when model NamespaceID is null")
	}
	if p.ProductTenantCode != nil {
		t.Error("expected ProductTenantCode nil when model ProductTenantCode is null")
	}
	if p.ProductAlias != nil {
		t.Error("expected ProductAlias nil when model ProductAlias is null")
	}
	if p.IsDisabled {
		t.Error("expected IsDisabled false")
	}
}

func TestTenantProductEnvironment_toPayload_withAllFields(t *testing.T) {
	r := &TenantProductEnvironmentResource{}
	m := &TenantProductEnvironmentResourceModel{
		ID:                   types.StringValue("tpe-001"),
		TenantID:             types.StringValue("tenant-full"),
		ProductEnvironmentID: types.StringValue("penv-full"),
		NamespaceID:          types.StringValue("ns-001"),
		ProductTenantCode:    types.StringValue("PTC001"),
		ProductAlias:         types.StringValue("alias-a"),
		IsDisabled:           types.BoolValue(true),
	}

	p := r.toPayload(m)

	if p.ID == nil || *p.ID != "tpe-001" {
		t.Errorf("expected ID tpe-001, got %v", p.ID)
	}
	if p.NamespaceID == nil || *p.NamespaceID != "ns-001" {
		t.Errorf("expected NamespaceID ns-001, got %v", p.NamespaceID)
	}
	if p.ProductTenantCode == nil || *p.ProductTenantCode != "PTC001" {
		t.Errorf("expected ProductTenantCode PTC001, got %v", p.ProductTenantCode)
	}
	if p.ProductAlias == nil || *p.ProductAlias != "alias-a" {
		t.Errorf("expected ProductAlias alias-a, got %v", p.ProductAlias)
	}
	if !p.IsDisabled {
		t.Error("expected IsDisabled true")
	}
}

func TestTenantProductEnvironment_applyResult_allFields(t *testing.T) {
	id := "tpe-resp-001"
	ns := "ns-resp"
	code := "PTCRESP"
	alias := "alias-resp"
	p := &tenantProductEnvironmentPayload{
		ID:                   &id,
		TenantID:             "tenant-resp",
		ProductEnvironmentID: "penv-resp",
		NamespaceID:          &ns,
		ProductTenantCode:    &code,
		ProductAlias:         &alias,
		IsDisabled:           false,
	}
	var m TenantProductEnvironmentResourceModel
	applyTenantProductEnvironmentResult(&m, p)

	if m.ID.ValueString() != "tpe-resp-001" {
		t.Errorf("expected ID tpe-resp-001, got %s", m.ID.ValueString())
	}
	if m.TenantID.ValueString() != "tenant-resp" {
		t.Errorf("expected TenantID tenant-resp, got %s", m.TenantID.ValueString())
	}
	if m.ProductEnvironmentID.ValueString() != "penv-resp" {
		t.Errorf("expected ProductEnvironmentID penv-resp, got %s", m.ProductEnvironmentID.ValueString())
	}
	if m.NamespaceID.ValueString() != "ns-resp" {
		t.Errorf("expected NamespaceID ns-resp, got %s", m.NamespaceID.ValueString())
	}
	if m.ProductTenantCode.ValueString() != "PTCRESP" {
		t.Errorf("expected ProductTenantCode PTCRESP, got %s", m.ProductTenantCode.ValueString())
	}
	if m.ProductAlias.ValueString() != "alias-resp" {
		t.Errorf("expected ProductAlias alias-resp, got %s", m.ProductAlias.ValueString())
	}
}

func TestTenantProductEnvironment_applyResult_nullableNils(t *testing.T) {
	id := "tpe-nil-001"
	p := &tenantProductEnvironmentPayload{
		ID:                   &id,
		TenantID:             "tenant-nil",
		ProductEnvironmentID: "penv-nil",
		NamespaceID:          nil,
		ProductTenantCode:    nil,
		ProductAlias:         nil,
		IsDisabled:           false,
	}
	var m TenantProductEnvironmentResourceModel
	applyTenantProductEnvironmentResult(&m, p)

	if !m.NamespaceID.IsNull() {
		t.Error("expected NamespaceID to be null when payload NamespaceID is nil")
	}
	if !m.ProductTenantCode.IsNull() {
		t.Error("expected ProductTenantCode to be null when payload ProductTenantCode is nil")
	}
	if !m.ProductAlias.IsNull() {
		t.Error("expected ProductAlias to be null when payload ProductAlias is nil")
	}
}
