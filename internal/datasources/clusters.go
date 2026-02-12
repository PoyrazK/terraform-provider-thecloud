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
var _ datasource.DataSource = &ClustersDataSource{}

func NewClustersDataSource() datasource.DataSource {
	return &ClustersDataSource{}
}

// ClustersDataSource defines the data source implementation.
type ClustersDataSource struct {
	client *client.Client
}

// ClustersDataSourceModel describes the data source data model.
type ClustersDataSourceModel struct {
	Clusters []ClusterDataSourceModel `tfsdk:"clusters"`
}

func (d *ClustersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clusters"
}

func (d *ClustersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Clusters data source allows you to list all available managed Kubernetes clusters.",

		Attributes: map[string]schema.Attribute{
			"clusters": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of clusters.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the cluster.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the cluster.",
						},
						"vpc_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The VPC ID of the cluster.",
						},
						"version": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The Kubernetes version of the cluster.",
						},
						"worker_count": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "The number of worker nodes in the cluster.",
						},
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The status of the cluster.",
						},
						"pod_cidr": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The CIDR block for pods.",
						},
						"service_cidr": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The CIDR block for services.",
						},
						"network_isolation": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether network isolation is enabled.",
						},
						"ha_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether high availability is enabled.",
						},
						"api_server_lb_address": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The address of the API server load balancer.",
						},
					},
				},
			},
		},
	}
}

func (d *ClustersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ClustersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ClustersDataSourceModel

	clusters, err := d.client.ListClusters(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list clusters, got error: %s", err))
		return
	}

	for _, c := range clusters {
		data.Clusters = append(data.Clusters, ClusterDataSourceModel{
			ID:                 types.StringValue(c.ID),
			Name:               types.StringValue(c.Name),
			VpcID:              types.StringValue(c.VpcID),
			Version:            types.StringValue(c.Version),
			WorkerCount:        types.Int64Value(int64(c.WorkerCount)),
			Status:             types.StringValue(c.Status),
			PodCIDR:            types.StringValue(c.PodCIDR),
			ServiceCIDR:        types.StringValue(c.ServiceCIDR),
			NetworkIsolation:   types.BoolValue(c.NetworkIsolation),
			HAEnabled:          types.BoolValue(c.HAEnabled),
			APIServerLBAddress: types.StringValue(c.APIServerLBAddress),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
