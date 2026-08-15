package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &AgePublicKeyResource{}

func NewAgePublicKeyResource() resource.Resource {
	return &AgePublicKeyResource{}
}

type AgePublicKeyResource struct{}

type AgePublicKeyResourceModel struct {
	PrivateKeyWo        types.String `tfsdk:"private_key_wo"`
	PrivateKeyWoVersion types.Int64  `tfsdk:"private_key_wo_version"`
	PublicKey           types.String `tfsdk:"public_key"`
}

func (r *AgePublicKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_age_public_key"
}

func (r *AgePublicKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Derives an age public key from a private key supplied via a write-only argument, persisting **only the public key** in Terraform state. Accepts ephemeral values, such as the private key of the `sops_age_private_key` ephemeral resource. Requires Terraform 1.11 or later.",

		Attributes: map[string]schema.Attribute{
			"private_key_wo": schema.StringAttribute{
				MarkdownDescription: "Age private key (identity) in `AGE-SECRET-KEY-1...` format. Write-only: never persisted in state or plan. Because Terraform cannot detect changes to write-only values, increment `private_key_wo_version` when supplying a different key.",
				Required:            true,
				WriteOnly:           true,
				Sensitive:           true,
				Validators: []validator.String{
					ageIdentityValidator{},
				},
			},
			"private_key_wo_version": schema.Int64Attribute{
				MarkdownDescription: "Version counter for `private_key_wo`. Increment this value to re-derive the public key after changing the private key.",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"public_key": schema.StringAttribute{
				MarkdownDescription: "Derived age public key (recipient) in `age1...` format.",
				Computed:            true,
			},
		},
	}
}

func (r *AgePublicKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AgePublicKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Write-only values are only available from the configuration, never the plan.
	var privateKey types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("private_key_wo"), &privateKey)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if privateKey.IsNull() || privateKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("private_key_wo"),
			"Missing Age Private Key",
			"The \"private_key_wo\" attribute must be a known value during apply.",
		)
		return
	}

	publicKey, err := deriveAgePublicKey(privateKey.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("private_key_wo"),
			"Invalid Age Private Key",
			fmt.Sprintf("Failed to derive public key: %s", err),
		)
		return
	}

	data.PublicKey = types.StringValue(publicKey)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AgePublicKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *AgePublicKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Unexpected Update Call",
		"This resource does not support updates. Changes to 'private_key_wo_version' should trigger replacement.",
	)
}

func (r *AgePublicKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
