package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/inmem"
)

func TestAccSegmentResourceMinimal(t *testing.T) {
	providerConf := getProviderConf(inmem.CreateTestServer().Start(t), "")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConf + `
resource "unleash_segment" "segment1" {
	project = "default"
	name = "segment1"
	description = "desc segment1"
	constraints = [{
			case_insensitive = true
			context_name = "userId"
			operator = "IN"
			inverted = false	
			values_json = jsonencode(["uid1", "uid2", "uid3"])
		},
		{
			case_insensitive = true
			context_name = "businessId"
			operator = "IN"
			inverted = false	
			values_json = "[\"m1\",\"m2\"]"
		},
	]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unleash_segment.segment1", "id", "1"),
					resource.TestCheckResourceAttr("unleash_segment.segment1", "project", "default"),
					resource.TestCheckResourceAttr("unleash_segment.segment1", "name", "segment1"),
					resource.TestCheckResourceAttr("unleash_segment.segment1", "description", "desc segment1"),
					resource.TestCheckResourceAttr("unleash_segment.segment1", "constraints.#", "2"),
				),
			},
			// ImportState testing
			{
				ResourceName:  "unleash_segment.segment1",
				ImportStateId: "1",
				Config: providerConf + `
resource "unleash_segment" "segment1" {
 project = "default"
 name = "segment1"
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
resource "unleash_segment" "segment1" {
	project = "default"
	name = "segment1"
	description = "desc segment1 mod"
	constraints = [{
			case_insensitive = true
			context_name = "userId"
			operator = "IN"
			inverted = false	
			values_json = jsonencode(["uid1", "uid2", "uid3", "uid4"])
		},
		{
			case_insensitive = true
			context_name = "businessId"
			operator = "IN"
			inverted = false	
			values_json = "[\"m1\",\"m2\",\"m3\"]"
		},
		{
			case_insensitive = true
			context_name = "device"
			operator = "IN"
			inverted = false	
			values_json = "[\"device1\",\"device2\"]"
		},
	]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unleash_segment.segment1", "id", "1"),
					resource.TestCheckResourceAttr("unleash_segment.segment1", "project", "default"),
					resource.TestCheckResourceAttr("unleash_segment.segment1", "name", "segment1"),
					resource.TestCheckResourceAttr("unleash_segment.segment1", "description", "desc segment1 mod"),
					resource.TestCheckResourceAttr("unleash_segment.segment1", "constraints.#", "3"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
