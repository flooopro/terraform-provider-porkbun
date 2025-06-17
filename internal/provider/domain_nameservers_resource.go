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

var _ resource.Resource = &domainNameserversResource{}
var _ resource.ResourceWithConfigure = &domainNameserversResource{}
var _ resource.ResourceWithImportState = &domainNameserversResource{}

var defaultPorkbunNameservers = []string{
	"curia.porkbun.com",
	"livia.porkbun.com",
	"pliny.porkbun.com",
	"salvia.porkbun.com",
}

func NewDomainNameserversResource() resource.Resource {
	return &domainNameserversResource{}
}

type domainNameserversResource struct {
	client *porkbun.Client
}

type domainNameserversResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Domain      types.String `tfsdk:"domain"`
	Nameservers types.List   `tfsdk:"nameservers"`
}

func (r *domainNameserversResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_nameservers"
}

func (r *domainNameserversResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Verwaltet die autoritativen Nameserver für eine Domain bei Porkbun.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				Description: "Der Domainname, dessen Nameserver verwaltet werden.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"nameservers": schema.ListAttribute{
				Description: "Eine Liste der Nameserver für die Domain.",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *domainNameserversResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *domainNameserversResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan domainNameserversResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ns []string
	resp.Diagnostics.Append(plan.Nameservers.ElementsAs(ctx, &ns, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateNameservers(plan.Domain.ValueString(), ns)
	if err != nil {
		resp.Diagnostics.AddError("Fehler beim Erstellen der Nameserver", "Konnte Nameserver nicht aktualisieren: "+err.Error())
		return
	}

	plan.ID = plan.Domain

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *domainNameserversResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state domainNameserversResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()
	foundNs, err := r.client.GetNameservers(domain)
	if err != nil {
		if strings.Contains(err.Error(), "Domain not found") || strings.Contains(err.Error(), "Domain does not exist") {
			tflog.Warn(ctx, "Domain nicht gefunden, wird aus dem State entfernt.", map[string]any{"domain": domain})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Fehler beim Lesen der Nameserver", "Konnte registrierte Nameserver nicht abrufen: "+err.Error())
		return
	}

	if len(foundNs) == 0 {
		tflog.Warn(ctx, "Keine NS-Einträge für die Domain gefunden. Die Ressource wird aus dem State entfernt.", map[string]any{"domain": domain})
		resp.State.RemoveResource(ctx)
		return
	}

	sort.Strings(foundNs)

	state.ID = types.StringValue(domain)
	state.Nameservers, resp.Diagnostics = types.ListValueFrom(ctx, types.StringType, foundNs)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *domainNameserversResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan domainNameserversResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ns []string
	resp.Diagnostics.Append(plan.Nameservers.ElementsAs(ctx, &ns, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateNameservers(plan.Domain.ValueString(), ns)
	if err != nil {
		resp.Diagnostics.AddError("Fehler beim Aktualisieren der Nameserver", "Konnte Nameserver nicht aktualisieren: "+err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *domainNameserversResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state domainNameserversResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Setze Nameserver auf Porkbun-Standard zurück", map[string]any{"domain": state.Domain.ValueString()})
	err := r.client.UpdateNameservers(state.Domain.ValueString(), defaultPorkbunNameservers)
	if err != nil {
		resp.Diagnostics.AddError("Fehler beim Zurücksetzen der Nameserver", "Konnte Nameserver nicht auf Standard zurücksetzen: "+err.Error())
		return
	}
}

func (r *domainNameserversResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
