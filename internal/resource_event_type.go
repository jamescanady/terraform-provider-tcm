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

var _ resource.Resource = &EventTypeResource{}
var _ resource.ResourceWithImportState = &EventTypeResource{}

type EventTypeResource struct {
	client *TcmClient
}

type EventTypeResourceModel struct {
	ID             types.String `tfsdk:"id"`
	ProductID      types.String `tfsdk:"product_id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	IsDisabled     types.Bool   `tfsdk:"is_disabled"`
	CreatedBy      types.String `tfsdk:"created_by"`
	Created        types.String `tfsdk:"created"`
	LastModifiedBy types.String `tfsdk:"last_modified_by"`
	LastModified   types.String `tfsdk:"last_modified"`
}

type eventTypePayload struct {
	ID             *string `json:"id,omitempty"`
	ProductID      string  `json:"productId"`
	Name           string  `json:"name"`
	Description    *string `json:"description,omitempty"`
	IsDisabled     bool    `json:"isDisabled"`
	CreatedBy      *string `json:"createdBy,omitempty"`
	Created        *string `json:"created,omitempty"`
	LastModifiedBy *string `json:"lastModifiedBy,omitempty"`
	LastModified   *string `json:"lastModified,omitempty"`
}

func NewEventTypeResource() resource.Resource {
	return &EventTypeResource{}
}

func (r *EventTypeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_type"
}

func (r *EventTypeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TCM EventType.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID assigned by TCM on creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"product_id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the product.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the event type.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description of the event type.",
			},
			"is_disabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the event type is disabled.",
			},
			"created_by": schema.StringAttribute{
				Computed:    true,
				Description: "User who created the event type (read-only).",
			},
			"created": schema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp (read-only).",
			},
			"last_modified_by": schema.StringAttribute{
				Computed:    true,
				Description: "User who last modified the event type (read-only).",
			},
			"last_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Last modification timestamp (read-only).",
			},
		},
	}
}

func (r *EventTypeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EventTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EventTypeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPost, "/v1/EventType", body)
	if err != nil {
		resp.Diagnostics.AddError("create event type failed", err.Error())
		return
	}

	applyEventTypeResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EventTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EventTypeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.request(ctx, http.MethodGet, "/v1/EventType/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("read event type failed", err.Error())
		return
	}

	applyEventTypeResult(&state, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *EventTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EventTypeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPut, "/v1/EventType/"+plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("update event type failed", err.Error())
		return
	}

	applyEventTypeResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EventTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EventTypeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.request(ctx, http.MethodDelete, "/v1/EventType/"+state.ID.ValueString(), nil); err != nil {
		resp.Diagnostics.AddError("delete event type failed", err.Error())
	}
}

// ImportState allows importing an existing event type by its TCM UUID.
// Usage: terraform import tcm_event_type.<name> <uuid>
func (r *EventTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
}

func (r *EventTypeResource) toPayload(m *EventTypeResourceModel) *eventTypePayload {
	p := &eventTypePayload{
		ProductID:  m.ProductID.ValueString(),
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

func (r *EventTypeResource) request(ctx context.Context, method, apiPath string, body []byte) (*eventTypePayload, error) {
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

	var result eventTypePayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func applyEventTypeResult(m *EventTypeResourceModel, p *eventTypePayload) {
	if p.ID != nil {
		m.ID = types.StringValue(*p.ID)
	}
	m.ProductID = types.StringValue(p.ProductID)
	m.Name = types.StringValue(p.Name)
	setOptionalString(&m.Description, p.Description)
	m.IsDisabled = types.BoolValue(p.IsDisabled)
	setOptionalString(&m.CreatedBy, p.CreatedBy)
	setOptionalString(&m.Created, p.Created)
	setOptionalString(&m.LastModifiedBy, p.LastModifiedBy)
	setOptionalString(&m.LastModified, p.LastModified)
}
