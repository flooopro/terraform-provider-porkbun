package provider

import (
	"context"
	"os"

	"github.com/flooopro/terraform-provider-porkbun/internal/porkbun"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &PorkbunProvider{}

type PorkbunProvider struct {
	version string
}

type PorkbunProviderModel struct {
	APIKey    types.String `tfsdk:"api_key"`
	SecretKey types.String `tfsdk:"secret_api_key"`
}

func (p *PorkbunProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "porkbun"
	resp.Version = p.version
}

func (p *PorkbunProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing Porkbun resources.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Porkbun API Key. May be provided via PORKBUN_API_KEY environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"secret_api_key": schema.StringAttribute{
				MarkdownDescription: "Porkbun Secret API Key. May be provided via PORKBUN_SECRET_API_KEY environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *PorkbunProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data PorkbunProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("PORKBUN_API_KEY")
	secretKey := os.Getenv("PORKBUN_SECRET_API_KEY")

	if !data.APIKey.IsNull() {
		apiKey = data.APIKey.ValueString()
	}

	if !data.SecretKey.IsNull() {
		secretKey = data.SecretKey.ValueString()
	}

	if apiKey == "" || secretKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Credentials",
			"The provider cannot be configured without an API key and secret API key. "+
				"Set the credentials in the provider block or via PORKBUN_API_KEY and PORKBUN_SECRET_API_KEY environment variables.",
		)
		return
	}

	client := porkbun.NewClient(apiKey, secretKey)
	resp.ResourceData = client
	resp.DataSourceData = client
}

func (p *PorkbunProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDnsRecordResource,
		NewDomainNameserversResource,
		NewGlueRecordResource,
		NewDnssecRecordResource,
	}
}

func (p *PorkbunProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewPingDataSource,
		NewDnsRecordsDataSource,
		NewTldsDataSource,
		NewDomainsDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PorkbunProvider{
			version: version,
		}
	}
}
