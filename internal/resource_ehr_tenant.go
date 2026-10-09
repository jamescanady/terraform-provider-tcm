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

var _ resource.Resource = &EHRTenantResource{}
var _ resource.ResourceWithImportState = &EHRTenantResource{}

type EHRTenantResource struct {
	client *TcmClient
}

type EHRTenantResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	TenantID           types.String `tfsdk:"tenant_id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	RedoxEHRIdentifier types.String `tfsdk:"redox_ehr_identifier"`
	SymplrURLSlug      types.String `tfsdk:"symplr_url_slug"`
	RedoxEnvironment   types.String `tfsdk:"redox_environment"`
	IsDisabled         types.Bool   `tfsdk:"is_disabled"`
}

type ehrTenantPayload struct {
	ID                 *string `json:"id,omitempty"`
	TenantID           *string `json:"tenantId,omitempty"`
	Name               string  `json:"name"`
	Description        *string `json:"description,omitempty"`
	RedoxEHRIdentifier string  `json:"redoxEHRIdentifier"`
	SymplrURLSlug      string  `json:"symplrUrlSlug"`
	RedoxEnvironment   string  `json:"redoxEnvironment"`
	IsDisabled         bool    `json:"isDisabled"`
}

func NewEHRTenantResource() resource.Resource {
	return &EHRTenantResource{}
}

func (r *EHRTenantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ehr_tenant"
}

func (r *EHRTenantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TCM EHR Tenant.",
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
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the EHR tenant.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description of the EHR tenant.",
			},
			"redox_ehr_identifier": schema.StringAttribute{
				Required:    true,
				Description: "Redox EHR identifier.",
			},
			"symplr_url_slug": schema.StringAttribute{
				Required:    true,
				Description: "Symplr URL slug.",
			},
			"redox_environment": schema.StringAttribute{
				Required:    true,
				Description: "Redox environment.",
			},
			"is_disabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the EHR tenant is disabled.",
			},
		},
	}
}

func (r *EHRTenantResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EHRTenantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EHRTenantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPost, "/v1/ehr/tenant", body)
	if err != nil {
		resp.Diagnostics.AddError("create ehr tenant failed", err.Error())
		return
	}

	applyEHRTenantResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EHRTenantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EHRTenantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.request(ctx, http.MethodGet, "/v1/ehr/tenant/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("read ehr tenant failed", err.Error())
		return
	}

	applyEHRTenantResult(&state, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *EHRTenantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EHRTenantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPut, "/v1/ehr/tenant/"+plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("update ehr tenant failed", err.Error())
		return
	}

	applyEHRTenantResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EHRTenantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EHRTenantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.request(ctx, http.MethodDelete, "/v1/ehr/tenant/"+state.ID.ValueString(), nil); err != nil {
		resp.Diagnostics.AddError("delete ehr tenant failed", err.Error())
	}
}

// ImportState allows importing an existing EHR tenant by its TCM UUID.
// Usage: terraform import tcm_ehr_tenant.<name> <uuid>
func (r *EHRTenantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
}

func (r *EHRTenantResource) toPayload(m *EHRTenantResourceModel) *ehrTenantPayload {
	p := &ehrTenantPayload{
		Name:               m.Name.ValueString(),
		RedoxEHRIdentifier: m.RedoxEHRIdentifier.ValueString(),
		SymplrURLSlug:      m.SymplrURLSlug.ValueString(),
		RedoxEnvironment:   m.RedoxEnvironment.ValueString(),
		IsDisabled:         m.IsDisabled.ValueBool(),
	}
	if !m.ID.IsNull() && !m.ID.IsUnknown() {
		v := m.ID.ValueString()
		p.ID = &v
	}
	if !m.TenantID.IsNull() && !m.TenantID.IsUnknown() {
		v := m.TenantID.ValueString()
		p.TenantID = &v
	}
	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		v := m.Description.ValueString()
		p.Description = &v
	}
	return p
}

func (r *EHRTenantResource) request(ctx context.Context, method, apiPath string, body []byte) (*ehrTenantPayload, error) {
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

	var result ehrTenantPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func applyEHRTenantResult(m *EHRTenantResourceModel, p *ehrTenantPayload) {
	if p.ID != nil {
		m.ID = types.StringValue(*p.ID)
	}
	setOptionalString(&m.TenantID, p.TenantID)
	m.Name = types.StringValue(p.Name)
	setOptionalString(&m.Description, p.Description)
	m.RedoxEHRIdentifier = types.StringValue(p.RedoxEHRIdentifier)
	m.SymplrURLSlug = types.StringValue(p.SymplrURLSlug)
	m.RedoxEnvironment = types.StringValue(p.RedoxEnvironment)
	m.IsDisabled = types.BoolValue(p.IsDisabled)
}
