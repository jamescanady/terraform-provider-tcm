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

var _ resource.Resource = &EHRSchemaResource{}
var _ resource.ResourceWithImportState = &EHRSchemaResource{}

type EHRSchemaResource struct {
	client *TcmClient
}

type EHRSchemaResourceModel struct {
	ID             types.String `tfsdk:"id"`
	TenantID       types.String `tfsdk:"tenant_id"`
	EndpointID     types.String `tfsdk:"endpoint_id"`
	RequiredFields types.List   `tfsdk:"required_fields"`
	IsDisabled     types.Bool   `tfsdk:"is_disabled"`
}

type ehrSchemaPayload struct {
	ID             *string  `json:"id,omitempty"`
	TenantID       *string  `json:"tenantId,omitempty"`
	EndpointID     *string  `json:"endpointId,omitempty"`
	RequiredFields []string `json:"requiredFields,omitempty"`
	IsDisabled     bool     `json:"isDisabled"`
}

func NewEHRSchemaResource() resource.Resource {
	return &EHRSchemaResource{}
}

func (r *EHRSchemaResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ehr_schema"
}

func (r *EHRSchemaResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TCM EHR Schema.",
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
			"required_fields": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "List of required field names.",
			},
			"is_disabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the EHR schema is disabled.",
			},
		},
	}
}

func (r *EHRSchemaResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EHRSchemaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EHRSchemaResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPost, "/v1/ehr/schema", body)
	if err != nil {
		resp.Diagnostics.AddError("create ehr schema failed", err.Error())
		return
	}

	applyEHRSchemaResult(ctx, &plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EHRSchemaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EHRSchemaResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.request(ctx, http.MethodGet, "/v1/ehr/schema/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("read ehr schema failed", err.Error())
		return
	}

	applyEHRSchemaResult(ctx, &state, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *EHRSchemaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EHRSchemaResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPut, "/v1/ehr/schema/"+plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("update ehr schema failed", err.Error())
		return
	}

	applyEHRSchemaResult(ctx, &plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EHRSchemaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EHRSchemaResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.request(ctx, http.MethodDelete, "/v1/ehr/schema/"+state.ID.ValueString(), nil); err != nil {
		resp.Diagnostics.AddError("delete ehr schema failed", err.Error())
	}
}

// ImportState allows importing an existing EHR schema by its TCM UUID.
// Usage: terraform import tcm_ehr_schema.<name> <uuid>
func (r *EHRSchemaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
}

func (r *EHRSchemaResource) toPayload(m *EHRSchemaResourceModel) *ehrSchemaPayload {
	p := &ehrSchemaPayload{
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
	if !m.RequiredFields.IsNull() && !m.RequiredFields.IsUnknown() {
		elems := m.RequiredFields.Elements()
		fields := make([]string, len(elems))
		for i, e := range elems {
			fields[i] = e.(types.String).ValueString()
		}
		p.RequiredFields = fields
	}
	return p
}

func (r *EHRSchemaResource) request(ctx context.Context, method, apiPath string, body []byte) (*ehrSchemaPayload, error) {
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

	var result ehrSchemaPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func applyEHRSchemaResult(ctx context.Context, m *EHRSchemaResourceModel, p *ehrSchemaPayload) {
	if p.ID != nil {
		m.ID = types.StringValue(*p.ID)
	}
	setOptionalString(&m.TenantID, p.TenantID)
	setOptionalString(&m.EndpointID, p.EndpointID)
	m.IsDisabled = types.BoolValue(p.IsDisabled)
	if p.RequiredFields != nil {
		listVal, _ := types.ListValueFrom(ctx, types.StringType, p.RequiredFields)
		m.RequiredFields = listVal
	} else {
		m.RequiredFields = types.ListNull(types.StringType)
	}
}
