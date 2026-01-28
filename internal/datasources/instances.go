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
var _ datasource.DataSource = &InstancesDataSource{}

func NewInstancesDataSource() datasource.DataSource {
	return &InstancesDataSource{}
}

// InstancesDataSource defines the data source implementation.
type InstancesDataSource struct {
	client *client.Client
}

// InstancesDataSourceModel describes the data source data model.
type InstancesDataSourceModel struct {
	Instances []InstanceDataSourceModel `tfsdk:"instances"`
}

func (d *InstancesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instances"
}

func (d *InstancesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Instances data source allows you to list all available compute instances.",

		Attributes: map[string]schema.Attribute{
			"instances": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of instances.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the instance.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the instance.",
						},
						"image": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The image of the instance.",
						},
						"ports": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The port mappings of the instance.",
						},
						"vpc_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The VPC ID of the instance.",
						},
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The status of the instance.",
						},
						"ip_address": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The IP address of the instance.",
						},
					},
				},
			},
		},
	}
}

func (d *InstancesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *InstancesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data InstancesDataSourceModel

	instances, err := d.client.ListInstances(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list instances, got error: %s", err))
		return
	}

	for _, inst := range instances {
		data.Instances = append(data.Instances, InstanceDataSourceModel{
			ID:        types.StringValue(inst.ID),
			Name:      types.StringValue(inst.Name),
			Image:     types.StringValue(inst.Image),
			Ports:     types.StringValue(inst.Ports),
			VpcID:     types.StringValue(inst.VpcID),
			Status:    types.StringValue(inst.Status),
			IPAddress: types.StringValue(inst.IPAddress),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
