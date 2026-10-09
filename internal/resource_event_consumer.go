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

var _ resource.Resource = &EventConsumerResource{}
var _ resource.ResourceWithImportState = &EventConsumerResource{}

type EventConsumerResource struct {
	client *TcmClient
}

type EventConsumerResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	TenantID                types.String `tfsdk:"tenant_id"`
	Name                    types.String `tfsdk:"name"`
	Description             types.String `tfsdk:"description"`
	Endpoint                types.String `tfsdk:"endpoint"`
	AuthorizationType       types.String `tfsdk:"authorization_type"`
	AuthorizationParameters types.String `tfsdk:"authorization_parameters"`
	IsDisabled              types.Bool   `tfsdk:"is_disabled"`
	LastSyncDate            types.String `tfsdk:"last_sync_date"`
	LastSyncMessage         types.String `tfsdk:"last_sync_message"`
	CreatedBy               types.String `tfsdk:"created_by"`
	Created                 types.String `tfsdk:"created"`
	LastModifiedBy          types.String `tfsdk:"last_modified_by"`
	LastModified            types.String `tfsdk:"last_modified"`
	Version                 types.Int64  `tfsdk:"version"`
}

type eventConsumerPayload struct {
	ID                      *string          `json:"id,omitempty"`
	TenantID                *string          `json:"tenantId,omitempty"`
	Name                    string           `json:"name"`
	Description             *string          `json:"description,omitempty"`
	Endpoint                string           `json:"endpoint"`
	AuthorizationType       *string          `json:"authorizationType,omitempty"`
	AuthorizationParameters *json.RawMessage `json:"authorizationParameters,omitempty"`
	IsDisabled              bool             `json:"isDisabled"`
	LastSyncDate            *string          `json:"lastSyncDate,omitempty"`
	LastSyncMessage         *string          `json:"lastSyncMessage,omitempty"`
	CreatedBy               *string          `json:"createdBy,omitempty"`
	Created                 *string          `json:"created,omitempty"`
	LastModifiedBy          *string          `json:"lastModifiedBy,omitempty"`
	LastModified            *string          `json:"lastModified,omitempty"`
	Version                 *int64           `json:"version,omitempty"`
}

func NewEventConsumerResource() resource.Resource {
	return &EventConsumerResource{}
}

func (r *EventConsumerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_consumer"
}

func (r *EventConsumerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TCM EventConsumer.",
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
				Description: "Name of the event consumer.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description of the event consumer.",
			},
			"endpoint": schema.StringAttribute{
				Required:    true,
				Description: "Endpoint URL of the event consumer.",
			},
			"authorization_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Auth type: API_KEY, BASIC, or OAUTH_CLIENT_CREDENTIALS",
			},
			"authorization_parameters": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "JSON-encoded authorization parameters object",
			},
			"is_disabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the event consumer is disabled.",
			},
			"last_sync_date": schema.StringAttribute{
				Computed:    true,
				Description: "Last sync date (read-only).",
			},
			"last_sync_message": schema.StringAttribute{
				Computed:    true,
				Description: "Last sync message (read-only).",
			},
			"created_by": schema.StringAttribute{
				Computed:    true,
				Description: "User who created the event consumer (read-only).",
			},
			"created": schema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp (read-only).",
			},
			"last_modified_by": schema.StringAttribute{
				Computed:    true,
				Description: "User who last modified the event consumer (read-only).",
			},
			"last_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Last modification timestamp (read-only).",
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "Version of the event consumer (read-only).",
			},
		},
	}
}

func (r *EventConsumerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EventConsumerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EventConsumerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPost, "/v1/EventConsumer", body)
	if err != nil {
		resp.Diagnostics.AddError("create event consumer failed", err.Error())
		return
	}

	applyEventConsumerResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EventConsumerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EventConsumerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.request(ctx, http.MethodGet, "/v1/EventConsumer/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("read event consumer failed", err.Error())
		return
	}

	applyEventConsumerResult(&state, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *EventConsumerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EventConsumerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.toPayload(&plan))
	if err != nil {
		resp.Diagnostics.AddError("serialization error", err.Error())
		return
	}

	result, err := r.request(ctx, http.MethodPut, "/v1/EventConsumer/"+plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("update event consumer failed", err.Error())
		return
	}

	applyEventConsumerResult(&plan, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EventConsumerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EventConsumerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.request(ctx, http.MethodDelete, "/v1/EventConsumer/"+state.ID.ValueString(), nil); err != nil {
		resp.Diagnostics.AddError("delete event consumer failed", err.Error())
	}
}

// ImportState allows importing an existing event consumer by its TCM UUID.
// Usage: terraform import tcm_event_consumer.<name> <uuid>
func (r *EventConsumerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
}

func (r *EventConsumerResource) toPayload(m *EventConsumerResourceModel) *eventConsumerPayload {
	p := &eventConsumerPayload{
		Name:       m.Name.ValueString(),
		Endpoint:   m.Endpoint.ValueString(),
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
	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		v := m.Description.ValueString()
		p.Description = &v
	}
	if !m.AuthorizationType.IsNull() && !m.AuthorizationType.IsUnknown() {
		v := m.AuthorizationType.ValueString()
		p.AuthorizationType = &v
	}
	if !m.AuthorizationParameters.IsNull() && !m.AuthorizationParameters.IsUnknown() {
		raw := json.RawMessage(m.AuthorizationParameters.ValueString())
		p.AuthorizationParameters = &raw
	}
	return p
}

func (r *EventConsumerResource) request(ctx context.Context, method, apiPath string, body []byte) (*eventConsumerPayload, error) {
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

	var result eventConsumerPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

func applyEventConsumerResult(m *EventConsumerResourceModel, p *eventConsumerPayload) {
	if p.ID != nil {
		m.ID = types.StringValue(*p.ID)
	}
	setOptionalString(&m.TenantID, p.TenantID)
	m.Name = types.StringValue(p.Name)
	setOptionalString(&m.Description, p.Description)
	m.Endpoint = types.StringValue(p.Endpoint)
	setOptionalString(&m.AuthorizationType, p.AuthorizationType)
	if p.AuthorizationParameters != nil {
		m.AuthorizationParameters = types.StringValue(string(*p.AuthorizationParameters))
	} else {
		m.AuthorizationParameters = types.StringNull()
	}
	m.IsDisabled = types.BoolValue(p.IsDisabled)
	setOptionalString(&m.LastSyncDate, p.LastSyncDate)
	setOptionalString(&m.LastSyncMessage, p.LastSyncMessage)
	setOptionalString(&m.CreatedBy, p.CreatedBy)
	setOptionalString(&m.Created, p.Created)
	setOptionalString(&m.LastModifiedBy, p.LastModifiedBy)
	setOptionalString(&m.LastModified, p.LastModified)
	if p.Version != nil {
		m.Version = types.Int64Value(*p.Version)
	} else {
		m.Version = types.Int64Null()
	}
}
