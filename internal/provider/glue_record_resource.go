// internal/provider/glue_record_resource.go
package provider

import (
	"context"
	"fmt"
	"sort"
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

var _ resource.Resource = &glueRecordResource{}
var _ resource.ResourceWithConfigure = &glueRecordResource{}
var _ resource.ResourceWithImportState = &glueRecordResource{}

func NewGlueRecordResource() resource.Resource {
	return &glueRecordResource{}
}

type glueRecordResource struct {
	client *porkbun.Client
}

type glueRecordResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Domain types.String `tfsdk:"domain"`
	Host   types.String `tfsdk:"host"`
	IPs    types.List   `tfsdk:"ips"`
}

func (r *glueRecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_glue_record"
}

func (r *glueRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Verwaltet einen Glue Record bei Porkbun. Glue Records sind notwendig, wenn die Nameserver Subdomains der Domain selbst sind.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Der eindeutige Bezeichner des Glue Records im Format 'domain/host'.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				Description: "Die Top-Level-Domain, für die der Glue Record erstellt wird (z.B. 'example.com').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"host": schema.StringAttribute{
				Description: "Der Host-Teil des Nameservers (z.B. 'ns1' für 'ns1.example.com').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ips": schema.ListAttribute{
				Description: "Eine Liste von IP-Adressen (IPv4 oder IPv6) für den Nameserver-Host.",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *glueRecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*porkbun.Client)
	if !ok {
		resp.Diagnostics.AddError("Unerwarteter Konfigurationstyp", fmt.Sprintf("Erwartet *porkbun.Client, erhalten: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *glueRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan glueRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := plan.Domain.ValueString()
	host := plan.Host.ValueString()

	var ipList []string
	resp.Diagnostics.Append(plan.IPs.ElementsAs(ctx, &ipList, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.AddGlueRecord(domain, host, ipList)
	if err != nil {
		resp.Diagnostics.AddError("Fehler beim Erstellen des Glue Records", "Konnte Glue Record nicht hinzufügen: "+err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", domain, host))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *glueRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state glueRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()
	host := state.Host.ValueString()

	allGlueRecords, err := r.client.GetGlueRecords(domain)
	if err != nil {
		resp.Diagnostics.AddError("Fehler beim Lesen der Glue Records", "Konnte Glue Records für die Domain nicht abrufen: "+err.Error())
		return
	}

	foundIPs, ok := allGlueRecords[host]
	if !ok {
		tflog.Warn(ctx, "Glue Record nicht gefunden, wird aus dem State entfernt.", map[string]any{"id": state.ID.ValueString()})
		resp.State.RemoveResource(ctx)
		return
	}

	sort.Strings(foundIPs)

	state.IPs, resp.Diagnostics = types.ListValueFrom(ctx, types.StringType, foundIPs)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *glueRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state glueRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()
	host := state.Host.ValueString()

	var newIPs []string
	resp.Diagnostics.Append(plan.IPs.ElementsAs(ctx, &newIPs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteGlueRecord(domain, host)
	if err != nil && !strings.Contains(err.Error(), "Could not find glue record") {
		resp.Diagnostics.AddError("Fehler beim Aktualisieren (Löschen) des Glue Records", err.Error())
		return
	}

	err = r.client.AddGlueRecord(domain, host, newIPs)
	if err != nil {
		resp.Diagnostics.AddError("Fehler beim Aktualisieren (Hinzufügen) des Glue Records", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", domain, host))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *glueRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state glueRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteGlueRecord(state.Domain.ValueString(), state.Host.ValueString())
	if err != nil && !strings.Contains(err.Error(), "Could not find glue record") {
		resp.Diagnostics.AddError("Fehler beim Löschen des Glue Records", err.Error())
		return
	}
}

func (r *glueRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Unerwarteter Import-Bezeichner",
			fmt.Sprintf("Erwarteter Bezeichner im Format 'domain/host'. Erhalten: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("host"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
