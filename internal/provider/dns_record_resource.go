package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/flooopro/terraform-provider-porkbun/internal/porkbun"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &dnsRecordResource{}
	_ resource.ResourceWithConfigure   = &dnsRecordResource{}
	_ resource.ResourceWithImportState = &dnsRecordResource{}
)

func NewDnsRecordResource() resource.Resource {
	return &dnsRecordResource{}
}

type dnsRecordResource struct {
	client *porkbun.Client
}

type dnsRecordResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Domain  types.String `tfsdk:"domain"`
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	Content types.String `tfsdk:"content"`
	TTL     types.String `tfsdk:"ttl"`
	Prio    types.String `tfsdk:"prio"`
}

func (r *dnsRecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (r *dnsRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DNS record on Porkbun.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the DNS record.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				Description: "The domain name for the record.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The subdomain for the record, if any. Use an empty string for the root domain.",
				Optional:    true,
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of the DNS record (e.g., A, CNAME, TXT).",
				Required:    true,
			},
			"content": schema.StringAttribute{
				Description: "The content/value of the DNS record.",
				Required:    true,
			},
			"ttl": schema.StringAttribute{
				Description: "The Time To Live (TTL) of the record in seconds.",
				Optional:    true,
				Computed:    true,
			},
			"prio": schema.StringAttribute{
				Description: "The priority of the record (for MX and SRV records only).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *dnsRecordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*porkbun.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *porkbun.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *dnsRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dnsRecordResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	record := porkbun.DnsRecord{
		Name:    plan.Name.ValueString(),
		Type:    plan.Type.ValueString(),
		Content: plan.Content.ValueString(),
		TTL:     plan.TTL.ValueString(),
		Prio:    plan.Prio.ValueString(),
	}

	recordID, err := r.client.CreateRecord(plan.Domain.ValueString(), record)
	if err != nil {
		resp.Diagnostics.AddError("Error creating DNS record", "Could not create record, unexpected error: "+err.Error())
		return
	}

	plan.ID = types.StringValue(recordID)

	if plan.TTL.IsUnknown() || plan.TTL.IsNull() {
		plan.TTL = types.StringValue("300")
	}
	if plan.Name.IsUnknown() || plan.Name.IsNull() {
		plan.Name = types.StringValue("")
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *dnsRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dnsRecordResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Reading all records for domain", map[string]interface{}{"domain": state.Domain.ValueString()})
	records, err := r.client.RetrieveRecords(state.Domain.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading Porkbun records", "Could not retrieve records for domain "+state.Domain.ValueString()+": "+err.Error())
		return
	}

	var foundRecord *porkbun.DnsRecord
	for _, record := range records {
		if record.ID == state.ID.ValueString() {
			rec := record
			foundRecord = &rec
			break
		}
	}

	if foundRecord == nil {
		tflog.Warn(ctx, "DNS record not found, removing from state", map[string]interface{}{"id": state.ID.ValueString()})
		resp.State.RemoveResource(ctx)
		return
	}

	recordName := foundRecord.Name
	domainName := state.Domain.ValueString()
	normalizedName := strings.TrimSuffix(recordName, "."+domainName)
	if normalizedName == domainName {
		normalizedName = ""
	}

	state.Name = types.StringValue(normalizedName)
	state.Type = types.StringValue(foundRecord.Type)
	state.Content = types.StringValue(foundRecord.Content)
	state.TTL = types.StringValue(foundRecord.TTL)
	state.Prio = types.StringValue(foundRecord.Prio)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *dnsRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dnsRecordResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	record := porkbun.DnsRecord{
		Name:    plan.Name.ValueString(),
		Type:    plan.Type.ValueString(),
		Content: plan.Content.ValueString(),
		TTL:     plan.TTL.ValueString(),
		Prio:    plan.Prio.ValueString(),
	}

	err := r.client.EditRecord(plan.Domain.ValueString(), plan.ID.ValueString(), record)
	if err != nil {
		resp.Diagnostics.AddError("Error updating DNS record", "Could not update record, unexpected error: "+err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *dnsRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dnsRecordResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRecord(state.Domain.ValueString(), state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			tflog.Warn(ctx, "Record to be deleted was not found on remote. Ignoring.")
			return
		}
		resp.Diagnostics.AddError("Error deleting DNS record", "Could not delete record, unexpected error: "+err.Error())
		return
	}
}

func (r *dnsRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: domain/record_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
