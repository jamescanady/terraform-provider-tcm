package internal

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSystemInfo_toPayload_requiredOnly(t *testing.T) {
	r := &SystemInfoResource{}
	m := &SystemInfoResourceModel{
		ID:                   types.StringNull(),
		ProductID:            types.StringValue("product-001"),
		TenantID:             types.StringNull(),
		NamespaceID:          types.StringNull(),
		ProductEnvironmentID: types.StringNull(),
		FlowID:               types.StringNull(),
		ConnectionType:       types.StringNull(),
		Description:          types.StringNull(),
		Host:                 types.StringValue("https://api.example.com"),
		FlowVersion:          types.StringValue("1.0"),
		BaseApiPath:          types.StringValue("/api/v1"),
		Endpoint:             types.StringNull(),
		OAuthScope:           types.StringValue("openid profile"),
		IsDisabled:           types.BoolValue(false),
		NamespaceName:        types.StringNull(),
		TenantName:           types.StringNull(),
		ProductName:          types.StringNull(),
	}

	p := r.toPayload(m)

	if p.ProductID != "product-001" {
		t.Errorf("expected ProductID product-001, got %s", p.ProductID)
	}
	if p.Host == nil || *p.Host != "https://api.example.com" {
		t.Errorf("expected Host https://api.example.com, got %v", p.Host)
	}
	if p.FlowVersion == nil || *p.FlowVersion != "1.0" {
		t.Errorf("expected FlowVersion 1.0, got %v", p.FlowVersion)
	}
	if p.BaseApiPath == nil || *p.BaseApiPath != "/api/v1" {
		t.Errorf("expected BaseApiPath /api/v1, got %v", p.BaseApiPath)
	}
	if p.OAuthScope == nil || *p.OAuthScope != "openid profile" {
		t.Errorf("expected OAuthScope openid profile, got %v", p.OAuthScope)
	}
	if p.IsDisabled {
		t.Error("expected IsDisabled false")
	}
	// optional fields should be nil
	if p.ID != nil {
		t.Errorf("expected ID nil when model ID is null, got %v", p.ID)
	}
	if p.TenantID != nil {
		t.Error("expected TenantID nil when model TenantID is null")
	}
	if p.NamespaceID != nil {
		t.Error("expected NamespaceID nil when model NamespaceID is null")
	}
	if p.ConnectionType != nil {
		t.Error("expected ConnectionType nil when model ConnectionType is null")
	}
	if p.Description != nil {
		t.Error("expected Description nil when model Description is null")
	}
	if p.Endpoint != nil {
		t.Error("expected Endpoint nil when model Endpoint is null")
	}
}

func TestSystemInfo_toPayload_allFields(t *testing.T) {
	r := &SystemInfoResource{}
	m := &SystemInfoResourceModel{
		ID:                   types.StringValue("si-001"),
		ProductID:            types.StringValue("product-full"),
		TenantID:             types.StringValue("tenant-full"),
		NamespaceID:          types.StringValue("ns-full"),
		ProductEnvironmentID: types.StringValue("penv-full"),
		FlowID:               types.StringValue("flow-full"),
		ConnectionType:       types.StringValue("Http"),
		Description:          types.StringValue("A full description"),
		Host:                 types.StringValue("https://full.example.com"),
		FlowVersion:          types.StringValue("2.0"),
		BaseApiPath:          types.StringValue("/api/v2"),
		Endpoint:             types.StringValue("/endpoint"),
		OAuthScope:           types.StringValue("openid profile email"),
		IsDisabled:           types.BoolValue(true),
		NamespaceName:        types.StringNull(),
		TenantName:           types.StringNull(),
		ProductName:          types.StringNull(),
	}

	p := r.toPayload(m)

	if p.ID == nil || *p.ID != "si-001" {
		t.Errorf("expected ID si-001, got %v", p.ID)
	}
	if p.TenantID == nil || *p.TenantID != "tenant-full" {
		t.Errorf("expected TenantID tenant-full, got %v", p.TenantID)
	}
	if p.NamespaceID == nil || *p.NamespaceID != "ns-full" {
		t.Errorf("expected NamespaceID ns-full, got %v", p.NamespaceID)
	}
	if p.ProductEnvironmentID == nil || *p.ProductEnvironmentID != "penv-full" {
		t.Errorf("expected ProductEnvironmentID penv-full, got %v", p.ProductEnvironmentID)
	}
	if p.FlowID == nil || *p.FlowID != "flow-full" {
		t.Errorf("expected FlowID flow-full, got %v", p.FlowID)
	}
	if p.ConnectionType == nil || *p.ConnectionType != "Http" {
		t.Errorf("expected ConnectionType Http, got %v", p.ConnectionType)
	}
	if p.Description == nil || *p.Description != "A full description" {
		t.Errorf("expected Description A full description, got %v", p.Description)
	}
	if p.Endpoint == nil || *p.Endpoint != "/endpoint" {
		t.Errorf("expected Endpoint /endpoint, got %v", p.Endpoint)
	}
	if !p.IsDisabled {
		t.Error("expected IsDisabled true")
	}
}

func TestSystemInfo_applySystemInfoResult_allFields(t *testing.T) {
	id := "si-resp-001"
	tenantID := "tenant-resp"
	nsID := "ns-resp"
	penvID := "penv-resp"
	flowID := "flow-resp"
	connType := "Http"
	desc := "Response description"
	host := "https://resp.example.com"
	flowVer := "3.0"
	basePath := "/api/v3"
	endpoint := "/ep"
	oAuth := "openid"
	nsName := "Namespace Name"
	tenantName := "Tenant Name"
	productName := "Product Name"

	p := &systemInfoPayload{
		ID:                   &id,
		ProductID:            "product-resp",
		TenantID:             &tenantID,
		NamespaceID:          &nsID,
		ProductEnvironmentID: &penvID,
		FlowID:               &flowID,
		ConnectionType:       &connType,
		Description:          &desc,
		Host:                 &host,
		FlowVersion:          &flowVer,
		BaseApiPath:          &basePath,
		Endpoint:             &endpoint,
		OAuthScope:           &oAuth,
		IsDisabled:           false,
		NamespaceName:        &nsName,
		TenantName:           &tenantName,
		ProductName:          &productName,
	}

	var m SystemInfoResourceModel
	applySystemInfoResult(&m, p)

	if m.ID.ValueString() != "si-resp-001" {
		t.Errorf("expected ID si-resp-001, got %s", m.ID.ValueString())
	}
	if m.ProductID.ValueString() != "product-resp" {
		t.Errorf("expected ProductID product-resp, got %s", m.ProductID.ValueString())
	}
	if m.TenantID.ValueString() != "tenant-resp" {
		t.Errorf("expected TenantID tenant-resp, got %s", m.TenantID.ValueString())
	}
	if m.Host.ValueString() != "https://resp.example.com" {
		t.Errorf("expected Host https://resp.example.com, got %s", m.Host.ValueString())
	}
	if m.FlowVersion.ValueString() != "3.0" {
		t.Errorf("expected FlowVersion 3.0, got %s", m.FlowVersion.ValueString())
	}
	if m.BaseApiPath.ValueString() != "/api/v3" {
		t.Errorf("expected BaseApiPath /api/v3, got %s", m.BaseApiPath.ValueString())
	}
	if m.OAuthScope.ValueString() != "openid" {
		t.Errorf("expected OAuthScope openid, got %s", m.OAuthScope.ValueString())
	}
	if m.NamespaceName.ValueString() != "Namespace Name" {
		t.Errorf("expected NamespaceName Namespace Name, got %s", m.NamespaceName.ValueString())
	}
	if m.TenantName.ValueString() != "Tenant Name" {
		t.Errorf("expected TenantName Tenant Name, got %s", m.TenantName.ValueString())
	}
	if m.ProductName.ValueString() != "Product Name" {
		t.Errorf("expected ProductName Product Name, got %s", m.ProductName.ValueString())
	}
}

func TestSystemInfo_applySystemInfoResult_nullableNils(t *testing.T) {
	id := "si-nil-001"
	host := "https://nil.example.com"
	flowVer := "1.0"
	basePath := "/api"
	oAuth := "openid"
	p := &systemInfoPayload{
		ID:                   &id,
		ProductID:            "product-nil",
		TenantID:             nil,
		NamespaceID:          nil,
		ProductEnvironmentID: nil,
		FlowID:               nil,
		ConnectionType:       nil,
		Description:          nil,
		Host:                 &host,
		FlowVersion:          &flowVer,
		BaseApiPath:          &basePath,
		Endpoint:             nil,
		OAuthScope:           &oAuth,
		IsDisabled:           false,
		NamespaceName:        nil,
		TenantName:           nil,
		ProductName:          nil,
	}

	var m SystemInfoResourceModel
	applySystemInfoResult(&m, p)

	if !m.TenantID.IsNull() {
		t.Error("expected TenantID null when payload TenantID is nil")
	}
	if !m.NamespaceID.IsNull() {
		t.Error("expected NamespaceID null when payload NamespaceID is nil")
	}
	if !m.ProductEnvironmentID.IsNull() {
		t.Error("expected ProductEnvironmentID null when payload ProductEnvironmentID is nil")
	}
	if !m.FlowID.IsNull() {
		t.Error("expected FlowID null when payload FlowID is nil")
	}
	if !m.ConnectionType.IsNull() {
		t.Error("expected ConnectionType null when payload ConnectionType is nil")
	}
	if !m.Description.IsNull() {
		t.Error("expected Description null when payload Description is nil")
	}
	if !m.Endpoint.IsNull() {
		t.Error("expected Endpoint null when payload Endpoint is nil")
	}
	if !m.NamespaceName.IsNull() {
		t.Error("expected NamespaceName null when payload NamespaceName is nil")
	}
	if !m.TenantName.IsNull() {
		t.Error("expected TenantName null when payload TenantName is nil")
	}
	if !m.ProductName.IsNull() {
		t.Error("expected ProductName null when payload ProductName is nil")
	}
}

func TestSetOptionalString_nonNil(t *testing.T) {
	val := "hello"
	var target types.String
	setOptionalString(&target, &val)

	if target.ValueString() != "hello" {
		t.Errorf("expected hello, got %s", target.ValueString())
	}
}

func TestSetOptionalString_nil(t *testing.T) {
	var target types.String
	setOptionalString(&target, nil)

	if !target.IsNull() {
		t.Error("expected null when val is nil")
	}
}
