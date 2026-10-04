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

var _ datasource.DataSource = &TenantProductEnvironmentDataSource{}

type TenantProductEnvironmentDataSource struct {
	client *TcmClient
}

type TenantProductEnvironmentDataSourceModel struct {
	ID                   types.String `tfsdk:"id"`
	TenantID             types.String `tfsdk:"tenant_id"`
	ProductEnvironmentID types.String `tfsdk:"product_environment_id"`
	NamespaceID          types.String `tfsdk:"namespace_id"`
	ProductTenantCode    types.String `tfsdk:"product_tenant_code"`
	ProductAlias         types.String `tfsdk:"product_alias"`
	IsDisabled           types.Bool   `tfsdk:"is_disabled"`
}

func NewTenantProductEnvironmentDataSource() datasource.DataSource {
	return &TenantProductEnvironmentDataSource{}
}

func (d *TenantProductEnvironmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_product_environment"
}

func (d *TenantProductEnvironmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a TCM TenantProductEnvironment by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the tenant product environment.",
			},
			"tenant_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the tenant.",
			},
			"product_environment_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the product environment.",
			},
			"namespace_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the namespace.",
			},
			"product_tenant_code": schema.StringAttribute{
				Computed:    true,
				Description: "Product tenant code.",
			},
			"product_alias": schema.StringAttribute{
				Computed:    true,
				Description: "Product alias.",
			},
			"is_disabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the tenant product environment is disabled.",
			},
		},
	}
}

func (d *TenantProductEnvironmentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TenantProductEnvironmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TenantProductEnvironmentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, d.client.BaseURL+"/v1/TenantProductEnvironment/"+state.ID.ValueString(), bytes.NewReader([]byte{}))
	if err != nil {
		resp.Diagnostics.AddError("request build failed", err.Error())
		return
	}
	httpReq.Header.Set("Authorization", "Bearer "+d.client.Token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := d.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("read tenant product environment failed", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 400 {
		var errBody map[string]any
		json.NewDecoder(httpResp.Body).Decode(&errBody) //nolint:errcheck
		resp.Diagnostics.AddError("read tenant product environment failed", fmt.Sprintf("HTTP %d: %v", httpResp.StatusCode, errBody))
		return
	}

	var result tenantProductEnvironmentPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		resp.Diagnostics.AddError("decoding response failed", err.Error())
		return
	}

	if result.ID != nil {
		state.ID = types.StringValue(*result.ID)
	}
	state.TenantID = types.StringValue(result.TenantID)
	state.ProductEnvironmentID = types.StringValue(result.ProductEnvironmentID)
	if result.NamespaceID != nil {
		state.NamespaceID = types.StringValue(*result.NamespaceID)
	} else {
		state.NamespaceID = types.StringNull()
	}
	if result.ProductTenantCode != nil {
		state.ProductTenantCode = types.StringValue(*result.ProductTenantCode)
	} else {
		state.ProductTenantCode = types.StringNull()
	}
	if result.ProductAlias != nil {
		state.ProductAlias = types.StringValue(*result.ProductAlias)
	} else {
		state.ProductAlias = types.StringNull()
	}
	state.IsDisabled = types.BoolValue(result.IsDisabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
