package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func convertDynamicValueToGo(dyn types.Dynamic) (interface{}, error) {
	underlyingVal := dyn.UnderlyingValue()

	goVal, err := convertAttrValueToGo(underlyingVal)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to Go value: %w", err)
	}

	return goVal, nil
}

func marshalDynamicValue(dyn types.Dynamic) ([]byte, error) {
	goVal, err := convertDynamicValueToGo(dyn)
	if err != nil {
		return nil, err
	}

	return json.Marshal(goVal)
}

func convertDynamicValueToBytes(dyn types.Dynamic) ([]byte, error) {
	underlyingVal := dyn.UnderlyingValue()

	if strVal, ok := underlyingVal.(types.String); ok {
		if strVal.IsNull() || strVal.IsUnknown() {
			return nil, fmt.Errorf("cannot convert null or unknown string value to bytes")
		}
		return []byte(strVal.ValueString()), nil
	}

	return marshalDynamicValue(dyn)
}

func convertAttrValueToGo(val attr.Value) (interface{}, error) {
	if val.IsNull() {
		return nil, nil
	}
	if val.IsUnknown() {
		return nil, fmt.Errorf("cannot convert unknown value to JSON")
	}

	switch v := val.(type) {
	case types.String:
		return v.ValueString(), nil
	case types.Number:
		bigFloat := v.ValueBigFloat()
		f, accuracy := bigFloat.Float64()
		if accuracy == big.Below && f == 0 {
			return nil, fmt.Errorf("number underflow: value too small to represent as float64")
		}
		if accuracy == big.Above && (f == math.Inf(1) || f == math.Inf(-1)) {
			return nil, fmt.Errorf("number overflow: value too large to represent as float64")
		}
		return f, nil
	case types.Bool:
		return v.ValueBool(), nil
	case types.Object:
		result := make(map[string]interface{})
		for key, attrVal := range v.Attributes() {
			goVal, err := convertAttrValueToGo(attrVal)
			if err != nil {
				return nil, err
			}
			result[key] = goVal
		}
		return result, nil
	case types.List:
		elements := v.Elements()
		result := make([]interface{}, len(elements))
		for i, elem := range elements {
			goVal, err := convertAttrValueToGo(elem)
			if err != nil {
				return nil, err
			}
			result[i] = goVal
		}
		return result, nil
	case types.Map:
		elements := v.Elements()
		result := make(map[string]interface{})
		for key, elem := range elements {
			goVal, err := convertAttrValueToGo(elem)
			if err != nil {
				return nil, err
			}
			result[key] = goVal
		}
		return result, nil
	case types.Tuple:
		elements := v.Elements()
		result := make([]interface{}, len(elements))
		for i, elem := range elements {
			goVal, err := convertAttrValueToGo(elem)
			if err != nil {
				return nil, err
			}
			result[i] = goVal
		}
		return result, nil
	case types.Set:
		elements := v.Elements()
		result := make([]interface{}, len(elements))
		for i, elem := range elements {
			goVal, err := convertAttrValueToGo(elem)
			if err != nil {
				return nil, err
			}
			result[i] = goVal
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported attr.Value type: %T", v)
	}
}

func unmarshalToDynamicValue(jsonBytes []byte) (types.Dynamic, error) {
	var raw interface{}
	if err := json.Unmarshal(jsonBytes, &raw); err != nil {
		return types.DynamicNull(), fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	attrVal, err := convertGoValueToAttr(raw)
	if err != nil {
		return types.DynamicNull(), fmt.Errorf("failed to convert to attr.Value: %w", err)
	}

	return types.DynamicValue(attrVal), nil
}

func convertGoValueToAttr(val interface{}) (attr.Value, error) {
	if val == nil {
		return types.StringNull(), nil
	}

	switch v := val.(type) {
	case string:
		return types.StringValue(v), nil
	case float64:
		return types.NumberValue(big.NewFloat(v)), nil
	case bool:
		return types.BoolValue(v), nil
	case map[string]interface{}:
		attrs := make(map[string]attr.Value)
		attrTypes := make(map[string]attr.Type)
		for key, value := range v {
			attrVal, err := convertGoValueToAttr(value)
			if err != nil {
				return nil, err
			}
			attrs[key] = attrVal
			attrTypes[key] = attrVal.Type(context.Background())
		}
		objVal, diags := types.ObjectValue(attrTypes, attrs)
		if diags.HasError() {
			return nil, fmt.Errorf("failed to create object value: %s", diags.Errors()[0].Detail())
		}
		return objVal, nil
	case []interface{}:
		if len(v) == 0 {
			return types.ListValueMust(types.StringType, []attr.Value{}), nil
		}

		elements := make([]attr.Value, len(v))
		elementTypes := make([]attr.Type, len(v))
		var firstType attr.Type
		allSameType := true

		for i, elem := range v {
			attrVal, err := convertGoValueToAttr(elem)
			if err != nil {
				return nil, err
			}
			elements[i] = attrVal
			elementTypes[i] = attrVal.Type(context.Background())

			if i == 0 {
				firstType = elementTypes[i]
			} else if !elementTypes[i].Equal(firstType) {
				allSameType = false
			}
		}

		if allSameType {
			listVal, diags := types.ListValue(firstType, elements)
			if diags.HasError() {
				return nil, fmt.Errorf("failed to create list value: %s", diags.Errors()[0].Detail())
			}
			return listVal, nil
		} else {
			tupleVal, diags := types.TupleValue(elementTypes, elements)
			if diags.HasError() {
				return nil, fmt.Errorf("failed to create tuple value: %s", diags.Errors()[0].Detail())
			}
			return tupleVal, nil
		}
	default:
		return types.StringValue(fmt.Sprintf("%v", v)), nil
	}
}

func hashDynamicValue(dyn types.Dynamic) (string, error) {
	jsonBytes, err := marshalDynamicValue(dyn)
	if err != nil {
		return "", fmt.Errorf("failed to marshal for hashing: %w", err)
	}

	hash := sha256.Sum256(jsonBytes)
	return fmt.Sprintf("%x", hash), nil
}
