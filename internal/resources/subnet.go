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
var _ resource.Resource = &SubnetResource{}
var _ resource.ResourceWithImportState = &SubnetResource{}

func NewSubnetResource() resource.Resource {
	return &SubnetResource{}
}

// SubnetResource defines the resource implementation.
type SubnetResource struct {
	client *client.Client
}

// SubnetResourceModel describes the resource data model.
type SubnetResourceModel struct {
	ID               types.String `tfsdk:"id"`
	VpcID            types.String `tfsdk:"vpc_id"`
	Name             types.String `tfsdk:"name"`
	CIDRBlock        types.String `tfsdk:"cidr_block"`
	AvailabilityZone types.String `tfsdk:"availability_zone"`
}

func (r *SubnetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (r *SubnetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Subnet resource allows you to manage network subnets within a VPC.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the subnet.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"vpc_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the VPC this subnet belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the subnet.",
			},
			"cidr_block": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The CIDR block for the subnet.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"availability_zone": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The availability zone for the subnet.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *SubnetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SubnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SubnetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	subnet, err := r.client.CreateSubnet(
		ctx,
		data.VpcID.ValueString(),
		data.Name.ValueString(),
		data.CIDRBlock.ValueString(),
		data.AvailabilityZone.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to create subnet, got error: %s", err))
		return
	}

	data.ID = types.StringValue(subnet.ID)
	data.VpcID = types.StringValue(subnet.VPCID)
	data.Name = types.StringValue(subnet.Name)
	data.CIDRBlock = types.StringValue(subnet.CIDRBlock)
	if !data.AvailabilityZone.IsNull() || subnet.AvailabilityZone != "" {
		data.AvailabilityZone = types.StringValue(subnet.AvailabilityZone)
	} else {
		data.AvailabilityZone = types.StringNull()
	}

	tflog.Trace(ctx, "created a Subnet resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SubnetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	subnet, err := r.client.GetSubnet(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to read subnet, got error: %s", err))
		return
	}

	if subnet == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(subnet.ID)
	data.VpcID = types.StringValue(subnet.VPCID)
	data.Name = types.StringValue(subnet.Name)
	data.CIDRBlock = types.StringValue(subnet.CIDRBlock)
	if !data.AvailabilityZone.IsNull() || subnet.AvailabilityZone != "" {
		data.AvailabilityZone = types.StringValue(subnet.AvailabilityZone)
	} else {
		data.AvailabilityZone = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddWarning("Update Not Supported", "Updating a subnet is not currently supported by the API.")
}

func (r *SubnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SubnetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSubnet(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to delete subnet, got error: %s", err))
		return
	}
}

func (r *SubnetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
