// internal/provider/dnssec_record_resource.go
package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/flooopro/terraform-provider-porkbun/internal/porkbun" // Passen Sie den Pfad an
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &dnssecRecordResource{}
var _ resource.ResourceWithConfigure = &dnssecRecordResource{}
var _ resource.ResourceWithImportState = &dnssecRecordResource{}

func NewDnssecRecordResource() resource.Resource {
	return &dnssecRecordResource{}
}

type dnssecRecordResource struct {
	client *porkbun.Client
}

type dnssecRecordResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Domain     types.String `tfsdk:"domain"`
	Algorithm  types.String `tfsdk:"algorithm"`
	DigestType types.String `tfsdk:"digest_type"`
	KeyTag     types.String `tfsdk:"key_tag"`
	Digest     types.String `tfsdk:"digest"`
}

// buildDnssecID erstellt eine eindeutige, deterministische ID für einen DS-Eintrag.
func buildDnssecID(domain, keyTag, algorithm, digestType, digest string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", domain, keyTag, algorithm, digestType, digest)
}

func (r *dnssecRecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dnssec_record"
}

func (r *dnssecRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Verwaltet einen DNSSEC DS (Delegation Signer) Eintrag für eine Domain.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				Description: "Der Domainname.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"algorithm":   schema.StringAttribute{Required: true},
			"digest_type": schema.StringAttribute{Required: true},
			"key_tag":     schema.StringAttribute{Required: true},
			"digest":      schema.StringAttribute{Required: true},
		},
	}
}

func (r *dnssecRecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *dnssecRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dnssecRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	record := porkbun.DnssecRecord{
		Algorithm:  plan.Algorithm.ValueString(),
		DigestType: plan.DigestType.ValueString(),
		KeyTag:     plan.KeyTag.ValueString(),
		Digest:     plan.Digest.ValueString(),
	}

	err := r.client.AddDnssecRecord(plan.Domain.ValueString(), record)
	if err != nil {
		resp.Diagnostics.AddError("Fehler beim Erstellen des DNSSEC-Eintrags", err.Error())
		return
	}

	plan.ID = types.StringValue(buildDnssecID(
		plan.Domain.ValueString(),
		plan.KeyTag.ValueString(),
		plan.Algorithm.ValueString(),
		plan.DigestType.ValueString(),
		plan.Digest.ValueString(),
	))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dnssecRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dnssecRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()
	allRecords, err := r.client.GetDnssecRecords(domain)
	if err != nil {
		resp.Diagnostics.AddError("Fehler beim Lesen der DNSSEC-Einträge", err.Error())
		return
	}

	var foundRecord *porkbun.DnssecRecord
	for _, record := range allRecords {
		// Vergleichen Sie den gelesenen Eintrag mit dem im State gespeicherten.
		if record.KeyTag == state.KeyTag.ValueString() &&
			record.Algorithm == state.Algorithm.ValueString() &&
			record.DigestType == state.DigestType.ValueString() &&
			record.Digest == state.Digest.ValueString() {
			rec := record
			foundRecord = &rec
			break
		}
	}

	if foundRecord == nil {
		tflog.Warn(ctx, "DNSSEC-Eintrag nicht gefunden, wird aus dem State entfernt.", map[string]any{"id": state.ID.ValueString()})
		resp.State.RemoveResource(ctx)
		return
	}

	// Der Zustand ist aktuell, keine weiteren Aktionen erforderlich.
}

func (r *dnssecRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state dnssecRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ein "Update" für einen DS-Eintrag ist ein atomares "Löschen" des alten
	// und "Hinzufügen" des neuen Eintrags.

	oldRecord := porkbun.DnssecRecord{
		Algorithm:  state.Algorithm.ValueString(),
		DigestType: state.DigestType.ValueString(),
		KeyTag:     state.KeyTag.ValueString(),
		Digest:     state.Digest.ValueString(),
	}
	err := r.client.DeleteDnssecRecord(state.Domain.ValueString(), oldRecord)
	if err != nil {
		resp.Diagnostics.AddError("Fehler beim Aktualisieren (Löschen) des alten DNSSEC-Eintrags", err.Error())
		return
	}

	newRecord := porkbun.DnssecRecord{
		Algorithm:  plan.Algorithm.ValueString(),
		DigestType: plan.DigestType.ValueString(),
		KeyTag:     plan.KeyTag.ValueString(),
		Digest:     plan.Digest.ValueString(),
	}
	err = r.client.AddDnssecRecord(plan.Domain.ValueString(), newRecord)
	if err != nil {
		resp.Diagnostics.AddError("Fehler beim Aktualisieren (Hinzufügen) des neuen DNSSEC-Eintrags", err.Error())
		return
	}

	plan.ID = types.StringValue(buildDnssecID(
		plan.Domain.ValueString(),
		plan.KeyTag.ValueString(),
		plan.Algorithm.ValueString(),
		plan.DigestType.ValueString(),
		plan.Digest.ValueString(),
	))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dnssecRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dnssecRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	record := porkbun.DnssecRecord{
		Algorithm:  state.Algorithm.ValueString(),
		DigestType: state.DigestType.ValueString(),
		KeyTag:     state.KeyTag.ValueString(),
		Digest:     state.Digest.ValueString(),
	}

	err := r.client.DeleteDnssecRecord(state.Domain.ValueString(), record)
	if err != nil {
		resp.Diagnostics.AddError("Fehler beim Löschen des DNSSEC-Eintrags", err.Error())
		return
	}
}

func (r *dnssecRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 5)
	if len(parts) != 5 {
		resp.Diagnostics.AddError(
			"Unerwarteter Import-Bezeichner",
			"Erwarteter ID im Format 'domain/keyTag/algorithm/digestType/digest'.",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key_tag"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("algorithm"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("digest_type"), parts[3])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("digest"), parts[4])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
