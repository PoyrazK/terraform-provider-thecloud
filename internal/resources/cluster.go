package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/poyrazk/terraform-provider-thecloud/internal/client"
)

// Ensure implementation of interfaces
var _ resource.Resource = &ClusterResource{}
var _ resource.ResourceWithImportState = &ClusterResource{}

func NewClusterResource() resource.Resource {
	return &ClusterResource{}
}

// ClusterResource defines the resource implementation.
type ClusterResource struct {
	client *client.Client
}

// ClusterResourceModel describes the resource data model.
type ClusterResourceModel struct {
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

func (r *ClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (r *ClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Cluster resource allows you to manage managed Kubernetes clusters.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the cluster.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the cluster.",
			},
			"vpc_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the VPC to launch the cluster in.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The Kubernetes version of the cluster.",
			},
			"worker_count": schema.Int64Attribute{
				Optional:            true,
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
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to enable network isolation for the cluster.",
			},
			"ha_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to enable high availability for the control plane.",
			},
			"api_server_lb_address": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The address of the API server load balancer.",
			},
		},
	}
}

func (r *ClusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Data Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *ClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ClusterResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	clusterReq := client.CreateClusterRequest{
		Name:             data.Name.ValueString(),
		VpcID:            data.VpcID.ValueString(),
		Version:          data.Version.ValueString(),
		Workers:          int(data.WorkerCount.ValueInt64()),
		NetworkIsolation: data.NetworkIsolation.ValueBool(),
		HA:               data.HAEnabled.ValueBool(),
	}

	cluster, err := r.client.CreateCluster(ctx, clusterReq)
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to create Cluster, got error: %s", err))
		return
	}

	data.ID = types.StringValue(cluster.ID)
	data.Status = types.StringValue(cluster.Status)
	data.PodCIDR = types.StringValue(cluster.PodCIDR)
	data.ServiceCIDR = types.StringValue(cluster.ServiceCIDR)
	data.APIServerLBAddress = types.StringValue(cluster.APIServerLBAddress)
	if data.Version.IsNull() {
		data.Version = types.StringValue(cluster.Version)
	}
	if data.WorkerCount.IsNull() {
		data.WorkerCount = types.Int64Value(int64(cluster.WorkerCount))
	}

	tflog.Trace(ctx, "created a Cluster resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ClusterResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	cluster, err := r.client.GetCluster(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to read Cluster, got error: %s", err))
		return
	}

	if cluster == nil {
		resp.State.RemoveResource(ctx)
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

func (r *ClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ClusterResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.WorkerCount.Equal(state.WorkerCount) {
		err := r.client.ScaleCluster(ctx, plan.ID.ValueString(), int(plan.WorkerCount.ValueInt64()))
		if err != nil {
			resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to scale Cluster, got error: %s", err))
			return
		}
	}

	if !plan.Version.Equal(state.Version) {
		err := r.client.UpgradeCluster(ctx, plan.ID.ValueString(), plan.Version.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to upgrade Cluster, got error: %s", err))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ClusterResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteCluster(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to delete Cluster, got error: %s", err))
		return
	}
}

func (r *ClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
