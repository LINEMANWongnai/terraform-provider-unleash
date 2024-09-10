package provider_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/inmem"
	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/ptr"
	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/unleash"
)

func TestAccFeatureResourceIgnoreStrategy(t *testing.T) {
	unleashTestServer := inmem.CreateTestServer()
	ctx := context.Background()
	_, _ = unleashTestServer.CreateFeature(ctx, unleash.CreateFeatureRequestObject{
		ProjectId: "default",
		Body: &unleash.CreateFeatureJSONRequestBody{
			Name: "test-feature.automated",
			Type: ptr.ToPtr("release"),
		},
	})
	_, _ = unleashTestServer.AddFeatureStrategy(ctx, unleash.AddFeatureStrategyRequestObject{
		ProjectId:   "default",
		FeatureName: "test-feature.automated",
		Environment: "production",
		Body: &unleash.AddFeatureStrategyJSONRequestBody{
			Name:     "standard",
			Title:    ptr.ToPtr("My AutomatedTest"),
			Disabled: ptr.ToPtr(false),
		},
	})
	_, _ = unleashTestServer.AddFeatureStrategy(ctx, unleash.AddFeatureStrategyRequestObject{
		ProjectId:   "default",
		FeatureName: "test-feature.automated",
		Environment: "development",
		Body: &unleash.AddFeatureStrategyJSONRequestBody{
			Name:     "standard",
			Title:    ptr.ToPtr("Standard Title"),
			Disabled: ptr.ToPtr(false),
		},
	})

	providerConf := getProviderConf(unleashTestServer.Start(t), ".*AutomatedTest.*")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConf + `
import {
	to = unleash_feature.automated
	id = "default.test-feature.automated"
}

resource "unleash_feature" "automated" {
	project = "default"
	name = "test-feature.automated"
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
						"groupId" = "test-feature.automated"
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
						"groupId" = "test-feature.automated"
					}
				},
				{
					name = "standard"
					disabled = false
					title = "Standard Title"
				},
			]
		}
	}
}`,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				ImportStateVerifyIgnore: []string{},
			},
			//	Update and Read testing
			{
				Config: providerConf + `
resource "unleash_feature" "automated" {
	project = "default"
	name = "test-feature.automated"
	description = "desc test-feature.automated"
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
						"groupId" = "test-feature.automated"
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
						"groupId" = "test-feature.automated"
					}
				},
			]
		}
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					func(state *terraform.State) error {
						resp, _ := unleashTestServer.GetFeatureStrategies(ctx, unleash.GetFeatureStrategiesRequestObject{
							ProjectId:   "default",
							FeatureName: "test-feature.automated",
							Environment: "production",
						})
						strategiesResp := resp.(unleash.GetFeatureStrategies200JSONResponse)
						if len(strategiesResp) != 2 {
							return fmt.Errorf("expected 2 strategies, got %d", len(strategiesResp))
						}
						found := false
						for _, strategy := range strategiesResp {
							if strategy.Title != nil && *strategy.Title == "My AutomatedTest" {
								found = true
								break
							}
						}
						if !found {
							return fmt.Errorf("expected strategy with title My AutomatedTest not found in %v", strategiesResp)
						}
						return nil
					},
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
