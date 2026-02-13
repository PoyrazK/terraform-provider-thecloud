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
var _ datasource.DataSource = &BucketsDataSource{}

func NewBucketsDataSource() datasource.DataSource {
	return &BucketsDataSource{}
}

// BucketsDataSource defines the data source implementation.
type BucketsDataSource struct {
	client *client.Client
}

// BucketsDataSourceModel describes the data source data model.
type BucketsDataSourceModel struct {
	Buckets []BucketDataSourceModel `tfsdk:"buckets"`
}

func (d *BucketsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_buckets"
}

func (d *BucketsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Buckets data source allows you to list all available storage buckets.",

		Attributes: map[string]schema.Attribute{
			"buckets": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of buckets.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the bucket.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the bucket.",
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
				},
			},
		},
	}
}

func (d *BucketsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *BucketsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BucketsDataSourceModel

	buckets, err := d.client.ListBuckets(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list buckets, got error: %s", err))
		return
	}

	for _, b := range buckets {
		data.Buckets = append(data.Buckets, BucketDataSourceModel{
			ID:                types.StringValue(b.ID),
			Name:              types.StringValue(b.Name),
			IsPublic:          types.BoolValue(b.IsPublic),
			VersioningEnabled: types.BoolValue(b.VersioningEnabled),
			EncryptionEnabled: types.BoolValue(b.EncryptionEnabled),
			CreatedAt:         types.StringValue(b.CreatedAt),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
