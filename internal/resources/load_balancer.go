package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/poyrazk/terraform-provider-thecloud/internal/client"
)

// Ensure implementation of interfaces
var _ resource.Resource = &LoadBalancerResource{}
var _ resource.ResourceWithImportState = &LoadBalancerResource{}

func NewLoadBalancerResource() resource.Resource {
	return &LoadBalancerResource{}
}

// LoadBalancerResource defines the resource implementation.
type LoadBalancerResource struct {
	client *client.Client
}

// LoadBalancerResourceModel describes the resource data model.
type LoadBalancerResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	VpcID     types.String `tfsdk:"vpc_id"`
	Port      types.Int64  `tfsdk:"port"`
	Algorithm types.String `tfsdk:"algorithm"`
	Status    types.String `tfsdk:"status"`
}

func (r *LoadBalancerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer"
}

func (r *LoadBalancerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Load Balancer resource allows you to manage traffic distribution.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the load balancer.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the load balancer.",
			},
			"vpc_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the VPC this load balancer belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"port": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The port the load balancer listens on.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"algorithm": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The load balancing algorithm (round-robin, least-connections).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the load balancer.",
			},
		},
	}
}

func (r *LoadBalancerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LoadBalancerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LoadBalancerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	lb, err := r.client.CreateLoadBalancer(
		data.Name.ValueString(),
		data.VpcID.ValueString(),
		int(data.Port.ValueInt64()),
		data.Algorithm.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to create load balancer, got error: %s", err))
		return
	}

	data.ID = types.StringValue(lb.ID)
	data.Name = types.StringValue(lb.Name)
	data.VpcID = types.StringValue(lb.VpcID)
	data.Port = types.Int64Value(int64(lb.Port))
	data.Algorithm = types.StringValue(lb.Algorithm)
	data.Status = types.StringValue(lb.Status)

	tflog.Trace(ctx, "created a Load Balancer resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LoadBalancerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LoadBalancerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	lb, err := r.client.GetLoadBalancer(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to read load balancer, got error: %s", err))
		return
	}

	if lb == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(lb.ID)
	data.Name = types.StringValue(lb.Name)
	data.VpcID = types.StringValue(lb.VpcID)
	data.Port = types.Int64Value(int64(lb.Port))
	data.Algorithm = types.StringValue(lb.Algorithm)
	data.Status = types.StringValue(lb.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LoadBalancerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddWarning("Update Not Supported", "Updating a load balancer is not currently supported by the API.")
}

func (r *LoadBalancerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LoadBalancerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteLoadBalancer(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to delete load balancer, got error: %s", err))
		return
	}
}

func (r *LoadBalancerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
