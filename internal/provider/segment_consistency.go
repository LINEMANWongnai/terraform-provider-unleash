package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

func ensureSegmentModelNullAndEmptyConsistency(segmentModel *SegmentModel, segmentModelBefore SegmentModel) {
	tryUpdateToEmptyStringIfBeforeEmpty(segmentModel.Project, segmentModelBefore.Project, func(value types.String) {
		segmentModel.Project = value
	})
	tryUpdateToEmptyStringIfBeforeEmpty(segmentModel.Description, segmentModelBefore.Description, func(value types.String) {
		segmentModel.Description = value
	})
	if isNullArrayAndExistingEmptyArray(segmentModel.Constraints, segmentModelBefore.Constraints) {
		segmentModel.Constraints = []ConstraintModel{}
	} else if len(segmentModel.Constraints) == len(segmentModelBefore.Constraints) {
		for i := range segmentModel.Constraints {
			constraint := &segmentModel.Constraints[i]
			constraintBefore := segmentModelBefore.Constraints[i]
			if !constraint.ContextName.Equal(constraintBefore.ContextName) {
				continue
			}
			ensureConstraintNullAndEmptyConsistency(constraint, constraintBefore)
		}
	}
}
