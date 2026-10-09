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

var _ datasource.DataSource = &EHREndpointDataSource{}

type EHREndpointDataSource struct {
	client *TcmClient
}

type EHREndpointDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	IsDisabled  types.Bool   `tfsdk:"is_disabled"`
}

func NewEHREndpointDataSource() datasource.DataSource {
	return &EHREndpointDataSource{}
}

func (d *EHREndpointDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ehr_endpoint"
}

func (d *EHREndpointDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a TCM EHR Endpoint by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the EHR endpoint.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the EHR endpoint.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the EHR endpoint.",
			},
			"is_disabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the EHR endpoint is disabled.",
			},
		},
	}
}

func (d *EHREndpointDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EHREndpointDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state EHREndpointDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, d.client.BaseURL+"/v1/ehr/endpoint/"+state.ID.ValueString(), bytes.NewReader([]byte{}))
	if err != nil {
		resp.Diagnostics.AddError("request build failed", err.Error())
		return
	}
	httpReq.Header.Set("Authorization", "Bearer "+d.client.Token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := d.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("read ehr endpoint failed", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 400 {
		var errBody map[string]any
		json.NewDecoder(httpResp.Body).Decode(&errBody) //nolint:errcheck
		resp.Diagnostics.AddError("read ehr endpoint failed", fmt.Sprintf("HTTP %d: %v", httpResp.StatusCode, errBody))
		return
	}

	var result ehrEndpointPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		resp.Diagnostics.AddError("decoding response failed", err.Error())
		return
	}

	if result.ID != nil {
		state.ID = types.StringValue(*result.ID)
	}
	state.Name = types.StringValue(result.Name)
	setOptionalString(&state.Description, result.Description)
	state.IsDisabled = types.BoolValue(result.IsDisabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
