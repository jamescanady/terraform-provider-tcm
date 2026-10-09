package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &EventConsumerDataSource{}

type EventConsumerDataSource struct {
	client *TcmClient
}

type EventConsumerDataSourceModel struct {
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

func NewEventConsumerDataSource() datasource.DataSource {
	return &EventConsumerDataSource{}
}

func (d *EventConsumerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_consumer"
}

func (d *EventConsumerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a TCM EventConsumer by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the event consumer.",
			},
			"tenant_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the tenant.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the event consumer.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the event consumer.",
			},
			"endpoint": schema.StringAttribute{
				Computed:    true,
				Description: "Endpoint URL of the event consumer.",
			},
			"authorization_type": schema.StringAttribute{
				Computed:    true,
				Description: "Auth type: API_KEY, BASIC, or OAUTH_CLIENT_CREDENTIALS",
			},
			"authorization_parameters": schema.StringAttribute{
				Computed:    true,
				Description: "JSON-encoded authorization parameters object",
			},
			"is_disabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the event consumer is disabled.",
			},
			"last_sync_date": schema.StringAttribute{
				Computed:    true,
				Description: "Last sync date.",
			},
			"last_sync_message": schema.StringAttribute{
				Computed:    true,
				Description: "Last sync message.",
			},
			"created_by": schema.StringAttribute{
				Computed:    true,
				Description: "User who created the event consumer.",
			},
			"created": schema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp.",
			},
			"last_modified_by": schema.StringAttribute{
				Computed:    true,
				Description: "User who last modified the event consumer.",
			},
			"last_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Last modification timestamp.",
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "Version of the event consumer.",
			},
		},
	}
}

func (d *EventConsumerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*TcmClient)
	if !ok {
		resp.Diagnostics.AddError("unexpected provider data type", fmt.Sprintf("expected *TcmClient, got %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *EventConsumerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state EventConsumerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, d.client.BaseURL+"/v1/EventConsumer/"+state.ID.ValueString(), bytes.NewReader([]byte{}))
	if err != nil {
		resp.Diagnostics.AddError("request build failed", err.Error())
		return
	}
	httpReq.Header.Set("Authorization", "Bearer "+d.client.Token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := d.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("read event consumer failed", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 400 {
		var errBody map[string]any
		json.NewDecoder(httpResp.Body).Decode(&errBody) //nolint:errcheck
		resp.Diagnostics.AddError("read event consumer failed", fmt.Sprintf("HTTP %d: %v", httpResp.StatusCode, errBody))
		return
	}

	var result eventConsumerPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		resp.Diagnostics.AddError("decoding response failed", err.Error())
		return
	}

	if result.ID != nil {
		state.ID = types.StringValue(*result.ID)
	}
	setOptionalString(&state.TenantID, result.TenantID)
	state.Name = types.StringValue(result.Name)
	setOptionalString(&state.Description, result.Description)
	state.Endpoint = types.StringValue(result.Endpoint)
	setOptionalString(&state.AuthorizationType, result.AuthorizationType)
	if result.AuthorizationParameters != nil {
		state.AuthorizationParameters = types.StringValue(string(*result.AuthorizationParameters))
	} else {
		state.AuthorizationParameters = types.StringNull()
	}
	state.IsDisabled = types.BoolValue(result.IsDisabled)
	setOptionalString(&state.LastSyncDate, result.LastSyncDate)
	setOptionalString(&state.LastSyncMessage, result.LastSyncMessage)
	setOptionalString(&state.CreatedBy, result.CreatedBy)
	setOptionalString(&state.Created, result.Created)
	setOptionalString(&state.LastModifiedBy, result.LastModifiedBy)
	setOptionalString(&state.LastModified, result.LastModified)
	if result.Version != nil {
		state.Version = types.Int64Value(*result.Version)
	} else {
		state.Version = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
