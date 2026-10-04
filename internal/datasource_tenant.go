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

var _ datasource.DataSource = &TenantDataSource{}

type TenantDataSource struct {
	client *TcmClient
}

type TenantDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	GlobalTenantCode types.String `tfsdk:"global_tenant_code"`
	TenantShortCode  types.String `tfsdk:"tenant_short_code"`
	IsDisabled       types.Bool   `tfsdk:"is_disabled"`
}

func NewTenantDataSource() datasource.DataSource {
	return &TenantDataSource{}
}

func (d *TenantDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant"
}

func (d *TenantDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a TCM Tenant by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the tenant.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Tenant name.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Tenant description.",
			},
			"global_tenant_code": schema.StringAttribute{
				Computed:    true,
				Description: "Global tenant code.",
			},
			"tenant_short_code": schema.StringAttribute{
				Computed:    true,
				Description: "Tenant short code.",
			},
			"is_disabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the tenant is disabled.",
			},
		},
	}
}

func (d *TenantDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TenantDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TenantDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, d.client.BaseURL+"/v1/Tenant/"+state.ID.ValueString(), bytes.NewReader([]byte{}))
	if err != nil {
		resp.Diagnostics.AddError("request build failed", err.Error())
		return
	}
	httpReq.Header.Set("Authorization", "Bearer "+d.client.Token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := d.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("read tenant failed", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 400 {
		var errBody map[string]any
		json.NewDecoder(httpResp.Body).Decode(&errBody) //nolint:errcheck
		resp.Diagnostics.AddError("read tenant failed", fmt.Sprintf("HTTP %d: %v", httpResp.StatusCode, errBody))
		return
	}

	var result tenantPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		resp.Diagnostics.AddError("decoding response failed", err.Error())
		return
	}

	if result.ID != nil {
		state.ID = types.StringValue(*result.ID)
	}
	state.Name = types.StringValue(result.Name)
	state.Description = types.StringValue(result.Description)
	state.GlobalTenantCode = types.StringValue(result.GlobalTenantCode)
	state.TenantShortCode = types.StringValue(result.TenantShortCode)
	state.IsDisabled = types.BoolValue(result.IsDisabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
