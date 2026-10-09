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

var _ resource.Resource = &EHRDestinationResource{}
var _ resource.ResourceWithImportState = &EHRDestinationResource{}

type EHRDestinationResource struct {
	client *TcmClient
}

type EHRDestinationResourceModel struct {
	ID             types.String `tfsdk:"id"`
	JsonMetadataID types.String `tfsdk:"json_metadata_id"`
	DestinationID  types.String `tfsdk:"destination_id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	IsDisabled     types.Bool   `tfsdk:"is_disabled"`
}

type ehrDestinationPayload struct {
	ID             *string `json:"id,omitempty"`
	JsonMetadataID *string `json:"jsonMetadataId,omitempty"`
	DestinationID  *string `json:"destinationId,omitempty"`
	Name           *string `json:"name,omitempty"`
	Description    *string `json:"description,omitempty"`
	IsDisabled     bool    `json:"isDisabled"`
}

func NewEHRDestinationResource() resource.Resource {
	return &EHRDestinationResource{}
}

func (r *EHRDestinationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ehr_destination"
}

func (r *EHRDestinationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TCM EHR Destination.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID assigned by TCM on creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"json_metadata_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of the JSON metadata.",
			},
			"destination_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of the destination.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Name of the EHR destination.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description of the EHR destination.",
			},
			"is_disabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the EHR destination is disabled.",
			},
		},
	}
}

func (r *EHRDestinationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EHRDestinationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EHRDestinationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPost, "/v1/ehr/destination", body)
	if err != nil {
		resp.Diagnostics.AddError("create ehr destination failed", err.Error())
		return
	}

	applyEHRDestinationResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EHRDestinationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EHRDestinationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.request(ctx, http.MethodGet, "/v1/ehr/destination/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("read ehr destination failed", err.Error())
		return
	}

	applyEHRDestinationResult(&state, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *EHRDestinationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EHRDestinationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPut, "/v1/ehr/destination/"+plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("update ehr destination failed", err.Error())
		return
	}

	applyEHRDestinationResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EHRDestinationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EHRDestinationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.request(ctx, http.MethodDelete, "/v1/ehr/destination/"+state.ID.ValueString(), nil); err != nil {
		resp.Diagnostics.AddError("delete ehr destination failed", err.Error())
	}
}

// ImportState allows importing an existing EHR destination by its TCM UUID.
// Usage: terraform import tcm_ehr_destination.<name> <uuid>
func (r *EHRDestinationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
}

func (r *EHRDestinationResource) toPayload(m *EHRDestinationResourceModel) *ehrDestinationPayload {
	p := &ehrDestinationPayload{
		IsDisabled: m.IsDisabled.ValueBool(),
	}
	if !m.ID.IsNull() && !m.ID.IsUnknown() {
		v := m.ID.ValueString()
		p.ID = &v
	}
	if !m.JsonMetadataID.IsNull() && !m.JsonMetadataID.IsUnknown() {
		v := m.JsonMetadataID.ValueString()
		p.JsonMetadataID = &v
	}
	if !m.DestinationID.IsNull() && !m.DestinationID.IsUnknown() {
		v := m.DestinationID.ValueString()
		p.DestinationID = &v
	}
	if !m.Name.IsNull() && !m.Name.IsUnknown() {
		v := m.Name.ValueString()
		p.Name = &v
	}
	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		v := m.Description.ValueString()
		p.Description = &v
	}
	return p
}

func (r *EHRDestinationResource) request(ctx context.Context, method, apiPath string, body []byte) (*ehrDestinationPayload, error) {
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

	var result ehrDestinationPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func applyEHRDestinationResult(m *EHRDestinationResourceModel, p *ehrDestinationPayload) {
	if p.ID != nil {
		m.ID = types.StringValue(*p.ID)
	}
	setOptionalString(&m.JsonMetadataID, p.JsonMetadataID)
	setOptionalString(&m.DestinationID, p.DestinationID)
	setOptionalString(&m.Name, p.Name)
	setOptionalString(&m.Description, p.Description)
	m.IsDisabled = types.BoolValue(p.IsDisabled)
}
