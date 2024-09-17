package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

func ensureConstraintNullAndEmptyConsistency(constraint *ConstraintModel, constraintBefore ConstraintModel) {
	tryUpdateToFalseIfBeforeFalse(constraint.CaseInsensitive, constraintBefore.CaseInsensitive, func(value types.Bool) {
		constraint.CaseInsensitive = value
	})
	tryUpdateToFalseIfBeforeFalse(constraint.Inverted, constraintBefore.Inverted, func(value types.Bool) {
		constraint.Inverted = value
	})
}
