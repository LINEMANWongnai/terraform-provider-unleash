package unleash

import (
	"context"
	"fmt"
)

type FetchedFeature struct {
	Feature FeatureSchema

	FetchedProject      string
	FetchedEnvironments []FetchedEnvironment
}

type FetchedEnvironment struct {
	Environment FeatureEnvironmentSchema

	FetchedVariants   []VariantSchema
	FetchedStrategies []FeatureStrategySchema
}

func GetFeatures(ctx context.Context, client ClientWithResponsesInterface, projectID string) ([]FetchedFeature, error) {
	featuresResp, err := client.GetFeaturesWithResponse(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if featuresResp.StatusCode() > 299 {
		return nil, fmt.Errorf("failed to get features from project %s with status %d %s", projectID, featuresResp.StatusCode(), string(featuresResp.Body))
	}

	fetchedFeatures := make([]FetchedFeature, 0, len(featuresResp.JSON200.Features))
	for _, feature := range featuresResp.JSON200.Features {
		fetched, err := fetchFeatureProperties(ctx, client, projectID, feature)
		if err != nil {
			return nil, err
		}
		fetchedFeatures = append(fetchedFeatures, fetched)
	}

	return fetchedFeatures, nil
}

func fetchFeatureProperties(ctx context.Context, client ClientWithResponsesInterface, projectID string, feature FeatureSchema) (FetchedFeature, error) {
	fetchedFeature := FetchedFeature{
		Feature: feature,

		FetchedProject: projectID,
	}

	if feature.Environments != nil {
		for _, env := range *feature.Environments {
			fetched, err := fetchEnvironmentProperties(ctx, client, projectID, feature.Name, env)
			if err != nil {
				return fetchedFeature, err
			}
			fetchedFeature.FetchedEnvironments = append(fetchedFeature.FetchedEnvironments, fetched)
		}
	}

	return fetchedFeature, nil
}

func fetchEnvironmentProperties(ctx context.Context, client ClientWithResponsesInterface, projectID string, featureName string, env FeatureEnvironmentSchema) (FetchedEnvironment, error) {
	fetchedEnv := FetchedEnvironment{
		Environment: env,
	}

	if env.Variants != nil {
		fetchedEnv.FetchedVariants = *env.Variants
	} else {
		variantsResp, err := client.GetEnvironmentFeatureVariantsWithResponse(ctx, projectID, featureName, env.Name)
		if err != nil {
			return fetchedEnv, err
		}
		if variantsResp.StatusCode() > 299 {
			return fetchedEnv, fmt.Errorf("failed to get variant for %s %s with status %d %s", projectID, featureName, variantsResp.StatusCode(), string(variantsResp.Body))
		}

		fetchedEnv.FetchedVariants = variantsResp.JSON200.Variants
	}

	if env.Strategies != nil {
		fetchedEnv.FetchedStrategies = *env.Strategies
	} else {
		strategiesResp, err := client.GetFeatureStrategiesWithResponse(ctx, projectID, featureName, env.Name)
		if err != nil {
			return fetchedEnv, err
		}
		if strategiesResp.StatusCode() > 299 {
			return fetchedEnv, fmt.Errorf("failed to get strategies for %s %s with status %d %s", projectID, featureName, strategiesResp.StatusCode(), string(strategiesResp.Body))
		}

		fetchedEnv.FetchedStrategies = *strategiesResp.JSON200
	}
	return fetchedEnv, nil
}

func GetFeature(ctx context.Context, client ClientWithResponsesInterface, projectID string, featureName string) (FetchedFeature, bool, error) {
	featureResp, err := client.GetFeatureWithResponse(ctx, projectID, featureName)
	if err != nil {
		return FetchedFeature{}, false, err
	}
	if featureResp.StatusCode() == 404 {
		return FetchedFeature{}, false, nil
	}
	if featureResp.StatusCode() > 299 {
		return FetchedFeature{}, false, fmt.Errorf("failed to get feature %s from project %s with status %d %s", featureName, projectID, featureResp.StatusCode(), string(featureResp.Body))
	}

	fetched, err := fetchFeatureProperties(ctx, client, projectID, *featureResp.JSON200)

	return fetched, true, err
}
