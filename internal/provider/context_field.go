// Copyright (c) HashiCorp, Inc.

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/unleash"
)

type ContextFieldModel struct {
	ID          types.String      `tfsdk:"id"`
	Name        types.String      `tfsdk:"name"`
	Description types.String      `tfsdk:"description"`
	SortOrder   types.Int32       `tfsdk:"sort_order"`
	Stickiness  types.Bool        `tfsdk:"stickiness"`
	LegalValues []LegalValueModel `tfsdk:"legal_values"`
}

type LegalValueModel struct {
	Description types.String `tfsdk:"description"`
	Value       types.String `tfsdk:"value"`
}

func createContextFieldResourceSchemaAttr() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "ID of this context field",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			Description: "The name of this context field",
			Required:    true,
		},
		"description": schema.StringAttribute{
			Description: "A description of what the context field is for",
			Optional:    true,
		},
		"sort_order": schema.Int32Attribute{
			Description: "Sort order",
			Optional:    true,
		},
		"stickiness": schema.BoolAttribute{
			Description: "Stickiness",
			Optional:    true,
		},
		"legal_values": schema.ListNestedAttribute{
			Description: "a list of allowed values for this context field",
			Optional:    true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: createLegalValueResourceSchemaAttrs(),
			},
		},
	}
}

func createLegalValueResourceSchemaAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"value": schema.StringAttribute{
			Description: "valid value",
			Required:    true,
		},
		"description": schema.StringAttribute{
			Description: "Description",
			Optional:    true,
		},
	}
}

func toContextFieldModel(contextField *unleash.ContextFieldSchema) ContextFieldModel {
	contextFieldModel := ContextFieldModel{
		ID:   types.StringValue(contextField.Name),
		Name: types.StringValue(contextField.Name),
	}
	if contextField.Description != nil {
		contextFieldModel.Description = types.StringValue(*contextField.Description)
	}
	if contextField.SortOrder != nil {
		contextFieldModel.SortOrder = types.Int32Value(int32(*contextField.SortOrder))
	}
	if contextField.Stickiness != nil {
		contextFieldModel.Stickiness = types.BoolValue(*contextField.Stickiness)
	}
	if contextField.LegalValues != nil && len(*contextField.LegalValues) > 0 {
		contextFieldModel.LegalValues = make([]LegalValueModel, len(*contextField.LegalValues))
		for i, legalValue := range *contextField.LegalValues {
			contextFieldModel.LegalValues[i] = toLegalValueModel(legalValue)
		}
	}

	return contextFieldModel
}

func toLegalValueModel(legalValue unleash.LegalValueSchema) LegalValueModel {
	legalValueModel := LegalValueModel{
		Value: types.StringValue(legalValue.Value),
	}
	if legalValue.Description != nil {
		legalValueModel.Description = types.StringValue(*legalValue.Description)
	}

	return legalValueModel
}
