package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ ephemeral.EphemeralResource = &DecryptEphemeralResource{}
var _ ephemeral.EphemeralResourceWithConfigure = &DecryptEphemeralResource{}

func NewDecryptEphemeralResource() ephemeral.EphemeralResource {
	return &DecryptEphemeralResource{}
}

type DecryptEphemeralResource struct {
	client *SopsProviderConfig
}

type DecryptEphemeralResourceModel struct {
	Input     types.Dynamic `tfsdk:"input"`
	InputType types.String  `tfsdk:"input_type"`
	Output    types.Dynamic `tfsdk:"output"`
}

func (r *DecryptEphemeralResource) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_decrypt"
}

func (r *DecryptEphemeralResource) Schema(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Decrypts SOPS-encrypted data without storing plaintext in Terraform state.",

		Attributes: map[string]schema.Attribute{
			"input": schema.DynamicAttribute{
				MarkdownDescription: "The encrypted data structure to decrypt.",
				Required:            true,
				Sensitive:           true,
			},
			"input_type": schema.StringAttribute{
				MarkdownDescription: "The format of the encrypted input. Valid values are \"json\" or \"yaml\". Defaults to \"json\".",
				Optional:            true,
			},
			"output": schema.DynamicAttribute{
				MarkdownDescription: "The decrypted data structure.",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (r *DecryptEphemeralResource) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(*SopsProviderConfig)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Ephemeral Resource Configure Type",
			fmt.Sprintf("Expected *SopsProviderConfig, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = config
}

func (r *DecryptEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data DecryptEphemeralResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Ephemeral Resource",
			"Expected configured SopsProviderConfig. Please report this issue to the provider developers.",
		)
		return
	}

	inputBytes, err := convertDynamicValueToBytes(data.Input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Input Conversion Failed",
			fmt.Sprintf("Failed to convert input to bytes: %s", err),
		)
		return
	}

	var ageIdentityPath, ageIdentityValue string
	if !r.client.AgeIdentityPath.IsNull() {
		ageIdentityPath = r.client.AgeIdentityPath.ValueString()
	}
	if !r.client.AgeIdentityValue.IsNull() {
		ageIdentityValue = r.client.AgeIdentityValue.ValueString()
	}

	inputType := "json"
	if !data.InputType.IsNull() && !data.InputType.IsUnknown() {
		inputType = data.InputType.ValueString()
	}
	if inputType == "" {
		inputType = "json"
	}

	decryptedJSON, err := decryptWithSops(ctx, inputBytes, SopsDecryptOptions{
		AgeIdentityPath:  ageIdentityPath,
		AgeIdentityValue: ageIdentityValue,
		InputType:        inputType,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"SOPS Decryption Failed",
			fmt.Sprintf("Failed to decrypt content: %s", err),
		)
		return
	}

	outputValue, err := unmarshalToDynamicValue(decryptedJSON)
	if err != nil {
		resp.Diagnostics.AddError(
			"JSON Parsing Failed",
			fmt.Sprintf("Failed to parse SOPS decrypted output: %s", err),
		)
		return
	}

	data.Output = outputValue

	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
