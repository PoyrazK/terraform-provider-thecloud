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
var _ datasource.DataSource = &VpcsDataSource{}

func NewVpcsDataSource() datasource.DataSource {
	return &VpcsDataSource{}
}

// VpcsDataSource defines the data source implementation.
type VpcsDataSource struct {
	client *client.Client
}

// VpcsDataSourceModel describes the data source data model.
type VpcsDataSourceModel struct {
	Vpcs []VpcDataSourceModel `tfsdk:"vpcs"`
}

func (d *VpcsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpcs"
}

func (d *VpcsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "VPCs data source allows you to list all available VPCs.",

		Attributes: map[string]schema.Attribute{
			"vpcs": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of VPCs.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the VPC.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the VPC.",
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
				},
			},
		},
	}
}

func (d *VpcsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VpcsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VpcsDataSourceModel

	vpcs, err := d.client.ListVPCs(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list VPCs, got error: %s", err))
		return
	}

	for _, v := range vpcs {
		data.Vpcs = append(data.Vpcs, VpcDataSourceModel{
			ID:        types.StringValue(v.ID),
			Name:      types.StringValue(v.Name),
			CIDRBlock: types.StringValue(v.CIDRBlock),
			Status:    types.StringValue(v.Status),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
