package umbrellaprovider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/olegunza/umbrella-api-go/umbrella"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var (
	_ provider.Provider = &umbrellaProvider{}
)

type umbrellaProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ScaffoldingProviderModel describes the provider data model.
type UmbrellaProviderModel struct {
	Host      types.String `tfsdk:"host"`
	Apikey    types.String `tfsdk:"apikey"`
	Apisecret types.String `tfsdk:"apisecret"`
}

func (p *umbrellaProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "umbrella"
	resp.Version = p.version
}

func (p *umbrellaProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "umbrella API host",
				Optional:            true,
			},
			"apikey": schema.StringAttribute{
				MarkdownDescription: "umbrella API key",
				Optional:            true,
			},
			"apisecret": schema.StringAttribute{
				MarkdownDescription: "umbrella API secret",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *umbrellaProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Umbrella client")
	var config UmbrellaProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Umbrella API Host",
			"The provider cannot create the Umbrella API client as there is an unknown configuration value for the HashiCups API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the UMBRELLA_HOST environment variable.",
		)
	}

	if config.Apikey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("apikey"),
			"Unknown Umbrella API key",
			"The provider cannot create the Umbrella API client as there is an unknown configuration value for the Umbrella API key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the UMBRELLA_APIKEY environment variable.",
		)
	}

	if config.Apisecret.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("apisecret"),
			"Unknown Umbrella API Secret",
			"The provider cannot create the Umbrella API client as there is an unknown configuration value for the Umbrella API secret. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HASHICUPS_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	host := os.Getenv("UMBRELLA_HOST")
	apikey := os.Getenv("UMBRELLA_APIKEY")
	apisecret := os.Getenv("UMBRELLA_APISECRET")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Apikey.IsNull() {
		apikey = config.Apikey.ValueString()
	}

	if !config.Apisecret.IsNull() {
		apisecret = config.Apisecret.ValueString()
	}

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Umbrella API Host",
			"The provider cannot create the Umbrella API client as there is a missing or empty value for the Umbrella API host. "+
				"Set the host value in the configuration or use the UMBRELLA_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if apikey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("apikey"),
			"Missing Umbrella API Key",
			"The provider cannot create the Umbrella API client as there is a missing or empty value for the Umbrella API key. "+
				"Set the username value in the configuration or use the UMBRELLA_APIKEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if apisecret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("apisecret"),
			"Missing Umbrella API Secret",
			"The provider cannot create the Umbrella API client as there is a missing or empty value for the Umbrella API secret. "+
				"Set the password value in the configuration or use the UMBRELLA_APISECRET environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := umbrella.NewClient(&host, &apikey, &apisecret)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Umbrella API Client",
			"An unexpected error occurred when creating the Umbrella API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Umbrella Client Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client

}

func (p *umbrellaProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExampleResource,
		NewSiteResource,
		NewVAResource,
		NewTunnelResource,
	}
}

func (p *umbrellaProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewExampleDataSource,
		NewSiteDataSource,
		NewVADataSource,
		NewDClistDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &umbrellaProvider{
			version: version,
		}
	}
}
