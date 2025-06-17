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

var _ datasource.DataSource = &tldsDataSource{}
var _ datasource.DataSourceWithConfigure = &tldsDataSource{}

func NewTldsDataSource() datasource.DataSource {
	return &tldsDataSource{}
}

type tldsDataSource struct {
	client *porkbun.Client
}

type tldsDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Tlds types.Map    `tfsdk:"tlds"`
}

func tldAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"registration_price": types.StringType,
		"renewal_price":      types.StringType,
		"transfer_price":     types.StringType,
		"sla":                types.Float64Type,
	}
}

func (d *tldsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tlds"
}

func (d *tldsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Ruft eine Liste aller von Porkbun unterstützten TLDs und deren Preise ab.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Ein eindeutiger Bezeichner, der auf den aktuellen Zeitstempel gesetzt wird, um bei jeder Ausführung eine Aktualisierung zu erzwingen.",
				Computed:    true,
			},
			"tlds": schema.MapNestedAttribute{
				Description: "Eine Karte von TLDs, wobei der Schlüssel der TLD-Name ist (z.B. 'com').",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"registration_price": schema.StringAttribute{Computed: true},
						"renewal_price":      schema.StringAttribute{Computed: true},
						"transfer_price":     schema.StringAttribute{Computed: true},
						"sla":                schema.Float64Attribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *tldsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *tldsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	pricing, err := d.client.GetPricing()
	if err != nil {
		resp.Diagnostics.AddError("Fehler beim Abrufen der TLD-Preise", err.Error())
		return
	}

	tldMap := make(map[string]attr.Value)
	for tld, prices := range pricing {
		tldMap[tld] = types.ObjectValueMust(
			tldAttributeTypes(),
			map[string]attr.Value{
				"registration_price": types.StringValue(prices.Registration),
				"renewal_price":      types.StringValue(prices.Renewal),
				"transfer_price":     types.StringValue(prices.Transfer),
				"sla":                types.Float64Value(prices.SLA),
			},
		)
	}

	var state tldsDataSourceModel
	state.ID = types.StringValue(time.Now().UTC().String())
	state.Tlds = types.MapValueMust(types.ObjectType{AttrTypes: tldAttributeTypes()}, tldMap)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
