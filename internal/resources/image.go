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
var _ resource.Resource = &ImageResource{}
var _ resource.ResourceWithImportState = &ImageResource{}

func NewImageResource() resource.Resource {
	return &ImageResource{}
}

// ImageResource defines the resource implementation.
type ImageResource struct {
	client *client.Client
}

// ImageResourceModel describes the resource data model.
type ImageResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	OS          types.String `tfsdk:"os"`
	Version     types.String `tfsdk:"version"`
	IsPublic    types.Bool   `tfsdk:"is_public"`
	Filename    types.String `tfsdk:"filename"`
	Status      types.String `tfsdk:"status"`
}

func (r *ImageResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image"
}

func (r *ImageResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Image resource allows you to manage virtual machine images.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the image.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the image.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The description of the image.",
			},
			"os": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The operating system of the image (e.g. ubuntu).",
			},
			"version": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The version of the operating system (e.g. 22.04).",
			},
			"is_public": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the image is public.",
			},
			"filename": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The path to the image file to upload.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the image.",
			},
		},
	}
}

func (r *ImageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ImageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ImageResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	registerReq := client.RegisterImageRequest{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		OS:          data.OS.ValueString(),
		Version:     data.Version.ValueString(),
		IsPublic:    data.IsPublic.ValueBool(),
	}

	image, err := r.client.RegisterImage(ctx, registerReq)
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to register Image, got error: %s", err))
		return
	}

	code, err := os.ReadFile(data.Filename.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("File Error", fmt.Sprintf("Unable to read image file %s, got error: %s", data.Filename.ValueString(), err))
		return
	}

	err = r.client.UploadImage(ctx, image.ID, code)
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to upload Image, got error: %s", err))
		return
	}

	data.ID = types.StringValue(image.ID)
	data.Status = types.StringValue(image.Status)

	tflog.Trace(ctx, "created an Image resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ImageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ImageResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	image, err := r.client.GetImage(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to read Image, got error: %s", err))
		return
	}

	if image == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(image.ID)
	data.Name = types.StringValue(image.Name)
	data.Description = types.StringValue(image.Description)
	data.OS = types.StringValue(image.OS)
	data.Version = types.StringValue(image.Version)
	data.IsPublic = types.BoolValue(image.IsPublic)
	data.Status = types.StringValue(image.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ImageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Not supported
}

func (r *ImageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ImageResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteImage(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to delete Image, got error: %s", err))
		return
	}
}

func (r *ImageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
