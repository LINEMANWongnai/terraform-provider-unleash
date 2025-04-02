// Copyright (c) HashiCorp, Inc.

package generator

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/unleash"
)

func genContextFields(ctx context.Context, client unleash.ClientWithResponsesInterface, hclBody *hclwrite.Body, importHclBody *hclwrite.Body) error {
	contextFieldsResp, err := client.GetContextFieldsWithResponse(ctx)
	if err != nil {
		return err
	}
	if contextFieldsResp.StatusCode() > 299 {
		return fmt.Errorf("failed to get context fields: %d %s", contextFieldsResp.StatusCode(), string(contextFieldsResp.Body))
	}
	if contextFieldsResp.JSON200 == nil {
		return nil
	}
	for _, contextField := range *contextFieldsResp.JSON200 {
		resourceName := strings.ToLower(contextField.Name)
		resourceName = strings.ReplaceAll(strings.ReplaceAll(resourceName, ".", "_"), " ", "_")

		resource := hclBody.AppendNewBlock("resource", []string{"unleash_context_field", resourceName})
		resourceBody := resource.Body()
		resourceBody.SetAttributeValue("name", cty.StringVal(contextField.Name))
		if contextField.Description != nil {
			resourceBody.SetAttributeValue("description", cty.StringVal(*contextField.Description))
		}
		if contextField.SortOrder != nil && *contextField.SortOrder != 0 {
			resourceBody.SetAttributeValue("sort_order", cty.NumberFloatVal(float64(*contextField.SortOrder)))
		}
		if contextField.Stickiness != nil && *contextField.Stickiness {
			resourceBody.SetAttributeValue("stickiness", cty.BoolVal(*contextField.Stickiness))
		}
		if contextField.LegalValues != nil {
			resourceBody.SetAttributeValue("legal_values", toLegalValues(contextField.LegalValues))
		}

		hclBody.AppendNewline()

		importBlock := importHclBody.AppendNewBlock("import", []string{})
		importBody := importBlock.Body()
		importBody.SetAttributeRaw("to", []*hclwrite.Token{
			{
				Type:         hclsyntax.TokenQuotedLit,
				Bytes:        []byte("unleash_context_field." + resourceName),
				SpacesBefore: 0,
			},
		})
		importBody.SetAttributeValue("id", cty.StringVal(contextField.Name))
		importHclBody.AppendNewline()
	}
	return nil
}

func toLegalValues(legalValues *[]unleash.LegalValueSchema) cty.Value {
	if legalValues == nil || len(*legalValues) == 0 {
		return cty.NullVal(cty.List(legalValueType))
	}
	legalValueValues := make([]cty.Value, 0, len(*legalValues))
	for _, legalValue := range *legalValues {
		attributes := map[string]cty.Value{
			"value":       cty.StringVal(legalValue.Value),
			"description": cty.NullVal(cty.String),
		}
		if legalValue.Description != nil {
			attributes["description"] = cty.StringVal(*legalValue.Description)
		}
		legalValueValues = append(legalValueValues, cty.ObjectVal(attributes))
	}
	return cty.ListVal(legalValueValues)
}
