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

var _ datasource.DataSource = &EHRSchemaDataSource{}

type EHRSchemaDataSource struct {
	client *TcmClient
}

type EHRSchemaDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	TenantID       types.String `tfsdk:"tenant_id"`
	EndpointID     types.String `tfsdk:"endpoint_id"`
	RequiredFields types.List   `tfsdk:"required_fields"`
	IsDisabled     types.Bool   `tfsdk:"is_disabled"`
}

func NewEHRSchemaDataSource() datasource.DataSource {
	return &EHRSchemaDataSource{}
}

func (d *EHRSchemaDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ehr_schema"
}

func (d *EHRSchemaDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a TCM EHR Schema by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the EHR schema.",
			},
			"tenant_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the tenant.",
			},
			"endpoint_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the endpoint.",
			},
			"required_fields": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "List of required field names.",
			},
			"is_disabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the EHR schema is disabled.",
			},
		},
	}
}

func (d *EHRSchemaDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EHRSchemaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state EHRSchemaDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, d.client.BaseURL+"/v1/ehr/schema/"+state.ID.ValueString(), bytes.NewReader([]byte{}))
	if err != nil {
		resp.Diagnostics.AddError("request build failed", err.Error())
		return
	}
	httpReq.Header.Set("Authorization", "Bearer "+d.client.Token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := d.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("read ehr schema failed", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 400 {
		var errBody map[string]any
		json.NewDecoder(httpResp.Body).Decode(&errBody) //nolint:errcheck
		resp.Diagnostics.AddError("read ehr schema failed", fmt.Sprintf("HTTP %d: %v", httpResp.StatusCode, errBody))
		return
	}

	var result ehrSchemaPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		resp.Diagnostics.AddError("decoding response failed", err.Error())
		return
	}

	if result.ID != nil {
		state.ID = types.StringValue(*result.ID)
	}
	setOptionalString(&state.TenantID, result.TenantID)
	setOptionalString(&state.EndpointID, result.EndpointID)
	state.IsDisabled = types.BoolValue(result.IsDisabled)
	if result.RequiredFields != nil {
		listVal, _ := types.ListValueFrom(ctx, types.StringType, result.RequiredFields)
		state.RequiredFields = listVal
	} else {
		state.RequiredFields = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
