package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/unleash"
)

type FeatureModel struct {
	ID             types.String                `tfsdk:"id"`
	Project        types.String                `tfsdk:"project"`
	Name           types.String                `tfsdk:"name"`
	Type           types.String                `tfsdk:"type"`
	Description    types.String                `tfsdk:"description"`
	ImpressionData types.Bool                  `tfsdk:"impression_data"`
	Environments   map[string]EnvironmentModel `tfsdk:"environments"`
}

type EnvironmentModel struct {
	Enabled    types.Bool      `tfsdk:"enabled"`
	Strategies []StrategyModel `tfsdk:"strategies"`
	Variants   []VariantModel  `tfsdk:"variants"`
}

type VariantModel struct {
	Name        types.String           `tfsdk:"name"`
	Payload     types.String           `tfsdk:"payload"`
	PayloadType types.String           `tfsdk:"payload_type"`
	Weight      types.Float32          `tfsdk:"weight"`
	WeightType  types.String           `tfsdk:"weight_type"`
	Stickiness  types.String           `tfsdk:"stickiness"`
	Overrides   []VariantOverrideModel `tfsdk:"overrides"`
}

type VariantOverrideModel struct {
	ContextName types.String   `tfsdk:"context_name"`
	Values      []types.String `tfsdk:"values"`
}

type StrategyModel struct {
	Id          types.String            `tfsdk:"id"`
	Name        types.String            `tfsdk:"name"`
	Disabled    types.Bool              `tfsdk:"disabled"`
	Title       types.String            `tfsdk:"title"`
	SortOrder   types.Float32           `tfsdk:"sort_order"`
	Constraints []ConstraintModel       `tfsdk:"constraints"`
	Parameters  map[string]types.String `tfsdk:"parameters"`
	Segments    []types.Float32         `tfsdk:"segments"`
	Variants    []StrategyVariantModel  `tfsdk:"variants"`
}

type ConstraintModel struct {
	CaseInsensitive types.Bool     `tfsdk:"case_insensitive"`
	ContextName     types.String   `tfsdk:"context_name"`
	Operator        types.String   `tfsdk:"operator"`
	Inverted        types.Bool     `tfsdk:"inverted"`
	Values          []types.String `tfsdk:"values"`
}

type StrategyVariantModel struct {
	Name        types.String `tfsdk:"name"`
	Payload     types.String `tfsdk:"payload"`
	PayloadType types.String `tfsdk:"payload_type"`
	Weight      types.Int64  `tfsdk:"weight"`
	WeightType  types.String `tfsdk:"weight_type"`
	Stickiness  types.String `tfsdk:"stickiness"`
}

func createFeatureResourceSchemaAttr() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "ID which is a combination of project and feature name",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				createFeatureIDPlanModifier("project", "name"),
			},
		},
		"project": schema.StringAttribute{
			Description: "The name of project this feature belongs to",
			Required:    true,
		},
		"name": schema.StringAttribute{
			Description: "The name of this feature",
			Required:    true,
		},
		"type": schema.StringAttribute{
			Description: "Type of the toggle e.g. experiment, kill-switch, release, operational, permission",
			Required:    true,
		},
		"description": schema.StringAttribute{
			Description: "Detailed description of the feature",
			Optional:    true,
		},
		"impression_data": schema.BoolAttribute{
			Description: "true if the impression data collection is enabled for the feature, otherwise false",
			Optional:    true,
		},
		"environments": schema.MapNestedAttribute{
			Description: "The list of environments where the feature can be used",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: createEnvironmentResourceSchemaAttrs(),
			},
		},
	}
}

func createEnvironmentResourceSchemaAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"enabled": schema.BoolAttribute{
			Description: "Is this environment enabled",
			Required:    true,
		},
		"variants": schema.SetNestedAttribute{
			Description: "Variants of this feature",
			NestedObject: schema.NestedAttributeObject{
				Attributes: createVariantResourceSchemaAttrs(),
			},
			Optional: true,
		},
		"strategies": schema.SetNestedAttribute{
			Description: "Strategies of this feature",
			NestedObject: schema.NestedAttributeObject{
				Attributes: createStrategyResourceSchemaAttrs(),
			},
			Required: true,
		},
	}
}

func createVariantResourceSchemaAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Description: "Name of this variant",
			Required:    true,
		},
		"payload": schema.StringAttribute{
			Description: "Payload value",
			Optional:    true,
		},
		"payload_type": schema.StringAttribute{
			Description: "Payload type",
			Optional:    true,
		},
		"weight": schema.Float32Attribute{
			Description: "Weight (1 - 1000). This is required only if weight_type is fix.",
			Optional:    true,
		},
		"weight_type": schema.StringAttribute{
			Description: "Weight type (fix, variable)",
			Optional:    true,
		},
		"stickiness": schema.StringAttribute{
			Description: "Stickiness",
			Optional:    true,
		},
		"overrides": schema.SetNestedAttribute{
			Description: "Overrides assigning specific variants to specific users. The weighting system automatically assigns users to specific groups for you, but any overrides in this list will take precedence.",
			Optional:    true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: createVariantOverrideResourceSchemaAttrs(),
			},
		},
	}
}

func createVariantOverrideResourceSchemaAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"context_name": schema.StringAttribute{
			Description: "The name of the context field used to determine overrides",
			Required:    true,
		},
		"values": schema.ListAttribute{
			Description: "Overriding values",
			Optional:    true,
			ElementType: types.StringType,
		},
	}
}

func createStrategyResourceSchemaAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "ID of this variant",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			Description: "Name of this strategy",
			Required:    true,
		},
		"disabled": schema.BoolAttribute{
			Description: "Disabled flag",
			Required:    true,
		},
		"title": schema.StringAttribute{
			Description: "Title of this strategy",
			Optional:    true,
		},
		"sort_order": schema.Float32Attribute{
			Description: "Sort order",
			Optional:    true,
		},
		"constraints": schema.SetNestedAttribute{
			Description: "Constraints of this strategy",
			Optional:    true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: createConstraintResourceSchemaAttrs(),
			},
		},
		"parameters": schema.MapAttribute{
			Description: "Parameters of this strategy",
			Optional:    true,
			ElementType: types.StringType,
		},
		"segments": schema.SetAttribute{
			Description: "Segment IDs of this strategy",
			Optional:    true,
			ElementType: types.Float32Type,
		},
		"variants": schema.SetNestedAttribute{
			Description: "Variants of this strategy",
			Optional:    true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: createStrategyVariantResourceSchemaAttrs(),
			},
		},
	}
}

func createConstraintResourceSchemaAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"case_insensitive": schema.BoolAttribute{
			Description: "Case insensitive flag",
			Optional:    true,
		},
		"context_name": schema.StringAttribute{
			Description: "Context name",
			Optional:    true,
		},
		"operator": schema.StringAttribute{
			Description: "Operator",
			Optional:    true,
		},
		"inverted": schema.BoolAttribute{
			Description: "Inverted flag",
			Optional:    true,
		},
		"values": schema.SetAttribute{
			Description: "values for constraint",
			Required:    true,
			ElementType: types.StringType,
		},
	}
}

func createStrategyVariantResourceSchemaAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Description: "Name of this variant",
			Required:    true,
		},
		"payload": schema.StringAttribute{
			Description: "Payload value",
			Optional:    true,
		},
		"payload_type": schema.StringAttribute{
			Description: "Payload type",
			Optional:    true,
		},
		"weight": schema.Int64Attribute{
			Description: "Weight (1 - 1000). This is required only if weight_type is fix.",
			Optional:    true,
		},
		"weight_type": schema.StringAttribute{
			Description: "Weight type (fix, variable)",
			Optional:    true,
		},
		"stickiness": schema.StringAttribute{
			Description: "Stickiness",
			Optional:    true,
		},
	}
}

func toFeatureModel(fetchedFeature unleash.FetchedFeature) FeatureModel {
	f := FeatureModel{
		ID:      types.StringValue(fetchedFeature.FetchedProject + "." + fetchedFeature.Feature.Name),
		Project: types.StringValue(fetchedFeature.FetchedProject),
		Name:    types.StringValue(fetchedFeature.Feature.Name),
	}
	if fetchedFeature.Feature.Type != nil {
		f.Type = types.StringValue(*fetchedFeature.Feature.Type)
	}
	if fetchedFeature.Feature.Description != nil {
		f.Description = types.StringValue(*fetchedFeature.Feature.Description)
	}
	if fetchedFeature.Feature.ImpressionData != nil {
		f.ImpressionData = types.BoolValue(*fetchedFeature.Feature.ImpressionData)
	}
	if len(fetchedFeature.FetchedEnvironments) > 0 {
		f.Environments = make(map[string]EnvironmentModel, len(fetchedFeature.FetchedEnvironments))
		for _, fetchedEnv := range fetchedFeature.FetchedEnvironments {
			f.Environments[fetchedEnv.Environment.Name] = toEnvironmentModel(fetchedEnv)
		}
	}

	return f
}

func toEnvironmentModel(fetchedEnv unleash.FetchedEnvironment) EnvironmentModel {
	envModel := EnvironmentModel{
		Enabled: types.BoolValue(fetchedEnv.Environment.Enabled),
	}
	for _, variant := range fetchedEnv.FetchedVariants {
		envModel.Variants = append(envModel.Variants, toVariantModel(variant))
	}
	for _, strategy := range fetchedEnv.FetchedStrategies {
		envModel.Strategies = append(envModel.Strategies, toStrategyModel(strategy))
	}

	return envModel
}

func toVariantModel(variant unleash.VariantSchema) VariantModel {
	variantModel := VariantModel{
		Name: types.StringValue(variant.Name),
	}
	if variant.Payload != nil {
		variantModel.Payload = types.StringValue(variant.Payload.Value)
		variantModel.PayloadType = types.StringValue(string(variant.Payload.Type))
	}
	if variant.WeightType != nil {
		variantModel.WeightType = types.StringValue(string(*variant.WeightType))
	}
	if variant.Stickiness != nil {
		variantModel.Stickiness = types.StringValue(*variant.Stickiness)
	}
	if variant.WeightType != nil && *variant.WeightType != unleash.Variable {
		variantModel.Weight = types.Float32Value(variant.Weight)
	}
	if variant.Overrides != nil && len(*variant.Overrides) > 0 {
		variantModel.Overrides = make([]VariantOverrideModel, len(*variant.Overrides))
		for i, override := range *variant.Overrides {
			variantModel.Overrides[i] = toVariantOverrideModel(override)
		}
	}

	return variantModel
}

func toVariantOverrideModel(override unleash.OverrideSchema) VariantOverrideModel {
	overrideModel := VariantOverrideModel{
		ContextName: types.StringValue(override.ContextName),
	}
	overrideModel.Values = make([]types.String, len(override.Values))
	for i, value := range override.Values {
		overrideModel.Values[i] = types.StringValue(value)
	}

	return overrideModel
}

func toStrategyModel(strategy unleash.FeatureStrategySchema) StrategyModel {
	strategyModel := StrategyModel{
		Name:     types.StringValue(strategy.Name),
		Disabled: types.BoolValue(false),
	}
	if strategy.Id != nil {
		strategyModel.Id = types.StringValue(*strategy.Id)
	}
	if strategy.Disabled != nil || *strategy.Disabled {
		strategyModel.Disabled = types.BoolValue(*strategy.Disabled)
	}
	if strategy.Title != nil && *strategy.Title != "" {
		strategyModel.Title = types.StringValue(*strategy.Title)
	}
	if strategy.SortOrder != nil && *strategy.SortOrder != 0 {
		strategyModel.SortOrder = types.Float32Value(*strategy.SortOrder)
	}
	if strategy.Constraints != nil {
		for _, constraint := range *strategy.Constraints {
			strategyModel.Constraints = append(strategyModel.Constraints, toConstraintModel(constraint))
		}
	}
	if strategy.Parameters != nil && len(*strategy.Parameters) > 0 {
		strategyModel.Parameters = make(map[string]types.String)
		for k, v := range *strategy.Parameters {
			strategyModel.Parameters[k] = types.StringValue(v)
		}
	}
	if strategy.Segments != nil {
		for _, segment := range *strategy.Segments {
			strategyModel.Segments = append(strategyModel.Segments, types.Float32Value(segment))
		}
	}
	if strategy.Variants != nil {
		for _, variant := range *strategy.Variants {
			strategyModel.Variants = append(strategyModel.Variants, toStrategyVariantModel(variant))
		}
	}

	return strategyModel
}

func toConstraintModel(constraint unleash.ConstraintSchema) ConstraintModel {
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
		constraintModel.Values = append(constraintModel.Values, types.StringValue(*constraint.Value))
	}
	if constraint.Values != nil {
		for _, value := range *constraint.Values {
			constraintModel.Values = append(constraintModel.Values, types.StringValue(value))
		}
	}

	return constraintModel
}

func toStrategyVariantModel(variant unleash.StrategyVariantSchema) StrategyVariantModel {
	strategyVariantModel := StrategyVariantModel{
		Name:       types.StringValue(variant.Name),
		Stickiness: types.StringValue(variant.Stickiness),
		WeightType: types.StringValue(string(variant.WeightType)),
	}
	if variant.Payload != nil {
		strategyVariantModel.Payload = types.StringValue(variant.Payload.Value)
		strategyVariantModel.PayloadType = types.StringValue(string(variant.Payload.Type))
	}
	if variant.WeightType == unleash.StrategyVariantSchemaWeightTypeFix {
		strategyVariantModel.Weight = types.Int64Value(int64(variant.Weight))
	}

	return strategyVariantModel
}
