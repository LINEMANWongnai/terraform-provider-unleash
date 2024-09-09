package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

func tryUpdateToEmptyStringIfBeforeEmpty(value types.String, valueBefore types.String, setterFn func(types.String)) {
	if value.IsNull() && !valueBefore.IsNull() && len(valueBefore.ValueString()) == 0 {
		setterFn(types.StringValue(""))
	}
}

func tryUpdateToFalseIfBeforeFalse(value types.Bool, valueBefore types.Bool, setterFn func(types.Bool)) {
	if value.IsNull() && !valueBefore.IsNull() && !valueBefore.ValueBool() {
		setterFn(types.BoolValue(false))
	}
}

func isNullArrayAndExistingEmptyArray[T any](current []T, before []T) bool {
	return current == nil && before != nil && len(before) == 0
}

func isNullMapAndExistingEmptyMap(value map[string]types.String, valueBefore map[string]types.String) bool {
	return value == nil && valueBefore != nil && len(valueBefore) == 0
}
