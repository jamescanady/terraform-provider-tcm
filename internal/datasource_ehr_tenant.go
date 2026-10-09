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

var _ datasource.DataSource = &EHRTenantDataSource{}

type EHRTenantDataSource struct {
	client *TcmClient
}

type EHRTenantDataSourceModel struct {
	ID                 types.String `tfsdk:"id"`
	TenantID           types.String `tfsdk:"tenant_id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	RedoxEHRIdentifier types.String `tfsdk:"redox_ehr_identifier"`
	SymplrURLSlug      types.String `tfsdk:"symplr_url_slug"`
	RedoxEnvironment   types.String `tfsdk:"redox_environment"`
	IsDisabled         types.Bool   `tfsdk:"is_disabled"`
}

func NewEHRTenantDataSource() datasource.DataSource {
	return &EHRTenantDataSource{}
}

func (d *EHRTenantDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ehr_tenant"
}

func (d *EHRTenantDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a TCM EHR Tenant by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the EHR tenant.",
			},
			"tenant_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the tenant.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the EHR tenant.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the EHR tenant.",
			},
			"redox_ehr_identifier": schema.StringAttribute{
				Computed:    true,
				Description: "Redox EHR identifier.",
			},
			"symplr_url_slug": schema.StringAttribute{
				Computed:    true,
				Description: "Symplr URL slug.",
			},
			"redox_environment": schema.StringAttribute{
				Computed:    true,
				Description: "Redox environment.",
			},
			"is_disabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the EHR tenant is disabled.",
			},
		},
	}
}

func (d *EHRTenantDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EHRTenantDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state EHRTenantDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, d.client.BaseURL+"/v1/ehr/tenant/"+state.ID.ValueString(), bytes.NewReader([]byte{}))
	if err != nil {
		resp.Diagnostics.AddError("request build failed", err.Error())
		return
	}
	httpReq.Header.Set("Authorization", "Bearer "+d.client.Token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := d.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("read ehr tenant failed", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 400 {
		var errBody map[string]any
		json.NewDecoder(httpResp.Body).Decode(&errBody) //nolint:errcheck
		resp.Diagnostics.AddError("read ehr tenant failed", fmt.Sprintf("HTTP %d: %v", httpResp.StatusCode, errBody))
		return
	}

	var result ehrTenantPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		resp.Diagnostics.AddError("decoding response failed", err.Error())
		return
	}

	if result.ID != nil {
		state.ID = types.StringValue(*result.ID)
	}
	setOptionalString(&state.TenantID, result.TenantID)
	state.Name = types.StringValue(result.Name)
	setOptionalString(&state.Description, result.Description)
	state.RedoxEHRIdentifier = types.StringValue(result.RedoxEHRIdentifier)
	state.SymplrURLSlug = types.StringValue(result.SymplrURLSlug)
	state.RedoxEnvironment = types.StringValue(result.RedoxEnvironment)
	state.IsDisabled = types.BoolValue(result.IsDisabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
