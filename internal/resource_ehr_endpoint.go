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

var _ resource.Resource = &EHREndpointResource{}
var _ resource.ResourceWithImportState = &EHREndpointResource{}

type EHREndpointResource struct {
	client *TcmClient
}

type EHREndpointResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	IsDisabled  types.Bool   `tfsdk:"is_disabled"`
}

type ehrEndpointPayload struct {
	ID          *string `json:"id,omitempty"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	IsDisabled  bool    `json:"isDisabled"`
}

func NewEHREndpointResource() resource.Resource {
	return &EHREndpointResource{}
}

func (r *EHREndpointResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ehr_endpoint"
}

func (r *EHREndpointResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TCM EHR Endpoint.",
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
				Description: "Name of the EHR endpoint.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description of the EHR endpoint.",
			},
			"is_disabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the EHR endpoint is disabled.",
			},
		},
	}
}

func (r *EHREndpointResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EHREndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EHREndpointResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPost, "/v1/ehr/endpoint", body)
	if err != nil {
		resp.Diagnostics.AddError("create ehr endpoint failed", err.Error())
		return
	}

	applyEHREndpointResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EHREndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EHREndpointResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.request(ctx, http.MethodGet, "/v1/ehr/endpoint/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("read ehr endpoint failed", err.Error())
		return
	}

	applyEHREndpointResult(&state, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *EHREndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EHREndpointResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPut, "/v1/ehr/endpoint/"+plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("update ehr endpoint failed", err.Error())
		return
	}

	applyEHREndpointResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EHREndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EHREndpointResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.request(ctx, http.MethodDelete, "/v1/ehr/endpoint/"+state.ID.ValueString(), nil); err != nil {
		resp.Diagnostics.AddError("delete ehr endpoint failed", err.Error())
	}
}

// ImportState allows importing an existing EHR endpoint by its TCM UUID.
// Usage: terraform import tcm_ehr_endpoint.<name> <uuid>
func (r *EHREndpointResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
}

func (r *EHREndpointResource) toPayload(m *EHREndpointResourceModel) *ehrEndpointPayload {
	p := &ehrEndpointPayload{
		Name:       m.Name.ValueString(),
		IsDisabled: m.IsDisabled.ValueBool(),
	}
	if !m.ID.IsNull() && !m.ID.IsUnknown() {
		id := m.ID.ValueString()
		p.ID = &id
	}
	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		v := m.Description.ValueString()
		p.Description = &v
	}
	return p
}

func (r *EHREndpointResource) request(ctx context.Context, method, apiPath string, body []byte) (*ehrEndpointPayload, error) {
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

	var result ehrEndpointPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func applyEHREndpointResult(m *EHREndpointResourceModel, p *ehrEndpointPayload) {
	if p.ID != nil {
		m.ID = types.StringValue(*p.ID)
	}
	m.Name = types.StringValue(p.Name)
	setOptionalString(&m.Description, p.Description)
	m.IsDisabled = types.BoolValue(p.IsDisabled)
}
