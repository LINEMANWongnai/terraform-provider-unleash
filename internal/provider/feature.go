package provider

import (
	"encoding/json"
	"sort"

	"github.com/google/go-cmp/cmp"
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
	ContextName types.String `tfsdk:"context_name"`
	JsonValues  types.String `tfsdk:"values_json"`
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
	ContextName     types.String `tfsdk:"context_name"`
	CaseInsensitive types.Bool   `tfsdk:"case_insensitive"`
	Operator        types.String `tfsdk:"operator"`
	Inverted        types.Bool   `tfsdk:"inverted"`
	Value           types.String `tfsdk:"value"`
	JsonValues      types.String `tfsdk:"values_json"`
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
			Description: "ID which is a combination of project , `.` and feature name. e.g. default.my-feature",
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
		"variants": schema.ListNestedAttribute{
			Description: "Variants of this feature",
			NestedObject: schema.NestedAttributeObject{
				Attributes: createVariantResourceSchemaAttrs(),
			},
			Optional: true,
		},
		"strategies": schema.ListNestedAttribute{
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
			Required:    true,
		},
		"stickiness": schema.StringAttribute{
			Description: "Stickiness",
			Optional:    true,
		},
		"overrides": schema.ListNestedAttribute{
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
		"values_json": schema.StringAttribute{
			Description: "An overriding array of string values encoded in JSON. This need to be JSON to avoid performance issue with large number of values.",
			Optional:    true,
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
		"constraints": schema.ListNestedAttribute{
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
		"variants": schema.ListNestedAttribute{
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
			Required:    true,
		},
	}
}

func toFeatureModel(fetchedFeature unleash.FetchedFeature) (FeatureModel, error) {
	f := FeatureModel{
		ID:      types.StringValue(fetchedFeature.FetchedProject + "." + fetchedFeature.Feature.Name),
		Project: types.StringValue(fetchedFeature.FetchedProject),
		Name:    types.StringValue(fetchedFeature.Feature.Name),
	}
	if fetchedFeature.Feature.Type != nil {
		f.Type = types.StringValue(*fetchedFeature.Feature.Type)
	}
	if fetchedFeature.Feature.Description != nil && *fetchedFeature.Feature.Description != "" {
		f.Description = types.StringValue(*fetchedFeature.Feature.Description)
	}
	if fetchedFeature.Feature.ImpressionData != nil {
		f.ImpressionData = types.BoolValue(*fetchedFeature.Feature.ImpressionData)
	}
	if len(fetchedFeature.FetchedEnvironments) > 0 {
		f.Environments = make(map[string]EnvironmentModel, len(fetchedFeature.FetchedEnvironments))
		for _, fetchedEnv := range fetchedFeature.FetchedEnvironments {
			environmentModel, err := toEnvironmentModel(fetchedEnv)
			if err != nil {
				return f, err
			}
			f.Environments[fetchedEnv.Environment.Name] = environmentModel
		}
	}

	return f, nil
}

func toEnvironmentModel(fetchedEnv unleash.FetchedEnvironment) (EnvironmentModel, error) {
	envModel := EnvironmentModel{
		Enabled: types.BoolValue(fetchedEnv.Environment.Enabled),
	}
	for _, variant := range fetchedEnv.FetchedVariants {
		variantModel, err := toVariantModel(variant)
		if err != nil {
			return envModel, err
		}
		envModel.Variants = append(envModel.Variants, variantModel)
	}
	for _, strategy := range fetchedEnv.FetchedStrategies {
		strategyModel, err := toStrategyModel(strategy)
		if err != nil {
			return envModel, err
		}
		envModel.Strategies = append(envModel.Strategies, strategyModel)
	}

	return envModel, nil
}

func toVariantModel(variant unleash.VariantSchema) (VariantModel, error) {
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
	if variant.Stickiness != nil && *variant.Stickiness != "" {
		variantModel.Stickiness = types.StringValue(*variant.Stickiness)
	}
	if variant.WeightType != nil && *variant.WeightType != unleash.Variable {
		variantModel.Weight = types.Float32Value(variant.Weight)
	}
	if variant.Overrides != nil && len(*variant.Overrides) > 0 {
		variantModel.Overrides = make([]VariantOverrideModel, len(*variant.Overrides))
		for i, override := range *variant.Overrides {
			variantOverrideModel, err := toVariantOverrideModel(override)
			if err != nil {
				return variantModel, err
			}

			variantModel.Overrides[i] = variantOverrideModel
		}
	}

	return variantModel, nil
}

func toVariantOverrideModel(override unleash.OverrideSchema) (VariantOverrideModel, error) {
	overrideModel := VariantOverrideModel{
		ContextName: types.StringValue(override.ContextName),
	}
	if override.Values != nil || len(override.Values) > 0 {
		b, err := json.Marshal(override.Values)
		if err != nil {
			return overrideModel, err
		}
		overrideModel.JsonValues = types.StringValue(string(b))
	}

	return overrideModel, nil
}

func toStrategyModel(strategy unleash.FeatureStrategySchema) (StrategyModel, error) {
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
	if strategy.Constraints != nil && len(*strategy.Constraints) > 0 {
		for _, constraint := range *strategy.Constraints {
			constraintModel, err := toConstraintModel(constraint)
			if err != nil {
				return strategyModel, err
			}
			strategyModel.Constraints = append(strategyModel.Constraints, constraintModel)
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

	return strategyModel, nil
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

func toStringValues(jsonValues string) ([]string, error) {
	var stringValues []string
	err := json.Unmarshal([]byte(jsonValues), &stringValues)
	if err != nil {
		return stringValues, err
	}
	return stringValues, nil
}

type variantModelWithIndex struct {
	Variant VariantModel
	Index   int
}

type variantsDiff struct {
	ToAdd     []variantModelWithIndex
	ToRemove  []variantModelWithIndex
	ToReplace []variantModelWithIndex
	Mode      variantDiffMode
}

func getVariantDiffMode(variants []VariantModel, existingVariants []VariantModel) variantsDiff {
	diff := variantsDiff{
		Mode: variantDiffModeMixed,
	}

	if len(variants) == len(existingVariants) {
		for i := 0; i < len(existingVariants); i++ {
			variant := variants[i]
			existingVariant := existingVariants[i]
			if cmp.Equal(variant, existingVariant) {
				continue
			}
			diff.ToReplace = append(diff.ToReplace, variantModelWithIndex{
				Variant: variant,
				Index:   i,
			})
		}

		if len(diff.ToReplace) == 0 {
			diff.Mode = variantDiffModeEqual
		} else {
			diff.Mode = variantDiffModeReplaceOnly
		}

		return diff
	}

	variantModelByName := toVariantModelByName(variants)
	existingVariantModelByName := toVariantModelByName(existingVariants)

	for i, variantModel := range variants {
		_, ok := existingVariantModelByName[variantModel.Name.ValueString()]
		if !ok {
			diff.ToAdd = append(diff.ToAdd, variantModelWithIndex{
				Variant: variantModel,
				Index:   i,
			})
		}
	}
	for i, existingVariant := range existingVariants {
		variantModel, ok := variantModelByName[existingVariant.Name.ValueString()]
		if !ok {
			diff.ToRemove = append(diff.ToRemove, variantModelWithIndex{
				Variant: existingVariant,
				Index:   i,
			})
			continue
		}
		if !cmp.Equal(variantModel, existingVariant) {
			diff.ToReplace = append(diff.ToReplace, variantModelWithIndex{
				Variant: variantModel,
				Index:   i,
			})
		}
	}

	toAddLen := len(diff.ToAdd)
	toRemoveLen := len(diff.ToRemove)
	if toAddLen > 0 {
		sort.Sort(byIndexAsc(diff.ToAdd))
	}
	if toRemoveLen > 0 {
		sort.Sort(byIndexDesc(diff.ToRemove))
	}

	if len(diff.ToReplace) == 0 {
		if toAddLen > 0 {
			if toRemoveLen == 0 {
				diff.Mode = variantDiffModeAddOnly
			}
		} else if toRemoveLen > 0 {
			diff.Mode = variantDiffModeRemoveOnly
		}
	}

	return diff
}

func toVariantModelByName(variants []VariantModel) map[string]VariantModel {
	variantModelByName := make(map[string]VariantModel)
	for _, variant := range variants {
		variantModelByName[variant.Name.ValueString()] = variant
	}

	return variantModelByName
}

type byIndexAsc []variantModelWithIndex

func (a byIndexAsc) Len() int           { return len(a) }
func (a byIndexAsc) Less(i, j int) bool { return a[i].Index < a[j].Index }
func (a byIndexAsc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type byIndexDesc []variantModelWithIndex

func (a byIndexDesc) Len() int           { return len(a) }
func (a byIndexDesc) Less(i, j int) bool { return a[i].Index > a[j].Index }
func (a byIndexDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
