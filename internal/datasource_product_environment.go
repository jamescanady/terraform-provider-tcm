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

var _ datasource.DataSource = &ProductEnvironmentDataSource{}

type ProductEnvironmentDataSource struct {
	client *TcmClient
}

type ProductEnvironmentDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	ProductID  types.String `tfsdk:"product_id"`
	Name       types.String `tfsdk:"name"`
	IsDisabled types.Bool   `tfsdk:"is_disabled"`
}

func NewProductEnvironmentDataSource() datasource.DataSource {
	return &ProductEnvironmentDataSource{}
}

func (d *ProductEnvironmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_product_environment"
}

func (d *ProductEnvironmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a TCM ProductEnvironment by id.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the product environment.",
			},
			"product_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the product.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Environment name.",
			},
			"is_disabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the product environment is disabled.",
			},
		},
	}
}

func (d *ProductEnvironmentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProductEnvironmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ProductEnvironmentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, d.client.BaseURL+"/v1/ProductEnvironment/"+state.ID.ValueString(), bytes.NewReader([]byte{}))
	if err != nil {
		resp.Diagnostics.AddError("request build failed", err.Error())
		return
	}
	httpReq.Header.Set("Authorization", "Bearer "+d.client.Token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := d.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("read product environment failed", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 400 {
		var errBody map[string]any
		json.NewDecoder(httpResp.Body).Decode(&errBody) //nolint:errcheck
		resp.Diagnostics.AddError("read product environment failed", fmt.Sprintf("HTTP %d: %v", httpResp.StatusCode, errBody))
		return
	}

	var item productEnvironmentPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&item); err != nil {
		resp.Diagnostics.AddError("decoding response failed", err.Error())
		return
	}

	if item.ID != nil {
		state.ID = types.StringValue(*item.ID)
	}
	state.ProductID = types.StringValue(item.ProductID)
	if item.Name != nil {
		state.Name = types.StringValue(*item.Name)
	} else {
		state.Name = types.StringNull()
	}
	state.IsDisabled = types.BoolValue(item.IsDisabled)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
