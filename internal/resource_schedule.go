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

var _ resource.Resource = &ScheduleResource{}
var _ resource.ResourceWithImportState = &ScheduleResource{}

type ScheduleResource struct {
	client *TcmClient
}

type ScheduleResourceModel struct {
	ID                         types.String `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
	Description                types.String `tfsdk:"description"`
	TenantProductEnvironmentID types.String `tfsdk:"tenant_product_environment_id"`
	ScheduleCategoryID         types.String `tfsdk:"schedule_category_id"`
	Enabled                    types.Bool   `tfsdk:"enabled"`
	ScheduleDetails            types.String `tfsdk:"schedule_details"`
}

type schedulePayload struct {
	ID                         *string          `json:"id,omitempty"`
	Name                       *string          `json:"name,omitempty"`
	Description                *string          `json:"description,omitempty"`
	TenantProductEnvironmentID string           `json:"tenantProductEnvironmentId"`
	ScheduleCategoryID         string           `json:"scheduleCategoryId"`
	Enabled                    bool             `json:"enabled"`
	ScheduleDetails            *json.RawMessage `json:"scheduleDetails,omitempty"`
}

func NewScheduleResource() resource.Resource {
	return &ScheduleResource{}
}

func (r *ScheduleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schedule"
}

func (r *ScheduleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TCM Schedule.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID assigned by TCM on creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Name of the schedule.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description of the schedule.",
			},
			"tenant_product_environment_id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the tenant product environment.",
			},
			"schedule_category_id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the schedule category.",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the schedule is enabled.",
			},
			"schedule_details": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "JSON-encoded schedule details object.",
			},
		},
	}
}

func (r *ScheduleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ScheduleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPost, "/v1/Schedule", body)
	if err != nil {
		resp.Diagnostics.AddError("create schedule failed", err.Error())
		return
	}

	applyScheduleResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ScheduleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.request(ctx, http.MethodGet, "/v1/Schedule/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("read schedule failed", err.Error())
		return
	}

	applyScheduleResult(&state, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ScheduleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPut, "/v1/Schedule/"+plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("update schedule failed", err.Error())
		return
	}

	applyScheduleResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ScheduleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.request(ctx, http.MethodDelete, "/v1/Schedule/"+state.ID.ValueString(), nil); err != nil {
		resp.Diagnostics.AddError("delete schedule failed", err.Error())
	}
}

// ImportState allows importing an existing schedule by its TCM UUID.
// Usage: terraform import tcm_schedule.<name> <uuid>
func (r *ScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
}

func (r *ScheduleResource) toPayload(m *ScheduleResourceModel) *schedulePayload {
	p := &schedulePayload{
		TenantProductEnvironmentID: m.TenantProductEnvironmentID.ValueString(),
		ScheduleCategoryID:         m.ScheduleCategoryID.ValueString(),
		Enabled:                    m.Enabled.ValueBool(),
	}
	if !m.ID.IsNull() && !m.ID.IsUnknown() {
		id := m.ID.ValueString()
		p.ID = &id
	}
	if !m.Name.IsNull() && !m.Name.IsUnknown() {
		v := m.Name.ValueString()
		p.Name = &v
	}
	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		v := m.Description.ValueString()
		p.Description = &v
	}
	if !m.ScheduleDetails.IsNull() && !m.ScheduleDetails.IsUnknown() {
		raw := json.RawMessage(m.ScheduleDetails.ValueString())
		p.ScheduleDetails = &raw
	}
	return p
}

func (r *ScheduleResource) request(ctx context.Context, method, apiPath string, body []byte) (*schedulePayload, error) {
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

	var result schedulePayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func applyScheduleResult(m *ScheduleResourceModel, p *schedulePayload) {
	if p.ID != nil {
		m.ID = types.StringValue(*p.ID)
	}
	setOptionalString(&m.Name, p.Name)
	setOptionalString(&m.Description, p.Description)
	m.TenantProductEnvironmentID = types.StringValue(p.TenantProductEnvironmentID)
	m.ScheduleCategoryID = types.StringValue(p.ScheduleCategoryID)
	m.Enabled = types.BoolValue(p.Enabled)
	if p.ScheduleDetails != nil {
		m.ScheduleDetails = types.StringValue(string(*p.ScheduleDetails))
	} else {
		m.ScheduleDetails = types.StringNull()
	}
}
