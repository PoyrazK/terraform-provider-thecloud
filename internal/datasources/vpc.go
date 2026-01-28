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
var _ datasource.DataSource = &VpcDataSource{}

func NewVpcDataSource() datasource.DataSource {
	return &VpcDataSource{}
}

// VpcDataSource defines the data source implementation.
type VpcDataSource struct {
	client *client.Client
}

// VpcDataSourceModel describes the data source data model.
type VpcDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	CIDRBlock types.String `tfsdk:"cidr_block"`
	Status    types.String `tfsdk:"status"`
}

func (d *VpcDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc"
}

func (d *VpcDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "VPC data source allows you to look up VPC details by name or ID.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The ID of the VPC to look up.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The name of the VPC to look up.",
			},
			"cidr_block": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The IPv4 CIDR block for the VPC.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the VPC.",
			},
		},
	}
}

func (d *VpcDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VpcDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VpcDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	foundVpc, err := d.lookupVpc(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read VPC, got error: %s", err))
		return
	}

	if foundVpc == nil {
		resp.Diagnostics.AddError("VPC Not Found", "No VPC matching the criteria was found.")
		return
	}

	data.ID = types.StringValue(foundVpc.ID)
	data.Name = types.StringValue(foundVpc.Name)
	data.CIDRBlock = types.StringValue(foundVpc.CIDRBlock)
	data.Status = types.StringValue(foundVpc.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *VpcDataSource) lookupVpc(ctx context.Context, data VpcDataSourceModel) (*client.VPC, error) {
	if !data.ID.IsNull() {
		return d.client.GetVPC(ctx, data.ID.ValueString())
	}

	if !data.Name.IsNull() {
		return d.lookupVpcByName(ctx, data.Name.ValueString())
	}

	return nil, fmt.Errorf("either 'id' or 'name' must be specified")
}

func (d *VpcDataSource) lookupVpcByName(ctx context.Context, name string) (*client.VPC, error) {
	vpcs, err := d.client.ListVPCs(ctx)
	if err != nil {
		return nil, err
	}

	for _, v := range vpcs {
		if v.Name == name {
			return &v, nil
		}
	}

	return nil, nil // nolint:nilnil
}
