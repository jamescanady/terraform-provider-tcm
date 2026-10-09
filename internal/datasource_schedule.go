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

var _ datasource.DataSource = &ScheduleDataSource{}

type ScheduleDataSource struct {
	client *TcmClient
}

type ScheduleDataSourceModel struct {
	ID                         types.String `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
	Description                types.String `tfsdk:"description"`
	TenantProductEnvironmentID types.String `tfsdk:"tenant_product_environment_id"`
	ScheduleCategoryID         types.String `tfsdk:"schedule_category_id"`
	Enabled                    types.Bool   `tfsdk:"enabled"`
	ScheduleDetails            types.String `tfsdk:"schedule_details"`
}

func NewScheduleDataSource() datasource.DataSource {
	return &ScheduleDataSource{}
}

func (d *ScheduleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schedule"
}

func (d *ScheduleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a TCM Schedule by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the schedule.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the schedule.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the schedule.",
			},
			"tenant_product_environment_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the tenant product environment.",
			},
			"schedule_category_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the schedule category.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the schedule is enabled.",
			},
			"schedule_details": schema.StringAttribute{
				Computed:    true,
				Description: "JSON-encoded schedule details object.",
			},
		},
	}
}

func (d *ScheduleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ScheduleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ScheduleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, d.client.BaseURL+"/v1/Schedule/"+state.ID.ValueString(), bytes.NewReader([]byte{}))
	if err != nil {
		resp.Diagnostics.AddError("request build failed", err.Error())
		return
	}
	httpReq.Header.Set("Authorization", "Bearer "+d.client.Token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := d.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("read schedule failed", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 400 {
		var errBody map[string]any
		json.NewDecoder(httpResp.Body).Decode(&errBody) //nolint:errcheck
		resp.Diagnostics.AddError("read schedule failed", fmt.Sprintf("HTTP %d: %v", httpResp.StatusCode, errBody))
		return
	}

	var result schedulePayload
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		resp.Diagnostics.AddError("decoding response failed", err.Error())
		return
	}

	if result.ID != nil {
		state.ID = types.StringValue(*result.ID)
	}
	setOptionalString(&state.Name, result.Name)
	setOptionalString(&state.Description, result.Description)
	state.TenantProductEnvironmentID = types.StringValue(result.TenantProductEnvironmentID)
	state.ScheduleCategoryID = types.StringValue(result.ScheduleCategoryID)
	state.Enabled = types.BoolValue(result.Enabled)
	if result.ScheduleDetails != nil {
		state.ScheduleDetails = types.StringValue(string(*result.ScheduleDetails))
	} else {
		state.ScheduleDetails = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
