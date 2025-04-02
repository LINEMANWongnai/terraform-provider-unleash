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

func genSegments(ctx context.Context, client unleash.ClientWithResponsesInterface, hclBody *hclwrite.Body, importHclBody *hclwrite.Body) error {
	segmentsResp, err := client.GetSegmentsWithResponse(ctx)
	if err != nil {
		return err
	}
	if segmentsResp.StatusCode() > 299 {
		return fmt.Errorf("failed to get segments: %d %s", segmentsResp.StatusCode(), string(segmentsResp.Body))
	}
	if segmentsResp.JSON200.Segments == nil {
		return nil
	}
	for _, segment := range *segmentsResp.JSON200.Segments {
		resourceName := strings.ToLower(segment.Name)
		resourceName = strings.ReplaceAll(strings.ReplaceAll(resourceName, ".", "_"), " ", "_")

		resource := hclBody.AppendNewBlock("resource", []string{"unleash_segment", resourceName})
		resourceBody := resource.Body()
		resourceBody.SetAttributeValue("name", cty.StringVal(segment.Name))
		if segment.Project != nil {
			resourceBody.SetAttributeValue("project", cty.StringVal(*segment.Project))
		}
		if segment.Description != nil {
			resourceBody.SetAttributeValue("description", cty.StringVal(*segment.Description))
		}
		constraints, err := toConstraints(&segment.Constraints)
		if err != nil {
			return err
		}
		resourceBody.SetAttributeValue("constraints", constraints)

		hclBody.AppendNewline()

		importBlock := importHclBody.AppendNewBlock("import", []string{})
		importBody := importBlock.Body()
		importBody.SetAttributeRaw("to", []*hclwrite.Token{
			{
				Type:         hclsyntax.TokenQuotedLit,
				Bytes:        []byte("unleash_segment." + resourceName),
				SpacesBefore: 0,
			},
		})
		importBody.SetAttributeValue("id", cty.StringVal(fmt.Sprintf("%d", segment.Id)))
		importHclBody.AppendNewline()
	}
	return nil
}
