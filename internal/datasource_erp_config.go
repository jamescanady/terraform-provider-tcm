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

var _ datasource.DataSource = &ERPConfigDataSource{}

type ERPConfigDataSource struct {
	client *TcmClient
}

type ERPConfigDataSourceModel struct {
	ID                         types.String `tfsdk:"id"`
	ERPID                      types.String `tfsdk:"erp_id"`
	TenantProductEnvironmentID types.String `tfsdk:"tenant_product_environment_id"`
	URL                        types.String `tfsdk:"url"`
	ClientID                   types.String `tfsdk:"client_id"`
	ClientSecret               types.String `tfsdk:"client_secret"`
	LookupID                   types.String `tfsdk:"lookup_id"`
	HostName                   types.String `tfsdk:"host_name"`
	TenantSlug                 types.String `tfsdk:"tenant_slug"`
	APIVersion                 types.String `tfsdk:"api_version"`
	OntologyTypeIDs            types.List   `tfsdk:"ontology_type_ids"`
	GlobalTenantCode           types.String `tfsdk:"global_tenant_code"`
	ERPName                    types.String `tfsdk:"erp_name"`
	IsDeleted                  types.Bool   `tfsdk:"is_deleted"`
}

func NewERPConfigDataSource() datasource.DataSource {
	return &ERPConfigDataSource{}
}

func (d *ERPConfigDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_erp_config"
}

func (d *ERPConfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a TCM ERP Config by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the ERP config.",
			},
			"erp_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the ERP.",
			},
			"tenant_product_environment_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the tenant product environment.",
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "URL of the ERP.",
			},
			"client_id": schema.StringAttribute{
				Computed:    true,
				Description: "Client ID for authentication.",
			},
			"client_secret": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "Client secret for authentication.",
			},
			"lookup_id": schema.StringAttribute{
				Computed:    true,
				Description: "Lookup ID.",
			},
			"host_name": schema.StringAttribute{
				Computed:    true,
				Description: "Host name.",
			},
			"tenant_slug": schema.StringAttribute{
				Computed:    true,
				Description: "Tenant slug.",
			},
			"api_version": schema.StringAttribute{
				Computed:    true,
				Description: "API version.",
			},
			"ontology_type_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "List of ontology type IDs.",
			},
			"global_tenant_code": schema.StringAttribute{
				Computed:    true,
				Description: "Global tenant code.",
			},
			"erp_name": schema.StringAttribute{
				Computed:    true,
				Description: "ERP name.",
			},
			"is_deleted": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the ERP config is deleted.",
			},
		},
	}
}

func (d *ERPConfigDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ERPConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ERPConfigDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, d.client.BaseURL+"/v1/ERPConfig/"+state.ID.ValueString(), bytes.NewReader([]byte{}))
	if err != nil {
		resp.Diagnostics.AddError("request build failed", err.Error())
		return
	}
	httpReq.Header.Set("Authorization", "Bearer "+d.client.Token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := d.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("read ERP config failed", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 400 {
		var errBody map[string]any
		json.NewDecoder(httpResp.Body).Decode(&errBody) //nolint:errcheck
		resp.Diagnostics.AddError("read ERP config failed", fmt.Sprintf("HTTP %d: %v", httpResp.StatusCode, errBody))
		return
	}

	var result erpConfigPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		resp.Diagnostics.AddError("decoding response failed", err.Error())
		return
	}

	if result.ID != nil {
		state.ID = types.StringValue(*result.ID)
	}
	setOptionalString(&state.ERPID, result.ERPID)
	setOptionalString(&state.TenantProductEnvironmentID, result.TenantProductEnvironmentID)
	setOptionalString(&state.URL, result.URL)
	setOptionalString(&state.ClientID, result.ClientID)
	setOptionalString(&state.ClientSecret, result.ClientSecret)
	setOptionalString(&state.LookupID, result.LookupID)
	setOptionalString(&state.HostName, result.HostName)
	setOptionalString(&state.TenantSlug, result.TenantSlug)
	setOptionalString(&state.APIVersion, result.APIVersion)
	// OntologyTypeIDs is not returned by the API; set to null.
	state.OntologyTypeIDs = types.ListNull(types.StringType)
	setOptionalString(&state.GlobalTenantCode, result.GlobalTenantCode)
	setOptionalString(&state.ERPName, result.ERPName)
	if result.IsDeleted != nil {
		state.IsDeleted = types.BoolValue(*result.IsDeleted)
	} else {
		state.IsDeleted = types.BoolValue(false)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
