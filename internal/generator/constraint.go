// Copyright (c) HashiCorp, Inc.

package generator

import (
	"encoding/json"

	"github.com/zclconf/go-cty/cty"

	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/unleash"
)

func toConstraints(constraints *[]unleash.ConstraintSchema) (cty.Value, error) {
	if constraints == nil || len(*constraints) == 0 {
		return cty.NullVal(cty.List(constraintType)), nil
	}
	constraintValues := make([]cty.Value, 0, len(*constraints))
	for _, constraint := range *constraints {
		attributes := map[string]cty.Value{
			"case_insensitive": cty.NullVal(cty.Bool),
			"context_name":     cty.StringVal(constraint.ContextName),
			"operator":         cty.StringVal(string(constraint.Operator)),
			"inverted":         cty.NullVal(cty.Bool),
			"value":            cty.NullVal(cty.String),
		}
		if constraint.CaseInsensitive != nil {
			attributes["case_insensitive"] = cty.BoolVal(*constraint.CaseInsensitive)
		}
		if constraint.Inverted != nil {
			attributes["inverted"] = cty.BoolVal(*constraint.Inverted)
		}
		values, err := toConstraintValues(constraint.Values)
		if err != nil {
			return cty.NullVal(cty.List(constraintType)), err
		}
		if constraint.Value != nil {
			attributes["value"] = cty.StringVal(*constraint.Value)
		}
		attributes["values_json"] = values

		constraintValues = append(constraintValues, cty.ObjectVal(attributes))
	}
	return cty.ListVal(constraintValues), nil
}

func toConstraintValues(values *[]string) (cty.Value, error) {
	var allValues []string
	if values != nil && len(*values) > 0 {
		allValues = append(allValues, *values...)
	}
	if len(allValues) == 0 {
		return cty.NullVal(cty.String), nil
	}
	b, err := json.Marshal(allValues)
	if err != nil {
		return cty.NullVal(cty.String), err
	}
	return cty.StringVal(string(b)), nil
}
