// internal/provider/ping_data_source.go
package provider

import (
	"context"
	"fmt"

	"github.com/flooopro/terraform-provider-porkbun/internal/porkbun"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &pingDataSource{}
	_ datasource.DataSourceWithConfigure = &pingDataSource{}
)

func NewPingDataSource() datasource.DataSource {
	return &pingDataSource{}
}

type pingDataSource struct {
	client *porkbun.Client
}

type pingDataSourceModel struct {
	ID     types.String `tfsdk:"id"`
	YourIP types.String `tfsdk:"your_ip"`
}

func (d *pingDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ping"
}

func (d *pingDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Führt einen Ping-Test gegen die Porkbun-API durch und gibt Ihre IP-Adresse zurück.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Ein eindeutiger Bezeichner für diese Data Source. Wird auf Ihre IP-Adresse gesetzt.",
				Computed:    true,
			},
			"your_ip": schema.StringAttribute{
				Description: "Die öffentliche IP-Adresse, die von der Porkbun-API gesehen wird.",
				Computed:    true,
			},
		},
	}
}

func (d *pingDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*porkbun.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unerwarteter Konfigurationstyp für Data Source",
			fmt.Sprintf("Erwartet wurde *porkbun.Client, erhalten: %T. Bitte melden Sie dieses Problem.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *pingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state pingDataSourceModel

	ip, err := d.client.Ping(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Ping an Porkbun-API fehlgeschlagen",
			"Konnte den Ping-Endpunkt nicht aufrufen: "+err.Error(),
		)
		return
	}

	state.YourIP = types.StringValue(ip)
	state.ID = types.StringValue(ip)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
