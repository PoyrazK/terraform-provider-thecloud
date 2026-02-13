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
var _ datasource.DataSource = &DatabasesDataSource{}

func NewDatabasesDataSource() datasource.DataSource {
	return &DatabasesDataSource{}
}

// DatabasesDataSource defines the data source implementation.
type DatabasesDataSource struct {
	client *client.Client
}

// DatabasesDataSourceModel describes the data source data model.
type DatabasesDataSourceModel struct {
	Databases []DatabaseDataSourceModel `tfsdk:"databases"`
}

func (d *DatabasesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_databases"
}

func (d *DatabasesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Databases data source allows you to list all available managed databases.",

		Attributes: map[string]schema.Attribute{
			"databases": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of databases.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the database.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the database.",
						},
						"engine": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The database engine.",
						},
						"version": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The engine version.",
						},
						"vpc_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The VPC ID of the database.",
						},
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The status of the database.",
						},
						"port": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "The port the database is listening on.",
						},
						"username": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The master username for the database.",
						},
						"connection_string": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The connection string for the database.",
							Sensitive:           true,
						},
					},
				},
			},
		},
	}
}

func (d *DatabasesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DatabasesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DatabasesDataSourceModel

	databases, err := d.client.ListDatabases(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list databases, got error: %s", err))
		return
	}

	for _, db := range databases {
		data.Databases = append(data.Databases, DatabaseDataSourceModel{
			ID:               types.StringValue(db.ID),
			Name:             types.StringValue(db.Name),
			Engine:           types.StringValue(db.Engine),
			Version:          types.StringValue(db.Version),
			VpcID:            types.StringValue(db.VpcID),
			Status:           types.StringValue(db.Status),
			Port:             types.Int64Value(int64(db.Port)),
			Username:         types.StringValue(db.Username),
			ConnectionString: types.StringValue(db.ConnectionString),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
