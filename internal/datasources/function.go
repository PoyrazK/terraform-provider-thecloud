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
var _ datasource.DataSource = &FunctionDataSource{}

func NewFunctionDataSource() datasource.DataSource {
	return &FunctionDataSource{}
}

// FunctionDataSource defines the data source implementation.
type FunctionDataSource struct {
	client *client.Client
}

// FunctionDataSourceModel describes the data source data model.
type FunctionDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Runtime   types.String `tfsdk:"runtime"`
	Handler   types.String `tfsdk:"handler"`
	CodePath  types.String `tfsdk:"code_path"`
	Status    types.String `tfsdk:"status"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (d *FunctionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_function"
}

func (d *FunctionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Function data source allows you to look up serverless function details by ID or Name.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The ID of the function to look up.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The name of the function to look up.",
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
	}
}

func (d *FunctionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FunctionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FunctionDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var found *client.Function
	var err error

	if !data.ID.IsNull() {
		found, err = d.client.GetFunction(ctx, data.ID.ValueString())
	} else if !data.Name.IsNull() {
		found, err = d.lookupFunctionByName(ctx, data.Name.ValueString())
	} else {
		resp.Diagnostics.AddError("Missing Required Attribute", "Either id or name must be specified.")
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read function, got error: %s", err))
		return
	}

	if found == nil {
		resp.Diagnostics.AddError("Function Not Found", "No function matching the criteria was found.")
		return
	}

	data.ID = types.StringValue(found.ID)
	data.Name = types.StringValue(found.Name)
	data.Runtime = types.StringValue(found.Runtime)
	data.Handler = types.StringValue(found.Handler)
	data.CodePath = types.StringValue(found.CodePath)
	data.Status = types.StringValue(found.Status)
	data.CreatedAt = types.StringValue(found.CreatedAt.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *FunctionDataSource) lookupFunctionByName(ctx context.Context, name string) (*client.Function, error) {
	functions, err := d.client.ListFunctions(ctx)
	if err != nil {
		return nil, err
	}

	for _, f := range functions {
		if f.Name == name {
			return &f, nil
		}
	}

	return nil, nil
}
