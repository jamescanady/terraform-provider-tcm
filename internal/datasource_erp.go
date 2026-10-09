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

var _ datasource.DataSource = &ERPDataSource{}

type ERPDataSource struct {
	client *TcmClient
}

type ERPDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	Version          types.String `tfsdk:"version"`
	IsDeleted        types.Bool   `tfsdk:"is_deleted"`
	CreatedBy        types.String `tfsdk:"created_by"`
	CreatedDate      types.String `tfsdk:"created_date"`
	LastModifiedBy   types.String `tfsdk:"last_modified_by"`
	LastModifiedDate types.String `tfsdk:"last_modified_date"`
}

func NewERPDataSource() datasource.DataSource {
	return &ERPDataSource{}
}

func (d *ERPDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_erp"
}

func (d *ERPDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a TCM ERP by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the ERP.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the ERP.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the ERP.",
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "Version of the ERP.",
			},
			"is_deleted": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the ERP is deleted.",
			},
			"created_by": schema.StringAttribute{
				Computed:    true,
				Description: "User who created the ERP.",
			},
			"created_date": schema.StringAttribute{
				Computed:    true,
				Description: "Date the ERP was created.",
			},
			"last_modified_by": schema.StringAttribute{
				Computed:    true,
				Description: "User who last modified the ERP.",
			},
			"last_modified_date": schema.StringAttribute{
				Computed:    true,
				Description: "Date the ERP was last modified.",
			},
		},
	}
}

func (d *ERPDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ERPDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ERPDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, d.client.BaseURL+"/v1/ERP/"+state.ID.ValueString(), bytes.NewReader([]byte{}))
	if err != nil {
		resp.Diagnostics.AddError("request build failed", err.Error())
		return
	}
	httpReq.Header.Set("Authorization", "Bearer "+d.client.Token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := d.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("read ERP failed", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 400 {
		var errBody map[string]any
		json.NewDecoder(httpResp.Body).Decode(&errBody) //nolint:errcheck
		resp.Diagnostics.AddError("read ERP failed", fmt.Sprintf("HTTP %d: %v", httpResp.StatusCode, errBody))
		return
	}

	var result erpPayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		resp.Diagnostics.AddError("decoding response failed", err.Error())
		return
	}

	if result.ID != nil {
		state.ID = types.StringValue(*result.ID)
	}
	setOptionalString(&state.Name, result.Name)
	setOptionalString(&state.Description, result.Description)
	setOptionalString(&state.Version, result.Version)
	if result.IsDeleted != nil {
		state.IsDeleted = types.BoolValue(*result.IsDeleted)
	} else {
		state.IsDeleted = types.BoolValue(false)
	}
	setOptionalString(&state.CreatedBy, result.CreatedBy)
	setOptionalString(&state.CreatedDate, result.CreatedDate)
	setOptionalString(&state.LastModifiedBy, result.LastModifiedBy)
	setOptionalString(&state.LastModifiedDate, result.LastModifiedDate)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
