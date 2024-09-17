package provider

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/unleash"
)

// Ensure UnleashProvider satisfies various provider interfaces.
var _ provider.Provider = &UnleashProvider{}
var _ provider.ProviderWithFunctions = &UnleashProvider{}

// UnleashProvider defines the provider implementation.
type UnleashProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// UnleashProviderModel describes the provider data model.
type UnleashProviderModel struct {
	BaseURL                  types.String `tfsdk:"base_url"`
	AuthorizationToken       types.String `tfsdk:"authorization"`
	StrategyTitleIgnoreRegEx types.String `tfsdk:"strategy_title_ignore_regexp"`
}

type UnleashProviderData struct {
	Client                   unleash.ClientWithResponsesInterface
	StrategyTitleIgnoreRegEx *regexp.Regexp
}

func (p *UnleashProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "unleash"
	resp.Version = p.version
}

func (p *UnleashProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				MarkdownDescription: "Unleash base URL (everything before `/api`)",
				Optional:            true,
			},
			"authorization": schema.StringAttribute{
				MarkdownDescription: "Authorization token for Unleash API",
				Optional:            true,
				Sensitive:           true,
			},
			"strategy_title_ignore_regexp": schema.StringAttribute{
				MarkdownDescription: "Regular expression to ignore strategies by title. The matched strategies will not be managed by this provider.",
				Optional:            true,
			},
		},
	}
}

func (p *UnleashProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data UnleashProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }
	var providerData UnleashProviderData
	var err error
	providerData.Client, err = unleash.CreateClient(data.BaseURL.ValueString(), data.AuthorizationToken.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to create unleash", err.Error())
		return
	}

	if data.StrategyTitleIgnoreRegEx.ValueString() != "" {
		providerData.StrategyTitleIgnoreRegEx, err = regexp.Compile(data.StrategyTitleIgnoreRegEx.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("failed to create unleash", err.Error())
		}
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *UnleashProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewFeatureResource,
		NewSegmentResource,
	}
}

func (p *UnleashProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *UnleashProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &UnleashProvider{
			version: version,
		}
	}
}
