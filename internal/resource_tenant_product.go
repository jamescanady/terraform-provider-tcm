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

var _ resource.Resource = &TenantProductResource{}
var _ resource.ResourceWithImportState = &TenantProductResource{}

type TenantProductResource struct {
	client *TcmClient
}

type TenantProductResourceModel struct {
	ID                types.String `tfsdk:"id"`
	TenantID          types.String `tfsdk:"tenant_id"`
	ProductID         types.String `tfsdk:"product_id"`
	TenantProductCode types.String `tfsdk:"tenant_product_code"`
	IsDisabled        types.Bool   `tfsdk:"is_disabled"`
}

type tenantProductPayload struct {
	ID                *string `json:"id,omitempty"`
	TenantID          string  `json:"tenantId"`
	ProductID         string  `json:"productId"`
	TenantProductCode *string `json:"tenantProductCode,omitempty"`
	IsDisabled        bool    `json:"isDisabled"`
}

type tenantProductUpdatePayload struct {
	ID                *string `json:"id,omitempty"`
	TenantProductCode *string `json:"tenantProductCode,omitempty"`
	IsDisabled        bool    `json:"isDisabled"`
}

func NewTenantProductResource() resource.Resource {
	return &TenantProductResource{}
}

func (r *TenantProductResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_product"
}

func (r *TenantProductResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TCM TenantProduct mapping.",
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
			"product_id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the product.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tenant_product_code": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Optional tenant product code.",
			},
			"is_disabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the tenant product mapping is disabled.",
			},
		},
	}
}

func (r *TenantProductResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TenantProductResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TenantProductResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var code *string
	if !plan.TenantProductCode.IsNull() && !plan.TenantProductCode.IsUnknown() {
		v := plan.TenantProductCode.ValueString()
		code = &v
	}

	body, err := json.Marshal(tenantProductPayload{
		TenantID:          plan.TenantID.ValueString(),
		ProductID:         plan.ProductID.ValueString(),
		TenantProductCode: code,
		IsDisabled:        plan.IsDisabled.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPost, "/v1/TenantProduct", body)
	if err != nil {
		resp.Diagnostics.AddError("create tenant product failed", err.Error())
		return
	}

	applyTenantProductResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *TenantProductResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TenantProductResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.request(ctx, http.MethodGet, "/v1/TenantProduct/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("read tenant product failed", err.Error())
		return
	}

	applyTenantProductResult(&state, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *TenantProductResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TenantProductResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := plan.ID.ValueString()
	var code *string
	if !plan.TenantProductCode.IsNull() && !plan.TenantProductCode.IsUnknown() {
		v := plan.TenantProductCode.ValueString()
		code = &v
	}

	body, err := json.Marshal(tenantProductUpdatePayload{
		ID:                &id,
		TenantProductCode: code,
		IsDisabled:        plan.IsDisabled.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPut, "/v1/TenantProduct/"+id, body)
	if err != nil {
		resp.Diagnostics.AddError("update tenant product failed", err.Error())
		return
	}

	applyTenantProductResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *TenantProductResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TenantProductResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.request(ctx, http.MethodDelete, "/v1/TenantProduct/"+state.ID.ValueString(), nil); err != nil {
		resp.Diagnostics.AddError("delete tenant product failed", err.Error())
	}
}

// ImportState allows importing an existing tenant product by its TCM UUID.
// Usage: terraform import tcm_tenant_product.<name> <uuid>
func (r *TenantProductResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
}

func (r *TenantProductResource) request(ctx context.Context, method, apiPath string, body []byte) (*tenantProductPayload, error) {
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

	var result tenantProductPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func applyTenantProductResult(m *TenantProductResourceModel, p *tenantProductPayload) {
	if p.ID != nil {
		m.ID = types.StringValue(*p.ID)
	}
	m.TenantID = types.StringValue(p.TenantID)
	m.ProductID = types.StringValue(p.ProductID)
	if p.TenantProductCode != nil {
		m.TenantProductCode = types.StringValue(*p.TenantProductCode)
	} else {
		m.TenantProductCode = types.StringNull()
	}
	m.IsDisabled = types.BoolValue(p.IsDisabled)
}
