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

var _ datasource.DataSource = &NamespaceDataSource{}

type NamespaceDataSource struct {
	client *TcmClient
}

type NamespaceDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	IsDisabled     types.Bool   `tfsdk:"is_disabled"`
	IsDefault      types.Bool   `tfsdk:"is_default"`
	CreatedDate    types.String `tfsdk:"created_date"`
	CreatedBy      types.String `tfsdk:"created_by"`
	LastModified   types.String `tfsdk:"last_modified"`
	LastModifiedBy types.String `tfsdk:"last_modified_by"`
}

func NewNamespaceDataSource() datasource.DataSource {
	return &NamespaceDataSource{}
}

func (d *NamespaceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_namespace"
}

func (d *NamespaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a TCM Namespace by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the namespace.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Namespace name.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Namespace description.",
			},
			"is_disabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the namespace is disabled.",
			},
			"is_default": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether this is the default namespace.",
			},
			"created_date": schema.StringAttribute{
				Computed:    true,
				Description: "UTC date and time the namespace was created.",
			},
			"created_by": schema.StringAttribute{
				Computed:    true,
				Description: "Identity that created the namespace.",
			},
			"last_modified": schema.StringAttribute{
				Computed:    true,
				Description: "UTC date and time the namespace was last modified.",
			},
			"last_modified_by": schema.StringAttribute{
				Computed:    true,
				Description: "Identity that last modified the namespace.",
			},
		},
	}
}

func (d *NamespaceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NamespaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state NamespaceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, d.client.BaseURL+"/v1/Namespace/"+state.ID.ValueString(), bytes.NewReader([]byte{}))
	if err != nil {
		resp.Diagnostics.AddError("request build failed", err.Error())
		return
	}
	httpReq.Header.Set("Authorization", "Bearer "+d.client.Token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := d.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("read namespace failed", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 400 {
		var errBody map[string]any
		json.NewDecoder(httpResp.Body).Decode(&errBody) //nolint:errcheck
		resp.Diagnostics.AddError("read namespace failed", fmt.Sprintf("HTTP %d: %v", httpResp.StatusCode, errBody))
		return
	}

	var result namespaceResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		resp.Diagnostics.AddError("decoding response failed", err.Error())
		return
	}

	state.ID = types.StringValue(result.ID)
	if result.Name != nil {
		state.Name = types.StringValue(*result.Name)
	} else {
		state.Name = types.StringNull()
	}
	if result.Description != nil {
		state.Description = types.StringValue(*result.Description)
	} else {
		state.Description = types.StringNull()
	}
	state.IsDisabled = types.BoolValue(result.IsDisabled)
	state.IsDefault = types.BoolValue(result.IsDefault)
	state.CreatedDate = types.StringValue(result.CreatedDate)
	if result.CreatedBy != nil {
		state.CreatedBy = types.StringValue(*result.CreatedBy)
	} else {
		state.CreatedBy = types.StringNull()
	}
	state.LastModified = types.StringValue(result.LastModified)
	if result.LastModifiedBy != nil {
		state.LastModifiedBy = types.StringValue(*result.LastModifiedBy)
	} else {
		state.LastModifiedBy = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
