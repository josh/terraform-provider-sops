package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type dynamicObjectValidator struct{}

func (v dynamicObjectValidator) Description(ctx context.Context) string {
	return "value must be an object/map"
}

func (v dynamicObjectValidator) MarkdownDescription(ctx context.Context) string {
	return "value must be an object/map"
}

func (v dynamicObjectValidator) ValidateDynamic(ctx context.Context, req validator.DynamicRequest, resp *validator.DynamicResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if containsUnknownValues(req.ConfigValue) {
		return
	}

	inputValue, err := convertDynamicValueToGo(req.ConfigValue)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Value Conversion Failed",
			fmt.Sprintf("Failed to convert input to Go value: %s", err),
		)
		return
	}

	if _, ok := inputValue.(map[string]interface{}); !ok {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Input Type",
			fmt.Sprintf("Input must be a map/object, got %T. SOPS can only encrypt JSON objects.", inputValue),
		)
		return
	}
}

type encryptResourceValidator struct{}

func (v encryptResourceValidator) Description(ctx context.Context) string {
	return "Ensures only one of 'input' or 'input_wo' is provided"
}

func (v encryptResourceValidator) MarkdownDescription(ctx context.Context) string {
	return "Ensures only one of `input` or `input_wo` is provided"
}

func (v encryptResourceValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data EncryptResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasInput := !data.Input.IsNull()
	hasInputWO := !data.InputWO.IsNull()

	if hasInput && hasInputWO {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("input"),
			"Both 'input' and 'input_wo' Provided",
			"When both 'input' and 'input_wo' are set, 'input_wo' takes precedence. "+
				"Consider removing 'input' and using only 'input_wo' for better security.",
		)
		return
	}

	if !hasInput && !hasInputWO {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'input' or 'input_wo' must be provided. Use 'input_wo' when passing "+
				"secrets from ephemeral resources to avoid storing plaintext in state.",
		)
		return
	}

	if !data.InputWOVersion.IsNull() && hasInput {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("input_wo_version"),
			"Unused Attribute",
			"The 'input_wo_version' attribute only applies when using 'input_wo', not 'input'.",
		)
	}
}
