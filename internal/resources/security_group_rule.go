package resources

import (
	"context"
	"fmt"

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
var _ resource.Resource = &SecurityGroupRuleResource{}
var _ resource.ResourceWithImportState = &SecurityGroupRuleResource{}

func NewSecurityGroupRuleResource() resource.Resource {
	return &SecurityGroupRuleResource{}
}

// SecurityGroupRuleResource defines the resource implementation.
type SecurityGroupRuleResource struct {
	client *client.Client
}

// SecurityGroupRuleResourceModel describes the resource data model.
type SecurityGroupRuleResourceModel struct {
	ID              types.String `tfsdk:"id"`
	SecurityGroupID types.String `tfsdk:"security_group_id"`
	Direction       types.String `tfsdk:"direction"`
	Protocol        types.String `tfsdk:"protocol"`
	PortMin         types.Int64  `tfsdk:"port_min"`
	PortMax         types.Int64  `tfsdk:"port_max"`
	CIDR            types.String `tfsdk:"cidr"`
	Priority        types.Int64  `tfsdk:"priority"`
}

func (r *SecurityGroupRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group_rule"
}

func (r *SecurityGroupRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Security Group Rule resource allows you to manage specific firewall rules.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the security group rule.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"security_group_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the security group to add this rule to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"direction": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The direction of traffic (ingress or egress).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"protocol": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The protocol (tcp, udp, icmp, all).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"port_min": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The minimum port number.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"port_max": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The maximum port number.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"cidr": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The CIDR block for the rule.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"priority": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The evaluation priority of the rule.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *SecurityGroupRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SecurityGroupRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SecurityGroupRuleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ruleReq := client.SecurityRule{
		GroupID:   data.SecurityGroupID.ValueString(),
		Direction: data.Direction.ValueString(),
		Protocol:  data.Protocol.ValueString(),
		PortMin:   int(data.PortMin.ValueInt64()),
		PortMax:   int(data.PortMax.ValueInt64()),
		CIDR:      data.CIDR.ValueString(),
		Priority:  int(data.Priority.ValueInt64()),
	}

	rule, err := r.client.AddSecurityRule(data.SecurityGroupID.ValueString(), ruleReq)
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to create security group rule, got error: %s", err))
		return
	}

	data.ID = types.StringValue(rule.ID)
	data.Priority = types.Int64Value(int64(rule.Priority))

	tflog.Trace(ctx, "created a Security Group Rule resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecurityGroupRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecurityGroupRuleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// The API doesn't have a direct "GetRule" by ID, it returns rules within GetGroup.
	// We need to fetch the group and find our rule.
	sg, err := r.client.GetSecurityGroup(data.SecurityGroupID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to read security group for rule, got error: %s", err))
		return
	}

	if sg == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	found := false
	for _, rule := range sg.Rules {
		if rule.ID == data.ID.ValueString() {
			data.Direction = types.StringValue(rule.Direction)
			data.Protocol = types.StringValue(rule.Protocol)
			data.PortMin = types.Int64Value(int64(rule.PortMin))
			data.PortMax = types.Int64Value(int64(rule.PortMax))
			data.CIDR = types.StringValue(rule.CIDR)
			data.Priority = types.Int64Value(int64(rule.Priority))
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

func (r *SecurityGroupRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddWarning("Update Not Supported", "Updating a security group rule is not currently supported by the API.")
}

func (r *SecurityGroupRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SecurityGroupRuleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RemoveSecurityRule(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to delete security group rule, got error: %s", err))
		return
	}
}

func (r *SecurityGroupRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
