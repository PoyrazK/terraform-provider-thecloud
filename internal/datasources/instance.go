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
var _ datasource.DataSource = &InstanceDataSource{}

func NewInstanceDataSource() datasource.DataSource {
	return &InstanceDataSource{}
}

// InstanceDataSource defines the data source implementation.
type InstanceDataSource struct {
	client *client.Client
}

// InstanceDataSourceModel describes the data source data model.
type InstanceDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Image     types.String `tfsdk:"image"`
	Ports     types.String `tfsdk:"ports"`
	VpcID     types.String `tfsdk:"vpc_id"`
	Status    types.String `tfsdk:"status"`
	IPAddress types.String `tfsdk:"ip_address"`
}

func (d *InstanceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

func (d *InstanceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Instance data source allows you to look up instance details by ID.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the instance to look up.",
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
	}
}

func (d *InstanceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *InstanceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data InstanceDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	instance, err := d.client.GetInstance(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read instance, got error: %s", err))
		return
	}

	if instance == nil {
		resp.Diagnostics.AddError("Instance Not Found", "No instance matching the ID was found.")
		return
	}

	data.Name = types.StringValue(instance.Name)
	data.Image = types.StringValue(instance.Image)
	data.Ports = types.StringValue(instance.Ports)
	data.VpcID = types.StringValue(instance.VpcID)
	data.Status = types.StringValue(instance.Status)
	data.IPAddress = types.StringValue(instance.IPAddress)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
