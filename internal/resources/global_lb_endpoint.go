package resources

import (
	"context"
	"fmt"
	"strings"

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
var _ resource.Resource = &GlobalLBEndpointResource{}
var _ resource.ResourceWithImportState = &GlobalLBEndpointResource{}

func NewGlobalLBEndpointResource() resource.Resource {
	return &GlobalLBEndpointResource{}
}

// GlobalLBEndpointResource defines the resource implementation.
type GlobalLBEndpointResource struct {
	client *client.Client
}

// GlobalLBEndpointResourceModel describes the resource data model.
type GlobalLBEndpointResourceModel struct {
	ID         types.String `tfsdk:"id"`
	GlobalLBID types.String `tfsdk:"global_lb_id"`
	Region     types.String `tfsdk:"region"`
	TargetType types.String `tfsdk:"target_type"`
	TargetID   types.String `tfsdk:"target_id"`
	TargetIP   types.String `tfsdk:"target_ip"`
	Weight     types.Int64  `tfsdk:"weight"`
	Priority   types.Int64  `tfsdk:"priority"`
	Healthy    types.Bool   `tfsdk:"healthy"`
}

func (r *GlobalLBEndpointResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_global_lb_endpoint"
}

func (r *GlobalLBEndpointResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Global Load Balancer Endpoint resource allows you to add targets to a GLB.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the endpoint.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"global_lb_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the Global LB.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The region of the target.",
			},
			"target_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Target type (LB or IP).",
			},
			"target_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the regional Load Balancer.",
			},
			"target_ip": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Static IP if target_type is IP.",
			},
			"weight": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Traffic weight (1-100).",
			},
			"priority": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Failover priority.",
			},
			"healthy": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Health status of the endpoint.",
			},
		},
	}
}

func (r *GlobalLBEndpointResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GlobalLBEndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GlobalLBEndpointResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	epReq := client.AddGlobalEndpointRequest{
		Region:     data.Region.ValueString(),
		TargetType: data.TargetType.ValueString(),
		TargetID:   data.TargetID.ValueString(),
		TargetIP:   data.TargetIP.ValueString(),
		Weight:     int(data.Weight.ValueInt64()),
		Priority:   int(data.Priority.ValueInt64()),
	}

	ep, err := r.client.AddGlobalEndpoint(ctx, data.GlobalLBID.ValueString(), epReq)
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to add GLB Endpoint, got error: %s", err))
		return
	}

	data.ID = types.StringValue(ep.ID)
	data.Healthy = types.BoolValue(ep.Healthy)
	if data.Weight.IsNull() {
		data.Weight = types.Int64Value(int64(ep.Weight))
	}
	if data.Priority.IsNull() {
		data.Priority = types.Int64Value(int64(ep.Priority))
	}

	tflog.Trace(ctx, "added a Global LB Endpoint")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GlobalLBEndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GlobalLBEndpointResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	glb, err := r.client.GetGlobalLB(ctx, data.GlobalLBID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to read Global LB for endpoint, got error: %s", err))
		return
	}

	if glb == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	var found *client.GlobalEndpoint
	for _, ep := range glb.Endpoints {
		if ep.ID == data.ID.ValueString() {
			found = &ep
			break
		}
	}

	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.Region = types.StringValue(found.Region)
	data.TargetType = types.StringValue(found.TargetType)
	data.TargetID = types.StringValue(found.TargetID)
	data.TargetIP = types.StringValue(found.TargetIP)
	data.Weight = types.Int64Value(int64(found.Weight))
	data.Priority = types.Int64Value(int64(found.Priority))
	data.Healthy = types.BoolValue(found.Healthy)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GlobalLBEndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Not supported, use RequiresReplace or API update if available
}

func (r *GlobalLBEndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GlobalLBEndpointResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RemoveGlobalEndpoint(ctx, data.GlobalLBID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to remove Global LB Endpoint, got error: %s", err))
		return
	}
}

func (r *GlobalLBEndpointResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Custom import because we need both GLB ID and Endpoint ID
	idParts := strings.Split(req.ID, ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: global_lb_id:endpoint_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("global_lb_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}
