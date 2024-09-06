package provider

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestGetVariantDiffMode(t *testing.T) {
	tests := []struct {
		name             string
		variants         []VariantModel
		existingVariants []VariantModel
		result           variantsDiff
	}{
		{
			name: "equal",
			variants: []VariantModel{
				{Name: types.StringValue("v1"), Payload: types.StringValue("v1 payload")},
				{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload")},
				{Name: types.StringValue("v3"), Payload: types.StringValue("v3 payload")},
			},
			existingVariants: []VariantModel{
				{Name: types.StringValue("v1"), Payload: types.StringValue("v1 payload")},
				{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload")},
				{Name: types.StringValue("v3"), Payload: types.StringValue("v3 payload")},
			},
			result: variantsDiff{
				Mode: variantDiffModeEqual,
			},
		},
		{
			name: "switch sequence",
			variants: []VariantModel{
				{Name: types.StringValue("v4"), Payload: types.StringValue("v4 payload")},
				{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload")},
				{Name: types.StringValue("v3"), Payload: types.StringValue("v3 payload")},
				{Name: types.StringValue("v1"), Payload: types.StringValue("v1 payload")},
			},
			existingVariants: []VariantModel{
				{Name: types.StringValue("v1"), Payload: types.StringValue("v1 payload")},
				{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload")},
				{Name: types.StringValue("v3"), Payload: types.StringValue("v3 payload")},
				{Name: types.StringValue("v4"), Payload: types.StringValue("v4 payload")},
			},
			result: variantsDiff{
				ToReplace: []variantModelWithIndex{
					{Variant: VariantModel{Name: types.StringValue("v4"), Payload: types.StringValue("v4 payload")}, Index: 0},
					{Variant: VariantModel{Name: types.StringValue("v1"), Payload: types.StringValue("v1 payload")}, Index: 3},
				},
				Mode: variantDiffModeReplaceOnly,
			},
		},
		{
			name: "change content only",
			variants: []VariantModel{
				{Name: types.StringValue("v1"), Payload: types.StringValue("v1 payload")},
				{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload 2")},
				{Name: types.StringValue("v3"), Payload: types.StringValue("v3 payload 2")},
				{Name: types.StringValue("v4"), Payload: types.StringValue("v4 payload")},
			},
			existingVariants: []VariantModel{
				{Name: types.StringValue("v1"), Payload: types.StringValue("v1 payload")},
				{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload")},
				{Name: types.StringValue("v3"), Payload: types.StringValue("v3 payload")},
				{Name: types.StringValue("v4"), Payload: types.StringValue("v4 payload")},
			},
			result: variantsDiff{
				ToReplace: []variantModelWithIndex{
					{Variant: VariantModel{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload 2")}, Index: 1},
					{Variant: VariantModel{Name: types.StringValue("v3"), Payload: types.StringValue("v3 payload 2")}, Index: 2},
				},
				Mode: variantDiffModeReplaceOnly,
			},
		},
		{
			name: "change content",
			variants: []VariantModel{
				{Name: types.StringValue("v4"), Payload: types.StringValue("v4 payload")},
				{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload 2")},
				{Name: types.StringValue("v3"), Payload: types.StringValue("v3 payload")},
				{Name: types.StringValue("v1"), Payload: types.StringValue("v1 payload")},
			},
			existingVariants: []VariantModel{
				{Name: types.StringValue("v1"), Payload: types.StringValue("v1 payload")},
				{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload")},
				{Name: types.StringValue("v3"), Payload: types.StringValue("v3 payload")},
				{Name: types.StringValue("v4"), Payload: types.StringValue("v4 payload")},
			},
			result: variantsDiff{
				ToReplace: []variantModelWithIndex{
					{Variant: VariantModel{Name: types.StringValue("v4"), Payload: types.StringValue("v4 payload")}, Index: 0},
					{Variant: VariantModel{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload 2")}, Index: 1},
					{Variant: VariantModel{Name: types.StringValue("v1"), Payload: types.StringValue("v1 payload")}, Index: 3},
				},
				Mode: variantDiffModeReplaceOnly,
			},
		},
		{
			name: "new variants",
			variants: []VariantModel{
				{Name: types.StringValue("v1"), Payload: types.StringValue("v1 payload")},
				{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload")},
				{Name: types.StringValue("v3"), Payload: types.StringValue("v3 payload")},
			},
			existingVariants: []VariantModel{},
			result: variantsDiff{
				ToAdd: []variantModelWithIndex{
					{Variant: VariantModel{Name: types.StringValue("v1"), Payload: types.StringValue("v1 payload")}, Index: 0},
					{Variant: VariantModel{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload")}, Index: 1},
					{Variant: VariantModel{Name: types.StringValue("v3"), Payload: types.StringValue("v3 payload")}, Index: 2},
				},
				Mode: variantDiffModeAddOnly,
			},
		},
		{
			name: "add variants only",
			variants: []VariantModel{
				{Name: types.StringValue("v1"), Payload: types.StringValue("v1 payload")},
				{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload")},
				{Name: types.StringValue("v3"), Payload: types.StringValue("v3 payload")},
			},
			existingVariants: []VariantModel{
				{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload")},
			},
			result: variantsDiff{
				ToAdd: []variantModelWithIndex{
					{Variant: VariantModel{Name: types.StringValue("v1"), Payload: types.StringValue("v1 payload")}, Index: 0},
					{Variant: VariantModel{Name: types.StringValue("v3"), Payload: types.StringValue("v3 payload")}, Index: 2},
				},
				Mode: variantDiffModeAddOnly,
			},
		},
		{
			name: "add and change",
			variants: []VariantModel{
				{Name: types.StringValue("v1"), Payload: types.StringValue("v1 payload")},
				{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload 2")},
				{Name: types.StringValue("v3"), Payload: types.StringValue("v3 payload")},
				{Name: types.StringValue("v4"), Payload: types.StringValue("v4 payload 2")},
			},
			existingVariants: []VariantModel{
				{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload")},
				{Name: types.StringValue("v4"), Payload: types.StringValue("v4 payload")},
			},
			result: variantsDiff{
				ToAdd: []variantModelWithIndex{
					{Variant: VariantModel{Name: types.StringValue("v1"), Payload: types.StringValue("v1 payload")}, Index: 0},
					{Variant: VariantModel{Name: types.StringValue("v3"), Payload: types.StringValue("v3 payload")}, Index: 2},
				},
				ToReplace: []variantModelWithIndex{
					{Variant: VariantModel{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload 2")}, Index: 0},
					{Variant: VariantModel{Name: types.StringValue("v4"), Payload: types.StringValue("v4 payload 2")}, Index: 1},
				},
				Mode: variantDiffModeMixed,
			},
		},
		{
			name: "remove",
			variants: []VariantModel{
				{Name: types.StringValue("v1"), Payload: types.StringValue("v1 payload")},
				{Name: types.StringValue("v4"), Payload: types.StringValue("v4 payload")},
			},
			existingVariants: []VariantModel{
				{Name: types.StringValue("v1"), Payload: types.StringValue("v1 payload")},
				{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload")},
				{Name: types.StringValue("v3"), Payload: types.StringValue("v3 payload")},
				{Name: types.StringValue("v4"), Payload: types.StringValue("v4 payload")},
			},
			result: variantsDiff{
				ToRemove: []variantModelWithIndex{
					{Variant: VariantModel{Name: types.StringValue("v3"), Payload: types.StringValue("v3 payload")}, Index: 2},
					{Variant: VariantModel{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload")}, Index: 1},
				},
				Mode: variantDiffModeRemoveOnly,
			},
		},
		{
			name: "remove and change",
			variants: []VariantModel{
				{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload 2")},
				{Name: types.StringValue("v4"), Payload: types.StringValue("v4 payload 2")},
			},
			existingVariants: []VariantModel{
				{Name: types.StringValue("v1"), Payload: types.StringValue("v1 payload")},
				{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload")},
				{Name: types.StringValue("v3"), Payload: types.StringValue("v3 payload")},
				{Name: types.StringValue("v4"), Payload: types.StringValue("v4 payload")},
			},
			result: variantsDiff{
				ToRemove: []variantModelWithIndex{
					{Variant: VariantModel{Name: types.StringValue("v3"), Payload: types.StringValue("v3 payload")}, Index: 2},
					{Variant: VariantModel{Name: types.StringValue("v1"), Payload: types.StringValue("v1 payload")}, Index: 0},
				},
				ToReplace: []variantModelWithIndex{
					{Variant: VariantModel{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload 2")}, Index: 1},
					{Variant: VariantModel{Name: types.StringValue("v4"), Payload: types.StringValue("v4 payload 2")}, Index: 3},
				},
				Mode: variantDiffModeMixed,
			},
		},
		{
			name: "add and remove",
			variants: []VariantModel{
				{Name: types.StringValue("v5"), Payload: types.StringValue("v5 payload")},
				{Name: types.StringValue("v6"), Payload: types.StringValue("v6 payload")},
				{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload")},
				{Name: types.StringValue("v4"), Payload: types.StringValue("v4 payload")},
				{Name: types.StringValue("v7"), Payload: types.StringValue("v7 payload")},
			},
			existingVariants: []VariantModel{
				{Name: types.StringValue("v1"), Payload: types.StringValue("v1 payload")},
				{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload")},
				{Name: types.StringValue("v3"), Payload: types.StringValue("v3 payload")},
				{Name: types.StringValue("v4"), Payload: types.StringValue("v4 payload")},
			},
			result: variantsDiff{
				ToAdd: []variantModelWithIndex{
					{Variant: VariantModel{Name: types.StringValue("v5"), Payload: types.StringValue("v5 payload")}, Index: 0},
					{Variant: VariantModel{Name: types.StringValue("v6"), Payload: types.StringValue("v6 payload")}, Index: 1},
					{Variant: VariantModel{Name: types.StringValue("v7"), Payload: types.StringValue("v7 payload")}, Index: 4},
				},
				ToRemove: []variantModelWithIndex{
					{Variant: VariantModel{Name: types.StringValue("v3"), Payload: types.StringValue("v3 payload")}, Index: 2},
					{Variant: VariantModel{Name: types.StringValue("v1"), Payload: types.StringValue("v1 payload")}, Index: 0},
				},
				Mode: variantDiffModeMixed,
			},
		},
		{
			name: "mixed",
			variants: []VariantModel{
				{Name: types.StringValue("v5"), Payload: types.StringValue("v5 payload")},
				{Name: types.StringValue("v6"), Payload: types.StringValue("v6 payload")},
				{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload 2")},
				{Name: types.StringValue("v4"), Payload: types.StringValue("v4 payload")},
				{Name: types.StringValue("v7"), Payload: types.StringValue("v7 payload")},
			},
			existingVariants: []VariantModel{
				{Name: types.StringValue("v1"), Payload: types.StringValue("v1 payload")},
				{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload")},
				{Name: types.StringValue("v3"), Payload: types.StringValue("v3 payload")},
				{Name: types.StringValue("v4"), Payload: types.StringValue("v4 payload")},
			},
			result: variantsDiff{
				ToAdd: []variantModelWithIndex{
					{Variant: VariantModel{Name: types.StringValue("v5"), Payload: types.StringValue("v5 payload")}, Index: 0},
					{Variant: VariantModel{Name: types.StringValue("v6"), Payload: types.StringValue("v6 payload")}, Index: 1},
					{Variant: VariantModel{Name: types.StringValue("v7"), Payload: types.StringValue("v7 payload")}, Index: 4},
				},
				ToRemove: []variantModelWithIndex{
					{Variant: VariantModel{Name: types.StringValue("v3"), Payload: types.StringValue("v3 payload")}, Index: 2},
					{Variant: VariantModel{Name: types.StringValue("v1"), Payload: types.StringValue("v1 payload")}, Index: 0},
				},
				ToReplace: []variantModelWithIndex{
					{Variant: VariantModel{Name: types.StringValue("v2"), Payload: types.StringValue("v2 payload 2")}, Index: 1},
				},
				Mode: variantDiffModeMixed,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getVariantDiffMode(tt.variants, tt.existingVariants); !reflect.DeepEqual(got, tt.result) {
				t.Errorf("getVariantDiffMode() = %v, want %v", got, tt.result)
			}
		})
	}
}
