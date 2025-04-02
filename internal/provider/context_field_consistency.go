// Copyright (c) HashiCorp, Inc.

package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

func ensureContextModelNullAndEmptyConsistency(contextModel *ContextFieldModel, contextModelBefore ContextFieldModel) {
	tryUpdateToEmptyStringIfBeforeEmpty(contextModel.Description, contextModelBefore.Description, func(value types.String) {
		contextModel.Description = value
	})
	tryUpdateToFalseIfBeforeFalse(contextModel.Stickiness, contextModelBefore.Stickiness, func(value types.Bool) {
		contextModel.Stickiness = value
	})
	if isNullArrayAndExistingEmptyArray(contextModel.LegalValues, contextModelBefore.LegalValues) {
		contextModel.LegalValues = []LegalValueModel{}
	} else if len(contextModel.LegalValues) == len(contextModelBefore.LegalValues) {
		for i := range contextModel.LegalValues {
			legalValue := &contextModel.LegalValues[i]
			legalValueBefore := contextModelBefore.LegalValues[i]
			if !legalValue.Value.Equal(legalValueBefore.Value) {
				continue
			}
			ensureLegalValueNullAndEmptyConsistency(legalValue, legalValueBefore)
		}
	}
}

func ensureLegalValueNullAndEmptyConsistency(valueModel *LegalValueModel, valueModelBefore LegalValueModel) {
	tryUpdateToEmptyStringIfBeforeEmpty(valueModel.Description, valueModelBefore.Description, func(value types.String) {
		valueModel.Description = value
	})
}
