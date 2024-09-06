package generator

import (
	"context"
	"encoding/json"
	"io"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/unleash"
)

func Generate(client unleash.ClientWithResponsesInterface, projectID string, tfWriter io.Writer, importWriter io.Writer) error {
	ctx := context.Background()
	fetchedFeatures, err := unleash.GetFeatures(ctx, client, projectID)
	if err != nil {
		return err
	}

	sort.Sort(byFeatureName(fetchedFeatures))

	hclFile := hclwrite.NewEmptyFile()
	hclBody := hclFile.Body()

	importHclFile := hclwrite.NewEmptyFile()
	importHclBody := importHclFile.Body()
	for _, fetchedFeature := range fetchedFeatures {
		if fetchedFeature.Feature.Archived != nil && *fetchedFeature.Feature.Archived {
			continue
		}
		resourceName := strings.ReplaceAll(fetchedFeature.Feature.Name, ".", "_")

		resource := hclBody.AppendNewBlock("resource", []string{"unleash_feature", resourceName})
		resourceBody := resource.Body()
		resourceBody.SetAttributeValue("project", cty.StringVal(fetchedFeature.FetchedProject))
		resourceBody.SetAttributeValue("name", cty.StringVal(fetchedFeature.Feature.Name))
		resourceBody.SetAttributeValue("type", cty.StringVal(*fetchedFeature.Feature.Type))
		if fetchedFeature.Feature.ImpressionData != nil {
			resourceBody.SetAttributeValue("impression_data", cty.BoolVal(*fetchedFeature.Feature.ImpressionData))
		}
		if fetchedFeature.Feature.Description != nil && *fetchedFeature.Feature.Description != "" {
			resourceBody.SetAttributeValue("description", cty.StringVal(*fetchedFeature.Feature.Description))
		}
		environments, err := toEnvironmentMaps(fetchedFeature.Feature.Name, fetchedFeature.FetchedEnvironments)
		if err != nil {
			return err
		}
		resourceBody.SetAttributeValue("environments", environments)
		hclBody.AppendNewline()

		importBlock := importHclBody.AppendNewBlock("import", []string{})
		importBody := importBlock.Body()
		importBody.SetAttributeRaw("to", []*hclwrite.Token{
			{
				Type:         hclsyntax.TokenQuotedLit,
				Bytes:        []byte("unleash_feature." + resourceName),
				SpacesBefore: 0,
			},
		})
		importBody.SetAttributeValue("id", cty.StringVal(fetchedFeature.FetchedProject+"."+fetchedFeature.Feature.Name))
		importHclBody.AppendNewline()
	}

	_, err = hclFile.WriteTo(tfWriter)
	if err != nil {
		return err
	}
	_, err = importHclFile.WriteTo(importWriter)
	return err
}

type byFeatureName []unleash.FetchedFeature

func (a byFeatureName) Len() int           { return len(a) }
func (a byFeatureName) Less(i, j int) bool { return a[i].Feature.Name < a[j].Feature.Name }
func (a byFeatureName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func toEnvironmentMaps(featureName string, environments []unleash.FetchedEnvironment) (cty.Value, error) {
	if len(environments) == 0 {
		return cty.NullVal(cty.Map(environmentType)), nil
	}
	environmentByName := make(map[string]cty.Value)
	for _, environment := range environments {
		envValue, err := toEnvironment(featureName, environment)
		if err != nil {
			return cty.NullVal(cty.Map(environmentType)), err
		}
		environmentByName[environment.Environment.Name] = envValue
	}
	return cty.MapVal(environmentByName), nil
}

func toEnvironment(featureName string, environment unleash.FetchedEnvironment) (cty.Value, error) {
	attributes := make(map[string]cty.Value)

	attributes["enabled"] = cty.BoolVal(environment.Environment.Enabled)
	strategies, err := toStrategies(featureName, environment.FetchedStrategies)
	if err != nil {
		return cty.NullVal(environmentType), err
	}
	attributes["strategies"] = strategies
	variants, err := toVariants(environment.FetchedVariants)
	if err != nil {
		return cty.NullVal(environmentType), err
	}
	attributes["variants"] = variants

	return cty.ObjectVal(attributes), nil
}

func toStrategies(featureName string, strategies []unleash.FeatureStrategySchema) (cty.Value, error) {
	// always add a default strategy if no strategies are defined to prevent conflicts when a feature is enabled
	// since Unleash automatically creates a default strategy when a feature is enabled
	if len(strategies) == 0 {
		parameters := unleash.ParametersSchema(map[string]string{
			"groupId":    featureName,
			"rollout":    "100",
			"stickiness": "default",
		})
		defaultStrategy := unleash.FeatureStrategySchema{
			Name:       "flexibleRollout",
			Parameters: &parameters,
		}
		attrs, err := toStrategyAttributes(defaultStrategy)
		if err != nil {
			return cty.SetVal([]cty.Value{cty.ObjectVal(attrs)}), err
		}

		return cty.SetVal([]cty.Value{cty.ObjectVal(attrs)}), nil
	}
	strategyValues := make([]cty.Value, len(strategies))
	for i, strategy := range strategies {
		attrs, err := toStrategyAttributes(strategy)
		if err != nil {
			return cty.SetVal([]cty.Value{cty.ObjectVal(attrs)}), err
		}
		strategyValues[i] = cty.ObjectVal(attrs)
	}
	return cty.SetVal(strategyValues), nil
}

func toStrategyAttributes(strategy unleash.FeatureStrategySchema) (map[string]cty.Value, error) {
	attributes := map[string]cty.Value{
		"name":       cty.StringVal(strategy.Name),
		"disabled":   cty.BoolVal(false),
		"title":      cty.NullVal(cty.String),
		"sort_order": cty.NullVal(cty.Number),
		"parameters": toParameters(strategy.Parameters),
		"segments":   toSegments(strategy.Segments),
		"variants":   toStrategyVariants(strategy.Variants),
	}
	if strategy.Disabled != nil && *strategy.Disabled {
		attributes["disabled"] = cty.BoolVal(*strategy.Disabled)
	}
	if strategy.Title != nil && *strategy.Title != "" {
		attributes["title"] = cty.StringVal(*strategy.Title)
	}
	if strategy.SortOrder != nil && *strategy.SortOrder != 0 {
		attributes["sort_order"] = cty.NumberFloatVal(float64(*strategy.SortOrder))
	}
	constraints, err := toConstraints(strategy.Constraints)
	if err != nil {
		return attributes, err
	}
	attributes["constraints"] = constraints

	return attributes, nil
}

func toVariants(variants []unleash.VariantSchema) (cty.Value, error) {
	if len(variants) == 0 {
		return cty.NullVal(cty.List(variantType)), nil
	}
	variantValues := make([]cty.Value, 0, len(variants))
	for _, variant := range variants {
		attributes := map[string]cty.Value{
			"name":         cty.StringVal(variant.Name),
			"weight_type":  cty.StringVal(string(*variant.WeightType)),
			"payload":      cty.NullVal(cty.String),
			"payload_type": cty.NullVal(cty.String),
			"stickiness":   cty.NullVal(cty.String),
			"weight":       cty.NullVal(cty.Number),
		}
		if variant.Payload != nil {
			attributes["payload"] = cty.StringVal(variant.Payload.Value)
			attributes["payload_type"] = cty.StringVal(string(variant.Payload.Type))
		}
		if variant.Stickiness != nil {
			attributes["stickiness"] = cty.StringVal(*variant.Stickiness)
		}
		if *variant.WeightType == unleash.Fix {
			attributes["weight"] = cty.NumberFloatVal(float64(variant.Weight))
		}
		overrides, err := toVariantOverrides(variant.Overrides)
		if err != nil {
			return cty.NullVal(cty.List(variantType)), err
		}
		attributes["overrides"] = overrides
		variantValues = append(variantValues, cty.ObjectVal(attributes))
	}
	return cty.ListVal(variantValues), nil
}

func toVariantOverrides(overrides *[]unleash.OverrideSchema) (cty.Value, error) {
	if overrides == nil || len(*overrides) == 0 {
		return cty.NullVal(cty.Set(variantOverrideType)), nil
	}
	overrideValues := make([]cty.Value, len(*overrides))
	for i, override := range *overrides {
		attributes := map[string]cty.Value{
			"context_name": cty.StringVal(override.ContextName),
		}
		values, err := toVariantOverrideValues(override.Values)
		if err != nil {
			return cty.NullVal(cty.Set(variantOverrideType)), err
		}
		attributes["values_json"] = values
		overrideValues[i] = cty.ObjectVal(attributes)
	}
	return cty.SetVal(overrideValues), nil
}

func toVariantOverrideValues(values []string) (cty.Value, error) {
	if len(values) == 0 {
		return cty.NullVal(cty.String), nil
	}
	b, err := json.Marshal(values)
	if err != nil {
		return cty.NullVal(cty.String), err
	}
	return cty.StringVal(string(b)), nil
}

func toConstraints(constraints *[]unleash.ConstraintSchema) (cty.Value, error) {
	if constraints == nil || len(*constraints) == 0 {
		return cty.NullVal(cty.Set(constraintType)), nil
	}
	constraintValues := make([]cty.Value, 0, len(*constraints))
	for _, constraint := range *constraints {
		attributes := map[string]cty.Value{
			"case_insensitive": cty.NullVal(cty.Bool),
			"context_name":     cty.StringVal(constraint.ContextName),
			"operator":         cty.StringVal(string(constraint.Operator)),
			"inverted":         cty.NullVal(cty.Bool),
		}
		if constraint.CaseInsensitive != nil {
			attributes["case_insensitive"] = cty.BoolVal(*constraint.CaseInsensitive)
		}
		if constraint.Inverted != nil {
			attributes["inverted"] = cty.BoolVal(*constraint.Inverted)
		}
		values, err := toConstraintValues(constraint.Value, constraint.Values)
		if err != nil {
			return cty.NullVal(cty.Set(constraintType)), err
		}
		attributes["values_json"] = values

		constraintValues = append(constraintValues, cty.ObjectVal(attributes))
	}
	return cty.SetVal(constraintValues), nil
}

func toConstraintValues(value *string, values *[]string) (cty.Value, error) {
	var allValues []string
	if value != nil {
		allValues = append(allValues, *value)
	}
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

func toParameters(parameters *unleash.ParametersSchema) cty.Value {
	if parameters == nil || len(*parameters) == 0 {
		return cty.NullVal(cty.Map(cty.String))
	}
	attributes := make(map[string]cty.Value)
	for name, value := range *parameters {
		attributes[name] = cty.StringVal(value)
	}
	return cty.MapVal(attributes)
}

func toSegments(segments *[]float32) cty.Value {
	if segments == nil || len(*segments) == 0 {
		return cty.NullVal(cty.List(cty.Number))
	}
	segmentValues := make([]cty.Value, 0, len(*segments))
	for _, segment := range *segments {
		segmentValues = append(segmentValues, cty.NumberFloatVal(float64(segment)))
	}
	return cty.ListVal(segmentValues)
}

func toStrategyVariants(variants *[]unleash.StrategyVariantSchema) cty.Value {
	if variants == nil || len(*variants) == 0 {
		return cty.NullVal(cty.List(strategyVariantType))
	}
	variantValues := make([]cty.Value, 0, len(*variants))
	for _, variant := range *variants {
		attributes := map[string]cty.Value{
			"name":         cty.StringVal(variant.Name),
			"weight_type":  cty.StringVal(string(variant.WeightType)),
			"payload":      cty.NullVal(cty.String),
			"payload_type": cty.NullVal(cty.String),
			"stickiness":   cty.StringVal(variant.Stickiness),
			"weight":       cty.NullVal(cty.Number),
		}
		if variant.Payload != nil {
			attributes["payload"] = cty.StringVal(variant.Payload.Value)
			attributes["payload_type"] = cty.StringVal(string(variant.Payload.Type))
		}
		if variant.WeightType == unleash.StrategyVariantSchemaWeightTypeFix {
			attributes["weight"] = cty.NumberFloatVal(float64(variant.Weight))
		}
		variantValues = append(variantValues, cty.ObjectVal(attributes))
	}
	return cty.ListVal(variantValues)
}
