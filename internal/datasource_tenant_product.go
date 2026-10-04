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

var _ datasource.DataSource = &TenantProductDataSource{}

type TenantProductDataSource struct {
	client *TcmClient
}

type TenantProductDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	TenantID          types.String `tfsdk:"tenant_id"`
	ProductID         types.String `tfsdk:"product_id"`
	TenantProductCode types.String `tfsdk:"tenant_product_code"`
	IsDisabled        types.Bool   `tfsdk:"is_disabled"`
}

func NewTenantProductDataSource() datasource.DataSource {
	return &TenantProductDataSource{}
}

func (d *TenantProductDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_product"
}

func (d *TenantProductDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a TCM TenantProduct mapping by id.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the tenant product mapping.",
			},
			"tenant_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the tenant.",
			},
			"product_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the product.",
			},
			"tenant_product_code": schema.StringAttribute{
				Computed:    true,
				Description: "Tenant product code.",
			},
			"is_disabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the tenant product mapping is disabled.",
			},
		},
	}
}

func (d *TenantProductDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TenantProductDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TenantProductDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, d.client.BaseURL+"/v1/TenantProduct/"+state.ID.ValueString(), bytes.NewReader([]byte{}))
	if err != nil {
		resp.Diagnostics.AddError("request build failed", err.Error())
		return
	}
	httpReq.Header.Set("Authorization", "Bearer "+d.client.Token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := d.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("read tenant product failed", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 400 {
		var errBody map[string]any
		json.NewDecoder(httpResp.Body).Decode(&errBody) //nolint:errcheck
		resp.Diagnostics.AddError("read tenant product failed", fmt.Sprintf("HTTP %d: %v", httpResp.StatusCode, errBody))
		return
	}

	var item tenantProductPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&item); err != nil {
		resp.Diagnostics.AddError("decoding response failed", err.Error())
		return
	}

	state.TenantID = types.StringValue(item.TenantID)
	state.ProductID = types.StringValue(item.ProductID)
	if item.TenantProductCode != nil {
		state.TenantProductCode = types.StringValue(*item.TenantProductCode)
	} else {
		state.TenantProductCode = types.StringNull()
	}
	state.IsDisabled = types.BoolValue(item.IsDisabled)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
