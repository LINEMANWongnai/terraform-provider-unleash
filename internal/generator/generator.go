// Copyright (c) HashiCorp, Inc.

package generator

import (
	"context"
	"io"

	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/unleash"
)

func Generate(client unleash.ClientWithResponsesInterface, projectID string, tfWriter io.Writer, importWriter io.Writer) error {
	ctx := context.Background()
	hclFile := hclwrite.NewEmptyFile()
	hclBody := hclFile.Body()
	importHclFile := hclwrite.NewEmptyFile()
	importHclBody := importHclFile.Body()

	err := genContextFields(ctx, client, hclBody, importHclBody)
	if err != nil {
		return err
	}
	err = genFeatures(ctx, client, projectID, hclBody, importHclBody)
	if err != nil {
		return err
	}
	err = genSegments(ctx, client, hclBody, importHclBody)
	if err != nil {
		return err
	}

	_, err = hclFile.WriteTo(tfWriter)
	if err != nil {
		return err
	}
	_, err = importHclFile.WriteTo(importWriter)
	return err
}
