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
var _ resource.Resource = &VpcResource{}
var _ resource.ResourceWithImportState = &VpcResource{}

func NewVpcResource() resource.Resource {
	return &VpcResource{}
}

// VpcResource defines the resource implementation.
type VpcResource struct {
	client *client.Client
}

// VpcResourceModel describes the resource data model.
type VpcResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	CIDRBlock types.String `tfsdk:"cidr_block"`
	Status    types.String `tfsdk:"status"`
}

func (r *VpcResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc"
}

func (r *VpcResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "VPC resource allows you to manage virtual private clouds.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the VPC.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the VPC.",
			},
			"cidr_block": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The IPv4 CIDR block for the VPC.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the VPC.",
			},
		},
	}
}

func (r *VpcResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VpcResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VpcResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	vpc, err := r.client.CreateVPC(data.Name.ValueString(), data.CIDRBlock.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to create VPC, got error: %s", err))
		return
	}

	data.ID = types.StringValue(vpc.ID)
	data.Name = types.StringValue(vpc.Name)
	data.CIDRBlock = types.StringValue(vpc.CIDRBlock)
	data.Status = types.StringValue(vpc.Status)

	tflog.Trace(ctx, "created a VPC resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VpcResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VpcResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	vpc, err := r.client.GetVPC(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to read VPC, got error: %s", err))
		return
	}

	if vpc == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(vpc.ID)
	data.Name = types.StringValue(vpc.Name)
	data.CIDRBlock = types.StringValue(vpc.CIDRBlock)
	data.Status = types.StringValue(vpc.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VpcResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// VPC update not supported by API yet, but we'll mark it as No-Op for now or error out
	// Actually, name might be updateable? Let's assume No-Op for now or RequiresReplace in schema.
	resp.Diagnostics.AddWarning("Update Not Supported", "Updating a VPC is not currently supported by the API. This will be a no-op.")
}

func (r *VpcResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VpcResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteVPC(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to delete VPC, got error: %s", err))
		return
	}
}

func (r *VpcResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
