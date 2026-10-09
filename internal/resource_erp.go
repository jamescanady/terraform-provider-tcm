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

var _ resource.Resource = &ERPResource{}
var _ resource.ResourceWithImportState = &ERPResource{}

type ERPResource struct {
	client *TcmClient
}

type ERPResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	Version          types.String `tfsdk:"version"`
	IsDeleted        types.Bool   `tfsdk:"is_deleted"`
	CreatedBy        types.String `tfsdk:"created_by"`
	CreatedDate      types.String `tfsdk:"created_date"`
	LastModifiedBy   types.String `tfsdk:"last_modified_by"`
	LastModifiedDate types.String `tfsdk:"last_modified_date"`
}

type erpPayload struct {
	ID               *string `json:"id,omitempty"`
	Name             *string `json:"name,omitempty"`
	Description      *string `json:"description,omitempty"`
	Version          *string `json:"version,omitempty"`
	IsDeleted        *bool   `json:"isDeleted,omitempty"`
	CreatedBy        *string `json:"createdBy,omitempty"`
	CreatedDate      *string `json:"createdDate,omitempty"`
	LastModifiedBy   *string `json:"lastModifiedBy,omitempty"`
	LastModifiedDate *string `json:"lastModifiedDate,omitempty"`
}

func NewERPResource() resource.Resource {
	return &ERPResource{}
}

func (r *ERPResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_erp"
}

func (r *ERPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TCM ERP.",
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
				Description: "Name of the ERP.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description of the ERP.",
			},
			"version": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Version of the ERP.",
			},
			"is_deleted": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the ERP is deleted (read-only).",
			},
			"created_by": schema.StringAttribute{
				Computed:    true,
				Description: "User who created the ERP.",
			},
			"created_date": schema.StringAttribute{
				Computed:    true,
				Description: "Date the ERP was created.",
			},
			"last_modified_by": schema.StringAttribute{
				Computed:    true,
				Description: "User who last modified the ERP.",
			},
			"last_modified_date": schema.StringAttribute{
				Computed:    true,
				Description: "Date the ERP was last modified.",
			},
		},
	}
}

func (r *ERPResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ERPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ERPResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPost, "/v1/ERP", body)
	if err != nil {
		resp.Diagnostics.AddError("create ERP failed", err.Error())
		return
	}

	applyERPResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ERPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ERPResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.request(ctx, http.MethodGet, "/v1/ERP/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("read ERP failed", err.Error())
		return
	}

	applyERPResult(&state, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ERPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ERPResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPut, "/v1/ERP/"+plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("update ERP failed", err.Error())
		return
	}

	applyERPResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ERPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ERPResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.request(ctx, http.MethodDelete, "/v1/ERP/"+state.ID.ValueString(), nil); err != nil {
		resp.Diagnostics.AddError("delete ERP failed", err.Error())
	}
}

// ImportState allows importing an existing ERP by its TCM UUID.
// Usage: terraform import tcm_erp.<name> <uuid>
func (r *ERPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
}

func (r *ERPResource) toPayload(m *ERPResourceModel) *erpPayload {
	p := &erpPayload{}
	if !m.Name.IsNull() && !m.Name.IsUnknown() {
		v := m.Name.ValueString()
		p.Name = &v
	}
	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		v := m.Description.ValueString()
		p.Description = &v
	}
	if !m.Version.IsNull() && !m.Version.IsUnknown() {
		v := m.Version.ValueString()
		p.Version = &v
	}
	return p
}

func (r *ERPResource) request(ctx context.Context, method, apiPath string, body []byte) (*erpPayload, error) {
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

	var result erpPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func applyERPResult(m *ERPResourceModel, p *erpPayload) {
	if p.ID != nil {
		m.ID = types.StringValue(*p.ID)
	}
	setOptionalString(&m.Name, p.Name)
	setOptionalString(&m.Description, p.Description)
	setOptionalString(&m.Version, p.Version)
	if p.IsDeleted != nil {
		m.IsDeleted = types.BoolValue(*p.IsDeleted)
	} else {
		m.IsDeleted = types.BoolValue(false)
	}
	setOptionalString(&m.CreatedBy, p.CreatedBy)
	setOptionalString(&m.CreatedDate, p.CreatedDate)
	setOptionalString(&m.LastModifiedBy, p.LastModifiedBy)
	setOptionalString(&m.LastModifiedDate, p.LastModifiedDate)
}
