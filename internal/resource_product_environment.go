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

var _ resource.Resource = &ProductEnvironmentResource{}
var _ resource.ResourceWithImportState = &ProductEnvironmentResource{}

type ProductEnvironmentResource struct {
	client *TcmClient
}

type ProductEnvironmentResourceModel struct {
	ID         types.String `tfsdk:"id"`
	ProductID  types.String `tfsdk:"product_id"`
	Name       types.String `tfsdk:"name"`
	IsDisabled types.Bool   `tfsdk:"is_disabled"`
}

type productEnvironmentPayload struct {
	ID         *string `json:"id,omitempty"`
	ProductID  string  `json:"productId"`
	Name       *string `json:"name,omitempty"`
	IsDisabled bool    `json:"isDisabled"`
}

func NewProductEnvironmentResource() resource.Resource {
	return &ProductEnvironmentResource{}
}

func (r *ProductEnvironmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_product_environment"
}

func (r *ProductEnvironmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TCM ProductEnvironment.",
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Environment name.",
			},
			"is_disabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the product environment is disabled.",
			},
		},
	}
}

func (r *ProductEnvironmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProductEnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProductEnvironmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPost, "/v1/ProductEnvironment", body)
	if err != nil {
		resp.Diagnostics.AddError("create product environment failed", err.Error())
		return
	}

	applyProductEnvironmentResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ProductEnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProductEnvironmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.request(ctx, http.MethodGet, "/v1/ProductEnvironment/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("read product environment failed", err.Error())
		return
	}

	applyProductEnvironmentResult(&state, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ProductEnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProductEnvironmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPut, "/v1/ProductEnvironment/"+plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("update product environment failed", err.Error())
		return
	}

	applyProductEnvironmentResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ProductEnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProductEnvironmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.request(ctx, http.MethodDelete, "/v1/ProductEnvironment/"+state.ID.ValueString(), nil); err != nil {
		resp.Diagnostics.AddError("delete product environment failed", err.Error())
	}
}

// ImportState allows importing an existing product environment by its TCM UUID.
// Usage: terraform import tcm_product_environment.<name> <uuid>
func (r *ProductEnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
}

func (r *ProductEnvironmentResource) toPayload(m *ProductEnvironmentResourceModel) *productEnvironmentPayload {
	p := &productEnvironmentPayload{
		ProductID:  m.ProductID.ValueString(),
		IsDisabled: m.IsDisabled.ValueBool(),
	}
	if !m.ID.IsNull() && !m.ID.IsUnknown() {
		id := m.ID.ValueString()
		p.ID = &id
	}
	if !m.Name.IsNull() && !m.Name.IsUnknown() {
		name := m.Name.ValueString()
		p.Name = &name
	}
	return p
}

func (r *ProductEnvironmentResource) request(ctx context.Context, method, apiPath string, body []byte) (*productEnvironmentPayload, error) {
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

	var result productEnvironmentPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func applyProductEnvironmentResult(m *ProductEnvironmentResourceModel, p *productEnvironmentPayload) {
	if p.ID != nil {
		m.ID = types.StringValue(*p.ID)
	}
	m.ProductID = types.StringValue(p.ProductID)
	if p.Name != nil {
		m.Name = types.StringValue(*p.Name)
	} else {
		m.Name = types.StringNull()
	}
	m.IsDisabled = types.BoolValue(p.IsDisabled)
}
