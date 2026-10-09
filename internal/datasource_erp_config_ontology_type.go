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

var _ datasource.DataSource = &ERPConfigOntologyTypeDataSource{}

type ERPConfigOntologyTypeDataSource struct {
	client *TcmClient
}

type ERPConfigOntologyTypeDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	ERPConfigID    types.String `tfsdk:"erp_config_id"`
	OntologyTypeID types.String `tfsdk:"ontology_type_id"`
}

func NewERPConfigOntologyTypeDataSource() datasource.DataSource {
	return &ERPConfigOntologyTypeDataSource{}
}

func (d *ERPConfigOntologyTypeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_erp_config_ontology_type"
}

func (d *ERPConfigOntologyTypeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a TCM ERPConfigOntologyType by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the ERP config ontology type.",
			},
			"erp_config_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the ERP config.",
			},
			"ontology_type_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the ontology type.",
			},
		},
	}
}

func (d *ERPConfigOntologyTypeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ERPConfigOntologyTypeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ERPConfigOntologyTypeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, d.client.BaseURL+"/v1/ERPConfigOntologyType/"+state.ID.ValueString(), bytes.NewReader([]byte{}))
	if err != nil {
		resp.Diagnostics.AddError("request build failed", err.Error())
		return
	}
	httpReq.Header.Set("Authorization", "Bearer "+d.client.Token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := d.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("read erp config ontology type failed", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 400 {
		var errBody map[string]any
		json.NewDecoder(httpResp.Body).Decode(&errBody) //nolint:errcheck
		resp.Diagnostics.AddError("read erp config ontology type failed", fmt.Sprintf("HTTP %d: %v", httpResp.StatusCode, errBody))
		return
	}

	var result erpConfigOntologyTypePayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		resp.Diagnostics.AddError("decoding response failed", err.Error())
		return
	}

	setOptionalString(&state.ID, result.ID)
	setOptionalString(&state.ERPConfigID, result.ERPConfigID)
	setOptionalString(&state.OntologyTypeID, result.OntologyTypeID)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
