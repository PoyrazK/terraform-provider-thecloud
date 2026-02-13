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
var _ resource.Resource = &CacheResource{}
var _ resource.ResourceWithImportState = &CacheResource{}

func NewCacheResource() resource.Resource {
	return &CacheResource{}
}

// CacheResource defines the resource implementation.
type CacheResource struct {
	client *client.Client
}

// CacheResourceModel describes the resource data model.
type CacheResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Engine           types.String `tfsdk:"engine"`
	Version          types.String `tfsdk:"version"`
	VpcID            types.String `tfsdk:"vpc_id"`
	MemoryMB         types.Int64  `tfsdk:"memory_mb"`
	Status           types.String `tfsdk:"status"`
	Port             types.Int64  `tfsdk:"port"`
	ConnectionString types.String `tfsdk:"connection_string"`
}

func (r *CacheResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cache"
}

func (r *CacheResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Cache resource allows you to manage managed caching instances (Redis).",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the cache.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the cache instance.",
			},
			"engine": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The cache engine (e.g. redis).",
			},
			"version": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The engine version (e.g. 7.0).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vpc_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the VPC to launch the cache in.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"memory_mb": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Memory allocation in MB.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the cache.",
			},
			"port": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The port the cache is listening on.",
			},
			"connection_string": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The connection string for the cache.",
				Sensitive:           true,
			},
		},
	}
}

func (r *CacheResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CacheResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CacheResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	cache, err := r.client.CreateCache(
		ctx,
		data.Name.ValueString(),
		data.Version.ValueString(),
		int(data.MemoryMB.ValueInt64()),
		data.VpcID.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to create Cache, got error: %s", err))
		return
	}

	data.ID = types.StringValue(cache.ID)
	data.Engine = types.StringValue(cache.Engine)
	data.Status = types.StringValue(cache.Status)
	data.Port = types.Int64Value(int64(cache.Port))
	data.ConnectionString = types.StringValue(cache.ConnectionString)

	tflog.Trace(ctx, "created a Cache resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CacheResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CacheResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	cache, err := r.client.GetCache(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to read Cache, got error: %s", err))
		return
	}

	if cache == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(cache.ID)
	data.Name = types.StringValue(cache.Name)
	data.Engine = types.StringValue(cache.Engine)
	data.Version = types.StringValue(cache.Version)
	data.VpcID = types.StringValue(cache.VpcID)
	data.MemoryMB = types.Int64Value(int64(cache.MemoryMB))
	data.Status = types.StringValue(cache.Status)
	data.Port = types.Int64Value(int64(cache.Port))
	data.ConnectionString = types.StringValue(cache.ConnectionString)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CacheResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Not supported, handled by RequiresReplace if needed
}

func (r *CacheResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CacheResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteCache(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to delete Cache, got error: %s", err))
		return
	}
}

func (r *CacheResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
