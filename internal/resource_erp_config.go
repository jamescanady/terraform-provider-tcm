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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ERPConfigResource{}
var _ resource.ResourceWithImportState = &ERPConfigResource{}

type ERPConfigResource struct {
	client *TcmClient
}

type ERPConfigResourceModel struct {
	ID                         types.String `tfsdk:"id"`
	ERPID                      types.String `tfsdk:"erp_id"`
	TenantProductEnvironmentID types.String `tfsdk:"tenant_product_environment_id"`
	URL                        types.String `tfsdk:"url"`
	ClientID                   types.String `tfsdk:"client_id"`
	ClientSecret               types.String `tfsdk:"client_secret"`
	LookupID                   types.String `tfsdk:"lookup_id"`
	HostName                   types.String `tfsdk:"host_name"`
	TenantSlug                 types.String `tfsdk:"tenant_slug"`
	APIVersion                 types.String `tfsdk:"api_version"`
	OntologyTypeIDs            types.List   `tfsdk:"ontology_type_ids"`
	GlobalTenantCode           types.String `tfsdk:"global_tenant_code"`
	ERPName                    types.String `tfsdk:"erp_name"`
	IsDeleted                  types.Bool   `tfsdk:"is_deleted"`
}

type erpConfigPayload struct {
	ID                         *string  `json:"id,omitempty"`
	ERPID                      *string  `json:"erpId,omitempty"`
	TenantProductEnvironmentID *string  `json:"tenantProductEnvironmentId,omitempty"`
	URL                        *string  `json:"url,omitempty"`
	ClientID                   *string  `json:"clientId,omitempty"`
	ClientSecret               *string  `json:"clientSecret,omitempty"`
	LookupID                   *string  `json:"lookupId,omitempty"`
	HostName                   *string  `json:"hostName,omitempty"`
	TenantSlug                 *string  `json:"tenantSlug,omitempty"`
	APIVersion                 *string  `json:"apiVersion,omitempty"`
	OntologyTypeIDs            []string `json:"ontologyTypeIds,omitempty"`
	GlobalTenantCode           *string  `json:"globalTenantCode,omitempty"`
	ERPName                    *string  `json:"erpName,omitempty"`
	IsDeleted                  *bool    `json:"isDeleted,omitempty"`
}

func NewERPConfigResource() resource.Resource {
	return &ERPConfigResource{}
}

func (r *ERPConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_erp_config"
}

func (r *ERPConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TCM ERP Config.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID assigned by TCM on creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"erp_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of the ERP.",
			},
			"tenant_product_environment_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of the tenant product environment.",
			},
			"url": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "URL of the ERP.",
			},
			"client_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Client ID for authentication.",
			},
			"client_secret": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				Description: "Client secret for authentication.",
			},
			"lookup_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Lookup ID.",
			},
			"host_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Host name.",
			},
			"tenant_slug": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Tenant slug.",
			},
			"api_version": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "API version.",
			},
			"ontology_type_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "List of ontology type IDs.",
			},
			"global_tenant_code": schema.StringAttribute{
				Computed:    true,
				Description: "Global tenant code (read-only).",
			},
			"erp_name": schema.StringAttribute{
				Computed:    true,
				Description: "ERP name (read-only).",
			},
			"is_deleted": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the ERP config is deleted (read-only).",
			},
		},
	}
}

func (r *ERPConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ERPConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ERPConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPost, "/v1/ERPConfig", body)
	if err != nil {
		resp.Diagnostics.AddError("create ERP config failed", err.Error())
		return
	}

	applyERPConfigResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ERPConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ERPConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.request(ctx, http.MethodGet, "/v1/ERPConfig/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("read ERP config failed", err.Error())
		return
	}

	applyERPConfigResult(&state, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ERPConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ERPConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPut, "/v1/ERPConfig/"+plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("update ERP config failed", err.Error())
		return
	}

	applyERPConfigResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ERPConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ERPConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.request(ctx, http.MethodDelete, "/v1/ERPConfig/"+state.ID.ValueString(), nil); err != nil {
		resp.Diagnostics.AddError("delete ERP config failed", err.Error())
	}
}

// ImportState allows importing an existing ERP config by its TCM UUID.
// Usage: terraform import tcm_erp_config.<name> <uuid>
func (r *ERPConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
}

func (r *ERPConfigResource) toPayload(m *ERPConfigResourceModel) *erpConfigPayload {
	p := &erpConfigPayload{}
	if !m.ERPID.IsNull() && !m.ERPID.IsUnknown() {
		v := m.ERPID.ValueString()
		p.ERPID = &v
	}
	if !m.TenantProductEnvironmentID.IsNull() && !m.TenantProductEnvironmentID.IsUnknown() {
		v := m.TenantProductEnvironmentID.ValueString()
		p.TenantProductEnvironmentID = &v
	}
	if !m.URL.IsNull() && !m.URL.IsUnknown() {
		v := m.URL.ValueString()
		p.URL = &v
	}
	if !m.ClientID.IsNull() && !m.ClientID.IsUnknown() {
		v := m.ClientID.ValueString()
		p.ClientID = &v
	}
	if !m.ClientSecret.IsNull() && !m.ClientSecret.IsUnknown() {
		v := m.ClientSecret.ValueString()
		p.ClientSecret = &v
	}
	if !m.LookupID.IsNull() && !m.LookupID.IsUnknown() {
		v := m.LookupID.ValueString()
		p.LookupID = &v
	}
	if !m.HostName.IsNull() && !m.HostName.IsUnknown() {
		v := m.HostName.ValueString()
		p.HostName = &v
	}
	if !m.TenantSlug.IsNull() && !m.TenantSlug.IsUnknown() {
		v := m.TenantSlug.ValueString()
		p.TenantSlug = &v
	}
	if !m.APIVersion.IsNull() && !m.APIVersion.IsUnknown() {
		v := m.APIVersion.ValueString()
		p.APIVersion = &v
	}
	if !m.OntologyTypeIDs.IsNull() && !m.OntologyTypeIDs.IsUnknown() {
		elems := m.OntologyTypeIDs.Elements()
		ids := make([]string, len(elems))
		for i, e := range elems {
			ids[i] = e.(types.String).ValueString()
		}
		p.OntologyTypeIDs = ids
	}
	return p
}

func (r *ERPConfigResource) request(ctx context.Context, method, apiPath string, body []byte) (*erpConfigPayload, error) {
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

	var result erpConfigPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func applyERPConfigResult(m *ERPConfigResourceModel, p *erpConfigPayload) {
	if p.ID != nil {
		m.ID = types.StringValue(*p.ID)
	}
	setOptionalString(&m.ERPID, p.ERPID)
	setOptionalString(&m.TenantProductEnvironmentID, p.TenantProductEnvironmentID)
	setOptionalString(&m.URL, p.URL)
	setOptionalString(&m.ClientID, p.ClientID)
	setOptionalString(&m.ClientSecret, p.ClientSecret)
	setOptionalString(&m.LookupID, p.LookupID)
	setOptionalString(&m.HostName, p.HostName)
	setOptionalString(&m.TenantSlug, p.TenantSlug)
	setOptionalString(&m.APIVersion, p.APIVersion)
	// OntologyTypeIDs is not returned in the response; preserve the state value.
	setOptionalString(&m.GlobalTenantCode, p.GlobalTenantCode)
	setOptionalString(&m.ERPName, p.ERPName)
	if p.IsDeleted != nil {
		m.IsDeleted = types.BoolValue(*p.IsDeleted)
	} else {
		m.IsDeleted = types.BoolValue(false)
	}
}
