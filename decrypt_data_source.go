package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DecryptDataSource{}

func NewDecryptDataSource() datasource.DataSource {
	return &DecryptDataSource{}
}

type DecryptDataSource struct {
	client *SopsProviderConfig
}

type DecryptDataSourceModel struct {
	Input     types.Dynamic `tfsdk:"input"`
	InputType types.String  `tfsdk:"input_type"`
	Output    types.Dynamic `tfsdk:"output"`
}

func (d *DecryptDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_decrypt"
}

func (d *DecryptDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Decrypts data using SOPS with Age decryption",

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

func (d *DecryptDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(*SopsProviderConfig)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *SopsProviderConfig, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = config
}

func (d *DecryptDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DecryptDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
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
	if d.client != nil {
		if !d.client.AgeIdentityPath.IsNull() {
			ageIdentityPath = d.client.AgeIdentityPath.ValueString()
		}
		if !d.client.AgeIdentityValue.IsNull() {
			ageIdentityValue = d.client.AgeIdentityValue.ValueString()
		}
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
