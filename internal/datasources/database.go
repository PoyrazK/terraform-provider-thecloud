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
var _ datasource.DataSource = &DatabaseDataSource{}

func NewDatabaseDataSource() datasource.DataSource {
	return &DatabaseDataSource{}
}

// DatabaseDataSource defines the data source implementation.
type DatabaseDataSource struct {
	client *client.Client
}

// DatabaseDataSourceModel describes the data source data model.
type DatabaseDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Engine           types.String `tfsdk:"engine"`
	Version          types.String `tfsdk:"version"`
	VpcID            types.String `tfsdk:"vpc_id"`
	Status           types.String `tfsdk:"status"`
	Port             types.Int64  `tfsdk:"port"`
	Username         types.String `tfsdk:"username"`
	ConnectionString types.String `tfsdk:"connection_string"`
}

func (d *DatabaseDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (d *DatabaseDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Database data source allows you to look up managed database details by ID or Name.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The ID of the database to look up.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The name of the database to look up.",
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
	}
}

func (d *DatabaseDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DatabaseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DatabaseDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var found *client.Database
	var err error

	if !data.ID.IsNull() {
		found, err = d.client.GetDatabase(ctx, data.ID.ValueString())
	} else if !data.Name.IsNull() {
		found, err = d.lookupDatabaseByName(ctx, data.Name.ValueString())
	} else {
		resp.Diagnostics.AddError("Missing Required Attribute", "Either id or name must be specified.")
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read database, got error: %s", err))
		return
	}

	if found == nil {
		resp.Diagnostics.AddError("Database Not Found", "No database matching the criteria was found.")
		return
	}

	data.ID = types.StringValue(found.ID)
	data.Name = types.StringValue(found.Name)
	data.Engine = types.StringValue(found.Engine)
	data.Version = types.StringValue(found.Version)
	data.VpcID = types.StringValue(found.VpcID)
	data.Status = types.StringValue(found.Status)
	data.Port = types.Int64Value(int64(found.Port))
	data.Username = types.StringValue(found.Username)
	data.ConnectionString = types.StringValue(found.ConnectionString)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *DatabaseDataSource) lookupDatabaseByName(ctx context.Context, name string) (*client.Database, error) {
	dbs, err := d.client.ListDatabases(ctx)
	if err != nil {
		return nil, err
	}

	for _, db := range dbs {
		if db.Name == name {
			// Get full details including connection string
			return d.client.GetDatabase(ctx, db.ID)
		}
	}

	return nil, nil
}
