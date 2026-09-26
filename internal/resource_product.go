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

var _ resource.Resource = &ProductResource{}
var _ resource.ResourceWithImportState = &ProductResource{}

type ProductResource struct {
	client *TcmClient
}

type ProductResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	IsDisabled  types.Bool   `tfsdk:"is_disabled"`
}

// productPayload maps to POST /v1/Product and PUT /v1/Product/{id}.
type productPayload struct {
	ID          *string `json:"id,omitempty"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	IsDisabled  bool    `json:"isDisabled"`
}

func NewProductResource() resource.Resource {
	return &ProductResource{}
}

func (r *ProductResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_product"
}

func (r *ProductResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TCM Product. TCM has no hard-delete for products; destroying this resource sets is_disabled = true.",
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
				Description: "Product name.",
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Product description (max 110 characters).",
			},
			"is_disabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the product is disabled.",
			},
		},
	}
}

func (r *ProductResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*TcmClient)
	if !ok {
		resp.Diagnostics.AddError(
			"unexpected provider data type",
			fmt.Sprintf("expected *TcmClient, got %T", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *ProductResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProductResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(productPayload{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		IsDisabled:  plan.IsDisabled.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPost, "/v1/Product", body)
	if err != nil {
		resp.Diagnostics.AddError("create product failed", err.Error())
		return
	}

	applyResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ProductResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProductResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.request(ctx, http.MethodGet, "/v1/Product/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("read product failed", err.Error())
		return
	}

	applyResult(&state, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ProductResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProductResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := plan.ID.ValueString()
	body, err := json.Marshal(productPayload{
		ID:          &id,
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		IsDisabled:  plan.IsDisabled.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPut, "/v1/Product/"+id, body)
	if err != nil {
		resp.Diagnostics.AddError("update product failed", err.Error())
		return
	}

	applyResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete soft-deletes by setting isDisabled = true, since TCM has no hard-delete endpoint for products.
func (r *ProductResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProductResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	body, err := json.Marshal(productPayload{
		ID:          &id,
		Name:        state.Name.ValueString(),
		Description: state.Description.ValueString(),
		IsDisabled:  true,
	})
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	if _, err := r.request(ctx, http.MethodPut, "/v1/Product/"+id, body); err != nil {
		resp.Diagnostics.AddError("disable product failed", err.Error())
	}
}

func (r *ProductResource) request(ctx context.Context, method, path string, body []byte) (*productPayload, error) {
	var bodyReader *bytes.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, r.client.BaseURL+path, bodyReader)
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

	var result productPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

// ImportState allows importing an existing product by its TCM UUID.
// Usage: terraform import tcm_product.<name> <uuid>
func (r *ProductResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
}

func applyResult(m *ProductResourceModel, p *productPayload) {
	if p.ID != nil {
		m.ID = types.StringValue(*p.ID)
	}
	m.Name = types.StringValue(p.Name)
	m.Description = types.StringValue(p.Description)
	m.IsDisabled = types.BoolValue(p.IsDisabled)
}
