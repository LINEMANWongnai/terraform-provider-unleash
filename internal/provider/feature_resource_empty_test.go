package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/inmem"
)

func TestAccFeatureResourceNullAndEmpty(t *testing.T) {
	providerConf := getProviderConf(inmem.CreateTestServer().Start(t), "")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConf + `
resource "unleash_feature" "nullandempty" {
	project = "default"
	name = "test-feature.nullandempty"
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
						"groupId" = "test-feature.nullandempty"
					},
					constraints = [
						{
							context_name = "userId"
							case_insensitive = true
							operator = "IN"
							inverted = true
							values_json = "[\"uid1\",\"uid2\"]"
						},
						{
							context_name = "deviceId"
							operator = "IN"
							values_json = jsonencode(["iphone"])
						}
					],
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
							weight_type = "variable"
							stickiness = "session"
						},
					]
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
							values_json = jsonencode(["uid1","uid2"])
						},
						{
							context_name = "device"
							values_json = jsonencode(["iphone"])
						},
						{
							context_name = "nullOrEmpty"
						},
					]
				},
				{
					name = "variant2"
					weight_type = "variable"
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
						"groupId" = "test-feature.nullandempty"
					}
				},
				{
					name = "default"
					disabled = false
				},
			]
		}
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unleash_feature.nullandempty", "id", "default.test-feature.nullandempty"),
					resource.TestCheckResourceAttr("unleash_feature.nullandempty", "project", "default"),
					resource.TestCheckResourceAttr("unleash_feature.nullandempty", "name", "test-feature.nullandempty"),
					resource.TestCheckResourceAttr("unleash_feature.nullandempty", "type", "release"),
				),
			},
			// ImportState testing
			{
				ResourceName:  "unleash_feature.nullandempty",
				ImportStateId: "default.test-feature.nullandempty",
				Config: providerConf + `
resource "unleash_feature" "nullandempty" {
 project = "default"
 name = "test-feature.nullandempty"
}`,
				ImportState:       true,
				ImportStateVerify: false,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				ImportStateVerifyIgnore: []string{},
			},
			{
				RefreshState: true,
			},
			//	Null to Empty testing
			{
				Config: providerConf + `
resource "unleash_feature" "nullandempty" {
	project = "default"
	name = "test-feature.nullandempty"
	type = "release"
	description = ""
	impression_data = false
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
						"groupId" = "test-feature.nullandempty"
					}
					constraints = [
						{
							case_insensitive = true
							context_name = "userId"
							operator = "IN"
							inverted = true	
							values_json = "[\"uid1\",\"uid2\"]"
						},
						{
							context_name = "deviceId"
							operator = "IN"
							values_json = jsonencode(["iphone"])
							case_insensitive = false
							inverted = false
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
							weight_type = "variable"
							stickiness = "session"
							payload = ""
							payload_type = ""
						},
					]
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
							values_json = jsonencode(["uid1","uid2"])
						},
						{
							context_name = "device"
							values_json = jsonencode(["iphone"])
						},
						{
							context_name = "nullOrEmpty"
							values_json = ""
						},
					]
				},
				{
					name = "variant2"  
					weight_type = "variable"
					payload = ""
					payload_type = ""
					stickiness = ""
					overrides = []
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
						"groupId" = "test-feature.nullandempty"
					}
				},
				{
					name = "default"
					disabled = false
					title = ""
					constraints = []
					parameters = {}
					segments = []
					variants = []
				},
			]
			variants = []
		}
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unleash_feature.nullandempty", "id", "default.test-feature.nullandempty"),
					resource.TestCheckResourceAttr("unleash_feature.nullandempty", "project", "default"),
					resource.TestCheckResourceAttr("unleash_feature.nullandempty", "name", "test-feature.nullandempty"),
					resource.TestCheckResourceAttr("unleash_feature.nullandempty", "type", "release"),
				),
			},
			//	Empty to Null testing
			{
				Config: providerConf + `
resource "unleash_feature" "nullandempty" {
	project = "default"
	name = "test-feature.nullandempty"
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
						"groupId" = "test-feature.nullandempty"
					}
					constraints = [
						{
							case_insensitive = true
							context_name = "userId"
							operator = "IN"
							inverted = true
							values_json = "[\"uid1\",\"uid2\"]"
						},
						{
							context_name = "deviceId"
							operator = "IN"
							values_json = jsonencode(["iphone"])
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
							weight_type = "variable"
							stickiness = "session"
						},
					]
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
							values_json = jsonencode(["uid1","uid2"])
						},
						{
							context_name = "device"
							values_json = jsonencode(["iphone"])
						},
						{
							context_name = "nullOrEmpty"
							values_json = null
						},
					]
				},
				{
					name = "variant2" 
					weight_type = "variable"
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
						"groupId" = "test-feature.nullandempty"
					}
				},
				{
					name = "default"
					disabled = false
				},
			]
		}
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unleash_feature.nullandempty", "id", "default.test-feature.nullandempty"),
					resource.TestCheckResourceAttr("unleash_feature.nullandempty", "project", "default"),
					resource.TestCheckResourceAttr("unleash_feature.nullandempty", "name", "test-feature.nullandempty"),
					resource.TestCheckResourceAttr("unleash_feature.nullandempty", "type", "release"),
				),
			},
			//	Strategies' order preservation
			{
				Config: providerConf + `
resource "unleash_feature" "nullandempty" {
	project = "default"
	name = "test-feature.nullandempty"
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
						"groupId" = "test-feature.nullandempty"
					}
					constraints = [
						{
							case_insensitive = true
							context_name = "userId"
							operator = "IN"
							inverted = true
							values_json = "[\"uid1\",\"uid2\"]"
						},
						{
							context_name = "deviceId"
							operator = "IN"
							values_json = jsonencode(["iphone"])
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
							weight_type = "variable"
							stickiness = "session"
						},
					]
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
							values_json = jsonencode(["uid1","uid2"])
						},
						{
							context_name = "device"
							values_json = jsonencode(["iphone"])
						},
						{
							context_name = "nullOrEmpty"
							values_json = null
						},
					]
				},
				{
					name = "variant2" 
					weight_type = "variable"
				},
			]
		}
		development = {
			enabled = false
			strategies = [
				{
					name = "default"
					disabled = false
				},
				{
					name = "flexibleRollout"
					disabled = false
					parameters = {
						"rollout" = "100"
						"stickiness" = "default"
						"groupId" = "test-feature.nullandempty"
					}
				},
			]
		}
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("unleash_feature.nullandempty", "id", "default.test-feature.nullandempty"),
					resource.TestCheckResourceAttr("unleash_feature.nullandempty", "project", "default"),
					resource.TestCheckResourceAttr("unleash_feature.nullandempty", "name", "test-feature.nullandempty"),
					resource.TestCheckResourceAttr("unleash_feature.nullandempty", "type", "release"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
