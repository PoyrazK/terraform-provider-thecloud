package resources

import (
	"context"
	"fmt"
	"os"

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
var _ resource.Resource = &FunctionResource{}
var _ resource.ResourceWithImportState = &FunctionResource{}

func NewFunctionResource() resource.Resource {
	return &FunctionResource{}
}

// FunctionResource defines the resource implementation.
type FunctionResource struct {
	client *client.Client
}

// FunctionResourceModel describes the resource data model.
type FunctionResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Runtime   types.String `tfsdk:"runtime"`
	Handler   types.String `tfsdk:"handler"`
	Filename  types.String `tfsdk:"filename"`
	Status    types.String `tfsdk:"status"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (r *FunctionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_function"
}

func (r *FunctionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Function resource allows you to manage serverless functions.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the function.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the function.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"runtime": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The runtime of the function (e.g., python3.9, go1.21).",
			},
			"handler": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The entry point of the function.",
			},
			"filename": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The path to the zip file containing the function code.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the function.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the function was created.",
			},
		},
	}
}

func (r *FunctionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FunctionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FunctionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	code, err := os.ReadFile(data.Filename.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("File Error", fmt.Sprintf("Unable to read code file %s, got error: %s", data.Filename.ValueString(), err))
		return
	}

	function, err := r.client.CreateFunction(
		ctx,
		data.Name.ValueString(),
		data.Runtime.ValueString(),
		data.Handler.ValueString(),
		code,
	)
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to create Function, got error: %s", err))
		return
	}

	data.ID = types.StringValue(function.ID)
	data.Status = types.StringValue(function.Status)
	data.CreatedAt = types.StringValue(function.CreatedAt.String())

	tflog.Trace(ctx, "created a Function resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FunctionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FunctionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	function, err := r.client.GetFunction(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to read Function, got error: %s", err))
		return
	}

	if function == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(function.ID)
	data.Name = types.StringValue(function.Name)
	data.Runtime = types.StringValue(function.Runtime)
	data.Handler = types.StringValue(function.Handler)
	data.Status = types.StringValue(function.Status)
	data.CreatedAt = types.StringValue(function.CreatedAt.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FunctionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// API might handle updates by re-creating or a separate endpoint.
	// For now, we'll mark fields as RequiresReplace where appropriate or let it be no-op.
}

func (r *FunctionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FunctionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteFunction(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to delete Function, got error: %s", err))
		return
	}
}

func (r *FunctionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
