package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &SystemInfoResource{}
var _ resource.ResourceWithImportState = &SystemInfoResource{}

type SystemInfoResource struct {
	client *TcmClient
}

type SystemInfoResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	ProductID            types.String `tfsdk:"product_id"`
	TenantID             types.String `tfsdk:"tenant_id"`
	NamespaceID          types.String `tfsdk:"namespace_id"`
	ProductEnvironmentID types.String `tfsdk:"product_environment_id"`
	FlowID               types.String `tfsdk:"flow_id"`
	ConnectionType       types.String `tfsdk:"connection_type"`
	Description          types.String `tfsdk:"description"`
	Host                 types.String `tfsdk:"host"`
	FlowVersion          types.String `tfsdk:"flow_version"`
	BaseApiPath          types.String `tfsdk:"base_api_path"`
	Endpoint             types.String `tfsdk:"endpoint"`
	OAuthScope           types.String `tfsdk:"o_auth_scope"`
	IsDisabled           types.Bool   `tfsdk:"is_disabled"`
	NamespaceName        types.String `tfsdk:"namespace_name"`
	TenantName           types.String `tfsdk:"tenant_name"`
	ProductName          types.String `tfsdk:"product_name"`
}

// systemInfoPayload is used for both request body and response decoding.
// Host, FlowVersion, BaseApiPath, OAuthScope are required in requests but nullable in responses.
type systemInfoPayload struct {
	ID                   *string `json:"id,omitempty"`
	ProductID            string  `json:"productId"`
	TenantID             *string `json:"tenantId,omitempty"`
	NamespaceID          *string `json:"namespaceId,omitempty"`
	ProductEnvironmentID *string `json:"productEnvironmentId,omitempty"`
	FlowID               *string `json:"flowId,omitempty"`
	ConnectionType       *string `json:"connectionType,omitempty"`
	Description          *string `json:"description,omitempty"`
	Host                 *string `json:"host,omitempty"`
	FlowVersion          *string `json:"flowVersion,omitempty"`
	BaseApiPath          *string `json:"baseApiPath,omitempty"`
	Endpoint             *string `json:"endpoint,omitempty"`
	OAuthScope           *string `json:"oAuthScope,omitempty"`
	IsDisabled           bool    `json:"isDisabled"`
	NamespaceName        *string `json:"namespaceName,omitempty"`
	TenantName           *string `json:"tenantName,omitempty"`
	ProductName          *string `json:"productName,omitempty"`
}

func NewSystemInfoResource() resource.Resource {
	return &SystemInfoResource{}
}

func (r *SystemInfoResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_info"
}

func (r *SystemInfoResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TCM SystemInfo.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID assigned by TCM on creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"product_id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the product.",
			},
			"tenant_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of the tenant.",
			},
			"namespace_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of the namespace.",
			},
			"product_environment_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of the product environment.",
			},
			"flow_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Flow identifier.",
			},
			"connection_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Connection type (e.g. Http).",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description (max 110 characters).",
			},
			"host": schema.StringAttribute{
				Required:    true,
				Description: "Host.",
			},
			"flow_version": schema.StringAttribute{
				Required:    true,
				Description: "Flow version.",
			},
			"base_api_path": schema.StringAttribute{
				Required:    true,
				Description: "Base API path.",
			},
			"endpoint": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Endpoint.",
			},
			"o_auth_scope": schema.StringAttribute{
				Required:    true,
				Description: "OAuth scope.",
			},
			"is_disabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the system info is disabled.",
			},
			"namespace_name": schema.StringAttribute{
				Computed:    true,
				Description: "Namespace name (read-only).",
			},
			"tenant_name": schema.StringAttribute{
				Computed:    true,
				Description: "Tenant name (read-only).",
			},
			"product_name": schema.StringAttribute{
				Computed:    true,
				Description: "Product name (read-only).",
			},
		},
	}
}

func (r *SystemInfoResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*TcmClient)
	if !ok {
		resp.Diagnostics.AddError("unexpected provider data type", fmt.Sprintf("expected *TcmClient, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *SystemInfoResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SystemInfoResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPost, "/v1/SystemInfo", body)
	if err != nil {
		resp.Diagnostics.AddError("create system info failed", err.Error())
		return
	}

	applySystemInfoResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *SystemInfoResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SystemInfoResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.request(ctx, http.MethodGet, "/v1/SystemInfo/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("read system info failed", err.Error())
		return
	}

	applySystemInfoResult(&state, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *SystemInfoResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SystemInfoResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPut, "/v1/SystemInfo/"+plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("update system info failed", err.Error())
		return
	}

	applySystemInfoResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *SystemInfoResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SystemInfoResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.request(ctx, http.MethodDelete, "/v1/SystemInfo/"+state.ID.ValueString(), nil); err != nil {
		resp.Diagnostics.AddError("delete system info failed", err.Error())
	}
}

// ImportState allows importing an existing system info by its TCM UUID.
// Usage: terraform import tcm_system_info.<name> <uuid>
func (r *SystemInfoResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
}

func (r *SystemInfoResource) toPayload(m *SystemInfoResourceModel) *systemInfoPayload {
	host := m.Host.ValueString()
	flowVersion := m.FlowVersion.ValueString()
	baseApiPath := m.BaseApiPath.ValueString()
	oAuthScope := m.OAuthScope.ValueString()

	p := &systemInfoPayload{
		ProductID:   m.ProductID.ValueString(),
		Host:        &host,
		FlowVersion: &flowVersion,
		BaseApiPath: &baseApiPath,
		OAuthScope:  &oAuthScope,
		IsDisabled:  m.IsDisabled.ValueBool(),
	}
	if !m.ID.IsNull() && !m.ID.IsUnknown() {
		id := m.ID.ValueString()
		p.ID = &id
	}
	if !m.TenantID.IsNull() && !m.TenantID.IsUnknown() {
		v := m.TenantID.ValueString()
		p.TenantID = &v
	}
	if !m.NamespaceID.IsNull() && !m.NamespaceID.IsUnknown() {
		v := m.NamespaceID.ValueString()
		p.NamespaceID = &v
	}
	if !m.ProductEnvironmentID.IsNull() && !m.ProductEnvironmentID.IsUnknown() {
		v := m.ProductEnvironmentID.ValueString()
		p.ProductEnvironmentID = &v
	}
	if !m.FlowID.IsNull() && !m.FlowID.IsUnknown() {
		v := m.FlowID.ValueString()
		p.FlowID = &v
	}
	if !m.ConnectionType.IsNull() && !m.ConnectionType.IsUnknown() {
		v := m.ConnectionType.ValueString()
		p.ConnectionType = &v
	}
	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		v := m.Description.ValueString()
		p.Description = &v
	}
	if !m.Endpoint.IsNull() && !m.Endpoint.IsUnknown() {
		v := m.Endpoint.ValueString()
		p.Endpoint = &v
	}
	return p
}

func (r *SystemInfoResource) request(ctx context.Context, method, apiPath string, body []byte) (*systemInfoPayload, error) {
	var bodyReader *bytes.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, r.client.BaseURL+apiPath, bodyReader)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+r.client.Token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 400 {
		var errBody map[string]any
		json.NewDecoder(httpResp.Body).Decode(&errBody) //nolint:errcheck
		return nil, fmt.Errorf("HTTP %d: %v", httpResp.StatusCode, errBody)
	}

	if httpResp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	var result systemInfoPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func applySystemInfoResult(m *SystemInfoResourceModel, p *systemInfoPayload) {
	if p.ID != nil {
		m.ID = types.StringValue(*p.ID)
	}
	m.ProductID = types.StringValue(p.ProductID)
	setOptionalString(&m.TenantID, p.TenantID)
	setOptionalString(&m.NamespaceID, p.NamespaceID)
	setOptionalString(&m.ProductEnvironmentID, p.ProductEnvironmentID)
	setOptionalString(&m.FlowID, p.FlowID)
	setOptionalString(&m.ConnectionType, p.ConnectionType)
	setOptionalString(&m.Description, p.Description)
	if p.Host != nil {
		m.Host = types.StringValue(*p.Host)
	}
	if p.FlowVersion != nil {
		m.FlowVersion = types.StringValue(*p.FlowVersion)
	}
	if p.BaseApiPath != nil {
		m.BaseApiPath = types.StringValue(*p.BaseApiPath)
	}
	setOptionalString(&m.Endpoint, p.Endpoint)
	if p.OAuthScope != nil {
		m.OAuthScope = types.StringValue(*p.OAuthScope)
	}
	m.IsDisabled = types.BoolValue(p.IsDisabled)
	setOptionalString(&m.NamespaceName, p.NamespaceName)
	setOptionalString(&m.TenantName, p.TenantName)
	setOptionalString(&m.ProductName, p.ProductName)
}

func setOptionalString(target *types.String, val *string) {
	if val != nil {
		*target = types.StringValue(*val)
	} else {
		*target = types.StringNull()
	}
}
