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

var _ resource.Resource = &EHRJsonMetadataResource{}
var _ resource.ResourceWithImportState = &EHRJsonMetadataResource{}

type EHRJsonMetadataResource struct {
	client *TcmClient
}

type EHRJsonMetadataResourceModel struct {
	ID           types.String `tfsdk:"id"`
	TenantID     types.String `tfsdk:"tenant_id"`
	EndpointID   types.String `tfsdk:"endpoint_id"`
	FacilityCode types.String `tfsdk:"facility_code"`
	IsDisabled   types.Bool   `tfsdk:"is_disabled"`
}

type ehrJsonMetadataPayload struct {
	ID           *string `json:"id,omitempty"`
	TenantID     *string `json:"tenantId,omitempty"`
	EndpointID   *string `json:"endpointId,omitempty"`
	FacilityCode *string `json:"facilityCode,omitempty"`
	IsDisabled   bool    `json:"isDisabled"`
}

func NewEHRJsonMetadataResource() resource.Resource {
	return &EHRJsonMetadataResource{}
}

func (r *EHRJsonMetadataResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ehr_json_metadata"
}

func (r *EHRJsonMetadataResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TCM EHR JSON Metadata.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID assigned by TCM on creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tenant_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of the tenant.",
			},
			"endpoint_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of the endpoint.",
			},
			"facility_code": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Facility code.",
			},
			"is_disabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the EHR JSON metadata is disabled.",
			},
		},
	}
}

func (r *EHRJsonMetadataResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EHRJsonMetadataResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EHRJsonMetadataResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPost, "/v1/ehr/jsonMetadata", body)
	if err != nil {
		resp.Diagnostics.AddError("create ehr json metadata failed", err.Error())
		return
	}

	applyEHRJsonMetadataResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EHRJsonMetadataResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EHRJsonMetadataResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.request(ctx, http.MethodGet, "/v1/ehr/jsonMetadata/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("read ehr json metadata failed", err.Error())
		return
	}

	applyEHRJsonMetadataResult(&state, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *EHRJsonMetadataResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EHRJsonMetadataResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPut, "/v1/ehr/jsonMetadata/"+plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("update ehr json metadata failed", err.Error())
		return
	}

	applyEHRJsonMetadataResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EHRJsonMetadataResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EHRJsonMetadataResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.request(ctx, http.MethodDelete, "/v1/ehr/jsonMetadata/"+state.ID.ValueString(), nil); err != nil {
		resp.Diagnostics.AddError("delete ehr json metadata failed", err.Error())
	}
}

// ImportState allows importing an existing EHR JSON metadata by its TCM UUID.
// Usage: terraform import tcm_ehr_json_metadata.<name> <uuid>
func (r *EHRJsonMetadataResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
}

func (r *EHRJsonMetadataResource) toPayload(m *EHRJsonMetadataResourceModel) *ehrJsonMetadataPayload {
	p := &ehrJsonMetadataPayload{
		IsDisabled: m.IsDisabled.ValueBool(),
	}
	if !m.ID.IsNull() && !m.ID.IsUnknown() {
		id := m.ID.ValueString()
		p.ID = &id
	}
	if !m.TenantID.IsNull() && !m.TenantID.IsUnknown() {
		v := m.TenantID.ValueString()
		p.TenantID = &v
	}
	if !m.EndpointID.IsNull() && !m.EndpointID.IsUnknown() {
		v := m.EndpointID.ValueString()
		p.EndpointID = &v
	}
	if !m.FacilityCode.IsNull() && !m.FacilityCode.IsUnknown() {
		v := m.FacilityCode.ValueString()
		p.FacilityCode = &v
	}
	return p
}

func (r *EHRJsonMetadataResource) request(ctx context.Context, method, apiPath string, body []byte) (*ehrJsonMetadataPayload, error) {
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

	var result ehrJsonMetadataPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func applyEHRJsonMetadataResult(m *EHRJsonMetadataResourceModel, p *ehrJsonMetadataPayload) {
	if p.ID != nil {
		m.ID = types.StringValue(*p.ID)
	}
	setOptionalString(&m.TenantID, p.TenantID)
	setOptionalString(&m.EndpointID, p.EndpointID)
	setOptionalString(&m.FacilityCode, p.FacilityCode)
	m.IsDisabled = types.BoolValue(p.IsDisabled)
}
