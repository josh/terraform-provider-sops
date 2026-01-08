package main

import (
	"context"
	"fmt"

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
