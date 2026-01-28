package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/poyrazk/terraform-provider-thecloud/internal/client"
	"github.com/poyrazk/terraform-provider-thecloud/internal/datasources"
	"github.com/poyrazk/terraform-provider-thecloud/internal/resources"
)

// TheCloudProvider implements the provider.Provider interface
type TheCloudProvider struct {
	version string
}

// TheCloudProviderModel describes the provider data model
type TheCloudProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	APIKey   types.String `tfsdk:"api_key"`
}

func (p *TheCloudProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "thecloud"
	resp.Version = p.version
}

func (p *TheCloudProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "The base URL for The Cloud API.",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The API key for authentication.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *TheCloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data TheCloudProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values with precedence:
	// 1. Provider configuration in .tf
	// 2. Environment variables
	// 3. Defaults

	endpoint := os.Getenv("THECLOUD_ENDPOINT")
	apiKey := os.Getenv("THECLOUD_API_KEY")

	if !data.Endpoint.IsNull() {
		endpoint = data.Endpoint.ValueString()
	}
	if !data.APIKey.IsNull() {
		apiKey = data.APIKey.ValueString()
	}

	if endpoint == "" {
		endpoint = "http://localhost:8080"
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key",
			"The provider requires an API key for authentication. Set it via 'api_key' in the provider block or THECLOUD_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	c := client.NewClient(endpoint, apiKey)

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *TheCloudProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewVpcResource,
		resources.NewInstanceResource,
		resources.NewVolumeResource,
		resources.NewSecurityGroupResource,
		resources.NewSecurityGroupRuleResource,
		resources.NewLoadBalancerResource,
		resources.NewLoadBalancerTargetResource,
		resources.NewSecretResource,
		resources.NewApiKeyResource,
		resources.NewScalingGroupResource,
	}
}

func (p *TheCloudProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewVpcDataSource,
		datasources.NewVpcsDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &TheCloudProvider{
			version: version,
		}
	}
}
