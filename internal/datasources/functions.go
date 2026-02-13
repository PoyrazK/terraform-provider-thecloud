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
var _ datasource.DataSource = &FunctionsDataSource{}

func NewFunctionsDataSource() datasource.DataSource {
	return &FunctionsDataSource{}
}

// FunctionsDataSource defines the data source implementation.
type FunctionsDataSource struct {
	client *client.Client
}

// FunctionsDataSourceModel describes the data source data model.
type FunctionsDataSourceModel struct {
	Functions []FunctionDataSourceModel `tfsdk:"functions"`
}

func (d *FunctionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_functions"
}

func (d *FunctionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Functions data source allows you to list all available serverless functions.",

		Attributes: map[string]schema.Attribute{
			"functions": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of functions.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the function.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the function.",
						},
						"runtime": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The runtime of the function.",
						},
						"handler": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The entry point of the function.",
						},
						"code_path": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The path to the function code artifact.",
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
				},
			},
		},
	}
}

func (d *FunctionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FunctionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FunctionsDataSourceModel

	functions, err := d.client.ListFunctions(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list functions, got error: %s", err))
		return
	}

	for _, f := range functions {
		data.Functions = append(data.Functions, FunctionDataSourceModel{
			ID:        types.StringValue(f.ID),
			Name:      types.StringValue(f.Name),
			Runtime:   types.StringValue(f.Runtime),
			Handler:   types.StringValue(f.Handler),
			CodePath:  types.StringValue(f.CodePath),
			Status:    types.StringValue(f.Status),
			CreatedAt: types.StringValue(f.CreatedAt.String()),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
