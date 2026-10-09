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

var _ datasource.DataSource = &EventTypeConsumerDataSource{}

type EventTypeConsumerDataSource struct {
	client *TcmClient
}

type EventTypeConsumerDataSourceModel struct {
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

func NewEventTypeConsumerDataSource() datasource.DataSource {
	return &EventTypeConsumerDataSource{}
}

func (d *EventTypeConsumerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_type_consumer"
}

func (d *EventTypeConsumerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a TCM EventTypeConsumer by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the event type consumer.",
			},
			"event_type_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the event type.",
			},
			"event_consumer_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the event consumer.",
			},
			"tenant_product_environment_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the tenant product environment.",
			},
			"is_disabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the event type consumer is disabled.",
			},
			"consumer_name": schema.StringAttribute{
				Computed:    true,
				Description: "Consumer name.",
			},
			"consumer_description": schema.StringAttribute{
				Computed:    true,
				Description: "Consumer description.",
			},
			"consumer_endpoint": schema.StringAttribute{
				Computed:    true,
				Description: "Consumer endpoint.",
			},
			"consumer_authorization_type": schema.StringAttribute{
				Computed:    true,
				Description: "Consumer authorization type.",
			},
			"event_type_name": schema.StringAttribute{
				Computed:    true,
				Description: "Event type name.",
			},
			"tenant_id": schema.StringAttribute{
				Computed:    true,
				Description: "Tenant UUID.",
			},
			"tenant_name": schema.StringAttribute{
				Computed:    true,
				Description: "Tenant name.",
			},
			"product_id": schema.StringAttribute{
				Computed:    true,
				Description: "Product UUID.",
			},
			"product_name": schema.StringAttribute{
				Computed:    true,
				Description: "Product name.",
			},
			"product_string": schema.StringAttribute{
				Computed:    true,
				Description: "Product string.",
			},
			"environment_name": schema.StringAttribute{
				Computed:    true,
				Description: "Environment name.",
			},
			"created_by": schema.StringAttribute{
				Computed:    true,
				Description: "User who created the record.",
			},
			"created": schema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp.",
			},
			"last_modified_by": schema.StringAttribute{
				Computed:    true,
				Description: "User who last modified the record.",
			},
			"last_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Last modification timestamp.",
			},
		},
	}
}

func (d *EventTypeConsumerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EventTypeConsumerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state EventTypeConsumerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, d.client.BaseURL+"/v1/EventTypeConsumer/"+state.ID.ValueString(), bytes.NewReader([]byte{}))
	if err != nil {
		resp.Diagnostics.AddError("request build failed", err.Error())
		return
	}
	httpReq.Header.Set("Authorization", "Bearer "+d.client.Token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := d.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("read event type consumer failed", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 400 {
		var errBody map[string]any
		json.NewDecoder(httpResp.Body).Decode(&errBody) //nolint:errcheck
		resp.Diagnostics.AddError("read event type consumer failed", fmt.Sprintf("HTTP %d: %v", httpResp.StatusCode, errBody))
		return
	}

	var result eventTypeConsumerPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		resp.Diagnostics.AddError("decoding response failed", err.Error())
		return
	}

	if result.ID != nil {
		state.ID = types.StringValue(*result.ID)
	}
	setOptionalString(&state.EventTypeID, result.EventTypeID)
	// Map consumerId (response) → event_consumer_id (model)
	if result.ConsumerID != nil {
		state.EventConsumerID = types.StringValue(*result.ConsumerID)
	} else {
		setOptionalString(&state.EventConsumerID, result.ConsumerID)
	}
	setOptionalString(&state.TenantProductEnvironmentID, result.TenantProductEnvironmentID)
	state.IsDisabled = types.BoolValue(result.IsDisabled)
	setOptionalString(&state.ConsumerName, result.ConsumerName)
	setOptionalString(&state.ConsumerDescription, result.ConsumerDescription)
	setOptionalString(&state.ConsumerEndpoint, result.ConsumerEndpoint)
	setOptionalString(&state.ConsumerAuthorizationType, result.ConsumerAuthorizationType)
	setOptionalString(&state.EventTypeName, result.EventTypeName)
	setOptionalString(&state.TenantID, result.TenantID)
	setOptionalString(&state.TenantName, result.TenantName)
	setOptionalString(&state.ProductID, result.ProductID)
	setOptionalString(&state.ProductName, result.ProductName)
	setOptionalString(&state.ProductString, result.ProductString)
	setOptionalString(&state.EnvironmentName, result.EnvironmentName)
	setOptionalString(&state.CreatedBy, result.CreatedBy)
	setOptionalString(&state.Created, result.Created)
	setOptionalString(&state.LastModifiedBy, result.LastModifiedBy)
	setOptionalString(&state.LastModified, result.LastModified)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
