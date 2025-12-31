package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/dynamicplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &EncryptResource{}

func NewEncryptResource() resource.Resource {
	return &EncryptResource{}
}

type EncryptResource struct {
	client *SopsProviderConfig
}

type EncryptResourceModel struct {
	Input          types.Dynamic `tfsdk:"input"`
	InputWO        types.Dynamic `tfsdk:"input_wo"`
	InputWOVersion types.String  `tfsdk:"input_wo_version"`
	InputHash      types.String  `tfsdk:"input_hash"`
	Age            types.List    `tfsdk:"age_recipients"`
	OutputType     types.String  `tfsdk:"output_type"`
	OutputIndent   types.Int64   `tfsdk:"output_indent"`
	Output         types.String  `tfsdk:"output"`
}

func (r *EncryptResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_encrypt"
}

func (r *EncryptResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Encrypts data using SOPS with Age encryption and manages it as a resource",

		Attributes: map[string]schema.Attribute{
			"input": schema.DynamicAttribute{
				MarkdownDescription: "Data structure to encrypt. Must be a map/object with string keys.",
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.Dynamic{
					dynamicObjectValidator{},
				},
				PlanModifiers: []planmodifier.Dynamic{
					dynamicplanmodifier.RequiresReplace(),
				},
			},
			"input_wo": schema.DynamicAttribute{
				MarkdownDescription: "Write-only data structure to encrypt. Plaintext is not stored in state. Must be a map/object with string keys.",
				Optional:            true,
				Sensitive:           true,
				WriteOnly:           true,
				Validators: []validator.Dynamic{
					dynamicObjectValidator{},
				},
				PlanModifiers: []planmodifier.Dynamic{
					dynamicplanmodifier.RequiresReplace(),
				},
			},
			"input_wo_version": schema.StringAttribute{
				MarkdownDescription: "Version string to trigger re-encryption when using input_wo.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"input_hash": schema.StringAttribute{
				MarkdownDescription: "Hash for automatic change detection with input_wo.",
				Computed:            true,
			},
			"age_recipients": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Age recipients for encryption. Each recipient can decrypt the output with their corresponding identity.",
				Required:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"output_type": schema.StringAttribute{
				MarkdownDescription: "Output format for encrypted data. Valid values are \"json\" or \"yaml\". Defaults to \"json\".",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"output_indent": schema.Int64Attribute{
				MarkdownDescription: "Number of spaces to indent the encrypted output.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"output": schema.StringAttribute{
				MarkdownDescription: "Encrypted data as serialized JSON or YAML string containing encrypted values and SOPS metadata.",
				Computed:            true,
			},
		},
	}
}

func (r *EncryptResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(*SopsProviderConfig)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *SopsProviderConfig, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = config
}

func (r *EncryptResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		encryptResourceValidator{},
		resourcevalidator.PreferWriteOnlyAttribute(
			path.MatchRoot("input"),
			path.MatchRoot("input_wo"),
		),
	}
}

func (r *EncryptResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() {
		var plan EncryptResourceModel
		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}

		var config EncryptResourceModel
		resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
		if resp.Diagnostics.HasError() {
			return
		}

		hasUnknownInput := false

		if !config.InputWO.IsNull() {
			if config.InputWO.IsUnknown() {
				hasUnknownInput = true
			} else {
				if containsUnknownValues(config.InputWO) {
					hasUnknownInput = true
				}
			}
		} else if !config.Input.IsNull() {
			if config.Input.IsUnknown() {
				hasUnknownInput = true
			} else {
				if containsUnknownValues(config.Input) {
					hasUnknownInput = true
				}
			}
		}

		if hasUnknownInput {
			plan.Output = types.StringUnknown()

			if !config.InputWO.IsNull() && config.InputWOVersion.IsNull() {
				plan.InputHash = types.StringUnknown()
			}

			resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
		}
	}
}

func (r *EncryptResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EncryptResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var configData EncryptResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &configData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var inputValue interface{}
	var err error

	if !configData.InputWO.IsNull() {
		inputValue, err = convertDynamicValueToGo(configData.InputWO)
		if err != nil {
			resp.Diagnostics.AddError(
				"Value Conversion Failed",
				fmt.Sprintf("Failed to convert input_wo to Go value: %s", err),
			)
			return
		}

		if !configData.InputWOVersion.IsNull() {
			data.InputHash = types.StringNull()
		} else {
			hash, err := hashDynamicValue(configData.InputWO)
			if err != nil {
				resp.Diagnostics.AddError(
					"Hash Generation Failed",
					fmt.Sprintf("Failed to generate hash for change detection: %s", err),
				)
				return
			}
			data.InputHash = types.StringValue(hash)
		}
	} else {
		if data.Input.IsNull() {
			resp.Diagnostics.AddError(
				"Missing Input",
				"Either 'input' or 'input_wo' must be provided.",
			)
			return
		}

		inputValue, err = convertDynamicValueToGo(data.Input)
		if err != nil {
			resp.Diagnostics.AddError(
				"Value Conversion Failed",
				fmt.Sprintf("Failed to convert input to Go value: %s", err),
			)
			return
		}

		data.InputHash = types.StringNull()
	}

	inputMap := inputValue.(map[string]interface{})

	var ageRecipients []string
	resp.Diagnostics.Append(data.Age.ElementsAs(ctx, &ageRecipients, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	encryptedBytes, err := encryptWithSops(ctx, inputMap, SopsEncryptOptions{
		AgeRecipients: ageRecipients,
		OutputType:    outputType,
		OutputIndent:  outputIndent,
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

func (r *EncryptResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EncryptResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EncryptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Unexpected Update Call",
		"This resource does not support updates. Changes to 'input' or 'age' should trigger replacement.",
	)
}

func (r *EncryptResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EncryptResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
