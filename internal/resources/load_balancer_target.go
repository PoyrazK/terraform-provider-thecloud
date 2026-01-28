package resources

import (
	"context"
	"fmt"
	"strings"

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
var _ resource.Resource = &LoadBalancerTargetResource{}

func NewLoadBalancerTargetResource() resource.Resource {
	return &LoadBalancerTargetResource{}
}

// LoadBalancerTargetResource defines the resource implementation.
type LoadBalancerTargetResource struct {
	client *client.Client
}

// LoadBalancerTargetResourceModel describes the resource data model.
type LoadBalancerTargetResourceModel struct {
	ID             types.String `tfsdk:"id"` // Format: {lb_id}:{instance_id}
	LoadBalancerID types.String `tfsdk:"load_balancer_id"`
	InstanceID     types.String `tfsdk:"instance_id"`
	Port           types.Int64  `tfsdk:"port"`
	Weight         types.Int64  `tfsdk:"weight"`
}

func (r *LoadBalancerTargetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_target"
}

func (r *LoadBalancerTargetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Load Balancer Target resource allows you to register instances to a load balancer.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The composite ID of the target (lb_id:instance_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"load_balancer_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the load balancer.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"instance_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the instance to register.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"port": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The port on the instance to send traffic to.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"weight": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The weight of this target.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *LoadBalancerTargetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LoadBalancerTargetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LoadBalancerTargetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	target := client.LBTarget{
		InstanceID: data.InstanceID.ValueString(),
		Port:       int(data.Port.ValueInt64()),
		Weight:     int(data.Weight.ValueInt64()),
	}

	err := r.client.AddLBTarget(ctx, data.LoadBalancerID.ValueString(), target)
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to add target to load balancer, got error: %s", err))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s:%s", data.LoadBalancerID.ValueString(), data.InstanceID.ValueString()))

	tflog.Trace(ctx, "added a Load Balancer Target resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LoadBalancerTargetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LoadBalancerTargetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	targets, err := r.client.ListLBTargets(ctx, data.LoadBalancerID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to read load balancer targets, got error: %s", err))
		return
	}

	found := false
	for _, t := range targets {
		if t.InstanceID == data.InstanceID.ValueString() {
			data.Port = types.Int64Value(int64(t.Port))
			data.Weight = types.Int64Value(int64(t.Weight))
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LoadBalancerTargetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddWarning("Update Not Supported", "Updating a load balancer target is not currently supported by the API.")
}

func (r *LoadBalancerTargetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LoadBalancerTargetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RemoveLBTarget(ctx, data.LoadBalancerID.ValueString(), data.InstanceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to remove target from load balancer, got error: %s", err))
		return
	}
}

func (r *LoadBalancerTargetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import requires load_balancer_id:instance_id
	idParts := strings.Split(req.ID, ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: load_balancer_id:instance_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("load_balancer_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("instance_id"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
