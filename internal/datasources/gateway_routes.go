package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/poyrazk/terraform-provider-thecloud/internal/client"
)

// Ensure implementation of interfaces
var _ datasource.DataSource = &GatewayRoutesDataSource{}

func NewGatewayRoutesDataSource() datasource.DataSource {
	return &GatewayRoutesDataSource{}
}

// GatewayRoutesDataSource defines the data source implementation.
type GatewayRoutesDataSource struct {
	client *client.Client
}

// GatewayRoutesDataSourceModel describes the data source data model.
type GatewayRoutesDataSourceModel struct {
	Routes []GatewayRouteDataSourceModel `tfsdk:"routes"`
}

func (d *GatewayRoutesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gateway_routes"
}

func (d *GatewayRoutesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gateway Routes data source allows you to list all available API gateway routes.",

		Attributes: map[string]schema.Attribute{
			"routes": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of gateway routes.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the route.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the route.",
						},
						"path_prefix": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The path pattern to match.",
						},
						"target_url": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The destination URL to proxy to.",
						},
						"methods": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "HTTP methods to match.",
						},
						"strip_prefix": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether to strip the path prefix.",
						},
						"rate_limit": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Maximum requests per second per IP.",
						},
						"priority": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Priority for route matching.",
						},
					},
				},
			},
		},
	}
}

func (d *GatewayRoutesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Data Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *GatewayRoutesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GatewayRoutesDataSourceModel

	routes, err := d.client.ListGatewayRoutes(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list gateway routes, got error: %s", err))
		return
	}

	for _, r := range routes {
		methods, diags := types.ListValueFrom(ctx, types.StringType, r.Methods)
		resp.Diagnostics.Append(diags...)

		data.Routes = append(data.Routes, GatewayRouteDataSourceModel{
			ID:          types.StringValue(r.ID),
			Name:        types.StringValue(r.Name),
			PathPrefix:  types.StringValue(r.PathPrefix),
			TargetURL:   types.StringValue(r.TargetURL),
			StripPrefix: types.BoolValue(r.StripPrefix),
			RateLimit:   types.Int64Value(int64(r.RateLimit)),
			Priority:    types.Int64Value(int64(r.Priority)),
			Methods:     methods,
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
