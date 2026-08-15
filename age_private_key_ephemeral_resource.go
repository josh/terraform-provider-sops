package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ ephemeral.EphemeralResource = &AgePrivateKeyEphemeralResource{}

func NewAgePrivateKeyEphemeralResource() ephemeral.EphemeralResource {
	return &AgePrivateKeyEphemeralResource{}
}

type AgePrivateKeyEphemeralResource struct{}

type AgePrivateKeyEphemeralResourceModel struct {
	PrivateKey types.String `tfsdk:"private_key"`
	PublicKey  types.String `tfsdk:"public_key"`
}

func (r *AgePrivateKeyEphemeralResource) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_age_private_key"
}

func (r *AgePrivateKeyEphemeralResource) Schema(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Generates an age X25519 key pair without storing it in Terraform state. A new key pair is generated on every plan and apply, so the private key must be written to a write-only argument of a managed resource in the same configuration to be retained. To persist only the public key in state, pass the private key to the `sops_age_public_key` resource's write-only `private_key_wo` argument. Requires Terraform 1.10 or later.",

		Attributes: map[string]schema.Attribute{
			"private_key": schema.StringAttribute{
				MarkdownDescription: "Generated age private key (identity) in `AGE-SECRET-KEY-1...` format.",
				Computed:            true,
				Sensitive:           true,
			},
			"public_key": schema.StringAttribute{
				MarkdownDescription: "Corresponding age public key (recipient) in `age1...` format.",
				Computed:            true,
			},
		},
	}
}

func (r *AgePrivateKeyEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	privateKey, publicKey, err := generateAgeKeyPair()
	if err != nil {
		resp.Diagnostics.AddError(
			"Age Key Generation Failed",
			fmt.Sprintf("Failed to generate age key pair: %s", err),
		)
		return
	}

	data := AgePrivateKeyEphemeralResourceModel{
		PrivateKey: types.StringValue(privateKey),
		PublicKey:  types.StringValue(publicKey),
	}

	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
