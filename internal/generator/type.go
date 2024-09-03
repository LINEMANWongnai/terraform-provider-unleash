package generator

import "github.com/zclconf/go-cty/cty"

var environmentType cty.Type
var variantType cty.Type
var variantOverrideType cty.Type
var constraintType cty.Type
var strategyVariantType cty.Type

func init() {
	environmentType = createEnvironmentType()
	variantType = createVariantType()
	variantOverrideType = createVariantOverrideType()
	constraintType = createConstraintType()
	strategyVariantType = createStrategyVariantType()
}

func createEnvironmentType() cty.Type {
	return cty.Object(map[string]cty.Type{
		"enabled":    cty.Bool,
		"strategies": cty.Set(createStrategyType()),
		"variants":   cty.Set(createVariantType()),
	})
}

func createVariantType() cty.Type {
	return cty.Object(map[string]cty.Type{
		"name":         cty.String,
		"weight_type":  cty.String,
		"payload":      cty.String,
		"payload_type": cty.String,
		"stickiness":   cty.String,
		"weight":       cty.Number,
		"overrides":    cty.Set(createVariantOverrideType()),
	})
}

func createVariantOverrideType() cty.Type {
	return cty.Object(map[string]cty.Type{
		"context_name": cty.String,
		"values_json":  cty.String,
	})
}

func createStrategyType() cty.Type {
	return cty.Object(map[string]cty.Type{
		"name":        cty.String,
		"disabled":    cty.Bool,
		"title":       cty.String,
		"sort_order":  cty.Number,
		"constraints": cty.Set(createConstraintType()),
		"parameters":  cty.Map(cty.String),
		"segments":    cty.List(cty.Number),
		"variants":    cty.Set(createStrategyVariantType()),
	})
}

func createConstraintType() cty.Type {
	return cty.Object(map[string]cty.Type{
		"case_insensitive": cty.Bool,
		"context_name":     cty.String,
		"operator":         cty.String,
		"inverted":         cty.Bool,
		"values_json":      cty.String,
	})
}

func createStrategyVariantType() cty.Type {
	return cty.Object(map[string]cty.Type{
		"name":         cty.String,
		"payload":      cty.String,
		"payload_type": cty.String,
		"weight":       cty.Number,
		"weight_type":  cty.String,
		"stickiness":   cty.String,
	})
}
