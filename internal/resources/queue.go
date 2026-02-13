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
var _ resource.Resource = &QueueResource{}
var _ resource.ResourceWithImportState = &QueueResource{}

func NewQueueResource() resource.Resource {
	return &QueueResource{}
}

// QueueResource defines the resource implementation.
type QueueResource struct {
	client *client.Client
}

// QueueResourceModel describes the resource data model.
type QueueResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	ARN               types.String `tfsdk:"arn"`
	VisibilityTimeout types.Int64  `tfsdk:"visibility_timeout"`
	RetentionDays     types.Int64  `tfsdk:"retention_days"`
	MaxMessageSize    types.Int64  `tfsdk:"max_message_size"`
	Status            types.String `tfsdk:"status"`
}

func (r *QueueResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_queue"
}

func (r *QueueResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Queue resource allows you to manage message queues.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the queue.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the queue.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"arn": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Amazon Resource Name (ARN) of the queue.",
			},
			"visibility_timeout": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Seconds a message remains hidden after retrieval.",
			},
			"retention_days": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Days before non-deleted messages are purged.",
			},
			"max_message_size": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Maximum payload size in bytes.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the queue.",
			},
		},
	}
}

func (r *QueueResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *QueueResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data QueueResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	opts := client.CreateQueueOptions{}
	if !data.VisibilityTimeout.IsNull() {
		v := int(data.VisibilityTimeout.ValueInt64())
		opts.VisibilityTimeout = &v
	}
	if !data.RetentionDays.IsNull() {
		v := int(data.RetentionDays.ValueInt64())
		opts.RetentionDays = &v
	}
	if !data.MaxMessageSize.IsNull() {
		v := int(data.MaxMessageSize.ValueInt64())
		opts.MaxMessageSize = &v
	}

	q, err := r.client.CreateQueue(ctx, data.Name.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to create Queue, got error: %s", err))
		return
	}

	data.ID = types.StringValue(q.ID)
	data.ARN = types.StringValue(q.ARN)
	data.Status = types.StringValue(q.Status)
	data.VisibilityTimeout = types.Int64Value(int64(q.VisibilityTimeout))
	data.RetentionDays = types.Int64Value(int64(q.RetentionDays))
	data.MaxMessageSize = types.Int64Value(int64(q.MaxMessageSize))

	tflog.Trace(ctx, "created a Queue resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *QueueResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data QueueResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	q, err := r.client.GetQueue(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to read Queue, got error: %s", err))
		return
	}

	if q == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(q.ID)
	data.Name = types.StringValue(q.Name)
	data.ARN = types.StringValue(q.ARN)
	data.VisibilityTimeout = types.Int64Value(int64(q.VisibilityTimeout))
	data.RetentionDays = types.Int64Value(int64(q.RetentionDays))
	data.MaxMessageSize = types.Int64Value(int64(q.MaxMessageSize))
	data.Status = types.StringValue(q.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *QueueResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Not supported by current API, use RequiresReplace
}

func (r *QueueResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data QueueResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteQueue(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to delete Queue, got error: %s", err))
		return
	}
}

func (r *QueueResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
