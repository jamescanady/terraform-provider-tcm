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

var _ resource.Resource = &EventTypeConsumerResource{}
var _ resource.ResourceWithImportState = &EventTypeConsumerResource{}

type EventTypeConsumerResource struct {
	client *TcmClient
}

type EventTypeConsumerResourceModel struct {
	ID                         types.String `tfsdk:"id"`
	EventTypeID                types.String `tfsdk:"event_type_id"`
	EventConsumerID            types.String `tfsdk:"event_consumer_id"`
	TenantProductEnvironmentID types.String `tfsdk:"tenant_product_environment_id"`
	IsDisabled                 types.Bool   `tfsdk:"is_disabled"`
	ConsumerName               types.String `tfsdk:"consumer_name"`
	ConsumerDescription        types.String `tfsdk:"consumer_description"`
	ConsumerEndpoint           types.String `tfsdk:"consumer_endpoint"`
	ConsumerAuthorizationType  types.String `tfsdk:"consumer_authorization_type"`
	EventTypeName              types.String `tfsdk:"event_type_name"`
	TenantID                   types.String `tfsdk:"tenant_id"`
	TenantName                 types.String `tfsdk:"tenant_name"`
	ProductID                  types.String `tfsdk:"product_id"`
	ProductName                types.String `tfsdk:"product_name"`
	ProductString              types.String `tfsdk:"product_string"`
	EnvironmentName            types.String `tfsdk:"environment_name"`
	CreatedBy                  types.String `tfsdk:"created_by"`
	Created                    types.String `tfsdk:"created"`
	LastModifiedBy             types.String `tfsdk:"last_modified_by"`
	LastModified               types.String `tfsdk:"last_modified"`
}

type eventTypeConsumerPayload struct {
	ID                         *string `json:"id,omitempty"`
	EventTypeID                *string `json:"eventTypeId,omitempty"`
	EventConsumerID            *string `json:"eventConsumerId,omitempty"` // write field
	ConsumerID                 *string `json:"consumerId,omitempty"`      // read field
	TenantProductEnvironmentID *string `json:"tenantProductEnvironmentId,omitempty"`
	IsDisabled                 bool    `json:"isDisabled"`
	// Response-only computed fields:
	ConsumerName              *string `json:"consumerName,omitempty"`
	ConsumerDescription       *string `json:"consumerDescription,omitempty"`
	ConsumerEndpoint          *string `json:"consumerEndpoint,omitempty"`
	ConsumerAuthorizationType *string `json:"consumerAuthorizationType,omitempty"`
	EventTypeName             *string `json:"eventTypeName,omitempty"`
	TenantID                  *string `json:"tenantId,omitempty"`
	TenantName                *string `json:"tenantName,omitempty"`
	ProductID                 *string `json:"productId,omitempty"`
	ProductName               *string `json:"productName,omitempty"`
	ProductString             *string `json:"productString,omitempty"`
	EnvironmentName           *string `json:"environmentName,omitempty"`
	CreatedBy                 *string `json:"createdBy,omitempty"`
	Created                   *string `json:"created,omitempty"`
	LastModifiedBy            *string `json:"lastModifiedBy,omitempty"`
	LastModified              *string `json:"lastModified,omitempty"`
}

func NewEventTypeConsumerResource() resource.Resource {
	return &EventTypeConsumerResource{}
}

func (r *EventTypeConsumerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_type_consumer"
}

func (r *EventTypeConsumerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TCM EventTypeConsumer.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID assigned by TCM on creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"event_type_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of the event type.",
			},
			"event_consumer_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of the event consumer.",
			},
			"tenant_product_environment_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of the tenant product environment.",
			},
			"is_disabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the event type consumer is disabled.",
			},
			"consumer_name": schema.StringAttribute{
				Computed:    true,
				Description: "Consumer name (read-only).",
			},
			"consumer_description": schema.StringAttribute{
				Computed:    true,
				Description: "Consumer description (read-only).",
			},
			"consumer_endpoint": schema.StringAttribute{
				Computed:    true,
				Description: "Consumer endpoint (read-only).",
			},
			"consumer_authorization_type": schema.StringAttribute{
				Computed:    true,
				Description: "Consumer authorization type (read-only).",
			},
			"event_type_name": schema.StringAttribute{
				Computed:    true,
				Description: "Event type name (read-only).",
			},
			"tenant_id": schema.StringAttribute{
				Computed:    true,
				Description: "Tenant UUID (read-only).",
			},
			"tenant_name": schema.StringAttribute{
				Computed:    true,
				Description: "Tenant name (read-only).",
			},
			"product_id": schema.StringAttribute{
				Computed:    true,
				Description: "Product UUID (read-only).",
			},
			"product_name": schema.StringAttribute{
				Computed:    true,
				Description: "Product name (read-only).",
			},
			"product_string": schema.StringAttribute{
				Computed:    true,
				Description: "Product string (read-only).",
			},
			"environment_name": schema.StringAttribute{
				Computed:    true,
				Description: "Environment name (read-only).",
			},
			"created_by": schema.StringAttribute{
				Computed:    true,
				Description: "User who created the record (read-only).",
			},
			"created": schema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp (read-only).",
			},
			"last_modified_by": schema.StringAttribute{
				Computed:    true,
				Description: "User who last modified the record (read-only).",
			},
			"last_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Last modification timestamp (read-only).",
			},
		},
	}
}

func (r *EventTypeConsumerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EventTypeConsumerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EventTypeConsumerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPost, "/v1/EventTypeConsumer", body)
	if err != nil {
		resp.Diagnostics.AddError("create event type consumer failed", err.Error())
		return
	}

	applyEventTypeConsumerResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EventTypeConsumerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EventTypeConsumerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.request(ctx, http.MethodGet, "/v1/EventTypeConsumer/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("read event type consumer failed", err.Error())
		return
	}

	applyEventTypeConsumerResult(&state, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *EventTypeConsumerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EventTypeConsumerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPut, "/v1/EventTypeConsumer/"+plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("update event type consumer failed", err.Error())
		return
	}

	applyEventTypeConsumerResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EventTypeConsumerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EventTypeConsumerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.request(ctx, http.MethodDelete, "/v1/EventTypeConsumer/"+state.ID.ValueString(), nil); err != nil {
		resp.Diagnostics.AddError("delete event type consumer failed", err.Error())
	}
}

// ImportState allows importing an existing event type consumer by its TCM UUID.
// Usage: terraform import tcm_event_type_consumer.<name> <uuid>
func (r *EventTypeConsumerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
}

func (r *EventTypeConsumerResource) toPayload(m *EventTypeConsumerResourceModel) *eventTypeConsumerPayload {
	p := &eventTypeConsumerPayload{
		IsDisabled: m.IsDisabled.ValueBool(),
	}
	if !m.ID.IsNull() && !m.ID.IsUnknown() {
		id := m.ID.ValueString()
		p.ID = &id
	}
	if !m.EventTypeID.IsNull() && !m.EventTypeID.IsUnknown() {
		v := m.EventTypeID.ValueString()
		p.EventTypeID = &v
	}
	if !m.EventConsumerID.IsNull() && !m.EventConsumerID.IsUnknown() {
		v := m.EventConsumerID.ValueString()
		p.EventConsumerID = &v
	}
	if !m.TenantProductEnvironmentID.IsNull() && !m.TenantProductEnvironmentID.IsUnknown() {
		v := m.TenantProductEnvironmentID.ValueString()
		p.TenantProductEnvironmentID = &v
	}
	return p
}

func (r *EventTypeConsumerResource) request(ctx context.Context, method, apiPath string, body []byte) (*eventTypeConsumerPayload, error) {
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

	var result eventTypeConsumerPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func applyEventTypeConsumerResult(m *EventTypeConsumerResourceModel, p *eventTypeConsumerPayload) {
	if p.ID != nil {
		m.ID = types.StringValue(*p.ID)
	}
	setOptionalString(&m.EventTypeID, p.EventTypeID)
	// Map consumerId (response) → event_consumer_id (model)
	if p.ConsumerID != nil {
		m.EventConsumerID = types.StringValue(*p.ConsumerID)
	} else {
		setOptionalString(&m.EventConsumerID, p.ConsumerID)
	}
	setOptionalString(&m.TenantProductEnvironmentID, p.TenantProductEnvironmentID)
	m.IsDisabled = types.BoolValue(p.IsDisabled)
	setOptionalString(&m.ConsumerName, p.ConsumerName)
	setOptionalString(&m.ConsumerDescription, p.ConsumerDescription)
	setOptionalString(&m.ConsumerEndpoint, p.ConsumerEndpoint)
	setOptionalString(&m.ConsumerAuthorizationType, p.ConsumerAuthorizationType)
	setOptionalString(&m.EventTypeName, p.EventTypeName)
	setOptionalString(&m.TenantID, p.TenantID)
	setOptionalString(&m.TenantName, p.TenantName)
	setOptionalString(&m.ProductID, p.ProductID)
	setOptionalString(&m.ProductName, p.ProductName)
	setOptionalString(&m.ProductString, p.ProductString)
	setOptionalString(&m.EnvironmentName, p.EnvironmentName)
	setOptionalString(&m.CreatedBy, p.CreatedBy)
	setOptionalString(&m.Created, p.Created)
	setOptionalString(&m.LastModifiedBy, p.LastModifiedBy)
	setOptionalString(&m.LastModified, p.LastModified)
}
