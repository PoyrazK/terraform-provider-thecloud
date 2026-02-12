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
var _ datasource.DataSource = &SubnetsDataSource{}

func NewSubnetsDataSource() datasource.DataSource {
	return &SubnetsDataSource{}
}

// SubnetsDataSource defines the data source implementation.
type SubnetsDataSource struct {
	client *client.Client
}

// SubnetsDataSourceModel describes the data source data model.
type SubnetsDataSourceModel struct {
	VpcID   types.String            `tfsdk:"vpc_id"`
	Subnets []SubnetDataSourceModel `tfsdk:"subnets"`
}

func (d *SubnetsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnets"
}

func (d *SubnetsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Subnets data source allows you to list all available subnets in a VPC.",

		Attributes: map[string]schema.Attribute{
			"vpc_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the VPC to list subnets for.",
			},
			"subnets": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of subnets.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the subnet.",
						},
						"vpc_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The VPC ID of the subnet.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the subnet.",
						},
						"cidr_block": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The IPv4 CIDR block for the subnet.",
						},
						"availability_zone": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The availability zone for the subnet.",
						},
					},
				},
			},
		},
	}
}

func (d *SubnetsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SubnetsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SubnetsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	subnets, err := d.client.ListSubnets(ctx, data.VpcID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list subnets, got error: %s", err))
		return
	}

	for _, s := range subnets {
		var az types.String
		if s.AvailabilityZone != "" {
			az = types.StringValue(s.AvailabilityZone)
		} else {
			az = types.StringNull()
		}

		data.Subnets = append(data.Subnets, SubnetDataSourceModel{
			ID:               types.StringValue(s.ID),
			VpcID:            types.StringValue(s.VPCID),
			Name:             types.StringValue(s.Name),
			CIDRBlock:        types.StringValue(s.CIDRBlock),
			AvailabilityZone: az,
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
