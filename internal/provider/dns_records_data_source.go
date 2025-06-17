package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/flooopro/terraform-provider-porkbun/internal/porkbun"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &dnsRecordsDataSource{}
	_ datasource.DataSourceWithConfigure = &dnsRecordsDataSource{}
)

func NewDnsRecordsDataSource() datasource.DataSource {
	return &dnsRecordsDataSource{}
}

type dnsRecordsDataSource struct {
	client *porkbun.Client
}

type dnsRecordsDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	Domain  types.String `tfsdk:"domain"`
	Records types.List   `tfsdk:"records"`
}

func recordAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":      types.StringType,
		"name":    types.StringType,
		"type":    types.StringType,
		"content": types.StringType,
		"ttl":     types.StringType,
		"prio":    types.StringType,
	}
}

func (d *dnsRecordsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_records"
}

func (d *dnsRecordsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Ruft alle DNS-Einträge für eine bestimmte Domain ab.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Wird auf den Domainnamen gesetzt.",
				Computed:    true,
			},
			"domain": schema.StringAttribute{
				Description: "Der Domainname, für den die Einträge abgerufen werden sollen.",
				Required:    true,
			},
			"records": schema.ListNestedAttribute{
				Description: "Die Liste der DNS-Einträge für die Domain.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":      schema.StringAttribute{Computed: true},
						"name":    schema.StringAttribute{Computed: true},
						"type":    schema.StringAttribute{Computed: true},
						"content": schema.StringAttribute{Computed: true},
						"ttl":     schema.StringAttribute{Computed: true},
						"prio":    schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *dnsRecordsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*porkbun.Client)
	if !ok {
		resp.Diagnostics.AddError("Unerwarteter Konfigurationstyp für Data Source", fmt.Sprintf("Erwartet wurde *porkbun.Client, erhalten: %T.", req.ProviderData))
		return
	}
	d.client = client
}

func (d *dnsRecordsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config dnsRecordsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := config.Domain.ValueString()
	records, err := d.client.RetrieveRecords(domain)
	if err != nil {
		resp.Diagnostics.AddError("Fehler beim Abrufen der DNS-Einträge", fmt.Sprintf("Konnte Einträge für Domain %s nicht abrufen: %s", domain, err.Error()))
		return
	}

	var state dnsRecordsDataSourceModel
	state.Domain = types.StringValue(domain)
	state.ID = types.StringValue(domain)

	var recordModels []attr.Value
	for _, record := range records {
		recordName := record.Name
		normalizedName := strings.TrimSuffix(recordName, "."+domain)
		if normalizedName == domain {
			normalizedName = ""
		}

		recordModels = append(recordModels, types.ObjectValueMust(
			recordAttributeTypes(),
			map[string]attr.Value{
				"id":      types.StringValue(record.ID),
				"name":    types.StringValue(normalizedName),
				"type":    types.StringValue(record.Type),
				"content": types.StringValue(record.Content),
				"ttl":     types.StringValue(record.TTL),
				"prio":    types.StringValue(record.Prio),
			},
		))
	}
	state.Records = types.ListValueMust(types.ObjectType{AttrTypes: recordAttributeTypes()}, recordModels)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
