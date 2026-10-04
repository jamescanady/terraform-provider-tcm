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

var _ resource.Resource = &TenantProductEnvironmentResource{}
var _ resource.ResourceWithImportState = &TenantProductEnvironmentResource{}

type TenantProductEnvironmentResource struct {
	client *TcmClient
}

type TenantProductEnvironmentResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	TenantID             types.String `tfsdk:"tenant_id"`
	ProductEnvironmentID types.String `tfsdk:"product_environment_id"`
	NamespaceID          types.String `tfsdk:"namespace_id"`
	ProductTenantCode    types.String `tfsdk:"product_tenant_code"`
	ProductAlias         types.String `tfsdk:"product_alias"`
	IsDisabled           types.Bool   `tfsdk:"is_disabled"`
}

type tenantProductEnvironmentPayload struct {
	ID                   *string `json:"id,omitempty"`
	TenantID             string  `json:"tenantId"`
	ProductEnvironmentID string  `json:"productEnvironmentId"`
	NamespaceID          *string `json:"namespaceId,omitempty"`
	ProductTenantCode    *string `json:"productTenantCode,omitempty"`
	ProductAlias         *string `json:"productAlias,omitempty"`
	IsDisabled           bool    `json:"isDisabled"`
}

func NewTenantProductEnvironmentResource() resource.Resource {
	return &TenantProductEnvironmentResource{}
}

func (r *TenantProductEnvironmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_product_environment"
}

func (r *TenantProductEnvironmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TCM TenantProductEnvironment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID assigned by TCM on creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tenant_id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the tenant.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"product_environment_id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the product environment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"namespace_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of the namespace.",
			},
			"product_tenant_code": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Product tenant code.",
			},
			"product_alias": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Product alias.",
			},
			"is_disabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the tenant product environment is disabled.",
			},
		},
	}
}

func (r *TenantProductEnvironmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TenantProductEnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TenantProductEnvironmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPost, "/v1/TenantProductEnvironment", body)
	if err != nil {
		resp.Diagnostics.AddError("create tenant product environment failed", err.Error())
		return
	}

	applyTenantProductEnvironmentResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *TenantProductEnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TenantProductEnvironmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.request(ctx, http.MethodGet, "/v1/TenantProductEnvironment/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("read tenant product environment failed", err.Error())
		return
	}

	applyTenantProductEnvironmentResult(&state, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *TenantProductEnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TenantProductEnvironmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPut, "/v1/TenantProductEnvironment/"+plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("update tenant product environment failed", err.Error())
		return
	}

	applyTenantProductEnvironmentResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *TenantProductEnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TenantProductEnvironmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.request(ctx, http.MethodDelete, "/v1/TenantProductEnvironment/"+state.ID.ValueString(), nil); err != nil {
		resp.Diagnostics.AddError("delete tenant product environment failed", err.Error())
	}
}

// ImportState allows importing an existing tenant product environment by its TCM UUID.
// Usage: terraform import tcm_tenant_product_environment.<name> <uuid>
func (r *TenantProductEnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
}

func (r *TenantProductEnvironmentResource) toPayload(m *TenantProductEnvironmentResourceModel) *tenantProductEnvironmentPayload {
	p := &tenantProductEnvironmentPayload{
		TenantID:             m.TenantID.ValueString(),
		ProductEnvironmentID: m.ProductEnvironmentID.ValueString(),
		IsDisabled:           m.IsDisabled.ValueBool(),
	}
	if !m.ID.IsNull() && !m.ID.IsUnknown() {
		id := m.ID.ValueString()
		p.ID = &id
	}
	if !m.NamespaceID.IsNull() && !m.NamespaceID.IsUnknown() {
		v := m.NamespaceID.ValueString()
		p.NamespaceID = &v
	}
	if !m.ProductTenantCode.IsNull() && !m.ProductTenantCode.IsUnknown() {
		v := m.ProductTenantCode.ValueString()
		p.ProductTenantCode = &v
	}
	if !m.ProductAlias.IsNull() && !m.ProductAlias.IsUnknown() {
		v := m.ProductAlias.ValueString()
		p.ProductAlias = &v
	}
	return p
}

func (r *TenantProductEnvironmentResource) request(ctx context.Context, method, apiPath string, body []byte) (*tenantProductEnvironmentPayload, error) {
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

	var result tenantProductEnvironmentPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func applyTenantProductEnvironmentResult(m *TenantProductEnvironmentResourceModel, p *tenantProductEnvironmentPayload) {
	if p.ID != nil {
		m.ID = types.StringValue(*p.ID)
	}
	m.TenantID = types.StringValue(p.TenantID)
	m.ProductEnvironmentID = types.StringValue(p.ProductEnvironmentID)
	if p.NamespaceID != nil {
		m.NamespaceID = types.StringValue(*p.NamespaceID)
	} else {
		m.NamespaceID = types.StringNull()
	}
	if p.ProductTenantCode != nil {
		m.ProductTenantCode = types.StringValue(*p.ProductTenantCode)
	} else {
		m.ProductTenantCode = types.StringNull()
	}
	if p.ProductAlias != nil {
		m.ProductAlias = types.StringValue(*p.ProductAlias)
	} else {
		m.ProductAlias = types.StringNull()
	}
	m.IsDisabled = types.BoolValue(p.IsDisabled)
}
