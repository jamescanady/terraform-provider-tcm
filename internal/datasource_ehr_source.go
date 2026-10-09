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

var _ datasource.DataSource = &EHRSourceDataSource{}

type EHRSourceDataSource struct {
	client *TcmClient
}

type EHRSourceDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	TenantID    types.String `tfsdk:"tenant_id"`
	SourceID    types.String `tfsdk:"source_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	IsDisabled  types.Bool   `tfsdk:"is_disabled"`
}

func NewEHRSourceDataSource() datasource.DataSource {
	return &EHRSourceDataSource{}
}

func (d *EHRSourceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ehr_source"
}

func (d *EHRSourceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a TCM EHR Source by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the EHR source.",
			},
			"tenant_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the tenant.",
			},
			"source_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the source.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the EHR source.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the EHR source.",
			},
			"is_disabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the EHR source is disabled.",
			},
		},
	}
}

func (d *EHRSourceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EHRSourceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state EHRSourceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, d.client.BaseURL+"/v1/ehr/source/"+state.ID.ValueString(), bytes.NewReader([]byte{}))
	if err != nil {
		resp.Diagnostics.AddError("request build failed", err.Error())
		return
	}
	httpReq.Header.Set("Authorization", "Bearer "+d.client.Token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := d.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("read ehr source failed", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 400 {
		var errBody map[string]any
		json.NewDecoder(httpResp.Body).Decode(&errBody) //nolint:errcheck
		resp.Diagnostics.AddError("read ehr source failed", fmt.Sprintf("HTTP %d: %v", httpResp.StatusCode, errBody))
		return
	}

	var result ehrSourcePayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		resp.Diagnostics.AddError("decoding response failed", err.Error())
		return
	}

	if result.ID != nil {
		state.ID = types.StringValue(*result.ID)
	}
	setOptionalString(&state.TenantID, result.TenantID)
	setOptionalString(&state.SourceID, result.SourceID)
	setOptionalString(&state.Name, result.Name)
	setOptionalString(&state.Description, result.Description)
	state.IsDisabled = types.BoolValue(result.IsDisabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
