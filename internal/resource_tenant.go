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

var _ resource.Resource = &TenantResource{}
var _ resource.ResourceWithImportState = &TenantResource{}

type TenantResource struct {
	client *TcmClient
}

type TenantResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	GlobalTenantCode types.String `tfsdk:"global_tenant_code"`
	TenantShortCode  types.String `tfsdk:"tenant_short_code"`
	IsDisabled       types.Bool   `tfsdk:"is_disabled"`
}

type tenantPayload struct {
	ID               *string `json:"id,omitempty"`
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	GlobalTenantCode string  `json:"globalTenantCode"`
	TenantShortCode  string  `json:"tenantShortCode,omitempty"`
	IsDisabled       bool    `json:"isDisabled"`
}

type tenantUpdatePayload struct {
	ID               *string `json:"id,omitempty"`
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	GlobalTenantCode string  `json:"globalTenantCode"`
	IsDisabled       bool    `json:"isDisabled"`
}

func NewTenantResource() resource.Resource {
	return &TenantResource{}
}

func (r *TenantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant"
}

func (r *TenantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TCM Tenant.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID assigned by TCM on creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Tenant name.",
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Tenant description (max 110 characters).",
			},
			"global_tenant_code": schema.StringAttribute{
				Required:    true,
				Description: "Global tenant code.",
			},
			"tenant_short_code": schema.StringAttribute{
				Required:    true,
				Description: "Tenant short code (lowercase alphanumeric and underscore). Cannot be changed after creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_disabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the tenant is disabled.",
			},
		},
	}
}

func (r *TenantResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TenantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TenantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(tenantPayload{
		Name:             plan.Name.ValueString(),
		Description:      plan.Description.ValueString(),
		GlobalTenantCode: plan.GlobalTenantCode.ValueString(),
		TenantShortCode:  plan.TenantShortCode.ValueString(),
		IsDisabled:       plan.IsDisabled.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPost, "/v1/Tenant", body)
	if err != nil {
		resp.Diagnostics.AddError("create tenant failed", err.Error())
		return
	}

	applyTenantResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *TenantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TenantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.request(ctx, http.MethodGet, "/v1/Tenant/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("read tenant failed", err.Error())
		return
	}

	applyTenantResult(&state, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *TenantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TenantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := plan.ID.ValueString()
	body, err := json.Marshal(tenantUpdatePayload{
		ID:               &id,
		Name:             plan.Name.ValueString(),
		Description:      plan.Description.ValueString(),
		GlobalTenantCode: plan.GlobalTenantCode.ValueString(),
		IsDisabled:       plan.IsDisabled.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPut, "/v1/Tenant/"+id, body)
	if err != nil {
		resp.Diagnostics.AddError("update tenant failed", err.Error())
		return
	}

	applyTenantResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *TenantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TenantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.request(ctx, http.MethodDelete, "/v1/Tenant/"+state.ID.ValueString(), nil); err != nil {
		resp.Diagnostics.AddError("delete tenant failed", err.Error())
	}
}

// ImportState allows importing an existing tenant by its TCM UUID.
// Usage: terraform import tcm_tenant.<name> <uuid>
func (r *TenantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
}

func (r *TenantResource) request(ctx context.Context, method, apiPath string, body []byte) (*tenantPayload, error) {
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

	var result tenantPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func applyTenantResult(m *TenantResourceModel, p *tenantPayload) {
	if p.ID != nil {
		m.ID = types.StringValue(*p.ID)
	}
	m.Name = types.StringValue(p.Name)
	m.Description = types.StringValue(p.Description)
	m.GlobalTenantCode = types.StringValue(p.GlobalTenantCode)
	if p.TenantShortCode != "" {
		m.TenantShortCode = types.StringValue(p.TenantShortCode)
	}
	m.IsDisabled = types.BoolValue(p.IsDisabled)
}
