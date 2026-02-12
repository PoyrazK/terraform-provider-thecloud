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
var _ datasource.DataSource = &ClusterDataSource{}

func NewClusterDataSource() datasource.DataSource {
	return &ClusterDataSource{}
}

// ClusterDataSource defines the data source implementation.
type ClusterDataSource struct {
	client *client.Client
}

// ClusterDataSourceModel describes the data source data model.
type ClusterDataSourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	VpcID              types.String `tfsdk:"vpc_id"`
	Version            types.String `tfsdk:"version"`
	WorkerCount        types.Int64  `tfsdk:"worker_count"`
	Status             types.String `tfsdk:"status"`
	PodCIDR            types.String `tfsdk:"pod_cidr"`
	ServiceCIDR        types.String `tfsdk:"service_cidr"`
	NetworkIsolation   types.Bool   `tfsdk:"network_isolation"`
	HAEnabled          types.Bool   `tfsdk:"ha_enabled"`
	APIServerLBAddress types.String `tfsdk:"api_server_lb_address"`
}

func (d *ClusterDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (d *ClusterDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Cluster data source allows you to look up cluster details by ID or Name.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The ID of the cluster to look up.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The name of the cluster to look up.",
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
	}
}

func (d *ClusterDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ClusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ClusterDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var cluster *client.Cluster
	var err error

	if !data.ID.IsNull() {
		cluster, err = d.client.GetCluster(ctx, data.ID.ValueString())
	} else if !data.Name.IsNull() {
		cluster, err = d.lookupClusterByName(ctx, data.Name.ValueString())
	} else {
		resp.Diagnostics.AddError("Missing Required Attribute", "Either id or name must be specified.")
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read cluster, got error: %s", err))
		return
	}

	if cluster == nil {
		resp.Diagnostics.AddError("Cluster Not Found", "No cluster matching the criteria was found.")
		return
	}

	data.ID = types.StringValue(cluster.ID)
	data.Name = types.StringValue(cluster.Name)
	data.VpcID = types.StringValue(cluster.VpcID)
	data.Version = types.StringValue(cluster.Version)
	data.WorkerCount = types.Int64Value(int64(cluster.WorkerCount))
	data.Status = types.StringValue(cluster.Status)
	data.PodCIDR = types.StringValue(cluster.PodCIDR)
	data.ServiceCIDR = types.StringValue(cluster.ServiceCIDR)
	data.NetworkIsolation = types.BoolValue(cluster.NetworkIsolation)
	data.HAEnabled = types.BoolValue(cluster.HAEnabled)
	data.APIServerLBAddress = types.StringValue(cluster.APIServerLBAddress)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *ClusterDataSource) lookupClusterByName(ctx context.Context, name string) (*client.Cluster, error) {
	clusters, err := d.client.ListClusters(ctx)
	if err != nil {
		return nil, err
	}

	for _, c := range clusters {
		if c.Name == name {
			return &c, nil
		}
	}

	return nil, nil
}
