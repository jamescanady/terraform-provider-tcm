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

var _ datasource.DataSource = &SystemInfoDataSource{}

type SystemInfoDataSource struct {
	client *TcmClient
}

type SystemInfoDataSourceModel struct {
	ID                   types.String `tfsdk:"id"`
	ProductID            types.String `tfsdk:"product_id"`
	TenantID             types.String `tfsdk:"tenant_id"`
	NamespaceID          types.String `tfsdk:"namespace_id"`
	ProductEnvironmentID types.String `tfsdk:"product_environment_id"`
	FlowID               types.String `tfsdk:"flow_id"`
	ConnectionType       types.String `tfsdk:"connection_type"`
	Description          types.String `tfsdk:"description"`
	Host                 types.String `tfsdk:"host"`
	FlowVersion          types.String `tfsdk:"flow_version"`
	BaseApiPath          types.String `tfsdk:"base_api_path"`
	Endpoint             types.String `tfsdk:"endpoint"`
	OAuthScope           types.String `tfsdk:"o_auth_scope"`
	IsDisabled           types.Bool   `tfsdk:"is_disabled"`
	NamespaceName        types.String `tfsdk:"namespace_name"`
	TenantName           types.String `tfsdk:"tenant_name"`
	ProductName          types.String `tfsdk:"product_name"`
}

func NewSystemInfoDataSource() datasource.DataSource {
	return &SystemInfoDataSource{}
}

func (d *SystemInfoDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_info"
}

func (d *SystemInfoDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a TCM SystemInfo by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the system info.",
			},
			"product_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the product.",
			},
			"tenant_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the tenant.",
			},
			"namespace_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the namespace.",
			},
			"product_environment_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the product environment.",
			},
			"flow_id": schema.StringAttribute{
				Computed:    true,
				Description: "Flow identifier.",
			},
			"connection_type": schema.StringAttribute{
				Computed:    true,
				Description: "Connection type.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description.",
			},
			"host": schema.StringAttribute{
				Computed:    true,
				Description: "Host.",
			},
			"flow_version": schema.StringAttribute{
				Computed:    true,
				Description: "Flow version.",
			},
			"base_api_path": schema.StringAttribute{
				Computed:    true,
				Description: "Base API path.",
			},
			"endpoint": schema.StringAttribute{
				Computed:    true,
				Description: "Endpoint.",
			},
			"o_auth_scope": schema.StringAttribute{
				Computed:    true,
				Description: "OAuth scope.",
			},
			"is_disabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the system info is disabled.",
			},
			"namespace_name": schema.StringAttribute{
				Computed:    true,
				Description: "Namespace name.",
			},
			"tenant_name": schema.StringAttribute{
				Computed:    true,
				Description: "Tenant name.",
			},
			"product_name": schema.StringAttribute{
				Computed:    true,
				Description: "Product name.",
			},
		},
	}
}

func (d *SystemInfoDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SystemInfoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state SystemInfoDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, d.client.BaseURL+"/v1/SystemInfo/"+state.ID.ValueString(), bytes.NewReader([]byte{}))
	if err != nil {
		resp.Diagnostics.AddError("request build failed", err.Error())
		return
	}
	httpReq.Header.Set("Authorization", "Bearer "+d.client.Token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := d.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("read system info failed", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 400 {
		var errBody map[string]any
		json.NewDecoder(httpResp.Body).Decode(&errBody) //nolint:errcheck
		resp.Diagnostics.AddError("read system info failed", fmt.Sprintf("HTTP %d: %v", httpResp.StatusCode, errBody))
		return
	}

	var result systemInfoPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		resp.Diagnostics.AddError("decoding response failed", err.Error())
		return
	}

	if result.ID != nil {
		state.ID = types.StringValue(*result.ID)
	}
	state.ProductID = types.StringValue(result.ProductID)
	setOptionalString(&state.TenantID, result.TenantID)
	setOptionalString(&state.NamespaceID, result.NamespaceID)
	setOptionalString(&state.ProductEnvironmentID, result.ProductEnvironmentID)
	setOptionalString(&state.FlowID, result.FlowID)
	setOptionalString(&state.ConnectionType, result.ConnectionType)
	setOptionalString(&state.Description, result.Description)
	if result.Host != nil {
		state.Host = types.StringValue(*result.Host)
	}
	if result.FlowVersion != nil {
		state.FlowVersion = types.StringValue(*result.FlowVersion)
	}
	if result.BaseApiPath != nil {
		state.BaseApiPath = types.StringValue(*result.BaseApiPath)
	}
	setOptionalString(&state.Endpoint, result.Endpoint)
	if result.OAuthScope != nil {
		state.OAuthScope = types.StringValue(*result.OAuthScope)
	}
	state.IsDisabled = types.BoolValue(result.IsDisabled)
	setOptionalString(&state.NamespaceName, result.NamespaceName)
	setOptionalString(&state.TenantName, result.TenantName)
	setOptionalString(&state.ProductName, result.ProductName)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
