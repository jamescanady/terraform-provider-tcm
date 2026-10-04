package internal

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &TcmProvider{}

// TcmClient is the configured HTTP client shared with all resources.
type TcmClient struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

type TcmProvider struct {
	version string
}

type tcmProviderModel struct {
	BaseURL types.String `tfsdk:"base_url"`
	Token   types.String `tfsdk:"token"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &TcmProvider{version: version}
	}
}

func (p *TcmProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "tcm"
	resp.Version = p.version
}

func (p *TcmProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with the symplr Tenant Configuration Management (TCM) service.",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Required:    true,
				Description: "TCM service base URL (e.g. https://dev-platform.symplr.com/ce-platform-tenant-configuration-service).",
			},
			"token": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "OAuth bearer token for authenticating with the TCM service.",
			},
		},
	}
}

func (p *TcmProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config tcmProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := &TcmClient{
		BaseURL:    config.BaseURL.ValueString(),
		Token:      config.Token.ValueString(),
		HTTPClient: &http.Client{},
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *TcmProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProductResource,
		NewTenantResource,
		NewTenantProductResource,
		NewTenantProductEnvironmentResource,
		NewSystemInfoResource,
		NewNamespaceResource,
	}
}

func (p *TcmProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewTenantDataSource,
		NewTenantProductDataSource,
		NewTenantProductEnvironmentDataSource,
		NewSystemInfoDataSource,
		NewNamespaceDataSource,
	}
}
