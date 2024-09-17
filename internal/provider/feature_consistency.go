package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ensureFeatureModelNullAndEmptyConsistency changes the given featureModel's properties which are null to matched null or empty properties of featureModelBefore.
//
// This avoids state conflicts when the user sets a property to empty in the configuration which does not match read model.
func ensureFeatureModelNullAndEmptyConsistency(featureModel *FeatureModel, featureModelBefore FeatureModel) {
	tryUpdateToEmptyStringIfBeforeEmpty(featureModel.Description, featureModelBefore.Description, func(value types.String) {
		featureModel.Description = value
	})
	tryUpdateToFalseIfBeforeFalse(featureModel.ImpressionData, featureModelBefore.ImpressionData, func(value types.Bool) {
		featureModel.ImpressionData = value
	})
	if len(featureModel.Environments) != len(featureModelBefore.Environments) {
		return
	}
	for name, env := range featureModel.Environments {
		envBefore, ok := featureModelBefore.Environments[name]
		if !ok {
			continue
		}
		ensureEnvironmentNullAndEmptyConsistency(&env, envBefore)
		featureModel.Environments[name] = env
	}
}

func ensureEnvironmentNullAndEmptyConsistency(env *EnvironmentModel, envBefore EnvironmentModel) {
	if isNullArrayAndExistingEmptyArray(env.Variants, envBefore.Variants) {
		env.Variants = []VariantModel{}
	} else if len(env.Variants) == len(envBefore.Variants) {
		for i := range env.Variants {
			variant := &env.Variants[i]
			variantBefore := envBefore.Variants[i]
			if !variant.Name.Equal(variantBefore.Name) {
				continue
			}
			ensureVariantNullAndEmptyConsistency(variant, variantBefore)
		}
	}
	if len(env.Strategies) == len(envBefore.Strategies) {
		allMatched := true
		strategyByName := toStrategyModelByIDName(env.Strategies)
		strategies := make([]StrategyModel, len(env.Strategies))
		for i, strategyBefore := range envBefore.Strategies {
			strategy, ok := strategyByName[toStrategyModelKey(strategyBefore)]
			if !ok {
				allMatched = false
				break
			}
			ensureStrategyNullAndEmptyConsistency(&strategy, strategyBefore)
			// Need to preserve order...
			strategies[i] = strategy
		}
		if allMatched {
			env.Strategies = strategies
		}
	}
}

func ensureVariantNullAndEmptyConsistency(variant *VariantModel, variantBefore VariantModel) {
	tryUpdateToEmptyStringIfBeforeEmpty(variant.Payload, variantBefore.Payload, func(value types.String) {
		variant.Payload = value
	})
	tryUpdateToEmptyStringIfBeforeEmpty(variant.PayloadType, variantBefore.PayloadType, func(value types.String) {
		variant.PayloadType = value
	})
	tryUpdateToEmptyStringIfBeforeEmpty(variant.Stickiness, variantBefore.Stickiness, func(value types.String) {
		variant.Stickiness = value
	})
	if isNullArrayAndExistingEmptyArray(variant.Overrides, variantBefore.Overrides) {
		variant.Overrides = []VariantOverrideModel{}
	} else if len(variant.Overrides) == len(variantBefore.Overrides) {
		for i := range variant.Overrides {
			override := &variant.Overrides[i]
			overrideBefore := variantBefore.Overrides[i]
			if !override.ContextName.Equal(overrideBefore.ContextName) {
				continue
			}
			ensureOverrideNullAndEmptyConsistency(override, overrideBefore)
		}
	}
}

func ensureOverrideNullAndEmptyConsistency(override *VariantOverrideModel, overrideBefore VariantOverrideModel) {
	tryUpdateToEmptyStringIfBeforeEmpty(override.JsonValues, overrideBefore.JsonValues, func(value types.String) {
		override.JsonValues = value
	})
}

func ensureStrategyNullAndEmptyConsistency(strategy *StrategyModel, strategyBefore StrategyModel) {
	tryUpdateToEmptyStringIfBeforeEmpty(strategy.Title, strategyBefore.Title, func(value types.String) {
		strategy.Title = value
	})
	if isNullArrayAndExistingEmptyArray(strategy.Constraints, strategyBefore.Constraints) {
		strategy.Constraints = []ConstraintModel{}
	} else if len(strategy.Constraints) == len(strategyBefore.Constraints) {
		for i := range strategy.Constraints {
			constraint := &strategy.Constraints[i]
			constraintBefore := strategyBefore.Constraints[i]
			if !constraint.ContextName.Equal(constraintBefore.ContextName) {
				continue
			}
			ensureConstraintNullAndEmptyConsistency(constraint, constraintBefore)
		}
	}
	if isNullMapAndExistingEmptyMap(strategy.Parameters, strategyBefore.Parameters) {
		strategy.Parameters = make(map[string]types.String)
	} else if len(strategy.Parameters) == len(strategyBefore.Parameters) {
		for name, parameter := range strategy.Parameters {
			existingParameter, ok := strategyBefore.Parameters[name]
			if !ok {
				continue
			}
			tryUpdateToEmptyStringIfBeforeEmpty(parameter, existingParameter, func(value types.String) {
				strategy.Parameters[name] = value
			})
		}
	}
	if isNullArrayAndExistingEmptyArray(strategy.Segments, strategyBefore.Segments) {
		strategy.Segments = []types.Float32{}
	}
	if isNullArrayAndExistingEmptyArray(strategy.Variants, strategyBefore.Variants) {
		strategy.Variants = []StrategyVariantModel{}
	} else if len(strategy.Variants) == len(strategyBefore.Variants) {
		for i := range strategy.Variants {
			variant := &strategy.Variants[i]
			existingVariant := strategyBefore.Variants[i]
			if !variant.Name.Equal(existingVariant.Name) {
				continue
			}
			ensureStrategyVariantNullAndEmptyConsistency(variant, existingVariant)
		}
	}
}

func ensureStrategyVariantNullAndEmptyConsistency(strategyVariant *StrategyVariantModel, strategyVariantBefore StrategyVariantModel) {
	tryUpdateToEmptyStringIfBeforeEmpty(strategyVariant.Payload, strategyVariantBefore.Payload, func(value types.String) {
		strategyVariant.Payload = value
	})
	tryUpdateToEmptyStringIfBeforeEmpty(strategyVariant.PayloadType, strategyVariantBefore.PayloadType, func(value types.String) {
		strategyVariant.PayloadType = value
	})
}
