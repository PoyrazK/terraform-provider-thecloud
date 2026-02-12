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
var _ resource.Resource = &DNSRecordResource{}
var _ resource.ResourceWithImportState = &DNSRecordResource{}

func NewDNSRecordResource() resource.Resource {
	return &DNSRecordResource{}
}

// DNSRecordResource defines the resource implementation.
type DNSRecordResource struct {
	client *client.Client
}

// DNSRecordResourceModel describes the resource data model.
type DNSRecordResourceModel struct {
	ID       types.String `tfsdk:"id"`
	ZoneID   types.String `tfsdk:"zone_id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Content  types.String `tfsdk:"content"`
	TTL      types.Int64  `tfsdk:"ttl"`
	Priority types.Int64  `tfsdk:"priority"`
}

func (r *DNSRecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (r *DNSRecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "DNS Record resource allows you to manage DNS records within a zone.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the DNS Record.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"zone_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the DNS Zone this record belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the DNS Record (e.g., www).",
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The type of the DNS Record (A, AAAA, CNAME, MX, TXT, SRV).",
			},
			"content": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The content of the DNS Record (e.g., IP address).",
			},
			"ttl": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The TTL of the DNS Record.",
			},
			"priority": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The priority of the DNS Record (for MX, SRV).",
			},
		},
	}
}

func (r *DNSRecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DNSRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DNSRecordResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	record := client.DNSRecord{
		Name:    data.Name.ValueString(),
		Type:    data.Type.ValueString(),
		Content: data.Content.ValueString(),
		TTL:     int(data.TTL.ValueInt64()),
	}

	if !data.Priority.IsNull() {
		p := int(data.Priority.ValueInt64())
		record.Priority = &p
	}

	res, err := r.client.CreateDNSRecord(ctx, data.ZoneID.ValueString(), record)
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to create DNS Record, got error: %s", err))
		return
	}

	data.ID = types.StringValue(res.ID)
	data.TTL = types.Int64Value(int64(res.TTL))
	if res.Priority != nil {
		data.Priority = types.Int64Value(int64(*res.Priority))
	} else {
		data.Priority = types.Int64Null()
	}

	tflog.Trace(ctx, "created a DNS Record resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DNSRecordResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	record, err := r.client.GetDNSRecord(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to read DNS Record, got error: %s", err))
		return
	}

	if record == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(record.ID)
	data.ZoneID = types.StringValue(record.ZoneID)
	data.Name = types.StringValue(record.Name)
	data.Type = types.StringValue(record.Type)
	data.Content = types.StringValue(record.Content)
	data.TTL = types.Int64Value(int64(record.TTL))
	if record.Priority != nil {
		data.Priority = types.Int64Value(int64(*record.Priority))
	} else {
		data.Priority = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DNSRecordResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	record := client.DNSRecord{
		ID:      data.ID.ValueString(),
		ZoneID:  data.ZoneID.ValueString(),
		Name:    data.Name.ValueString(),
		Type:    data.Type.ValueString(),
		Content: data.Content.ValueString(),
		TTL:     int(data.TTL.ValueInt64()),
	}

	if !data.Priority.IsNull() {
		p := int(data.Priority.ValueInt64())
		record.Priority = &p
	}

	res, err := r.client.UpdateDNSRecord(ctx, data.ID.ValueString(), record)
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to update DNS Record, got error: %s", err))
		return
	}

	data.TTL = types.Int64Value(int64(res.TTL))
	if res.Priority != nil {
		data.Priority = types.Int64Value(int64(*res.Priority))
	} else {
		data.Priority = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DNSRecordResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDNSRecord(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errClient, fmt.Sprintf("Unable to delete DNS Record, got error: %s", err))
		return
	}
}

func (r *DNSRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
