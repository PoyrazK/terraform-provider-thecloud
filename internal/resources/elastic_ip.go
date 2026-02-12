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
var _ resource.Resource = &ElasticIPResource{}
var _ resource.ResourceWithImportState = &ElasticIPResource{}

func NewElasticIPResource() resource.Resource {
	return &ElasticIPResource{}
}

// ElasticIPResource defines the resource implementation.
type ElasticIPResource struct {
	client *client.Client
}

// ElasticIPResourceModel describes the resource data model.
type ElasticIPResourceModel struct {
	ID         types.String `tfsdk:"id"`
	PublicIP   types.String `tfsdk:"public_ip"`
	InstanceID types.String `tfsdk:"instance_id"`
	Status     types.String `tfsdk:"status"`
}

func (r *ElasticIPResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elastic_ip"
}

func (r *ElasticIPResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Elastic IP resource allows you to allocate and manage static public IP addresses.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the Elastic IP.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"public_ip": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The allocated public IP address.",
			},
			"instance_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the instance this IP is associated with.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the Elastic IP.",
			},
		},
	}
}

func (r *ElasticIPResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ElasticIPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ElasticIPResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	eip, err := r.client.AllocateElasticIP(ctx)
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to allocate Elastic IP, got error: %s", err))
		return
	}

	data.ID = types.StringValue(eip.ID)
	data.PublicIP = types.StringValue(eip.PublicIP)
	data.Status = types.StringValue(eip.Status)
	if eip.InstanceID != "" {
		data.InstanceID = types.StringValue(eip.InstanceID)
	} else {
		data.InstanceID = types.StringNull()
	}

	tflog.Trace(ctx, "allocated an Elastic IP resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticIPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ElasticIPResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	eip, err := r.client.GetElasticIP(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to read Elastic IP, got error: %s", err))
		return
	}

	if eip == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(eip.ID)
	data.PublicIP = types.StringValue(eip.PublicIP)
	data.Status = types.StringValue(eip.Status)
	if eip.InstanceID != "" {
		data.InstanceID = types.StringValue(eip.InstanceID)
	} else {
		data.InstanceID = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticIPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Elastic IP itself doesn't have many updatable fields. Association is a separate resource or action.
}

func (r *ElasticIPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ElasticIPResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.ReleaseElasticIP(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to release Elastic IP, got error: %s", err))
		return
	}
}

func (r *ElasticIPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
