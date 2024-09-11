package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/ptr"
	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/unleash"
)

var _ resource.Resource = &FeatureResource{}
var _ resource.ResourceWithImportState = &FeatureResource{}

func NewFeatureResource() resource.Resource {
	return &FeatureResource{}
}

type FeatureResource struct {
	providerData UnleashProviderData
}

type FeatureResourceModel struct {
	FeatureModel
}

func resolveID(FeatureResourceModel FeatureResourceModel) string {
	return FeatureResourceModel.Project.ValueString() + "." + FeatureResourceModel.Name.ValueString()
}

func (r *FeatureResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_feature"
}

func (r *FeatureResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Project resource",

		Attributes: createFeatureResourceSchemaAttr(),
	}
}

func (r *FeatureResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(UnleashProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.providerData = providerData
}

func (r *FeatureResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FeatureResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(resolveID(data))

	body := unleash.CreateFeatureJSONRequestBody{
		Name: data.Name.ValueString(),
		Type: data.Type.ValueStringPointer(),
	}
	if !data.Description.IsNull() {
		body.Description = data.Description.ValueStringPointer()
	}

	tflog.Debug(ctx, "Creating feature", map[string]interface{}{"body": body})
	createResp, err := r.providerData.Client.CreateFeatureWithResponse(ctx, data.Project.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("failed to create feature "+data.ID.String(), err.Error())
		return
	}
	if createResp.StatusCode() > 299 {
		resp.Diagnostics.AddError("failed to create feature "+data.ID.String(), fmt.Sprintf(" with status %d %s", createResp.StatusCode(), string(createResp.Body)))
		return
	}

	err = r.updateEnvironments(ctx, data.Project.ValueString(), data.Name.ValueString(), data.Environments, map[string]EnvironmentModel{})
	if err != nil {
		resp.Diagnostics.AddError("failed to create environments", err.Error())
		return
	}

	tflog.Trace(ctx, "created a resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FeatureResource) updateEnvironments(ctx context.Context, projectID string, featureName string, environments map[string]EnvironmentModel, existingEnvironmentByName map[string]EnvironmentModel) error {
	for name, env := range environments {
		existingEnv, ok := existingEnvironmentByName[name]
		if !ok {
			var err error
			existingEnv, err = toEnvironmentModel(unleash.FetchedEnvironment{})
			if err != nil {
				return err
			}
		}
		updatedEnv, err := r.updateEnvironment(ctx, projectID, featureName, name, env, existingEnv)
		if err != nil {
			return err
		}
		environments[name] = updatedEnv
	}
	return nil
}

func (r *FeatureResource) updateEnvironment(ctx context.Context, projectID string, featureName string, environmentID string, environment EnvironmentModel, existingEnv EnvironmentModel) (EnvironmentModel, error) {
	environment, err := r.updateStrategies(ctx, projectID, featureName, environmentID, environment, existingEnv)
	if err != nil {
		return environment, err
	}

	environment, err = r.updateEnvironmentVariants(ctx, projectID, featureName, environmentID, environment, existingEnv)
	if err != nil {
		return environment, err
	}

	return r.updateEnvironmentStatus(ctx, projectID, featureName, environmentID, environment, existingEnv)
}

func (r *FeatureResource) updateStrategies(ctx context.Context, projectID string, featureName string, environmentID string, environment EnvironmentModel, existingEnv EnvironmentModel) (EnvironmentModel, error) {
	existingStrategyByKey := toStrategyModelByIDName(existingEnv.Strategies)
	newStrategyByKey := toStrategyModelByIDName(environment.Strategies)

	for key, strategy := range existingStrategyByKey {
		_, ok := newStrategyByKey[key]
		if ok {
			continue
		}
		err := r.deleteStrategy(ctx, projectID, featureName, environmentID, strategy)
		if err != nil {
			return environment, err
		}
	}
	for i, strategy := range environment.Strategies {
		if r.shouldIgnoreStrategy(strategy.Title.ValueString()) {
			return environment, fmt.Errorf("strategy title %s matches ignore regex. This strategy should not be managed by terraform", strategy.Title.ValueString())
		}
		key := toStrategyModelKey(strategy)
		existingStrategy, ok := existingStrategyByKey[key]
		if ok {
			err := r.updateStrategy(ctx, projectID, featureName, environmentID, strategy, existingStrategy)
			if err != nil {
				return environment, err
			}
		} else {
			id, err := r.addStrategy(ctx, projectID, featureName, environmentID, strategy)
			if err != nil {
				return environment, err
			}
			strategy.Id = types.StringValue(id)
			environment.Strategies[i] = strategy
		}
	}

	return environment, nil
}

func (r *FeatureResource) addStrategy(ctx context.Context, projectID string, featureName string, environmentID string, strategy StrategyModel) (string, error) {
	strategyBody := unleash.AddFeatureStrategyJSONRequestBody{
		Name:      strategy.Name.ValueString(),
		Title:     strategy.Title.ValueStringPointer(),
		SortOrder: strategy.SortOrder.ValueFloat32Pointer(),
		Disabled:  strategy.Disabled.ValueBoolPointer(),
	}
	if len(strategy.Parameters) > 0 {
		parameters := make(unleash.ParametersSchema)
		for parameterK, parameterV := range strategy.Parameters {
			parameters[parameterK] = parameterV.ValueString()
		}
		strategyBody.Parameters = &parameters
	}
	if len(strategy.Constraints) > 0 {
		constraints := make([]unleash.ConstraintSchema, 0, len(strategy.Constraints))
		for _, constraint := range strategy.Constraints {
			constraintBody := unleash.ConstraintSchema{
				CaseInsensitive: constraint.CaseInsensitive.ValueBoolPointer(),
				ContextName:     constraint.ContextName.ValueString(),
				Inverted:        constraint.Inverted.ValueBoolPointer(),
				Operator:        unleash.ConstraintSchemaOperator(constraint.Operator.ValueString()),
			}
			if !constraint.Value.IsNull() {
				constraintBody.Value = ptr.ToPtr(constraint.Value.ValueString())
			}
			if !constraint.JsonValues.IsNull() {
				values, err := toStringValues(constraint.JsonValues.ValueString())
				if err != nil {
					return "", err
				}
				constraintBody.Values = &values
			}

			constraints = append(constraints, constraintBody)
		}
		strategyBody.Constraints = &constraints
	}
	if len(strategy.Segments) > 0 {
		segments := make([]float32, 0, len(strategy.Segments))
		for _, segment := range strategy.Segments {
			segments = append(segments, segment.ValueFloat32())
		}
		strategyBody.Segments = &segments
	}
	if len(strategy.Variants) > 0 {
		variants := make([]unleash.CreateStrategyVariantSchema, 0, len(strategy.Variants))
		for _, variant := range strategy.Variants {
			variantBody := unleash.CreateStrategyVariantSchema{
				Name:       variant.Name.ValueString(),
				Stickiness: variant.Stickiness.ValueString(),
				WeightType: unleash.CreateStrategyVariantSchemaWeightType(variant.WeightType.ValueString()),
			}
			if variantBody.WeightType == unleash.CreateStrategyVariantSchemaWeightTypeFix {
				variantBody.Weight = int(variant.Weight.ValueInt64())
			}
			if !variant.Payload.IsNull() && !variant.PayloadType.IsNull() {
				variantBody.Payload = &struct {
					Type  unleash.CreateStrategyVariantSchemaPayloadType `json:"type"`
					Value string                                         `json:"value"`
				}{
					Type:  unleash.CreateStrategyVariantSchemaPayloadType(variant.PayloadType.ValueString()),
					Value: variant.Payload.ValueString(),
				}
			}

			variants = append(variants, variantBody)
		}
		strategyBody.Variants = &variants
	}
	resp, err := r.providerData.Client.AddFeatureStrategyWithResponse(ctx, projectID, featureName, environmentID, strategyBody)
	if err != nil {
		return "", err
	}
	if resp.StatusCode() > 299 {
		return "", fmt.Errorf("failed to add strategy for %s %s %s %s with status %d %s", projectID, featureName, environmentID, strategy.Name.ValueString(), resp.StatusCode(), string(resp.Body))
	}

	return *resp.JSON200.Id, nil
}

func (r *FeatureResource) updateStrategy(ctx context.Context, projectID string, featureName string, environmentID string, strategy StrategyModel, existingStrategy StrategyModel) error {
	body, err := toUpdateStrategyBody(strategy)
	if err != nil {
		return err
	}
	existingBody, err := toUpdateStrategyBody(existingStrategy)
	if err != nil {
		return err
	}
	if !cmp.Equal(body, existingBody) {
		tflog.Debug(ctx, "Updating strategy", map[string]interface{}{
			"projectID":     projectID,
			"featureName":   featureName,
			"environmentID": environmentID,
			"body":          body})
		resp, err := r.providerData.Client.UpdateFeatureStrategyWithResponse(ctx, projectID, featureName, environmentID, existingStrategy.Id.ValueString(), body)
		if err != nil {
			return err
		}
		if resp.StatusCode() > 299 {
			return fmt.Errorf("failed to update strategy for %s %s %s %s with status %d %s", projectID, featureName, environmentID, strategy.Name.ValueString(), resp.StatusCode(), string(resp.Body))
		}
	}
	if !strategy.SortOrder.Equal(existingStrategy.SortOrder) {
		var order float32 = 0
		if !strategy.SortOrder.IsNull() {
			order = strategy.SortOrder.ValueFloat32()
		}
		tflog.Debug(ctx, "Setting strategy sort order", map[string]interface{}{
			"projectID":     projectID,
			"featureName":   featureName,
			"environmentID": environmentID,
			"strategy":      strategy,
			"order":         order})
		resp, err := r.providerData.Client.SetStrategySortOrderWithResponse(ctx, projectID, featureName, environmentID, []struct {
			Id        string  `json:"id"`
			SortOrder float32 `json:"sortOrder"`
		}{{
			Id:        strategy.Id.ValueString(),
			SortOrder: order,
		}})
		if err != nil {
			return err
		}
		if resp.StatusCode() > 299 {
			return fmt.Errorf("failed to set strategy sort order for %s %s %s %s with status %d %s", projectID, featureName, environmentID, strategy.Name.ValueString(), resp.StatusCode(), string(resp.Body))
		}
	}
	if !cmp.Equal(strategy.Segments, existingStrategy.Segments) {
		updateStrategySegmentBody := toUpdateStrategySegmentsBody(projectID, environmentID, strategy.Id.ValueString(), strategy.Segments)
		tflog.Debug(ctx, "Updating strategy segments", map[string]interface{}{
			"projectID":     projectID,
			"featureName":   featureName,
			"environmentID": environmentID,
			"strategy":      strategy.Name.ValueString(),
			"body":          updateStrategySegmentBody})
		resp, err := r.providerData.Client.UpdateFeatureStrategySegmentsWithResponse(ctx, updateStrategySegmentBody)
		if err != nil {
			return err
		}
		if resp.StatusCode() > 299 {
			return fmt.Errorf("failed to update strategy segments for %s %s %s %s with status %d %s", projectID, featureName, environmentID, strategy.Name.ValueString(), resp.StatusCode(), string(resp.Body))
		}
	}

	return nil
}

func toUpdateStrategyBody(strategy StrategyModel) (unleash.UpdateFeatureStrategyJSONRequestBody, error) {
	body := unleash.UpdateFeatureStrategyJSONRequestBody{
		Name:     strategy.Name.ValueStringPointer(),
		Title:    strategy.Title.ValueStringPointer(),
		Disabled: strategy.Disabled.ValueBoolPointer(),
	}
	if body.Title == nil {
		body.Title = ptr.ToPtr("")
	}
	if body.Disabled == nil {
		body.Disabled = ptr.ToPtr(false)
	}
	parameters := make(unleash.ParametersSchema)
	for parameterK, parameterV := range strategy.Parameters {
		parameters[parameterK] = parameterV.ValueString()
	}
	body.Parameters = &parameters
	constraints := make([]unleash.ConstraintSchema, len(strategy.Constraints))
	for i, constraint := range strategy.Constraints {
		constraintBody := unleash.ConstraintSchema{
			CaseInsensitive: constraint.CaseInsensitive.ValueBoolPointer(),
			ContextName:     constraint.ContextName.ValueString(),
			Inverted:        constraint.Inverted.ValueBoolPointer(),
			Operator:        unleash.ConstraintSchemaOperator(constraint.Operator.ValueString()),
		}
		if !constraint.Value.IsNull() {
			constraintBody.Value = ptr.ToPtr(constraint.Value.ValueString())
		}
		if !constraint.JsonValues.IsNull() {
			values, err := toStringValues(constraint.JsonValues.ValueString())
			if err != nil {
				return body, err
			}
			if len(values) > 0 {
				constraintBody.Values = &values
			}
		}

		constraints[i] = constraintBody
	}
	body.Constraints = &constraints
	variants := make([]unleash.CreateStrategyVariantSchema, len(strategy.Variants))
	for i, variant := range strategy.Variants {
		variantBody := unleash.CreateStrategyVariantSchema{
			Name:       variant.Name.ValueString(),
			Stickiness: variant.Stickiness.ValueString(),
			WeightType: unleash.CreateStrategyVariantSchemaWeightType(variant.WeightType.ValueString()),
		}
		if variantBody.WeightType == unleash.CreateStrategyVariantSchemaWeightTypeFix {
			variantBody.Weight = int(variant.Weight.ValueInt64())
		}
		if !variant.PayloadType.IsNull() && !variant.Payload.IsNull() {
			variantBody.Payload = &struct {
				Type  unleash.CreateStrategyVariantSchemaPayloadType `json:"type"`
				Value string                                         `json:"value"`
			}{
				Type:  unleash.CreateStrategyVariantSchemaPayloadType(variant.PayloadType.ValueString()),
				Value: variant.Payload.ValueString(),
			}
		}

		variants[i] = variantBody
	}
	body.Variants = &variants

	return body, nil
}

func toUpdateStrategySegmentsBody(projectID string, environmentID string, strategyID string, segments []types.Float32) unleash.UpdateFeatureStrategySegmentsJSONRequestBody {
	body := unleash.UpdateFeatureStrategySegmentsJSONRequestBody{
		ProjectId:     projectID,
		StrategyId:    strategyID,
		EnvironmentId: environmentID,
	}
	for _, segment := range segments {
		body.SegmentIds = append(body.SegmentIds, int(segment.ValueFloat32()))
	}

	return body
}

func (r *FeatureResource) deleteStrategy(ctx context.Context, projectID string, featureName string, environmentID string, strategy StrategyModel) error {
	tflog.Debug(ctx, "Deleting strategy", map[string]interface{}{
		"projectID":     projectID,
		"featureName":   featureName,
		"environmentID": environmentID,
		"strategy":      strategy,
	})
	resp, err := r.providerData.Client.DeleteFeatureStrategyWithResponse(ctx, projectID, featureName, environmentID, strategy.Id.ValueString())
	if err != nil {
		return err
	}
	if resp.StatusCode() > 299 {
		return fmt.Errorf("failed to delete strategy for %s %s %s %s with status %d %s", projectID, featureName, environmentID, strategy.Name.ValueString(), resp.StatusCode(), string(resp.Body))
	}
	return nil
}

func toStrategyModelByIDName(strategies []StrategyModel) map[string]StrategyModel {
	strategyByName := make(map[string]StrategyModel)
	for _, strategy := range strategies {
		strategyByName[toStrategyModelKey(strategy)] = strategy
	}
	return strategyByName
}

func toStrategyModelKey(strategy StrategyModel) string {
	key := strategy.Name.ValueString()
	if !strategy.Id.IsNull() {
		key = strategy.Id.ValueString()
	}

	return key
}

func (r *FeatureResource) updateEnvironmentVariants(ctx context.Context, projectID string, featureName string, environmentID string, environment EnvironmentModel, existingEnv EnvironmentModel) (EnvironmentModel, error) {
	diff := getVariantDiffMode(environment.Variants, existingEnv.Variants)
	if diff.Mode == variantDiffModeEqual {
		return environment, nil
	}
	variantsBody := unleash.OverwriteFeatureVariantsOnEnvironmentsJSONRequestBody{
		Environments: &[]string{environmentID},
	}
	variants := make([]unleash.VariantSchema, 0, len(environment.Variants))
	for _, variant := range environment.Variants {
		variantBody, err := toVariantBody(variant)
		if err != nil {
			return environment, err
		}
		variants = append(variants, variantBody)
	}
	variantsBody.Variants = &variants

	tflog.Debug(ctx, "Overwriting variants", map[string]interface{}{
		"projectID":     projectID,
		"featureName":   featureName,
		"environmentID": environmentID,
		"body":          variantsBody,
	})
	// try to do everything in one go first. this however may face a problem when the body is too large...
	resp, err := r.providerData.Client.OverwriteFeatureVariantsOnEnvironmentsWithResponse(ctx, projectID, featureName, variantsBody)
	if err != nil {
		return environment, err
	}
	if resp.StatusCode() == 413 {
		tflog.Warn(ctx, "Cannot overwriting variants since the body is too large, try patching instead", map[string]interface{}{
			"projectID":     projectID,
			"featureName":   featureName,
			"environmentID": environmentID,
		})
		err := r.handleTooLargeEnvironmentVariants(ctx, projectID, featureName, environmentID, variants, diff)
		return environment, err
	}
	if resp.StatusCode() > 299 {
		return environment, fmt.Errorf("failed to overwrite variants for %s %s %s with status %d %s", projectID, featureName, environmentID, resp.StatusCode(), string(resp.Body))
	}

	return environment, nil
}

func toVariantBody(variant VariantModel) (unleash.VariantSchema, error) {
	variantBody := unleash.VariantSchema{
		Name:       variant.Name.ValueString(),
		Stickiness: variant.Stickiness.ValueStringPointer(),
	}
	if !variant.PayloadType.IsNull() && !variant.Payload.IsNull() {
		variantBody.Payload = &struct {
			Type  unleash.VariantSchemaPayloadType `json:"type"`
			Value string                           `json:"value"`
		}{
			Type:  unleash.VariantSchemaPayloadType(variant.PayloadType.ValueString()),
			Value: variant.Payload.ValueString(),
		}
	}
	if !variant.WeightType.IsNull() {
		t := unleash.VariantSchemaWeightType(variant.WeightType.ValueString())
		variantBody.WeightType = &t
		if *variantBody.WeightType == unleash.Fix {
			variantBody.Weight = variant.Weight.ValueFloat32()
		}
	}
	overrides := make([]unleash.OverrideSchema, len(variant.Overrides))
	for i, override := range variant.Overrides {
		overrideBody := unleash.OverrideSchema{
			ContextName: override.ContextName.ValueString(),
		}
		if !override.JsonValues.IsNull() && len(override.JsonValues.ValueString()) > 0 {
			values, err := toStringValues(override.JsonValues.ValueString())
			if err != nil {
				return variantBody, err
			}
			if len(values) > 0 {
				overrideBody.Values = values
			}
		}
		overrides[i] = overrideBody
	}
	variantBody.Overrides = &overrides
	return variantBody, nil
}

type variantDiffMode int

const variantDiffModeEqual variantDiffMode = 0
const variantDiffModeAddOnly variantDiffMode = 1
const variantDiffModeReplaceOnly variantDiffMode = 2
const variantDiffModeRemoveOnly variantDiffMode = 3
const variantDiffModeMixed variantDiffMode = 4

func (r *FeatureResource) handleTooLargeEnvironmentVariants(ctx context.Context, projectID string, featureName string, environmentID string, largeVariants []unleash.VariantSchema, diff variantsDiff) error {
	patches := make([]unleash.PatchSchema, 0, len(largeVariants))
	switch diff.Mode {
	case variantDiffModeMixed, variantDiffModeAddOnly:
		{
			if diff.Mode == variantDiffModeMixed {
				tflog.Warn(ctx, "Too many variant changes in one commit. There may be some unwanted behavior during patching changes...", map[string]interface{}{
					"projectID":     projectID,
					"featureName":   featureName,
					"environmentID": environmentID,
				})
			}
			smallVariants := make([]unleash.VariantSchema, len(largeVariants))

			// It is very hard to do multiple add patching requests since there should be at least 1 variant with variable type.
			// We then overwrite the whole variants and patch the large ones later. This surely causes bad payload or override values during update.
			for i, variant := range largeVariants {
				tooLarge := false
				smallVariant := variant
				if smallVariant.Payload != nil && len(smallVariant.Payload.Value) > 100000 {
					tooLarge = true
					smallVariant.Payload = nil
				}
				if smallVariant.Overrides != nil {
					for _, override := range *smallVariant.Overrides {
						if len(override.Values) > 100 {
							tooLarge = true
							smallVariant.Overrides = nil
							break
						}
					}
				}
				if tooLarge {
					patchPath := fmt.Sprintf("/%d", i)
					patches = append(patches, unleash.PatchSchema{
						Op:    "replace",
						Path:  patchPath,
						From:  &patchPath,
						Value: toPatchValue(variant),
					})
				}

				smallVariants[i] = smallVariant
			}

			resp, err := r.providerData.Client.OverwriteFeatureVariantsOnEnvironmentsWithResponse(ctx, projectID, featureName, unleash.OverwriteFeatureVariantsOnEnvironmentsJSONRequestBody{
				Environments: &[]string{environmentID},
				Variants:     &smallVariants,
			})
			if err != nil {
				return err
			}
			if resp.StatusCode() > 299 {
				return fmt.Errorf("failed to overwrite small variants for %s %s %s with status %d %s", projectID, featureName, environmentID, resp.StatusCode(), string(resp.Body))
			}
			break
		}
	case variantDiffModeRemoveOnly:
		{
			for _, variant := range diff.ToRemove {
				patchPath := fmt.Sprintf("/%d", variant.Index)
				patches = append(patches, unleash.PatchSchema{
					Op:   "remove",
					Path: patchPath,
					From: &patchPath,
				})
			}
			break
		}
	case variantDiffModeReplaceOnly:
		{
			for _, variant := range diff.ToReplace {
				// We can improve this for more efficient patching by checking which properties are changed and specify the path to patch.
				// For example, the change may just want to add a new overriding item to large number of existing items.
				patchPath := fmt.Sprintf("/%d", variant.Index)
				variantBody, err := toVariantBody(variant.Variant)
				if err != nil {
					return err
				}
				patches = append(patches, unleash.PatchSchema{
					Op:    "replace",
					Path:  patchPath,
					From:  &patchPath,
					Value: toPatchValue(variantBody),
				})
			}
			break
		}
	}

	for _, patch := range patches {
		tflog.Debug(ctx, "Patching variant", map[string]interface{}{
			"projectID":     projectID,
			"featureName":   featureName,
			"environmentID": environmentID,
			"body":          patch,
		})
		resp, err := r.providerData.Client.PatchEnvironmentsFeatureVariantsWithResponse(ctx, projectID, featureName, environmentID, unleash.PatchEnvironmentsFeatureVariantsJSONRequestBody{patch})
		if err != nil {
			return err
		}
		if resp.StatusCode() > 299 {
			return fmt.Errorf("failed to patch variant for %s %s %s with status %d %s", projectID, featureName, environmentID, resp.StatusCode(), string(resp.Body))
		}
	}

	return nil
}

func toPatchValue(body unleash.VariantSchema) *interface{} {
	var value interface{} = body

	return &value
}

func (r *FeatureResource) updateEnvironmentStatus(ctx context.Context, projectID string, featureName string, environmentID string, environment EnvironmentModel, existingEnv EnvironmentModel) (EnvironmentModel, error) {
	if environment.Enabled == existingEnv.Enabled {
		return environment, nil
	}

	if environment.Enabled.ValueBool() {
		tflog.Debug(ctx, "Enabling environment", map[string]interface{}{
			"projectID":     projectID,
			"featureName":   featureName,
			"environmentID": environmentID,
		})
		resp, err := r.providerData.Client.ToggleFeatureEnvironmentOnWithResponse(ctx, projectID, featureName, environmentID)
		if err != nil {
			return environment, err
		}
		if resp.StatusCode() > 299 {
			return environment, fmt.Errorf("failed to enable environment for %s %s %s with status %d %s", projectID, featureName, environmentID, resp.StatusCode(), string(resp.Body))
		}
	} else {
		tflog.Debug(ctx, "Disabling environment", map[string]interface{}{
			"projectID":     projectID,
			"featureName":   featureName,
			"environmentID": environmentID,
		})
		resp, err := r.providerData.Client.ToggleFeatureEnvironmentOffWithResponse(ctx, projectID, featureName, environmentID)
		if err != nil {
			return environment, err
		}
		if resp.StatusCode() > 299 {
			return environment, fmt.Errorf("failed to disable environment for %s %s %s with status %d %s", projectID, featureName, environmentID, resp.StatusCode(), string(resp.Body))
		}
	}

	return environment, nil
}

func (r *FeatureResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FeatureResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID, featureName := extractProjectAndFeatureName(data)

	fetchedFeature, found, err := unleash.GetFeature(ctx, r.providerData.Client, projectID, featureName)
	if err != nil {
		resp.Diagnostics.AddError("failed to get feature", err.Error())
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}
	removeIgnoredStrategies(ctx, &fetchedFeature, r.providerData.StrategyTitleIgnoreRegEx)

	featureModel, err := toFeatureModel(fetchedFeature)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert feature", err.Error())
		return
	}
	ensureFeatureModelNullAndEmptyConsistency(&featureModel, data.FeatureModel)
	data.FeatureModel = featureModel

	tflog.Trace(ctx, "read resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func extractProjectAndFeatureName(data FeatureResourceModel) (string, string) {
	id := data.ID.ValueString()
	firstDot := strings.Index(id, ".")

	return id[0:firstDot], id[firstDot+1:]
}

func (r *FeatureResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FeatureResourceModel
	var existingData FeatureResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &existingData)...)

	if resp.Diagnostics.HasError() {
		return
	}

	featureBody := toFeatureBody(data)
	existingFeatureBody := toFeatureBody(existingData)
	if !cmp.Equal(featureBody, existingFeatureBody) {
		tflog.Debug(ctx, "Updating feature", map[string]interface{}{
			"projectID":   data.Project.ValueString(),
			"featureName": data.Name.ValueString(),
			"body":        featureBody,
		})
		updateResp, err := r.providerData.Client.UpdateFeatureWithResponse(ctx, data.Project.ValueString(), data.Name.ValueString(), featureBody)
		if err != nil {
			resp.Diagnostics.AddError("failed to update feature "+data.ID.String(), err.Error())
			return
		}
		if updateResp.StatusCode() > 299 {
			resp.Diagnostics.AddError("failed to update feature "+data.ID.String(), fmt.Sprintf(" with status %d %s", updateResp.StatusCode(), string(updateResp.Body)))
			return
		}
	}

	err := r.updateEnvironments(ctx, data.Project.ValueString(), data.Name.ValueString(), data.Environments, existingData.Environments)
	if err != nil {
		resp.Diagnostics.AddError("failed to update environment", err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func toFeatureBody(data FeatureResourceModel) unleash.UpdateFeatureJSONRequestBody {
	body := unleash.UpdateFeatureJSONRequestBody{
		Type:           data.Type.ValueStringPointer(),
		Description:    data.Description.ValueStringPointer(),
		ImpressionData: data.ImpressionData.ValueBoolPointer(),
	}
	if body.Description == nil {
		body.Description = ptr.ToPtr("")
	}
	if body.ImpressionData == nil {
		body.ImpressionData = ptr.ToPtr(false)
	}

	return body
}

func (r *FeatureResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FeatureResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(resolveID(data))

	tflog.Debug(ctx, "Archiving feature", map[string]interface{}{"projectID": data.Project.ValueString(), "featureName": data.Name.ValueString()})
	archiveResp, err := r.providerData.Client.ArchiveFeatureWithResponse(ctx, data.Project.ValueString(), data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to archive feature "+data.ID.String(), err.Error())
		return
	}
	if archiveResp.StatusCode() > 299 && archiveResp.StatusCode() != 404 {
		resp.Diagnostics.AddError("failed to archive feature "+data.ID.String(), fmt.Sprintf(" with status %d %s", archiveResp.StatusCode(), string(archiveResp.Body)))
		return
	}
	tflog.Debug(ctx, "Deleting feature", map[string]interface{}{"projectID": data.Project.ValueString(), "featureName": data.Name.ValueString()})
	deleteResp, err := r.providerData.Client.DeleteFeatureWithResponse(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete feature "+data.ID.String(), err.Error())
		return
	}
	if deleteResp.StatusCode() > 299 && deleteResp.StatusCode() != 404 {
		resp.Diagnostics.AddError("failed to delete feature "+data.ID.String(), fmt.Sprintf(" with status %d %s", deleteResp.StatusCode(), string(deleteResp.Body)))
		return
	}
}

func (r *FeatureResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *FeatureResource) shouldIgnoreStrategy(title string) bool {
	if r.providerData.StrategyTitleIgnoreRegEx == nil {
		return false
	}
	return r.providerData.StrategyTitleIgnoreRegEx.MatchString(title)
}

func removeIgnoredStrategies(ctx context.Context, fetchedFeature *unleash.FetchedFeature, strategyTitleIgnoreRegEx *regexp.Regexp) {
	if strategyTitleIgnoreRegEx == nil {
		return
	}
	for i := range fetchedFeature.FetchedEnvironments {
		env := &fetchedFeature.FetchedEnvironments[i]
		if len(env.FetchedStrategies) == 0 {
			continue
		}
		strategiesWithoutIgnore := make([]unleash.FeatureStrategySchema, 0, len(env.FetchedStrategies))
		for _, strategy := range env.FetchedStrategies {
			if strategy.Title != nil && strategyTitleIgnoreRegEx.MatchString(*strategy.Title) {
				tflog.Info(ctx, "Ignoring strategy", map[string]interface{}{
					"projectID":     fetchedFeature.FetchedProject,
					"feature":       fetchedFeature.Feature.Name,
					"environmentID": env.Environment.Name,
					"strategy":      strategy,
				})
				continue
			}
			strategiesWithoutIgnore = append(strategiesWithoutIgnore, strategy)
		}
		env.FetchedStrategies = strategiesWithoutIgnore
	}
}
