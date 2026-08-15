package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &AgePublicKeyDataSource{}

func NewAgePublicKeyDataSource() datasource.DataSource {
	return &AgePublicKeyDataSource{}
}

type AgePublicKeyDataSource struct{}

type AgePublicKeyDataSourceModel struct {
	PrivateKey types.String `tfsdk:"private_key"`
	PublicKey  types.String `tfsdk:"public_key"`
}

func (d *AgePublicKeyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_age_public_key"
}

func (d *AgePublicKeyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Derives the age public key from an existing age private key. Note that data source arguments, including the private key, are persisted in the Terraform state; use the `sops_age_public_key` resource with its write-only `private_key_wo` argument to keep the private key out of state.",

		Attributes: map[string]schema.Attribute{
			"private_key": schema.StringAttribute{
				MarkdownDescription: "Age private key (identity) in `AGE-SECRET-KEY-1...` format.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					ageIdentityValidator{},
				},
			},
			"public_key": schema.StringAttribute{
				MarkdownDescription: "Derived age public key (recipient) in `age1...` format.",
				Computed:            true,
			},
		},
	}
}

func (d *AgePublicKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AgePublicKeyDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	publicKey, err := deriveAgePublicKey(data.PrivateKey.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("private_key"),
			"Invalid Age Private Key",
			fmt.Sprintf("Failed to derive public key: %s", err),
		)
		return
	}

	data.PublicKey = types.StringValue(publicKey)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
