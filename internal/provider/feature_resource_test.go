package provider_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/inmem"
)

func TestAccFeatureResourceMinimal(t *testing.T) {
	providerConf := getProviderConf(inmem.CreateTestServer().Start(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConf + `
resource "unleash_feature" "minimal" {
	project = "default"
	name = "test-feature.minimal"
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
						"stickiness" = "default"
						"groupId" = "test-feature.minimal"
					}
				},
			]
		}
		development = {
			enabled = true
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
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unleash_feature.minimal", "id", "default.test-feature.minimal"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "project", "default"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "name", "test-feature.minimal"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "type", "release"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "environments.production.enabled", "false"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "environments.production.strategies.0.name", "flexibleRollout"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "environments.production.strategies.0.disabled", "false"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "environments.production.strategies.0.parameters.rollout", "100"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "environments.production.strategies.0.parameters.stickiness", "default"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "environments.production.strategies.0.parameters.groupId", "test-feature.minimal"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "environments.development.enabled", "true"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "environments.development.strategies.0.name", "flexibleRollout"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "environments.development.strategies.0.disabled", "false"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "environments.development.strategies.0.parameters.rollout", "100"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "environments.development.strategies.0.parameters.stickiness", "default"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "environments.development.strategies.0.parameters.groupId", "test-feature.minimal"),
				),
			},
			// ImportState testing
			{
				ResourceName:  "unleash_feature.minimal",
				ImportStateId: "default.test-feature.minimal",
				Config: providerConf + `
resource "unleash_feature" "minimal" {
 project = "default"
 name = "test-feature.minimal"
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
resource "unleash_feature" "minimal" {
	project = "default"
	name = "test-feature.minimal"
	description = "desc test-feature.minimal"
	type = "release"
	environments = {
		production = {
			enabled = true
			strategies = [
				{
					name = "flexibleRollout"
					disabled = false
					parameters = {
						"rollout" = "100"
						"stickiness" = "session"
						"groupId" = "test-feature.minimal"
					}
				},
			]
		}
		development = {
			enabled = false
			strategies = [
				{
					name = "flexibleRollout"
					disabled = false
					parameters = {
						"rollout" = "50"
						"stickiness" = "default"
						"groupId" = "test-feature.minimal"
					}
				},
			]
		}
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unleash_feature.minimal", "id", "default.test-feature.minimal"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "project", "default"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "name", "test-feature.minimal"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "description", "desc test-feature.minimal"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "type", "release"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "environments.production.enabled", "true"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "environments.production.strategies.0.name", "flexibleRollout"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "environments.production.strategies.0.disabled", "false"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "environments.production.strategies.0.parameters.rollout", "100"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "environments.production.strategies.0.parameters.stickiness", "session"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "environments.production.strategies.0.parameters.groupId", "test-feature.minimal"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "environments.development.enabled", "false"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "environments.development.strategies.0.name", "flexibleRollout"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "environments.development.strategies.0.disabled", "false"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "environments.development.strategies.0.parameters.rollout", "50"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "environments.development.strategies.0.parameters.stickiness", "default"),
					resource.TestCheckResourceAttr("unleash_feature.minimal", "environments.development.strategies.0.parameters.groupId", "test-feature.minimal"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccFeatureResourceFull(t *testing.T) {
	providerConf := getProviderConf(inmem.CreateTestServer().Start(t))

	fullConfig := `
resource "unleash_feature" "full" {
	project = "default"
	name = "test-feature.full"
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
					constraints = [{
							case_insensitive = true
							context_name = "userId"
							operator = "IN"
							inverted = false	
							values_json = "[\"uid1\",\"uid2\"]"
						}
					]
					variants = [
						{
							name = "strategy_variant1"
							payload = "payload1"
							payload_type = "string"
							weight_type = "variable"
							stickiness = "default"
						},
						{
							name = "strategy_variant2"
							payload = "payload2"
							payload_type = "string"
							weight = 500
							weight_type = "fix"
							stickiness = "default"
						},
					]
				},
				{
					name = "flexibleRollout"
					title = "another rollout with session"
					disabled = false
					sort_order = 1
					parameters = {
						"rollout" = "50"
						"stickiness" = "default"
						"groupId" = "test-feature.full"
					}
				},
			]
			variants = [
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
							values_json = "[\"uid1\",\"uid2\"]"
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
						"groupId" = "test-feature.full"
					}
				},
			]
		}
	}
}
`

	notEmptyRegex := regexp.MustCompile(`.+`)

	fullChecker := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("unleash_feature.full", "id", "default.test-feature.full"),
		resource.TestCheckResourceAttr("unleash_feature.full", "project", "default"),
		resource.TestCheckResourceAttr("unleash_feature.full", "name", "test-feature.full"),
		resource.TestCheckResourceAttr("unleash_feature.full", "type", "release"),
		resource.TestCheckResourceAttr("unleash_feature.full", "environments.production.enabled", "false"),
		resource.TestCheckResourceAttr("unleash_feature.full", "environments.production.strategies.#", "2"),
		resource.TestCheckTypeSetElemNestedAttrs("unleash_feature.full", "environments.production.strategies.*", map[string]string{
			"name":                  "flexibleRollout",
			"disabled":              "false",
			"parameters.rollout":    "100",
			"parameters.stickiness": "session",
			"parameters.groupId":    "test-feature.full",
		}),
		resource.TestMatchResourceAttr("unleash_feature.full", "environments.production.strategies.0.id", notEmptyRegex),
		resource.TestCheckTypeSetElemNestedAttrs("unleash_feature.full", "environments.production.strategies.*", map[string]string{
			"name":                  "flexibleRollout",
			"disabled":              "false",
			"title":                 "another rollout with session",
			"parameters.rollout":    "50",
			"parameters.stickiness": "default",
			"parameters.groupId":    "test-feature.full",
		}),
		resource.TestMatchResourceAttr("unleash_feature.full", "environments.production.strategies.1.id", notEmptyRegex),
		resource.TestCheckResourceAttr("unleash_feature.full", "environments.development.enabled", "false"),
		resource.TestCheckResourceAttr("unleash_feature.full", "environments.development.strategies.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs("unleash_feature.full", "environments.development.strategies.*", map[string]string{
			"name":                  "flexibleRollout",
			"disabled":              "false",
			"parameters.rollout":    "100",
			"parameters.stickiness": "default",
			"parameters.groupId":    "test-feature.full",
		}),
		resource.TestMatchResourceAttr("unleash_feature.full", "environments.development.strategies.0.id", notEmptyRegex),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConf + fullConfig,
				Check:  fullChecker,
			},
			// ImportState testing
			{
				ResourceName:  "unleash_feature.full",
				ImportStateId: "default.test-feature.full",
				Config: providerConf + `
resource "unleash_feature" "full" {
  project = "default"
  name = "test-feature.full"
}`,
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				ImportStateVerifyIgnore: []string{},
			},
			// No change
			{
				Config: providerConf + fullConfig,
				Check:  fullChecker,
			},
			// Add/Remove some
			{
				Config: providerConf + `
resource "unleash_feature" "full" {
	project = "default"
	name = "test-feature.full"
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
					variants = [
						{
							name = "strategy_variant1"
							payload = "payload1"
							payload_type = "string"
							weight_type = "variable"
							stickiness = "default"
						},
						{
							name = "strategy_variant2"
							payload = "payload2"
							payload_type = "string"
							weight = 500
							weight_type = "fix"
							stickiness = "default"
						},
						{
							name = "strategy_variant3"
							payload = "payload3"
							payload_type = "string"
							weight = 100
							weight_type = "fix"
							stickiness = "default"
						},
					]
				},
				{
					name = "flexibleRollout"
					title = "another rollout with session"
					disabled = false
					sort_order = 1
					parameters = {
						"rollout" = "50"
						"stickiness" = "default"
						"groupId" = "test-feature.full"
					}
				},
				{
					name = "flexibleRollout"
					title = "another rollout with session 2"
					disabled = false
					sort_order = 2
					parameters = {
						"rollout" = "10"
						"stickiness" = "default"
						"groupId" = "test-feature.full"
					}
				},
			]
			variants = [
				{
					name = "variant2"
					payload = "payload2"
					payload_type = "string"
					weight_type = "variable"
					stickiness = "default"
				},
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
						"groupId" = "test-feature.full"
					}
				},
				{
					name = "flexibleRollout"
					disabled = false
					title = "another rollout with session"
					sort_order = 1
					parameters = {
						"rollout" = "50"
						"stickiness" = "session"
						"groupId" = "test-feature.full"
					}
				},
			]
		}
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unleash_feature.full", "id", "default.test-feature.full"),
					resource.TestCheckResourceAttr("unleash_feature.full", "project", "default"),
					resource.TestCheckResourceAttr("unleash_feature.full", "name", "test-feature.full"),
					resource.TestCheckResourceAttr("unleash_feature.full", "type", "release"),
					resource.TestCheckResourceAttr("unleash_feature.full", "environments.production.enabled", "false"),
					resource.TestCheckResourceAttr("unleash_feature.full", "environments.production.strategies.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs("unleash_feature.full", "environments.production.strategies.*", map[string]string{
						"name":                  "flexibleRollout",
						"disabled":              "false",
						"parameters.rollout":    "100",
						"parameters.stickiness": "session",
						"parameters.groupId":    "test-feature.full",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("unleash_feature.full", "environments.production.strategies.*", map[string]string{
						"name":                  "flexibleRollout",
						"disabled":              "false",
						"title":                 "another rollout with session",
						"parameters.rollout":    "50",
						"parameters.stickiness": "default",
						"parameters.groupId":    "test-feature.full",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("unleash_feature.full", "environments.production.strategies.*", map[string]string{
						"name":                  "flexibleRollout",
						"disabled":              "false",
						"title":                 "another rollout with session 2",
						"parameters.rollout":    "10",
						"parameters.stickiness": "default",
						"parameters.groupId":    "test-feature.full",
					}),
					resource.TestCheckResourceAttr("unleash_feature.full", "environments.development.enabled", "false"),
					resource.TestCheckResourceAttr("unleash_feature.full", "environments.development.strategies.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("unleash_feature.full", "environments.development.strategies.*", map[string]string{
						"name":                  "flexibleRollout",
						"disabled":              "false",
						"parameters.rollout":    "100",
						"parameters.stickiness": "default",
						"parameters.groupId":    "test-feature.full",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("unleash_feature.full", "environments.development.strategies.*", map[string]string{
						"name":                  "flexibleRollout",
						"disabled":              "false",
						"parameters.rollout":    "50",
						"parameters.stickiness": "session",
						"parameters.groupId":    "test-feature.full",
					}),
				),
			},
			// More update some
			{
				Config: providerConf + `
resource "unleash_feature" "full" {
	project = "default"
	name = "test-feature.full"
	description = "desc test-feature.full"
	type = "release"
	environments = {
		production = {
			enabled = true
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
		}
		development = {
			enabled = true
			strategies = [
				{
					name = "flexibleRollout"
					disabled = false
					parameters = {
						"rollout" = "50"
						"stickiness" = "default"
						"groupId" = "test-feature.full"
					}
				},
			]
		}
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unleash_feature.full", "id", "default.test-feature.full"),
					resource.TestCheckResourceAttr("unleash_feature.full", "project", "default"),
					resource.TestCheckResourceAttr("unleash_feature.full", "name", "test-feature.full"),
					resource.TestCheckResourceAttr("unleash_feature.full", "description", "desc test-feature.full"),
					resource.TestCheckResourceAttr("unleash_feature.full", "type", "release"),
					resource.TestCheckResourceAttr("unleash_feature.full", "environments.production.enabled", "true"),
					resource.TestCheckResourceAttr("unleash_feature.full", "environments.production.strategies.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("unleash_feature.full", "environments.production.strategies.*", map[string]string{
						"name":                  "flexibleRollout",
						"disabled":              "false",
						"parameters.rollout":    "100",
						"parameters.stickiness": "session",
						"parameters.groupId":    "test-feature.full",
					}),
					resource.TestCheckResourceAttr("unleash_feature.full", "environments.development.enabled", "true"),
					resource.TestCheckResourceAttr("unleash_feature.full", "environments.development.strategies.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("unleash_feature.full", "environments.development.strategies.*", map[string]string{
						"name":                  "flexibleRollout",
						"disabled":              "false",
						"parameters.rollout":    "50",
						"parameters.stickiness": "default",
						"parameters.groupId":    "test-feature.full",
					}),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
