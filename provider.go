package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &SopsProvider{}
var _ provider.ProviderWithEphemeralResources = &SopsProvider{}

type SopsProvider struct {
	version string
}

type SopsProviderModel struct {
	AgeIdentityPath  types.String `tfsdk:"age_identity_path"`
	AgeIdentityValue types.String `tfsdk:"age_identity_value"`
}

func (p *SopsProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "sops"
	resp.Version = p.version
}

func (p *SopsProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"age_identity_path": schema.StringAttribute{
				MarkdownDescription: "Path to the age identity file for SOPS decryption.",
				Optional:            true,
			},
			"age_identity_value": schema.StringAttribute{
				MarkdownDescription: "Raw age identity value for SOPS decryption. If both path and value are provided, value takes precedence.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *SopsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data SopsProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	config := &SopsProviderConfig{
		AgeIdentityPath:  data.AgeIdentityPath,
		AgeIdentityValue: data.AgeIdentityValue,
	}

	resp.DataSourceData = config
	resp.ResourceData = config
	resp.EphemeralResourceData = config
}

func (p *SopsProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewEncryptResource,
	}
}

func (p *SopsProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		NewDecryptEphemeralResource,
		NewTestDynamicEphemeralResource,
	}
}

func (p *SopsProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDecryptDataSource,
		NewEncryptDataSource,
	}
}

type SopsProviderConfig struct {
	AgeIdentityPath  types.String
	AgeIdentityValue types.String
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SopsProvider{
			version: version,
		}
	}
}
