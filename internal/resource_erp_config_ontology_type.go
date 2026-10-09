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

var _ resource.Resource = &ERPConfigOntologyTypeResource{}
var _ resource.ResourceWithImportState = &ERPConfigOntologyTypeResource{}

type ERPConfigOntologyTypeResource struct {
	client *TcmClient
}

type ERPConfigOntologyTypeResourceModel struct {
	ID             types.String `tfsdk:"id"`
	ERPConfigID    types.String `tfsdk:"erp_config_id"`
	OntologyTypeID types.String `tfsdk:"ontology_type_id"`
}

type erpConfigOntologyTypePayload struct {
	ID             *string `json:"id,omitempty"`
	ERPConfigID    *string `json:"erpConfigId,omitempty"`
	OntologyTypeID *string `json:"ontologyTypeId,omitempty"`
}

func NewERPConfigOntologyTypeResource() resource.Resource {
	return &ERPConfigOntologyTypeResource{}
}

func (r *ERPConfigOntologyTypeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_erp_config_ontology_type"
}

func (r *ERPConfigOntologyTypeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TCM ERPConfigOntologyType.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID assigned by TCM on creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"erp_config_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of the ERP config.",
			},
			"ontology_type_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of the ontology type.",
			},
		},
	}
}

func (r *ERPConfigOntologyTypeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ERPConfigOntologyTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ERPConfigOntologyTypeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPost, "/v1/ERPConfigOntologyType", body)
	if err != nil {
		resp.Diagnostics.AddError("create erp config ontology type failed", err.Error())
		return
	}

	applyERPConfigOntologyTypeResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ERPConfigOntologyTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ERPConfigOntologyTypeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.request(ctx, http.MethodGet, "/v1/ERPConfigOntologyType/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("read erp config ontology type failed", err.Error())
		return
	}

	applyERPConfigOntologyTypeResult(&state, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ERPConfigOntologyTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ERPConfigOntologyTypeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPut, "/v1/ERPConfigOntologyType/"+plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("update erp config ontology type failed", err.Error())
		return
	}

	applyERPConfigOntologyTypeResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ERPConfigOntologyTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ERPConfigOntologyTypeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.request(ctx, http.MethodDelete, "/v1/ERPConfigOntologyType/"+state.ID.ValueString(), nil); err != nil {
		resp.Diagnostics.AddError("delete erp config ontology type failed", err.Error())
	}
}

// ImportState allows importing an existing ERP config ontology type by its TCM UUID.
// Usage: terraform import tcm_erp_config_ontology_type.<name> <uuid>
func (r *ERPConfigOntologyTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
}

func (r *ERPConfigOntologyTypeResource) toPayload(m *ERPConfigOntologyTypeResourceModel) *erpConfigOntologyTypePayload {
	p := &erpConfigOntologyTypePayload{}
	if !m.ERPConfigID.IsNull() && !m.ERPConfigID.IsUnknown() {
		v := m.ERPConfigID.ValueString()
		p.ERPConfigID = &v
	}
	if !m.OntologyTypeID.IsNull() && !m.OntologyTypeID.IsUnknown() {
		v := m.OntologyTypeID.ValueString()
		p.OntologyTypeID = &v
	}
	return p
}

func (r *ERPConfigOntologyTypeResource) request(ctx context.Context, method, apiPath string, body []byte) (*erpConfigOntologyTypePayload, error) {
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

	var result erpConfigOntologyTypePayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func applyERPConfigOntologyTypeResult(m *ERPConfigOntologyTypeResourceModel, p *erpConfigOntologyTypePayload) {
	setOptionalString(&m.ID, p.ID)
	setOptionalString(&m.ERPConfigID, p.ERPConfigID)
	setOptionalString(&m.OntologyTypeID, p.OntologyTypeID)
}
