package provider

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/unleash"
)

type ConstraintModel struct {
	ContextName     types.String `tfsdk:"context_name"`
	CaseInsensitive types.Bool   `tfsdk:"case_insensitive"`
	Operator        types.String `tfsdk:"operator"`
	Inverted        types.Bool   `tfsdk:"inverted"`
	Value           types.String `tfsdk:"value"`
	JsonValues      types.String `tfsdk:"values_json"`
}

func createConstraintResourceSchemaAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"context_name": schema.StringAttribute{
			Description: "Context name",
			Required:    true,
		},
		"case_insensitive": schema.BoolAttribute{
			Description: "Case insensitive flag",
			Optional:    true,
		},
		"operator": schema.StringAttribute{
			Description: "Operator",
			Required:    true,
		},
		"inverted": schema.BoolAttribute{
			Description: "Inverted flag",
			Optional:    true,
		},
		"value": schema.StringAttribute{
			Description: "Value The context value that should be used for constraint evaluation. Use this property instead of `values` for properties that only accept single values.",
			Optional:    true,
		},
		"values_json": schema.StringAttribute{
			Description: "An array of string values encoded in JSON. This need to be JSON to avoid performance issue with large number of values.",
			Optional:    true,
		},
	}
}

func toConstraintModel(constraint unleash.ConstraintSchema) (ConstraintModel, error) {
	constraintModel := ConstraintModel{
		ContextName: types.StringValue(constraint.ContextName),
		Operator:    types.StringValue(string(constraint.Operator)),
	}
	if constraint.CaseInsensitive != nil {
		constraintModel.CaseInsensitive = types.BoolValue(*constraint.CaseInsensitive)
	}
	if constraint.Inverted != nil {
		constraintModel.Inverted = types.BoolValue(*constraint.Inverted)
	}
	if constraint.Value != nil {
		constraintModel.Value = types.StringValue(*constraint.Value)
	}
	if constraint.Values != nil && len(*constraint.Values) > 0 {
		b, err := json.Marshal(*constraint.Values)
		if err != nil {
			return constraintModel, err
		}
		constraintModel.JsonValues = types.StringValue(string(b))
	}

	return constraintModel, nil
}
