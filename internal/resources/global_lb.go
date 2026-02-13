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
var _ resource.Resource = &GlobalLBResource{}
var _ resource.ResourceWithImportState = &GlobalLBResource{}

func NewGlobalLBResource() resource.Resource {
	return &GlobalLBResource{}
}

// GlobalLBResource defines the resource implementation.
type GlobalLBResource struct {
	client *client.Client
}

// GlobalLBResourceModel describes the resource data model.
type GlobalLBResourceModel struct {
	ID          types.String           `tfsdk:"id"`
	Name        types.String           `tfsdk:"name"`
	Hostname    types.String           `tfsdk:"hostname"`
	Policy      types.String           `tfsdk:"policy"`
	Status      types.String           `tfsdk:"status"`
	HealthCheck GlobalHealthCheckModel `tfsdk:"health_check"`
}

type GlobalHealthCheckModel struct {
	Protocol       types.String `tfsdk:"protocol"`
	Port           types.Int64  `tfsdk:"port"`
	Path           types.String `tfsdk:"path"`
	IntervalSec    types.Int64  `tfsdk:"interval_sec"`
	TimeoutSec     types.Int64  `tfsdk:"timeout_sec"`
	HealthyCount   types.Int64  `tfsdk:"healthy_count"`
	UnhealthyCount types.Int64  `tfsdk:"unhealthy_count"`
}

func (r *GlobalLBResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_global_lb"
}

func (r *GlobalLBResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Global Load Balancer resource allows you to manage multi-region load balancing.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the GLB.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the GLB.",
			},
			"hostname": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The hostname for the GLB.",
			},
			"policy": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The routing policy (LATENCY, GEOLOCATION, WEIGHTED, FAILOVER).",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the GLB.",
			},
			"health_check": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"protocol": schema.StringAttribute{
						Required: true,
					},
					"port": schema.Int64Attribute{
						Required: true,
					},
					"path": schema.StringAttribute{
						Optional: true,
					},
					"interval_sec": schema.Int64Attribute{
						Required: true,
					},
					"timeout_sec": schema.Int64Attribute{
						Required: true,
					},
					"healthy_count": schema.Int64Attribute{
						Required: true,
					},
					"unhealthy_count": schema.Int64Attribute{
						Required: true,
					},
				},
			},
		},
	}
}

func (r *GlobalLBResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GlobalLBResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GlobalLBResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	glbReq := client.CreateGlobalLBRequest{
		Name:     data.Name.ValueString(),
		Hostname: data.Hostname.ValueString(),
		Policy:   data.Policy.ValueString(),
		HealthCheck: client.GlobalHealthCheck{
			Protocol:       data.HealthCheck.Protocol.ValueString(),
			Port:           int(data.HealthCheck.Port.ValueInt64()),
			Path:           data.HealthCheck.Path.ValueString(),
			IntervalSec:    int(data.HealthCheck.IntervalSec.ValueInt64()),
			TimeoutSec:     int(data.HealthCheck.TimeoutSec.ValueInt64()),
			HealthyCount:   int(data.HealthCheck.HealthyCount.ValueInt64()),
			UnhealthyCount: int(data.HealthCheck.UnhealthyCount.ValueInt64()),
		},
	}

	glb, err := r.client.CreateGlobalLB(ctx, glbReq)
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to create Global LB, got error: %s", err))
		return
	}

	data.ID = types.StringValue(glb.ID)
	data.Status = types.StringValue(glb.Status)

	tflog.Trace(ctx, "created a Global LB resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GlobalLBResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GlobalLBResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	glb, err := r.client.GetGlobalLB(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to read Global LB, got error: %s", err))
		return
	}

	if glb == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(glb.ID)
	data.Name = types.StringValue(glb.Name)
	data.Hostname = types.StringValue(glb.Hostname)
	data.Policy = types.StringValue(glb.Policy)
	data.Status = types.StringValue(glb.Status)
	data.HealthCheck = GlobalHealthCheckModel{
		Protocol:       types.StringValue(glb.HealthCheck.Protocol),
		Port:           types.Int64Value(int64(glb.HealthCheck.Port)),
		Path:           types.StringValue(glb.HealthCheck.Path),
		IntervalSec:    types.Int64Value(int64(glb.HealthCheck.IntervalSec)),
		TimeoutSec:     types.Int64Value(int64(glb.HealthCheck.TimeoutSec)),
		HealthyCount:   types.Int64Value(int64(glb.HealthCheck.HealthyCount)),
		UnhealthyCount: types.Int64Value(int64(glb.HealthCheck.UnhealthyCount)),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GlobalLBResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Not fully supported by API yet
}

func (r *GlobalLBResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GlobalLBResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteGlobalLB(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to delete Global LB, got error: %s", err))
		return
	}
}

func (r *GlobalLBResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
