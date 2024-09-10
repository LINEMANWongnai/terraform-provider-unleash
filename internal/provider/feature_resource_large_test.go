package provider_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/inmem"
)

func TestAccFeatureResourceLarge(t *testing.T) {
	providerConf := getProviderConf(inmem.CreateTestServer().Start(t), "")

	largeConfigFmt := `
resource "unleash_feature" "large" {
	project = "default"
	name = "test-feature.large"
	type = "release"
	environments = {
		production = {
			enabled = false
			strategies = [
				{
					name = "flexibleRollout"
					disabled = false
					parameters = {
						"rollout" = "100"
						"stickiness" = "session"
						"groupId" = "test-feature.full"
					}
				},
			]
			variants = [
				%s
				{
					name = "variant1"
					payload = "payload1"
					payload_type = "string"
					weight = 5
					weight_type = "fix"
					stickiness = "default"
					overrides = [
						{
							context_name = "userId"
							values_json = jsonencode(%s)
						},
						{
							context_name = "device"
							values_json = jsonencode(["iphone"])
						},
					]
				},
				{
					name = "variant2"
					payload = "payload2"
					payload_type = "string"
					weight_type = "variable"
					stickiness = "default"
				},
				%s
			]
		}
		development = {
			enabled = false
			strategies = [
				{
					name = "flexibleRollout"
					disabled = false
					parameters = {
						"rollout" = "100"
						"stickiness" = "default"
						"groupId" = "test-feature.minimal"
					}
				},
			]
		}
	}
}
`

	notEmptyRegex := regexp.MustCompile(`.+`)

	largeCheckFn := func(variantsLen int) resource.TestCheckFunc {
		return resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("unleash_feature.large", "id", "default.test-feature.large"),
			resource.TestCheckResourceAttr("unleash_feature.large", "project", "default"),
			resource.TestCheckResourceAttr("unleash_feature.large", "name", "test-feature.large"),
			resource.TestCheckResourceAttr("unleash_feature.large", "type", "release"),
			resource.TestCheckResourceAttr("unleash_feature.large", "environments.production.enabled", "false"),
			resource.TestCheckResourceAttr("unleash_feature.large", "environments.production.strategies.#", "1"),
			resource.TestCheckTypeSetElemNestedAttrs("unleash_feature.large", "environments.production.strategies.*", map[string]string{
				"name":                  "flexibleRollout",
				"disabled":              "false",
				"parameters.rollout":    "100",
				"parameters.stickiness": "session",
				"parameters.groupId":    "test-feature.full",
			}),
			resource.TestMatchResourceAttr("unleash_feature.large", "environments.production.strategies.0.id", notEmptyRegex),
			resource.TestCheckResourceAttr("unleash_feature.large", "environments.production.variants.#", strconv.Itoa(variantsLen)),
		)
	}

	beginArray := `["-"`
	largeOverrideValues := ""
	for i := 0; i < 101; i++ {
		largeOverrideValues += fmt.Sprintf(`,"%d"`, i)
	}
	endArray := "]"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConf + fmt.Sprintf(largeConfigFmt, "", beginArray+largeOverrideValues+endArray, ""),
				Check:  largeCheckFn(2),
			},
			// ImportState testing
			{
				ResourceName:  "unleash_feature.large",
				ImportStateId: "default.test-feature.large",
				Config: providerConf + `
resource "unleash_feature" "large" {
  project = "default"
  name = "test-feature.large"
}`,
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: providerConf + fmt.Sprintf(largeConfigFmt,
					`{
					name = "variant3"
					payload = "payload3"
					payload_type = "string"
					weight_type = "variable"
					stickiness = "default"
				},`,
					beginArray+largeOverrideValues+endArray,
					`{
					name = "variant4"
					payload = "payload4"
					payload_type = "string"
					weight_type = "variable"
					stickiness = "default"
				},`),
				Check: largeCheckFn(4),
			},
			{
				Config: providerConf + fmt.Sprintf(largeConfigFmt,
					`{
					name = "variant3"
					payload = "payload3 mod"
					payload_type = "string"
					weight_type = "variable"
					stickiness = "default"
				},`,
					beginArray+largeOverrideValues+endArray,
					`{
					name = "variant4"
					payload = "payload4 mod"
					payload_type = "string"
					weight_type = "variable"
					stickiness = "default"
				},`),
				Check: largeCheckFn(4),
			},
			{
				Config: providerConf + fmt.Sprintf(largeConfigFmt, "", beginArray+largeOverrideValues+endArray, `{
					name = "variant3"
					payload = "payload3 mod 2"
					payload_type = "string"
					weight_type = "variable"
					stickiness = "default"
				},
				{
					name = "variant4"
					payload = "payload4 mod 2"
					payload_type = "string"
					weight_type = "variable"
					stickiness = "default"
				},`),
				Check: largeCheckFn(4),
			},
			{
				Config: providerConf + fmt.Sprintf(largeConfigFmt, `{
					name = "variant3"
					payload = "payload3 mod 3"
					payload_type = "string"
					weight_type = "variable"
					stickiness = "default"
				},`,
					beginArray+largeOverrideValues+endArray, `
				{
					name = "variant4"
					payload = "payload4 mod 3"
					payload_type = "string"
					weight_type = "variable"
					stickiness = "default"
				},
				{	
					name = "variant5"
					payload = "payload5"
					payload_type = "string"
					weight_type = "variable"
					stickiness = "default"
				},`),
				Check: largeCheckFn(5),
			},
			{
				Config: providerConf + fmt.Sprintf(largeConfigFmt,
					"",
					beginArray+largeOverrideValues+endArray, `
				{
					name = "variant4"
					payload = "payload4 mod 4"
					payload_type = "string"
					weight_type = "variable"
					stickiness = "default"
				},
				{	
					name = "variant5"
					payload = "payload5 mod 2"
					payload_type = "string"
					weight_type = "variable"
					stickiness = "default"
				},`),
				Check: largeCheckFn(4),
			},
			{
				Config: providerConf + fmt.Sprintf(largeConfigFmt, "", beginArray+largeOverrideValues+endArray, ""),
				Check:  largeCheckFn(2),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
