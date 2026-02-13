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
var _ resource.Resource = &BucketResource{}
var _ resource.ResourceWithImportState = &BucketResource{}

func NewBucketResource() resource.Resource {
	return &BucketResource{}
}

// BucketResource defines the resource implementation.
type BucketResource struct {
	client *client.Client
}

// BucketResourceModel describes the resource data model.
type BucketResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	IsPublic          types.Bool   `tfsdk:"is_public"`
	VersioningEnabled types.Bool   `tfsdk:"versioning_enabled"`
	EncryptionEnabled types.Bool   `tfsdk:"encryption_enabled"`
	CreatedAt         types.String `tfsdk:"created_at"`
}

func (r *BucketResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket"
}

func (r *BucketResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Bucket resource allows you to manage object storage buckets.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the bucket.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the bucket.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_public": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the bucket is public.",
			},
			"versioning_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether versioning is enabled.",
			},
			"encryption_enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether encryption is enabled.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the bucket was created.",
			},
		},
	}
}

func (r *BucketResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BucketResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	bucket, err := r.client.CreateBucket(ctx, data.Name.ValueString(), data.IsPublic.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to create Bucket, got error: %s", err))
		return
	}

	data.ID = types.StringValue(bucket.ID)
	data.IsPublic = types.BoolValue(bucket.IsPublic)
	data.EncryptionEnabled = types.BoolValue(bucket.EncryptionEnabled)
	data.CreatedAt = types.StringValue(bucket.CreatedAt)

	// Update versioning if requested (API Create doesn't seem to set it directly)
	if !data.VersioningEnabled.IsNull() && data.VersioningEnabled.ValueBool() {
		err = r.client.SetBucketVersioning(ctx, bucket.Name, true)
		if err != nil {
			resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to set Bucket versioning, got error: %s", err))
			return
		}
		data.VersioningEnabled = types.BoolValue(true)
	} else {
		data.VersioningEnabled = types.BoolValue(false)
	}

	tflog.Trace(ctx, "created a Bucket resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BucketResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	bucket, err := r.client.GetBucket(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to read Bucket, got error: %s", err))
		return
	}

	if bucket == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(bucket.ID)
	data.Name = types.StringValue(bucket.Name)
	data.IsPublic = types.BoolValue(bucket.IsPublic)
	data.VersioningEnabled = types.BoolValue(bucket.VersioningEnabled)
	data.EncryptionEnabled = types.BoolValue(bucket.EncryptionEnabled)
	data.CreatedAt = types.StringValue(bucket.CreatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state BucketResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.VersioningEnabled.Equal(state.VersioningEnabled) {
		err := r.client.SetBucketVersioning(ctx, plan.Name.ValueString(), plan.VersioningEnabled.ValueBool())
		if err != nil {
			resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to update Bucket versioning, got error: %s", err))
			return
		}
	}

	// is_public update not clearly supported by single PATCH, but let's assume it might be or handled via Recreate
	// For now we only handle versioning as updateable field based on handler code.

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BucketResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteBucket(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to delete Bucket, got error: %s", err))
		return
	}
}

func (r *BucketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
