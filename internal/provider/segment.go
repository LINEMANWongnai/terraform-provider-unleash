package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/unleash"
)

type SegmentModel struct {
	ID          types.String      `tfsdk:"id"`
	IDInt       types.Int64       `tfsdk:"id_int"`
	Project     types.String      `tfsdk:"project"`
	Name        types.String      `tfsdk:"name"`
	Description types.String      `tfsdk:"description"`
	Constraints []ConstraintModel `tfsdk:"constraints"`
}

func createSegmentResourceSchemaAttr() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "ID of this segment",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"id_int": schema.Int64Attribute{
			Description: "ID of this segment as int",
			Computed:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"project": schema.StringAttribute{
			Description: "The name of project this segment belongs to",
			Optional:    true,
		},
		"name": schema.StringAttribute{
			Description: "The name of this segment",
			Required:    true,
		},
		"description": schema.StringAttribute{
			Description: "A description of what the segment is for",
			Optional:    true,
		},
		"constraints": schema.ListNestedAttribute{
			Description: "The list of constraints that make up this segment",
			Optional:    true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: createConstraintResourceSchemaAttrs(),
			},
		},
	}
}

func toSegmentModel(segment *unleash.AdminSegmentSchema) (SegmentModel, error) {
	segmentModel := SegmentModel{
		ID:    types.StringValue(fmt.Sprintf("%d", segment.Id)),
		IDInt: types.Int64Value(int64(segment.Id)),
		Name:  types.StringValue(segment.Name),
	}
	if segment.Project != nil {
		segmentModel.Project = types.StringValue(*segment.Project)
	}
	if segment.Description != nil {
		segmentModel.Description = types.StringValue(*segment.Description)
	}
	if len(segment.Constraints) > 0 {
		segmentModel.Constraints = make([]ConstraintModel, len(segment.Constraints))
		for i, constraint := range segment.Constraints {
			constraintModel, err := toConstraintModel(constraint)
			if err != nil {
				return SegmentModel{}, err
			}
			segmentModel.Constraints[i] = constraintModel
		}
	}

	return segmentModel, nil
}
