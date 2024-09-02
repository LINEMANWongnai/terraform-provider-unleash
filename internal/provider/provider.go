package provider

import (
	"context"

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
	BaseURL            types.String `tfsdk:"base_url"`
	AuthorizationToken types.String `tfsdk:"authorization"`
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

	c, err := unleash.CreateClient(data.BaseURL.ValueString(), data.AuthorizationToken.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to create unleash", err.Error())
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *UnleashProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewFeatureResource,
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
