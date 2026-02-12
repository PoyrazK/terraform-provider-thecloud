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
var _ resource.Resource = &ElasticIPAssociationResource{}
var _ resource.ResourceWithImportState = &ElasticIPAssociationResource{}

func NewElasticIPAssociationResource() resource.Resource {
	return &ElasticIPAssociationResource{}
}

// ElasticIPAssociationResource defines the resource implementation.
type ElasticIPAssociationResource struct {
	client *client.Client
}

// ElasticIPAssociationResourceModel describes the resource data model.
type ElasticIPAssociationResourceModel struct {
	ID         types.String `tfsdk:"id"`
	EipID      types.String `tfsdk:"eip_id"`
	InstanceID types.String `tfsdk:"instance_id"`
}

func (r *ElasticIPAssociationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elastic_ip_association"
}

func (r *ElasticIPAssociationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Elastic IP Association resource allows you to associate an Elastic IP with an instance.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the association (matches eip_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"eip_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the Elastic IP to associate.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"instance_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the instance to associate with.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *ElasticIPAssociationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ElasticIPAssociationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ElasticIPAssociationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	eip, err := r.client.AssociateElasticIP(ctx, data.EipID.ValueString(), data.InstanceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to associate Elastic IP, got error: %s", err))
		return
	}

	data.ID = types.StringValue(eip.ID)

	tflog.Trace(ctx, "associated an Elastic IP resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticIPAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ElasticIPAssociationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	eip, err := r.client.GetElasticIP(ctx, data.EipID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to read Elastic IP for association, got error: %s", err))
		return
	}

	if eip == nil || eip.InstanceID == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(eip.ID)
	data.EipID = types.StringValue(eip.ID)
	data.InstanceID = types.StringValue(eip.InstanceID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ElasticIPAssociationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Not supported, handled by RequiresReplace
}

func (r *ElasticIPAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ElasticIPAssociationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DisassociateElasticIP(ctx, data.EipID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to disassociate Elastic IP, got error: %s", err))
		return
	}
}

func (r *ElasticIPAssociationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
