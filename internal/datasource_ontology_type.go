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

var _ datasource.DataSource = &OntologyTypeDataSource{}

type OntologyTypeDataSource struct {
	client *TcmClient
}

type OntologyTypeDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	IsDisabled  types.Bool   `tfsdk:"is_disabled"`
}

func NewOntologyTypeDataSource() datasource.DataSource {
	return &OntologyTypeDataSource{}
}

func (d *OntologyTypeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ontology_type"
}

func (d *OntologyTypeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a TCM OntologyType by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the ontology type.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the ontology type.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the ontology type.",
			},
			"is_disabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the ontology type is disabled.",
			},
		},
	}
}

func (d *OntologyTypeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OntologyTypeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state OntologyTypeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, d.client.BaseURL+"/v1/OntologyType/"+state.ID.ValueString(), bytes.NewReader([]byte{}))
	if err != nil {
		resp.Diagnostics.AddError("request build failed", err.Error())
		return
	}
	httpReq.Header.Set("Authorization", "Bearer "+d.client.Token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := d.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("read ontology type failed", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 400 {
		var errBody map[string]any
		json.NewDecoder(httpResp.Body).Decode(&errBody) //nolint:errcheck
		resp.Diagnostics.AddError("read ontology type failed", fmt.Sprintf("HTTP %d: %v", httpResp.StatusCode, errBody))
		return
	}

	var result ontologyTypePayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		resp.Diagnostics.AddError("decoding response failed", err.Error())
		return
	}

	if result.ID != nil {
		state.ID = types.StringValue(*result.ID)
	}
	setOptionalString(&state.Name, result.Name)
	setOptionalString(&state.Description, result.Description)
	if result.IsDisabled != nil {
		state.IsDisabled = types.BoolValue(*result.IsDisabled)
	} else {
		state.IsDisabled = types.BoolValue(false)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
