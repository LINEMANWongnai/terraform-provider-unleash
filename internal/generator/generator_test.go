package generator_test

import (
	"bytes"
	"context"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/generator"
	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/inmem"
	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/ptr"
	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/unleash"
)

func TestGenerate(t *testing.T) {
	server := inmem.CreateTestServer()
	port := server.Start(t)

	client, err := unleash.CreateClient("http://localhost:"+strconv.Itoa(port), "any")
	require.NoError(t, err)

	var testCases = []struct {
		name                           string
		projectID                      string
		featureName                    string
		environmentToggleByEnvironment map[string]bool
		strategiesByEnvironment        map[string][]unleash.AddFeatureStrategyJSONRequestBody
		variantsByEnvironment          map[string][]unleash.VariantSchema
		expectedTf                     string
		expectedImportTf               string
	}{
		{
			name:                  "minimal",
			projectID:             "default",
			featureName:           "test.feature.minimal",
			variantsByEnvironment: map[string][]unleash.VariantSchema{},
			expectedTf: `resource "unleash_feature" "test_feature_minimal" {
  project = "default"
  name    = "test.feature.minimal"
  type    = "release"
  environments = {
    development = {
      enabled = false
      strategies = [{
        constraints = null
        disabled    = false
        name        = "flexibleRollout"
        parameters = {
          groupId    = "test.feature.minimal"
          rollout    = "100"
          stickiness = "default"
        }
        segments   = null
        sort_order = null
        title      = null
        variants   = null
      }]
      variants = null
    }
    production = {
      enabled = false
      strategies = [{
        constraints = null
        disabled    = false
        name        = "flexibleRollout"
        parameters = {
          groupId    = "test.feature.minimal"
          rollout    = "100"
          stickiness = "default"
        }
        segments   = null
        sort_order = null
        title      = null
        variants   = null
      }]
      variants = null
    }
  }
}`,
			expectedImportTf: `import {
  to =unleash_feature.test_feature_minimal
  id = "default.test.feature.minimal"
}`,
		},
		{
			name:        "full",
			projectID:   "myproject",
			featureName: "test.feature.full",
			environmentToggleByEnvironment: map[string]bool{
				"development": true,
				"production":  true,
			},
			strategiesByEnvironment: map[string][]unleash.AddFeatureStrategyJSONRequestBody{
				"development": {
					{
						Name: "myRolloutStrategy",
						Constraints: ptr.ToPtr([]unleash.ConstraintSchema{
							{
								CaseInsensitive: ptr.ToPtr(true),
								ContextName:     "userId",
								Inverted:        ptr.ToPtr(true),
								Operator:        "IN",
								Value:           ptr.ToPtr("1"),
								Values:          ptr.ToPtr([]string{"3", "4"}),
							},
							{
								ContextName: "restaurantID",
								Operator:    "IN",
								Value:       ptr.ToPtr("a"),
							},
						}),
						Disabled:   ptr.ToPtr(true),
						Parameters: ptr.ToPtr[unleash.ParametersSchema](map[string]string{"groupId": "test.feature.full"}),
						Segments:   ptr.ToPtr([]float32{1, 2, 3}),
						SortOrder:  ptr.ToPtr(float32(992)),
						Title:      ptr.ToPtr("My Rollout Strategy1"),
						Variants: ptr.ToPtr([]unleash.CreateStrategyVariantSchema{
							{
								Name: "variant1",
								Payload: &struct {
									Type  unleash.CreateStrategyVariantSchemaPayloadType `json:"type"`
									Value string                                         `json:"value"`
								}{
									Type:  "string",
									Value: "variant1 value",
								},
								Stickiness: "default",
								Weight:     100,
								WeightType: "fix",
							},
							{
								Name: "variant2",
								Payload: &struct {
									Type  unleash.CreateStrategyVariantSchemaPayloadType `json:"type"`
									Value string                                         `json:"value"`
								}{
									Type:  "string",
									Value: "variant1 value",
								},
								Stickiness: "default",
								WeightType: "variable",
							},
						}),
					},
					{
						Name: "myRolloutStrategy2",
						Constraints: ptr.ToPtr([]unleash.ConstraintSchema{
							{
								CaseInsensitive: ptr.ToPtr(false),
								ContextName:     "userId",
								Inverted:        ptr.ToPtr(false),
								Operator:        "IN",
								Values:          ptr.ToPtr([]string{"9", "10"}),
							},
						}),
						Disabled:   ptr.ToPtr(false),
						Parameters: ptr.ToPtr[unleash.ParametersSchema](map[string]string{"groupId": "test.feature.full"}),
						SortOrder:  ptr.ToPtr(float32(0)),
						Title:      ptr.ToPtr("My Rollout Strategy 2"),
						Variants: ptr.ToPtr([]unleash.CreateStrategyVariantSchema{
							{
								Name: "variant1",
								Payload: &struct {
									Type  unleash.CreateStrategyVariantSchemaPayloadType `json:"type"`
									Value string                                         `json:"value"`
								}{
									Type:  "string",
									Value: "variant1 value",
								},
								Stickiness: "session",
								WeightType: "variable",
							},
						}),
					},
				},
				"production": {
					{
						Name: "standard",
						Constraints: ptr.ToPtr([]unleash.ConstraintSchema{
							{
								CaseInsensitive: ptr.ToPtr(false),
								ContextName:     "userId",
								Inverted:        ptr.ToPtr(false),
								Operator:        "IN",
							},
						}),
						Parameters: ptr.ToPtr[unleash.ParametersSchema](map[string]string{"groupId": "test.feature.full"}),
						Variants: ptr.ToPtr([]unleash.CreateStrategyVariantSchema{
							{
								Name: "variant",
								Payload: &struct {
									Type  unleash.CreateStrategyVariantSchemaPayloadType `json:"type"`
									Value string                                         `json:"value"`
								}{
									Type:  "string",
									Value: "variant value",
								},
								Stickiness: "device",
								WeightType: "variable",
							},
						}),
					},
				},
			},
			variantsByEnvironment: map[string][]unleash.VariantSchema{
				"development": {
					{
						Name: "dev variant1",
						Overrides: ptr.ToPtr([]unleash.OverrideSchema{
							{
								ContextName: "userID",
								Values:      []string{"1", "2", "3"},
							},
							{
								ContextName: "UA",
								Values:      []string{"x", "y"},
							},
						}),
						Payload: &struct {
							Type  unleash.VariantSchemaPayloadType `json:"type"`
							Value string                           `json:"value"`
						}{
							Type:  "json",
							Value: "{value}",
						},
						Stickiness: ptr.ToPtr("default"),
						Weight:     10,
						WeightType: ptr.ToPtr[unleash.VariantSchemaWeightType]("fix"),
					},
					{
						Name: "dev variant2",
						Payload: &struct {
							Type  unleash.VariantSchemaPayloadType `json:"type"`
							Value string                           `json:"value"`
						}{
							Type:  "csv",
							Value: "variant value",
						},
						Stickiness: ptr.ToPtr("session"),
						WeightType: ptr.ToPtr[unleash.VariantSchemaWeightType]("variable"),
					},
				},
				"production": {
					{
						Name: "prod variant1",
						Overrides: ptr.ToPtr([]unleash.OverrideSchema{
							{
								ContextName: "UA",
								Values:      []string{"x", "y"},
							},
						}),
						Payload: &struct {
							Type  unleash.VariantSchemaPayloadType `json:"type"`
							Value string                           `json:"value"`
						}{
							Type:  "json",
							Value: "{value}",
						},
						WeightType: ptr.ToPtr[unleash.VariantSchemaWeightType]("variable"),
					},
				},
			},
			expectedTf: `resource "unleash_feature" "test_feature_full" {
  project = "myproject"
  name    = "test.feature.full"
  type    = "release"
  environments = {
    development = {
      enabled = true
      strategies = [{
        constraints = [{
          case_insensitive = true
          context_name     = "userId"
          inverted         = true
          operator         = "IN"
          values_json      = "[\"1\",\"3\",\"4\"]"
          }, {
          case_insensitive = null
          context_name     = "restaurantID"
          inverted         = null
          operator         = "IN"
          values_json      = "[\"a\"]"
        }]
        disabled = true
        name     = "myRolloutStrategy"
        parameters = {
          groupId = "test.feature.full"
        }
        segments   = [1, 2, 3]
        sort_order = 992
        title      = "My Rollout Strategy1"
        variants = [{
          name         = "variant1"
          payload      = "variant1 value"
          payload_type = "string"
          stickiness   = "default"
          weight       = 100
          weight_type  = "fix"
          }, {
          name         = "variant2"
          payload      = "variant1 value"
          payload_type = "string"
          stickiness   = "default"
          weight       = null
          weight_type  = "variable"
        }]
        }, {
        constraints = [{
          case_insensitive = false
          context_name     = "userId"
          inverted         = false
          operator         = "IN"
          values_json      = "[\"9\",\"10\"]"
        }]
        disabled = false
        name     = "myRolloutStrategy2"
        parameters = {
          groupId = "test.feature.full"
        }
        segments   = null
        sort_order = null
        title      = "My Rollout Strategy 2"
        variants = [{
          name         = "variant1"
          payload      = "variant1 value"
          payload_type = "string"
          stickiness   = "session"
          weight       = null
          weight_type  = "variable"
        }]
      }]
      variants = [{
        name = "dev variant1"
        overrides = [{
          context_name = "userID"
          values_json  = "[\"1\",\"2\",\"3\"]"
          }, {
          context_name = "UA"
          values_json  = "[\"x\",\"y\"]"
        }]
        payload      = "{value}"
        payload_type = "json"
        stickiness   = "default"
        weight       = 10
        weight_type  = "fix"
        }, {
        name         = "dev variant2"
        overrides    = null
        payload      = "variant value"
        payload_type = "csv"
        stickiness   = "session"
        weight       = null
        weight_type  = "variable"
      }]
    }
    production = {
      enabled = true
      strategies = [{
        constraints = [{
          case_insensitive = false
          context_name     = "userId"
          inverted         = false
          operator         = "IN"
          values_json      = null
        }]
        disabled = false
        name     = "standard"
        parameters = {
          groupId = "test.feature.full"
        }
        segments   = null
        sort_order = null
        title      = null
        variants = [{
          name         = "variant"
          payload      = "variant value"
          payload_type = "string"
          stickiness   = "device"
          weight       = null
          weight_type  = "variable"
        }]
      }]
      variants = [{
        name = "prod variant1"
        overrides = [{
          context_name = "UA"
          values_json  = "[\"x\",\"y\"]"
        }]
        payload      = "{value}"
        payload_type = "json"
        stickiness   = null
        weight       = null
        weight_type  = "variable"
      }]
    }
  }
}`,
			expectedImportTf: `import {
  to =unleash_feature.test_feature_full
  id = "myproject.test.feature.full"
}`,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(tt *testing.T) {
			tfWriter := &bytes.Buffer{}
			importWriter := &bytes.Buffer{}
			ctx := context.Background()

			_, _ = server.CreateFeature(ctx, unleash.CreateFeatureRequestObject{
				ProjectId: testCase.projectID,
				Body: &unleash.CreateFeatureJSONRequestBody{
					Description:    nil,
					ImpressionData: nil,
					Name:           testCase.featureName,
					Type:           ptr.ToPtr("release"),
				},
			})
			for name, toggle := range testCase.environmentToggleByEnvironment {
				if toggle {
					_, _ = server.ToggleFeatureEnvironmentOn(ctx, unleash.ToggleFeatureEnvironmentOnRequestObject{
						ProjectId:   testCase.projectID,
						FeatureName: testCase.featureName,
						Environment: name,
					})
				} else {
					_, _ = server.ToggleFeatureEnvironmentOff(ctx, unleash.ToggleFeatureEnvironmentOffRequestObject{
						ProjectId:   testCase.projectID,
						FeatureName: testCase.featureName,
						Environment: name,
					})
				}
			}
			for name, strategies := range testCase.strategiesByEnvironment {
				for _, strategy := range strategies {
					_, _ = server.AddFeatureStrategy(ctx, unleash.AddFeatureStrategyRequestObject{
						ProjectId:   testCase.projectID,
						FeatureName: testCase.featureName,
						Environment: name,
						Body:        &strategy,
					})
				}
			}
			for name, variants := range testCase.variantsByEnvironment {
				_, _ = server.OverwriteFeatureVariantsOnEnvironments(ctx, unleash.OverwriteFeatureVariantsOnEnvironmentsRequestObject{
					ProjectId:   testCase.projectID,
					FeatureName: testCase.featureName,
					Body: &unleash.OverwriteFeatureVariantsOnEnvironmentsJSONRequestBody{
						Environments: ptr.ToPtr([]string{name}),
						Variants:     ptr.ToPtr(variants),
					},
				})
			}

			err = generator.Generate(client, testCase.projectID, tfWriter, importWriter)
			require.NoError(tt, err)

			assertTfEqual(tt, testCase.expectedTf, tfWriter.String())
			assertTfEqual(tt, testCase.expectedImportTf, importWriter.String())
		})
	}
}

func assertTfEqual(tt *testing.T, expected string, actual string) {
	assert.Equal(tt, convertMultipleSpacesToSingleSpace(expected), convertMultipleSpacesToSingleSpace(actual))
}

var multipleSpacesRegex = regexp.MustCompile(`\s\s+`)
var newLineRegex = regexp.MustCompile("[\r\n]+")

func convertMultipleSpacesToSingleSpace(v string) string {
	return string(multipleSpacesRegex.ReplaceAll(newLineRegex.ReplaceAll([]byte(v), []byte("")), []byte(" ")))
}
