package inmem

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/ptr"
	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/unleash"
)

var _ unleash.StrictServerInterface = &TestServer{}

type TestServer struct {
	features map[string]map[string]unleash.FeatureSchema
	lock     *sync.RWMutex
	next     *atomic.Int32
}

func CreateTestServer() *TestServer {
	return &TestServer{
		features: make(map[string]map[string]unleash.FeatureSchema),
		lock:     &sync.RWMutex{},
		next:     &atomic.Int32{},
	}
}

func (t TestServer) Start(tt *testing.T) int {
	return startHTTPServer(tt, t.register)
}

func (t TestServer) register(engine *gin.Engine) error {
	unleash.RegisterHandlers(engine, unleash.NewStrictHandler(t, nil))
	return nil
}

func (t TestServer) getProjectFeatures(projectID string) map[string]unleash.FeatureSchema {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.getProjectFeaturesNoLock(projectID)
}

func (t TestServer) getProjectFeaturesNoLock(projectID string) map[string]unleash.FeatureSchema {
	features, ok := t.features[projectID]
	if !ok {
		t.features[projectID] = make(map[string]unleash.FeatureSchema)
		return t.features[projectID]
	}

	return features
}

func (t TestServer) getFeature(projectID string, featureName string) (unleash.FeatureSchema, bool) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	projectFeatures := t.getProjectFeaturesNoLock(projectID)

	feature, ok := projectFeatures[featureName]

	return feature, ok
}

func (t TestServer) getFeatures(projectID string) []unleash.FeatureSchema {
	t.lock.RLock()
	defer t.lock.RUnlock()

	projectFeatures := t.getProjectFeaturesNoLock(projectID)

	features := make([]unleash.FeatureSchema, 0, len(projectFeatures))
	for _, f := range projectFeatures {
		features = append(features, f)
	}

	return features
}

func (t TestServer) replaceFeature(feature unleash.FeatureSchema) {
	t.lock.Lock()
	defer t.lock.Unlock()

	projectFeatures := t.getProjectFeaturesNoLock(*feature.Project)

	projectFeatures[feature.Name] = feature
}

func (t TestServer) deleteFeature(projectID string, featureName string) bool {
	t.lock.Lock()
	defer t.lock.Unlock()

	projectFeatures := t.getProjectFeaturesNoLock(projectID)

	_, ok := projectFeatures[featureName]
	if !ok {
		return false
	}
	delete(projectFeatures, featureName)

	return true
}

func (t TestServer) getNext(name string) string {
	t.next.Add(1)

	return name + "_" + strconv.Itoa(int(t.next.Load()))
}

func (t TestServer) CreateFeature(_ context.Context, request unleash.CreateFeatureRequestObject) (unleash.CreateFeatureResponseObject, error) {
	_, ok := t.getFeature(request.ProjectId, request.Body.Name)
	if ok {
		return unleash.CreateFeature404JSONResponse{}, nil
	}
	projectID := request.ProjectId
	environments := []unleash.FeatureEnvironmentSchema{
		{
			Name:        "development",
			FeatureName: ptr.ToPtr(request.Body.Name),
		},
		{
			Name:        "production",
			FeatureName: ptr.ToPtr(request.Body.Name),
		},
	}
	feature := unleash.FeatureSchema{
		Project:      &projectID,
		Name:         request.Body.Name,
		Type:         request.Body.Type,
		Description:  request.Body.Description,
		Environments: &environments,
	}
	t.replaceFeature(feature)

	return unleash.CreateFeature200JSONResponse(feature), nil
}

func (t TestServer) GetFeature(_ context.Context, request unleash.GetFeatureRequestObject) (unleash.GetFeatureResponseObject, error) {
	feature, ok := t.getFeature(request.ProjectId, request.FeatureName)
	if !ok {
		return unleash.GetFeature404JSONResponse{}, nil
	}

	return unleash.GetFeature200JSONResponse(removeProperties(feature)), nil
}

func removeProperties(feature unleash.FeatureSchema) unleash.FeatureSchema {
	var copiedEnv []unleash.FeatureEnvironmentSchema
	for _, environment := range *feature.Environments {
		// the real server never returns strategies and variants with feature
		environment.Strategies = nil
		environment.Variants = nil
		copiedEnv = append(copiedEnv, environment)
	}
	feature.Environments = &copiedEnv

	return feature
}

func (t TestServer) GetFeatures(_ context.Context, request unleash.GetFeaturesRequestObject) (unleash.GetFeaturesResponseObject, error) {
	features := t.getFeatures(request.ProjectId)
	for i, feature := range features {
		features[i] = removeProperties(feature)
	}
	return unleash.GetFeatures200JSONResponse{
		Features: features,
	}, nil
}

func (t TestServer) UpdateFeature(_ context.Context, request unleash.UpdateFeatureRequestObject) (unleash.UpdateFeatureResponseObject, error) {
	feature, ok := t.getFeature(request.ProjectId, request.FeatureName)
	if !ok {
		return unleash.UpdateFeature404JSONResponse{}, nil
	}
	feature.Description = request.Body.Description
	feature.Type = request.Body.Type
	feature.Archived = request.Body.Archived
	t.replaceFeature(feature)

	return unleash.UpdateFeature200JSONResponse(feature), nil
}

func (t TestServer) ArchiveFeature(_ context.Context, request unleash.ArchiveFeatureRequestObject) (unleash.ArchiveFeatureResponseObject, error) {
	feature, ok := t.getFeature(request.ProjectId, request.FeatureName)
	if !ok {
		return unleash.ArchiveFeature404JSONResponse{}, nil
	}
	feature.Archived = ptr.ToPtr(true)
	t.replaceFeature(feature)

	return unleash.ArchiveFeature202Response{}, nil
}

func (t TestServer) DeleteFeature(_ context.Context, request unleash.DeleteFeatureRequestObject) (unleash.DeleteFeatureResponseObject, error) {
	ok := t.deleteFeature("default", request.FeatureName)
	if !ok {
		return unleash.DeleteFeature403JSONResponse{}, nil
	}

	return unleash.DeleteFeature200Response{}, nil
}

func (t TestServer) ToggleFeatureEnvironmentOn(_ context.Context, request unleash.ToggleFeatureEnvironmentOnRequestObject) (unleash.ToggleFeatureEnvironmentOnResponseObject, error) {
	feature, ok := t.getFeature(request.ProjectId, request.FeatureName)
	if !ok {
		return unleash.ToggleFeatureEnvironmentOn404JSONResponse{}, nil
	}
	_, ok = updateEnvironment(feature, request.Environment, func(environment *unleash.FeatureEnvironmentSchema) any {
		environment.Enabled = true
		return nil
	})
	if !ok {
		return unleash.ToggleFeatureEnvironmentOn404JSONResponse{}, nil
	}
	t.replaceFeature(feature)

	return unleash.ToggleFeatureEnvironmentOn200JSONResponse{}, nil
}

func (t TestServer) ToggleFeatureEnvironmentOff(_ context.Context, request unleash.ToggleFeatureEnvironmentOffRequestObject) (unleash.ToggleFeatureEnvironmentOffResponseObject, error) {
	feature, ok := t.getFeature(request.ProjectId, request.FeatureName)
	if !ok {
		return unleash.ToggleFeatureEnvironmentOff404JSONResponse{}, nil
	}
	_, ok = updateEnvironment(feature, request.Environment, func(environment *unleash.FeatureEnvironmentSchema) any {
		environment.Enabled = false
		return nil
	})
	if !ok {
		return unleash.ToggleFeatureEnvironmentOff404JSONResponse{}, nil
	}
	t.replaceFeature(feature)

	return unleash.ToggleFeatureEnvironmentOff200JSONResponse{}, nil
}

func (t TestServer) AddFeatureStrategy(_ context.Context, request unleash.AddFeatureStrategyRequestObject) (unleash.AddFeatureStrategyResponseObject, error) {
	feature, ok := t.getFeature(request.ProjectId, request.FeatureName)
	if !ok {
		return unleash.AddFeatureStrategy404JSONResponse{}, nil
	}
	strategy, ok := updateEnvironment(feature, request.Environment, func(environment *unleash.FeatureEnvironmentSchema) unleash.FeatureStrategySchema {
		var strategies []unleash.FeatureStrategySchema
		if environment.Strategies != nil {
			strategies = *environment.Strategies
		}

		strategy := unleash.FeatureStrategySchema{
			Id:          ptr.ToPtr(t.getNext("strategy")),
			Name:        request.Body.Name,
			FeatureName: ptr.ToPtr(request.FeatureName),
			Constraints: request.Body.Constraints,
			Disabled:    request.Body.Disabled,
			Parameters:  request.Body.Parameters,
			Segments:    request.Body.Segments,
			SortOrder:   request.Body.SortOrder,
			Title:       request.Body.Title,
		}
		var variants []unleash.StrategyVariantSchema
		if request.Body.Variants != nil && len(*request.Body.Variants) > 0 {
			for _, reqVariant := range *request.Body.Variants {
				variant := unleash.StrategyVariantSchema{
					Name:       reqVariant.Name,
					Stickiness: reqVariant.Stickiness,
					Weight:     reqVariant.Weight,
					WeightType: unleash.StrategyVariantSchemaWeightType(reqVariant.WeightType),
				}
				if reqVariant.Payload != nil {
					variant.Payload = &struct {
						Type  unleash.StrategyVariantSchemaPayloadType `json:"type"`
						Value string                                   `json:"value"`
					}{
						Type:  unleash.StrategyVariantSchemaPayloadType(reqVariant.Payload.Type),
						Value: reqVariant.Payload.Value,
					}
				}

				variants = append(variants, variant)
			}
		}
		strategy.Variants = &variants

		strategies = append(strategies, strategy)
		environment.Strategies = &strategies

		return strategy
	})
	if !ok {
		return unleash.AddFeatureStrategy404JSONResponse{}, nil
	}
	t.replaceFeature(feature)

	return unleash.AddFeatureStrategy200JSONResponse(strategy), nil
}

func updateEnvironment[T any](feature unleash.FeatureSchema, environmentName string, updateFn func(environment *unleash.FeatureEnvironmentSchema) T) (T, bool) {
	var t T
	var environments []unleash.FeatureEnvironmentSchema
	if feature.Environments != nil {
		environments = *feature.Environments
	}
	for i, environment := range environments {
		if environment.Name == environmentName {
			t = updateFn(&environment)
			environments[i] = environment
			return t, true
		}
	}
	feature.Environments = &environments
	return t, false
}

func (t TestServer) GetFeatureStrategies(_ context.Context, request unleash.GetFeatureStrategiesRequestObject) (unleash.GetFeatureStrategiesResponseObject, error) {
	feature, ok := t.getFeature(request.ProjectId, request.FeatureName)
	if !ok {
		return unleash.GetFeatureStrategies404JSONResponse{}, nil
	}
	environment, ok := getEnvironment(request.Environment, *feature.Environments)
	if !ok {
		return unleash.GetFeatureStrategies404JSONResponse{}, nil
	}

	return unleash.GetFeatureStrategies200JSONResponse(ptr.ToValue(environment.Strategies, func() []unleash.FeatureStrategySchema {
		return []unleash.FeatureStrategySchema{}
	})), nil
}

func getEnvironment(environmentName string, environments []unleash.FeatureEnvironmentSchema) (unleash.FeatureEnvironmentSchema, bool) {
	for _, environment := range environments {
		if environment.Name == environmentName {
			return environment, true
		}
	}
	return unleash.FeatureEnvironmentSchema{}, false
}

func (t TestServer) UpdateFeatureStrategy(_ context.Context, request unleash.UpdateFeatureStrategyRequestObject) (unleash.UpdateFeatureStrategyResponseObject, error) {
	feature, ok := t.getFeature(request.ProjectId, request.FeatureName)
	if !ok {
		return unleash.UpdateFeatureStrategy404JSONResponse{}, nil
	}
	strategy, ok := updateEnvironment(feature, request.Environment, func(environment *unleash.FeatureEnvironmentSchema) unleash.FeatureStrategySchema {
		var strategies []unleash.FeatureStrategySchema
		if environment.Strategies != nil {
			strategies = *environment.Strategies
		}
		for i := range strategies {
			strategy := strategies[i]
			if *strategy.Id != request.StrategyId {
				continue
			}
			strategy.Name = *request.Body.Name
			strategy.Constraints = request.Body.Constraints
			strategy.Disabled = request.Body.Disabled
			strategy.Parameters = request.Body.Parameters
			strategy.Title = request.Body.Title
			var variants []unleash.StrategyVariantSchema
			if request.Body.Variants != nil && len(*request.Body.Variants) > 0 {
				for _, reqVariant := range *request.Body.Variants {
					variant := unleash.StrategyVariantSchema{
						Name:       reqVariant.Name,
						Stickiness: reqVariant.Stickiness,
						Weight:     reqVariant.Weight,
						WeightType: unleash.StrategyVariantSchemaWeightType(reqVariant.WeightType),
					}
					if reqVariant.Payload != nil {
						variant.Payload = &struct {
							Type  unleash.StrategyVariantSchemaPayloadType `json:"type"`
							Value string                                   `json:"value"`
						}{
							Type:  unleash.StrategyVariantSchemaPayloadType(reqVariant.Payload.Type),
							Value: reqVariant.Payload.Value,
						}
					}

					variants = append(variants, variant)
				}
			}
			strategy.Variants = &variants
			strategies[i] = strategy
			environment.Strategies = &strategies

			return strategy
		}

		return unleash.FeatureStrategySchema{}
	})
	if !ok {
		return unleash.UpdateFeatureStrategy404JSONResponse{}, nil
	}
	t.replaceFeature(feature)

	return unleash.UpdateFeatureStrategy200JSONResponse(strategy), nil
}

func (t TestServer) DeleteFeatureStrategy(_ context.Context, request unleash.DeleteFeatureStrategyRequestObject) (unleash.DeleteFeatureStrategyResponseObject, error) {
	feature, ok := t.getFeature(request.ProjectId, request.FeatureName)
	if !ok {
		return unleash.DeleteFeatureStrategy404JSONResponse{}, nil
	}
	_, ok = updateEnvironment(feature, request.Environment, func(environment *unleash.FeatureEnvironmentSchema) any {
		var strategies []unleash.FeatureStrategySchema
		if environment.Strategies != nil {
			strategies = *environment.Strategies
		}
		for i := range strategies {
			strategy := strategies[i]
			if *strategy.Id != request.StrategyId {
				continue
			}
			strategies = append(strategies[:i], strategies[i+1:]...)
			environment.Strategies = &strategies
			return nil
		}

		return nil
	})
	if !ok {
		return unleash.DeleteFeatureStrategy404JSONResponse{}, nil
	}
	t.replaceFeature(feature)

	return unleash.DeleteFeatureStrategy200Response{}, nil
}

func (t TestServer) SetStrategySortOrder(_ context.Context, request unleash.SetStrategySortOrderRequestObject) (unleash.SetStrategySortOrderResponseObject, error) {
	feature, ok := t.getFeature(request.ProjectId, request.FeatureName)
	if !ok {
		return unleash.SetStrategySortOrder400JSONResponse{}, nil
	}
	_, ok = updateEnvironment(feature, request.Environment, func(environment *unleash.FeatureEnvironmentSchema) unleash.FeatureStrategySchema {
		var strategies []unleash.FeatureStrategySchema
		if environment.Strategies != nil {
			strategies = *environment.Strategies
		}
		for i := range strategies {
			strategy := strategies[i]
			for _, sortOrderWithID := range *request.Body {
				if *strategy.Id == sortOrderWithID.Id {
					strategy.SortOrder = ptr.ToPtr(sortOrderWithID.SortOrder)
					break
				}
			}
		}
		environment.Strategies = &strategies

		return unleash.FeatureStrategySchema{}
	})
	if !ok {
		return unleash.SetStrategySortOrder400JSONResponse{}, nil
	}
	t.replaceFeature(feature)

	return unleash.SetStrategySortOrder200Response{}, nil
}

func (t TestServer) OverwriteFeatureVariantsOnEnvironments(_ context.Context, request unleash.OverwriteFeatureVariantsOnEnvironmentsRequestObject) (unleash.OverwriteFeatureVariantsOnEnvironmentsResponseObject, error) {
	feature, ok := t.getFeature(request.ProjectId, request.FeatureName)
	if !ok {
		return unleash.OverwriteFeatureVariantsOnEnvironments400JSONResponse{}, nil
	}
	for _, requestEnvironment := range *request.Body.Environments {
		if request.Body.Variants != nil {
			for _, v := range *request.Body.Variants {
				if v.Overrides == nil {
					continue
				}
				for _, o := range *v.Overrides {
					// emulate too large body error
					if len(o.Values) > 100 {
						return OverwriteFeatureVariantsOnEnvironments413JSONResponse{}, nil
					}
				}
			}
		}

		_, ok := updateEnvironment(feature, requestEnvironment, func(environment *unleash.FeatureEnvironmentSchema) unleash.FeatureStrategySchema {
			environment.Variants = request.Body.Variants

			return unleash.FeatureStrategySchema{}
		})
		if !ok {
			return unleash.OverwriteFeatureVariantsOnEnvironments400JSONResponse{}, nil
		}
	}
	t.replaceFeature(feature)

	return unleash.OverwriteFeatureVariantsOnEnvironments200JSONResponse{}, nil
}

func (t TestServer) GetEnvironmentFeatureVariants(_ context.Context, request unleash.GetEnvironmentFeatureVariantsRequestObject) (unleash.GetEnvironmentFeatureVariantsResponseObject, error) {
	feature, ok := t.getFeature(request.ProjectId, request.FeatureName)
	if !ok {
		return unleash.GetEnvironmentFeatureVariants404JSONResponse{}, nil
	}
	environment, ok := getEnvironment(request.Environment, *feature.Environments)
	if !ok {
		return unleash.GetEnvironmentFeatureVariants404JSONResponse{}, nil
	}
	featureVariants := unleash.FeatureVariantsSchema{}
	if environment.Variants != nil {
		featureVariants.Variants = *environment.Variants
	}

	return unleash.GetEnvironmentFeatureVariants200JSONResponse(featureVariants), nil
}

func (t TestServer) UpdateFeatureStrategySegments(_ context.Context, request unleash.UpdateFeatureStrategySegmentsRequestObject) (unleash.UpdateFeatureStrategySegmentsResponseObject, error) {
	reqBody := *request.Body

	projectFeatures := t.getProjectFeatures(reqBody.ProjectId)
	for _, feature := range projectFeatures {
		updateEnvironment(feature, reqBody.EnvironmentId, func(environment *unleash.FeatureEnvironmentSchema) any {
			if environment.Strategies == nil {
				return ""
			}
			strategies := *environment.Strategies
			for i := range strategies {
				strategy := strategies[i]
				if *strategy.Id != reqBody.StrategyId {
					continue
				}
				segments := make([]float32, len(reqBody.SegmentIds))
				for i, segmentID := range reqBody.SegmentIds {
					segments[i] = float32(segmentID)
				}
				strategy.Segments = &segments
				strategies[i] = strategy
			}
			environment.Strategies = &strategies

			return nil
		})
		t.replaceFeature(feature)
	}

	return unleash.UpdateFeatureStrategySegments201JSONResponse{}, nil
}

func (t TestServer) PatchEnvironmentsFeatureVariants(_ context.Context, request unleash.PatchEnvironmentsFeatureVariantsRequestObject) (unleash.PatchEnvironmentsFeatureVariantsResponseObject, error) {
	for _, patch := range *request.Body {
		switch patch.Op {
		case "remove":
			{
				feature, ok := t.getFeature(request.ProjectId, request.FeatureName)
				if !ok {
					return unleash.PatchEnvironmentsFeatureVariants404JSONResponse{}, nil
				}
				_, ok = updateEnvironment(feature, request.Environment, func(environment *unleash.FeatureEnvironmentSchema) unleash.FeatureStrategySchema {
					variants := *environment.Variants
					index, _ := strconv.Atoi(patch.Path[1:])
					variants = append(variants[:index], variants[index+1:]...)
					environment.Variants = &variants

					return unleash.FeatureStrategySchema{}
				})
				if !ok {
					return unleash.PatchEnvironmentsFeatureVariants404JSONResponse{}, nil
				}
				t.replaceFeature(feature)
				break
			}
		case "replace":
			{
				feature, ok := t.getFeature(request.ProjectId, request.FeatureName)
				if !ok {
					return unleash.PatchEnvironmentsFeatureVariants404JSONResponse{}, nil
				}
				_, ok = updateEnvironment(feature, request.Environment, func(environment *unleash.FeatureEnvironmentSchema) unleash.FeatureStrategySchema {
					variants := *environment.Variants
					index, _ := strconv.Atoi(patch.Path[1:])

					v := unleash.VariantSchema{}
					m, _ := (*patch.Value).(map[string]interface{})
					b, _ := json.Marshal(m)
					_ = json.Unmarshal(b, &v)

					variants[index] = v
					environment.Variants = &variants

					return unleash.FeatureStrategySchema{}
				})
				if !ok {
					return unleash.PatchEnvironmentsFeatureVariants404JSONResponse{}, nil
				}
				t.replaceFeature(feature)
				break
			}
		default:
			return unleash.PatchEnvironmentsFeatureVariants400JSONResponse{
				Message: ptr.ToPtr("operation " + string(patch.Op) + " is not supported"),
			}, nil
		}
	}

	return unleash.PatchEnvironmentsFeatureVariants200JSONResponse{}, nil
}

type OverwriteFeatureVariantsOnEnvironments413JSONResponse struct {
	// Id The ID of the error instance
	Id *string `json:"id,omitempty"`

	// Message A description of what went wrong.
	Message *string `json:"message,omitempty"`

	// Name The name of the error kind
	Name *string `json:"name,omitempty"`
}

func (response OverwriteFeatureVariantsOnEnvironments413JSONResponse) VisitOverwriteFeatureVariantsOnEnvironmentsResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(413)

	return json.NewEncoder(w).Encode(response)
}
