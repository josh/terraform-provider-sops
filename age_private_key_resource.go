package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &AgePrivateKeyResource{}
var _ resource.ResourceWithImportState = &AgePrivateKeyResource{}

func NewAgePrivateKeyResource() resource.Resource {
	return &AgePrivateKeyResource{}
}

type AgePrivateKeyResource struct{}

type AgePrivateKeyResourceModel struct {
	PrivateKey types.String `tfsdk:"private_key"`
	PublicKey  types.String `tfsdk:"public_key"`
}

func (r *AgePrivateKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_age_private_key"
}

func (r *AgePrivateKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Generates an age X25519 key pair. The private key is stored **unencrypted** in the Terraform state; use the `sops_age_private_key` ephemeral resource if the private key should not be persisted. Regenerate the key pair with `terraform apply -replace`. An existing key can be adopted with `terraform import` using the private key as the import ID.",

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

func (r *AgePrivateKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	privateKey, publicKey, err := generateAgeKeyPair()
	if err != nil {
		resp.Diagnostics.AddError(
			"Age Key Generation Failed",
			fmt.Sprintf("Failed to generate age key pair: %s", err),
		)
		return
	}

	data := AgePrivateKeyResourceModel{
		PrivateKey: types.StringValue(privateKey),
		PublicKey:  types.StringValue(publicKey),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AgePrivateKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *AgePrivateKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Unexpected Update Call",
		"This resource does not support updates. The key pair can only be regenerated via replacement.",
	)
}

func (r *AgePrivateKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *AgePrivateKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	privateKey := strings.TrimSpace(req.ID)

	publicKey, err := deriveAgePublicKey(privateKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Age Private Key",
			fmt.Sprintf("The import ID must be an age private key in AGE-SECRET-KEY-1... format. Failed to derive public key: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("private_key"), privateKey)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("public_key"), publicKey)...)
}
