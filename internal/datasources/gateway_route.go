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
var _ datasource.DataSource = &GatewayRouteDataSource{}

func NewGatewayRouteDataSource() datasource.DataSource {
	return &GatewayRouteDataSource{}
}

// GatewayRouteDataSource defines the data source implementation.
type GatewayRouteDataSource struct {
	client *client.Client
}

// GatewayRouteDataSourceModel describes the data source data model.
type GatewayRouteDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	PathPrefix  types.String `tfsdk:"path_prefix"`
	TargetURL   types.String `tfsdk:"target_url"`
	Methods     types.List   `tfsdk:"methods"`
	StripPrefix types.Bool   `tfsdk:"strip_prefix"`
	RateLimit   types.Int64  `tfsdk:"rate_limit"`
	Priority    types.Int64  `tfsdk:"priority"`
}

func (d *GatewayRouteDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gateway_route"
}

func (d *GatewayRouteDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gateway Route data source allows you to look up gateway route details by ID or Name.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The ID of the route to look up.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The name of the route to look up.",
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
	}
}

func (d *GatewayRouteDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GatewayRouteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GatewayRouteDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var found *client.GatewayRoute
	var err error

	if !data.ID.IsNull() {
		found, err = d.client.GetGatewayRoute(ctx, data.ID.ValueString())
	} else if !data.Name.IsNull() {
		found, err = d.lookupRouteByName(ctx, data.Name.ValueString())
	} else {
		resp.Diagnostics.AddError("Missing Required Attribute", "Either id or name must be specified.")
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read gateway route, got error: %s", err))
		return
	}

	if found == nil {
		resp.Diagnostics.AddError("Gateway Route Not Found", "No route matching the criteria was found.")
		return
	}

	data.ID = types.StringValue(found.ID)
	data.Name = types.StringValue(found.Name)
	data.PathPrefix = types.StringValue(found.PathPrefix)
	data.TargetURL = types.StringValue(found.TargetURL)
	data.StripPrefix = types.BoolValue(found.StripPrefix)
	data.RateLimit = types.Int64Value(int64(found.RateLimit))
	data.Priority = types.Int64Value(int64(found.Priority))

	methods, diags := types.ListValueFrom(ctx, types.StringType, found.Methods)
	resp.Diagnostics.Append(diags...)
	data.Methods = methods

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *GatewayRouteDataSource) lookupRouteByName(ctx context.Context, name string) (*client.GatewayRoute, error) {
	routes, err := d.client.ListGatewayRoutes(ctx)
	if err != nil {
		return nil, err
	}

	for _, r := range routes {
		if r.Name == name {
			return &r, nil
		}
	}

	return nil, nil
}
