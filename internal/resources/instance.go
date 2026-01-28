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
var _ resource.Resource = &InstanceResource{}
var _ resource.ResourceWithImportState = &InstanceResource{}

func NewInstanceResource() resource.Resource {
	return &InstanceResource{}
}

// InstanceResource defines the resource implementation.
type InstanceResource struct {
	client *client.Client
}

// InstanceResourceModel describes the resource data model.
type InstanceResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Image     types.String `tfsdk:"image"`
	Ports     types.String `tfsdk:"ports"`
	VpcID     types.String `tfsdk:"vpc_id"`
	Status    types.String `tfsdk:"status"`
	IPAddress types.String `tfsdk:"ip_address"`
}

func (r *InstanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

func (r *InstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Instance resource allows you to manage compute instances.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the instance.",
			},
			"image": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The image to use for the instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ports": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The port mappings for the instance (e.g. '80:80,443:443').",
			},
			"vpc_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the VPC to launch the instance in.",
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

func (r *InstanceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *InstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data InstanceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.LaunchInstanceRequest{
		Name:  data.Name.ValueString(),
		Image: data.Image.ValueString(),
		Ports: data.Ports.ValueString(),
		VpcID: data.VpcID.ValueString(),
	}

	instance, err := r.client.CreateInstance(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to create instance, got error: %s", err))
		return
	}

	data.ID = types.StringValue(instance.ID)
	data.Name = types.StringValue(instance.Name)
	data.Image = types.StringValue(instance.Image)
	data.Ports = types.StringValue(instance.Ports)
	data.VpcID = types.StringValue(instance.VpcID)
	data.Status = types.StringValue(instance.Status)
	data.IPAddress = types.StringValue(instance.IPAddress)

	tflog.Trace(ctx, "created an Instance resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data InstanceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	instance, err := r.client.GetInstance(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to read instance, got error: %s", err))
		return
	}

	if instance == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(instance.ID)
	data.Name = types.StringValue(instance.Name)
	data.Image = types.StringValue(instance.Image)
	data.Ports = types.StringValue(instance.Ports)
	data.VpcID = types.StringValue(instance.VpcID)
	data.Status = types.StringValue(instance.Status)
	data.IPAddress = types.StringValue(instance.IPAddress)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Instance update not supported yet, but we'll mark it as No-Op for now
	resp.Diagnostics.AddWarning("Update Not Supported", "Updating an instance is not currently supported by the API.")
}

func (r *InstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data InstanceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteInstance(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to delete instance, got error: %s", err))
		return
	}
}

func (r *InstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
