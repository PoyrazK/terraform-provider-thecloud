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
var _ resource.Resource = &DeploymentResource{}
var _ resource.ResourceWithImportState = &DeploymentResource{}

func NewDeploymentResource() resource.Resource {
	return &DeploymentResource{}
}

// DeploymentResource defines the resource implementation.
type DeploymentResource struct {
	client *client.Client
}

// DeploymentResourceModel describes the resource data model.
type DeploymentResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Image        types.String `tfsdk:"image"`
	Replicas     types.Int64  `tfsdk:"replicas"`
	CurrentCount types.Int64  `tfsdk:"current_count"`
	Ports        types.String `tfsdk:"ports"`
	Status       types.String `tfsdk:"status"`
}

func (r *DeploymentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

func (r *DeploymentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Deployment resource allows you to manage container deployments (replicas).",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the deployment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the deployment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"image": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The container image to deploy (e.g. redis:alpine).",
			},
			"replicas": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Desired number of replicas.",
			},
			"current_count": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Actual number of running replicas.",
			},
			"ports": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Exposed ports (e.g. 80:8080).",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the deployment.",
			},
		},
	}
}

func (r *DeploymentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DeploymentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	deployReq := client.CreateDeploymentRequest{
		Name:     data.Name.ValueString(),
		Image:    data.Image.ValueString(),
		Replicas: int(data.Replicas.ValueInt64()),
		Ports:    data.Ports.ValueString(),
	}

	dep, err := r.client.CreateDeployment(ctx, deployReq)
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to create Deployment, got error: %s", err))
		return
	}

	data.ID = types.StringValue(dep.ID)
	data.Status = types.StringValue(dep.Status)
	data.CurrentCount = types.Int64Value(int64(dep.CurrentCount))

	tflog.Trace(ctx, "created a Deployment resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DeploymentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	dep, err := r.client.GetDeployment(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to read Deployment, got error: %s", err))
		return
	}

	if dep == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(dep.ID)
	data.Name = types.StringValue(dep.Name)
	data.Image = types.StringValue(dep.Image)
	data.Replicas = types.Int64Value(int64(dep.Replicas))
	data.CurrentCount = types.Int64Value(int64(dep.CurrentCount))
	data.Ports = types.StringValue(dep.Ports)
	data.Status = types.StringValue(dep.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state DeploymentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Replicas.Equal(state.Replicas) {
		err := r.client.ScaleDeployment(ctx, plan.ID.ValueString(), int(plan.Replicas.ValueInt64()))
		if err != nil {
			resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to scale Deployment, got error: %s", err))
			return
		}
	}

	// For Image or Ports, the current backend Scale only handles replicas.
	// We might need RequiresReplace for those if Update isn't supported.

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DeploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DeploymentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDeployment(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to delete Deployment, got error: %s", err))
		return
	}
}

func (r *DeploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
