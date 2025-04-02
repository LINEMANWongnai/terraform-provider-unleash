// Copyright (c) HashiCorp, Inc.

package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/inmem"
)

func TestAccContextFieldResourceMinimal(t *testing.T) {
	providerConf := getProviderConf(inmem.CreateTestServer().Start(t), "")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConf + `
resource "unleash_context_field" "context1" { 
	name = "context1"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unleash_context_field.context1", "id", "context1"),
					resource.TestCheckResourceAttr("unleash_context_field.context1", "name", "context1"),
					resource.TestCheckResourceAttr("unleash_context_field.context1", "legal_values.#", "0"),
				),
			},
			// ImportState testing
			{
				ResourceName:  "unleash_context_field.context1",
				ImportStateId: "context1",
				Config: providerConf + `
resource "unleash_context_field" "context1" { 
	name = "context1"
}`,
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				ImportStateVerifyIgnore: []string{},
			},
			//	Update and Read testing
			{
				Config: providerConf + `
resource "unleash_context_field" "context1" { 
	name = "context1"
	description = "desc context1"
	sort_order = 10
	stickiness = true
	legal_values = [
		{
			value = "abc"
		},
		{
			description = "second"
			value = "def"
		}
	]
}

resource "unleash_context_field" "context2" { 
	name = "context2"
	description = ""
	sort_order = 0
	stickiness = false
	legal_values = []
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unleash_context_field.context1", "id", "context1"),
					resource.TestCheckResourceAttr("unleash_context_field.context1", "name", "context1"),
					resource.TestCheckResourceAttr("unleash_context_field.context1", "description", "desc context1"),
					resource.TestCheckResourceAttr("unleash_context_field.context1", "sort_order", "10"),
					resource.TestCheckResourceAttr("unleash_context_field.context1", "stickiness", "true"),
					resource.TestCheckResourceAttr("unleash_context_field.context1", "legal_values.#", "2"),

					resource.TestCheckResourceAttr("unleash_context_field.context2", "id", "context2"),
					resource.TestCheckResourceAttr("unleash_context_field.context2", "name", "context2"),
					resource.TestCheckResourceAttr("unleash_context_field.context2", "description", ""),
					resource.TestCheckResourceAttr("unleash_context_field.context2", "sort_order", "0"),
					resource.TestCheckResourceAttr("unleash_context_field.context2", "stickiness", "false"),
					resource.TestCheckResourceAttr("unleash_context_field.context2", "legal_values.#", "0"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
