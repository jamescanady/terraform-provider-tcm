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

var _ resource.Resource = &NamespaceResource{}
var _ resource.ResourceWithImportState = &NamespaceResource{}

type NamespaceResource struct {
	client *TcmClient
}

type NamespaceResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	IsDisabled     types.Bool   `tfsdk:"is_disabled"`
	IsDefault      types.Bool   `tfsdk:"is_default"`
	CreatedDate    types.String `tfsdk:"created_date"`
	CreatedBy      types.String `tfsdk:"created_by"`
	LastModified   types.String `tfsdk:"last_modified"`
	LastModifiedBy types.String `tfsdk:"last_modified_by"`
}

type namespacePayload struct {
	ID          string  `json:"id,omitempty"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	IsDisabled  bool    `json:"isDisabled"`
}

type namespaceResponse struct {
	ID             string  `json:"id"`
	Name           *string `json:"name,omitempty"`
	Description    *string `json:"description,omitempty"`
	IsDefault      bool    `json:"isDefault"`
	IsDisabled     bool    `json:"isDisabled"`
	CreatedDate    string  `json:"createdDate"`
	CreatedBy      *string `json:"createdBy,omitempty"`
	LastModified   string  `json:"lastModified"`
	LastModifiedBy *string `json:"lastModifiedBy,omitempty"`
}

func NewNamespaceResource() resource.Resource {
	return &NamespaceResource{}
}

func (r *NamespaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_namespace"
}

func (r *NamespaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TCM Namespace.",
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
				Description: "Namespace name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Namespace description.",
			},
			"is_disabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the namespace is disabled.",
			},
			"is_default": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether this is the default namespace (read-only).",
			},
			"created_date": schema.StringAttribute{
				Computed:    true,
				Description: "UTC date and time the namespace was created (read-only).",
			},
			"created_by": schema.StringAttribute{
				Computed:    true,
				Description: "Identity that created the namespace (read-only).",
			},
			"last_modified": schema.StringAttribute{
				Computed:    true,
				Description: "UTC date and time the namespace was last modified (read-only).",
			},
			"last_modified_by": schema.StringAttribute{
				Computed:    true,
				Description: "Identity that last modified the namespace (read-only).",
			},
		},
	}
}

func (r *NamespaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NamespaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NamespaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPost, "/v1/Namespace", body)
	if err != nil {
		resp.Diagnostics.AddError("create namespace failed", err.Error())
		return
	}

	applyNamespaceResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *NamespaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NamespaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.request(ctx, http.MethodGet, "/v1/Namespace/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("read namespace failed", err.Error())
		return
	}

	applyNamespaceResult(&state, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *NamespaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NamespaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPut, "/v1/Namespace/"+plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("update namespace failed", err.Error())
		return
	}

	applyNamespaceResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *NamespaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NamespaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.request(ctx, http.MethodDelete, "/v1/Namespace/"+state.ID.ValueString(), nil); err != nil {
		resp.Diagnostics.AddError("delete namespace failed", err.Error())
	}
}

// ImportState allows importing an existing namespace by its TCM UUID.
// Usage: terraform import tcm_namespace.<name> <uuid>
func (r *NamespaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
}

func (r *NamespaceResource) toPayload(m *NamespaceResourceModel) *namespacePayload {
	p := &namespacePayload{
		Name:       m.Name.ValueString(),
		IsDisabled: m.IsDisabled.ValueBool(),
	}
	if !m.ID.IsNull() && !m.ID.IsUnknown() {
		p.ID = m.ID.ValueString()
	}
	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		v := m.Description.ValueString()
		p.Description = &v
	}
	return p
}

func (r *NamespaceResource) request(ctx context.Context, method, apiPath string, body []byte) (*namespaceResponse, error) {
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

	var result namespaceResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func applyNamespaceResult(m *NamespaceResourceModel, r *namespaceResponse) {
	m.ID = types.StringValue(r.ID)
	if r.Name != nil {
		m.Name = types.StringValue(*r.Name)
	}
	if r.Description != nil {
		m.Description = types.StringValue(*r.Description)
	} else {
		m.Description = types.StringNull()
	}
	m.IsDisabled = types.BoolValue(r.IsDisabled)
	m.IsDefault = types.BoolValue(r.IsDefault)
	m.CreatedDate = types.StringValue(r.CreatedDate)
	if r.CreatedBy != nil {
		m.CreatedBy = types.StringValue(*r.CreatedBy)
	} else {
		m.CreatedBy = types.StringNull()
	}
	m.LastModified = types.StringValue(r.LastModified)
	if r.LastModifiedBy != nil {
		m.LastModifiedBy = types.StringValue(*r.LastModifiedBy)
	} else {
		m.LastModifiedBy = types.StringNull()
	}
}
