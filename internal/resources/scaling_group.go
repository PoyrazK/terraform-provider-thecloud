package resources

import (
	"context"
	"fmt"
	"time"

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
var _ resource.Resource = &ScalingGroupResource{}
var _ resource.ResourceWithImportState = &ScalingGroupResource{}

func NewScalingGroupResource() resource.Resource {
	return &ScalingGroupResource{}
}

// ScalingGroupResource defines the resource implementation.
type ScalingGroupResource struct {
	client *client.Client
}

// ScalingGroupResourceModel describes the resource data model.
type ScalingGroupResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	VpcID          types.String `tfsdk:"vpc_id"`
	LoadBalancerID types.String `tfsdk:"load_balancer_id"`
	Image          types.String `tfsdk:"image"`
	Ports          types.String `tfsdk:"ports"`
	MinInstances   types.Int64  `tfsdk:"min_instances"`
	MaxInstances   types.Int64  `tfsdk:"max_instances"`
	DesiredCount   types.Int64  `tfsdk:"desired_count"`
	Status         types.String `tfsdk:"status"`
}

func (r *ScalingGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scaling_group"
}

func (r *ScalingGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Scaling Group resource allows you to manage auto-scaling instance groups.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the scaling group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the scaling group.",
			},
			"vpc_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the VPC this scaling group belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"load_balancer_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the load balancer to associate with this group.",
			},
			"image": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The image to use for instances in the group.",
			},
			"ports": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The port mappings for instances in the group.",
			},
			"min_instances": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The minimum number of instances in the group.",
			},
			"max_instances": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The maximum number of instances in the group.",
			},
			"desired_count": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The desired number of instances in the group.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the scaling group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ScalingGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ScalingGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ScalingGroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := map[string]interface{}{
		"name":          data.Name.ValueString(),
		"vpc_id":        data.VpcID.ValueString(),
		"image":         data.Image.ValueString(),
		"ports":         data.Ports.ValueString(),
		"min_instances": int(data.MinInstances.ValueInt64()),
		"max_instances": int(data.MaxInstances.ValueInt64()),
		"desired_count": int(data.DesiredCount.ValueInt64()),
	}

	if !data.LoadBalancerID.IsNull() {
		params["load_balancer_id"] = data.LoadBalancerID.ValueString()
	}

	group, err := r.client.CreateScalingGroup(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to create scaling group, got error: %s", err))
		return
	}

	data.ID = types.StringValue(group.ID)
	if !data.Ports.IsNull() || group.Ports != "" {
		data.Ports = types.StringValue(group.Ports)
	} else {
		data.Ports = types.StringNull()
	}
	if !data.LoadBalancerID.IsNull() || group.LoadBalancerID != "" {
		data.LoadBalancerID = types.StringValue(group.LoadBalancerID)
	} else {
		data.LoadBalancerID = types.StringNull()
	}
	data.Status = types.StringValue(group.Status)

	tflog.Trace(ctx, "created a Scaling Group resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ScalingGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ScalingGroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.GetScalingGroup(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to read scaling group, got error: %s", err))
		return
	}

	if group == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(group.ID)
	data.Name = types.StringValue(group.Name)
	data.VpcID = types.StringValue(group.VpcID)
	if !data.LoadBalancerID.IsNull() || group.LoadBalancerID != "" {
		data.LoadBalancerID = types.StringValue(group.LoadBalancerID)
	} else {
		data.LoadBalancerID = types.StringNull()
	}
	data.Image = types.StringValue(group.Image)
	if !data.Ports.IsNull() || group.Ports != "" {
		data.Ports = types.StringValue(group.Ports)
	} else {
		data.Ports = types.StringNull()
	}
	data.MinInstances = types.Int64Value(int64(group.MinInstances))
	data.MaxInstances = types.Int64Value(int64(group.MaxInstances))
	data.DesiredCount = types.Int64Value(int64(group.DesiredCount))
	data.Status = types.StringValue(group.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ScalingGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddWarning("Update Not Supported", "Updating a scaling group is not supported. It will be recreated if changed.")
}

func (r *ScalingGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ScalingGroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteScalingGroup(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to delete scaling group, got error: %s", err))
		return
	}

	// Wait for group to be gone from API (async deletion in backend)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			resp.Diagnostics.AddError("Delete Timeout", "Timed out waiting for scaling group to be deleted.")
			return
		case <-ticker.C:
			group, err := r.client.GetScalingGroup(ctx, data.ID.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(errClient, fmt.Sprintf("Error checking scaling group status: %s", err))
				return
			}
			if group == nil {
				tflog.Trace(ctx, "scaling group successfully deleted")
				return
			}
		}
	}
}

func (r *ScalingGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
