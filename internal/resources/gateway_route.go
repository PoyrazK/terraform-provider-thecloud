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
var _ resource.Resource = &GatewayRouteResource{}
var _ resource.ResourceWithImportState = &GatewayRouteResource{}

func NewGatewayRouteResource() resource.Resource {
	return &GatewayRouteResource{}
}

// GatewayRouteResource defines the resource implementation.
type GatewayRouteResource struct {
	client *client.Client
}

// GatewayRouteResourceModel describes the resource data model.
type GatewayRouteResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	PathPrefix  types.String `tfsdk:"path_prefix"`
	TargetURL   types.String `tfsdk:"target_url"`
	Methods     types.List   `tfsdk:"methods"`
	StripPrefix types.Bool   `tfsdk:"strip_prefix"`
	RateLimit   types.Int64  `tfsdk:"rate_limit"`
	Priority    types.Int64  `tfsdk:"priority"`
}

func (r *GatewayRouteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gateway_route"
}

func (r *GatewayRouteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gateway Route resource allows you to manage API gateway routing rules.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the route.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the route.",
			},
			"path_prefix": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The path pattern to match.",
			},
			"target_url": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The destination URL to proxy to.",
			},
			"methods": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "HTTP methods to match (e.g., GET, POST).",
			},
			"strip_prefix": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to strip the path prefix before forwarding.",
			},
			"rate_limit": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Maximum requests per second per IP.",
			},
			"priority": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Priority for route matching.",
			},
		},
	}
}

func (r *GatewayRouteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GatewayRouteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GatewayRouteResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var methods []string
	if !data.Methods.IsNull() {
		resp.Diagnostics.Append(data.Methods.ElementsAs(ctx, &methods, false)...)
	}

	routeReq := client.CreateRouteRequest{
		Name:        data.Name.ValueString(),
		PathPrefix:  data.PathPrefix.ValueString(),
		TargetURL:   data.TargetURL.ValueString(),
		Methods:     methods,
		StripPrefix: data.StripPrefix.ValueBool(),
		RateLimit:   int(data.RateLimit.ValueInt64()),
		Priority:    int(data.Priority.ValueInt64()),
	}

	route, err := r.client.CreateGatewayRoute(ctx, routeReq)
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to create Gateway Route, got error: %s", err))
		return
	}

	data.ID = types.StringValue(route.ID)
	data.StripPrefix = types.BoolValue(route.StripPrefix)
	data.RateLimit = types.Int64Value(int64(route.RateLimit))
	data.Priority = types.Int64Value(int64(route.Priority))

	tflog.Trace(ctx, "created a Gateway Route resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GatewayRouteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GatewayRouteResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	route, err := r.client.GetGatewayRoute(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to read Gateway Route, got error: %s", err))
		return
	}

	if route == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(route.ID)
	data.Name = types.StringValue(route.Name)
	data.PathPrefix = types.StringValue(route.PathPrefix)
	data.TargetURL = types.StringValue(route.TargetURL)
	data.StripPrefix = types.BoolValue(route.StripPrefix)
	data.RateLimit = types.Int64Value(int64(route.RateLimit))
	data.Priority = types.Int64Value(int64(route.Priority))

	methods, diags := types.ListValueFrom(ctx, types.StringType, route.Methods)
	resp.Diagnostics.Append(diags...)
	data.Methods = methods

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GatewayRouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// API might not support updates, handled by RequiresReplace if needed or no-op
}

func (r *GatewayRouteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GatewayRouteResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteGatewayRoute(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to delete Gateway Route, got error: %s", err))
		return
	}
}

func (r *GatewayRouteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
