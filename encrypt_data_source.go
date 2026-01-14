package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &EncryptDataSource{}

func NewEncryptDataSource() datasource.DataSource {
	return &EncryptDataSource{}
}

type EncryptDataSource struct{}

type EncryptDataSourceModel struct {
	Input             types.Dynamic `tfsdk:"input"`
	Age               types.List    `tfsdk:"age_recipients"`
	OutputType        types.String  `tfsdk:"output_type"`
	OutputIndent      types.Int64   `tfsdk:"output_indent"`
	UnencryptedSuffix types.String  `tfsdk:"unencrypted_suffix"`
	EncryptedSuffix   types.String  `tfsdk:"encrypted_suffix"`
	UnencryptedRegex  types.String  `tfsdk:"unencrypted_regex"`
	EncryptedRegex    types.String  `tfsdk:"encrypted_regex"`
	Output            types.String  `tfsdk:"output"`
}

func (d *EncryptDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_encrypt"
}

func (d *EncryptDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Encrypts data using SOPS with Age encryption",
		Attributes: map[string]schema.Attribute{
			"input": schema.DynamicAttribute{
				MarkdownDescription: "The data structure to encrypt. Must be a map/object with string keys. Will be automatically converted to JSON before encryption.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.Dynamic{
					dynamicObjectValidator{},
				},
			},
			"age_recipients": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of age recipients to encrypt the data for. Each recipient can decrypt the encrypted output with their corresponding age identity.",
				Required:            true,
			},
			"output_type": schema.StringAttribute{
				MarkdownDescription: "The output format for the encrypted data. Valid values are \"json\" or \"yaml\". Defaults to \"json\".",
				Optional:            true,
			},
			"output_indent": schema.Int64Attribute{
				MarkdownDescription: "Number of spaces to indent the encrypted output.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"unencrypted_suffix": schema.StringAttribute{
				MarkdownDescription: "Override the unencrypted key suffix. Keys with this suffix will not be encrypted.",
				Optional:            true,
			},
			"encrypted_suffix": schema.StringAttribute{
				MarkdownDescription: "Override the encrypted key suffix. When set, only keys with this suffix will be encrypted.",
				Optional:            true,
			},
			"unencrypted_regex": schema.StringAttribute{
				MarkdownDescription: "Set the unencrypted key regex. When specified, only keys matching this regex will be left unencrypted.",
				Optional:            true,
			},
			"encrypted_regex": schema.StringAttribute{
				MarkdownDescription: "Set the encrypted key regex. When specified, only keys matching this regex will be encrypted.",
				Optional:            true,
			},
			"output": schema.StringAttribute{
				MarkdownDescription: "The encrypted data as a raw string (JSON or YAML serialized). Contains the original structure with encrypted values (ENC[...]) and SOPS metadata. Use `jsondecode()` or `yamldecode()` to parse the output string.",
				Computed:            true,
			},
		},
	}
}

func (d *EncryptDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EncryptDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ageRecipients []string
	resp.Diagnostics.Append(data.Age.ElementsAs(ctx, &ageRecipients, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inputValue, err := convertDynamicValueToGo(data.Input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Value Conversion Failed",
			fmt.Sprintf("Failed to convert input to Go value: %s", err),
		)
		return
	}

	inputMap := inputValue.(map[string]interface{})

	outputType := "json"
	if !data.OutputType.IsNull() && !data.OutputType.IsUnknown() {
		outputType = data.OutputType.ValueString()
	}
	if outputType == "" {
		outputType = "json"
	}

	var outputIndent *int64
	if !data.OutputIndent.IsNull() && !data.OutputIndent.IsUnknown() {
		value := data.OutputIndent.ValueInt64()
		outputIndent = &value
	}

	var unencryptedSuffix *string
	if !data.UnencryptedSuffix.IsNull() && !data.UnencryptedSuffix.IsUnknown() {
		value := data.UnencryptedSuffix.ValueString()
		unencryptedSuffix = &value
	}

	var encryptedSuffix *string
	if !data.EncryptedSuffix.IsNull() && !data.EncryptedSuffix.IsUnknown() {
		value := data.EncryptedSuffix.ValueString()
		encryptedSuffix = &value
	}

	var unencryptedRegex *string
	if !data.UnencryptedRegex.IsNull() && !data.UnencryptedRegex.IsUnknown() {
		value := data.UnencryptedRegex.ValueString()
		unencryptedRegex = &value
	}

	var encryptedRegex *string
	if !data.EncryptedRegex.IsNull() && !data.EncryptedRegex.IsUnknown() {
		value := data.EncryptedRegex.ValueString()
		encryptedRegex = &value
	}

	encryptedBytes, err := encryptWithSops(ctx, inputMap, SopsEncryptOptions{
		AgeRecipients:     ageRecipients,
		OutputType:        outputType,
		OutputIndent:      outputIndent,
		UnencryptedSuffix: unencryptedSuffix,
		EncryptedSuffix:   encryptedSuffix,
		UnencryptedRegex:  unencryptedRegex,
		EncryptedRegex:    encryptedRegex,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"SOPS Encryption Failed",
			fmt.Sprintf("Failed to encrypt content: %s", err),
		)
		return
	}

	data.Output = types.StringValue(string(encryptedBytes))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
