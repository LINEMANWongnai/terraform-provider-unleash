package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ planmodifier.String = &featureIDPlanModifier{}

type featureIDPlanModifier struct {
	projectAttrName     string
	featureNameAttrName string
}

func createFeatureIDPlanModifier(projectAttrName string, nameAttrName string) planmodifier.String {
	return featureIDPlanModifier{
		projectAttrName:     projectAttrName,
		featureNameAttrName: nameAttrName,
	}
}

func (f featureIDPlanModifier) Description(_ context.Context) string {
	return "combination of project and feature name"
}

func (f featureIDPlanModifier) MarkdownDescription(_ context.Context) string {
	return "combination of project and feature name"
}

func (f featureIDPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}

	var projectID string
	req.Plan.GetAttribute(ctx, path.Root(f.projectAttrName), &projectID)
	var featureName string
	req.Plan.GetAttribute(ctx, path.Root(f.featureNameAttrName), &featureName)

	resp.PlanValue = types.StringValue(projectID + "." + featureName)
}
