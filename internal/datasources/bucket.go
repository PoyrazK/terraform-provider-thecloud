package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/poyrazk/terraform-provider-thecloud/internal/client"
)

// Ensure implementation of interfaces
var _ datasource.DataSource = &BucketDataSource{}

func NewBucketDataSource() datasource.DataSource {
	return &BucketDataSource{}
}

// BucketDataSource defines the data source implementation.
type BucketDataSource struct {
	client *client.Client
}

// BucketDataSourceModel describes the data source data model.
type BucketDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	IsPublic          types.Bool   `tfsdk:"is_public"`
	VersioningEnabled types.Bool   `tfsdk:"versioning_enabled"`
	EncryptionEnabled types.Bool   `tfsdk:"encryption_enabled"`
	CreatedAt         types.String `tfsdk:"created_at"`
}

func (d *BucketDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket"
}

func (d *BucketDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Bucket data source allows you to look up bucket details by Name.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the bucket.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the bucket to look up.",
			},
			"is_public": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the bucket is public.",
			},
			"versioning_enabled": schema.BoolAttribute{
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

func (d *BucketDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Data Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *BucketDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BucketDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	bucket, err := d.client.GetBucket(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read bucket, got error: %s", err))
		return
	}

	if bucket == nil {
		resp.Diagnostics.AddError("Bucket Not Found", "No bucket matching the name was found.")
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
