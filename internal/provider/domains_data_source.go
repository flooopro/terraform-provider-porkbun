// internal/provider/domains_data_source.go
package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/flooopro/terraform-provider-porkbun/internal/porkbun"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &domainsDataSource{}
var _ datasource.DataSourceWithConfigure = &domainsDataSource{}

func NewDomainsDataSource() datasource.DataSource {
	return &domainsDataSource{}
}

type domainsDataSource struct {
	client *porkbun.Client
}

type domainsDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	Domains types.List   `tfsdk:"domains"`
}

func domainListingAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"domain":        types.StringType,
		"status":        types.StringType,
		"tld":           types.StringType,
		"create_date":   types.StringType,
		"expire_date":   types.StringType,
		"security_lock": types.BoolType,
		"whois_privacy": types.BoolType,
		"auto_renew":    types.BoolType,
	}
}

func toBool(v interface{}) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val == "true"
	case float64:
		return val == 1
	default:
		return false
	}
}

func (d *domainsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domains"
}

func (d *domainsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Ruft eine Liste aller Domains ab, die sich im Porkbun-Konto befinden.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Ein eindeutiger Bezeichner, der auf den aktuellen Zeitstempel gesetzt wird.",
				Computed:    true,
			},
			"domains": schema.ListNestedAttribute{
				Description: "Die Liste der Domains im Konto.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"domain":        schema.StringAttribute{Computed: true},
						"status":        schema.StringAttribute{Computed: true},
						"tld":           schema.StringAttribute{Computed: true},
						"create_date":   schema.StringAttribute{Computed: true},
						"expire_date":   schema.StringAttribute{Computed: true},
						"security_lock": schema.BoolAttribute{Computed: true},
						"whois_privacy": schema.BoolAttribute{Computed: true},
						"auto_renew":    schema.BoolAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *domainsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*porkbun.Client)
	if !ok {
		resp.Diagnostics.AddError("Unerwarteter Konfigurationstyp", fmt.Sprintf("Erwartet *porkbun.Client, erhalten: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *domainsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	domainListings, err := d.client.ListAllDomains()
	if err != nil {
		resp.Diagnostics.AddError("Fehler beim Abrufen der Domain-Liste", err.Error())
		return
	}

	var domainModels []attr.Value
	for _, domain := range domainListings {

		domainModels = append(domainModels, types.ObjectValueMust(
			domainListingAttributeTypes(),
			map[string]attr.Value{
				"domain":        types.StringValue(domain.Domain),
				"status":        types.StringValue(domain.Status),
				"tld":           types.StringValue(domain.Tld),
				"create_date":   types.StringValue(domain.CreateDate),
				"expire_date":   types.StringValue(domain.ExpireDate),
				"security_lock": types.BoolValue(toBool(domain.SecurityLock)),
				"whois_privacy": types.BoolValue(toBool(domain.WhoisPrivacy)),
				"auto_renew":    types.BoolValue(toBool(domain.AutoRenew)),
			},
		))
	}

	var state domainsDataSourceModel
	state.ID = types.StringValue(time.Now().UTC().String())
	state.Domains = types.ListValueMust(types.ObjectType{AttrTypes: domainListingAttributeTypes()}, domainModels)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
