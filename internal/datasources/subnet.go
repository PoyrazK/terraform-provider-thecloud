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
var _ datasource.DataSource = &SubnetDataSource{}

func NewSubnetDataSource() datasource.DataSource {
	return &SubnetDataSource{}
}

// SubnetDataSource defines the data source implementation.
type SubnetDataSource struct {
	client *client.Client
}

// SubnetDataSourceModel describes the data source data model.
type SubnetDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	VpcID            types.String `tfsdk:"vpc_id"`
	Name             types.String `tfsdk:"name"`
	CIDRBlock        types.String `tfsdk:"cidr_block"`
	AvailabilityZone types.String `tfsdk:"availability_zone"`
}

func (d *SubnetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (d *SubnetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Subnet data source allows you to look up subnet details by ID or by name within a VPC.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The ID of the subnet to look up.",
			},
			"vpc_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The ID of the VPC the subnet belongs to. Required when looking up by name.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The name of the subnet to look up.",
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
	}
}

func (d *SubnetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SubnetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SubnetDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var subnet *client.Subnet
	var err error

	if !data.ID.IsNull() {
		subnet, err = d.client.GetSubnet(ctx, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read subnet, got error: %s", err))
			return
		}
	} else if !data.Name.IsNull() {
		if data.VpcID.IsNull() {
			resp.Diagnostics.AddError("Missing Required Attribute", "vpc_id is required when looking up subnet by name.")
			return
		}
		subnet, err = d.lookupSubnetByName(ctx, data.VpcID.ValueString(), data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to lookup subnet, got error: %s", err))
			return
		}
	} else {
		resp.Diagnostics.AddError("Missing Required Attribute", "Either id or name (with vpc_id) must be specified.")
		return
	}

	if subnet == nil {
		resp.Diagnostics.AddError("Subnet Not Found", "No subnet matching the criteria was found.")
		return
	}

	data.ID = types.StringValue(subnet.ID)
	data.VpcID = types.StringValue(subnet.VPCID)
	data.Name = types.StringValue(subnet.Name)
	data.CIDRBlock = types.StringValue(subnet.CIDRBlock)
	if subnet.AvailabilityZone != "" {
		data.AvailabilityZone = types.StringValue(subnet.AvailabilityZone)
	} else {
		data.AvailabilityZone = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *SubnetDataSource) lookupSubnetByName(ctx context.Context, vpcID, name string) (*client.Subnet, error) {
	subnets, err := d.client.ListSubnets(ctx, vpcID)
	if err != nil {
		return nil, err
	}

	for _, s := range subnets {
		if s.Name == name {
			return &s, nil
		}
	}

	return nil, nil
}
