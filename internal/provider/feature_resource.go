package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"git.wndv.co/LINEMANWongnai/terraform-provider-unleash/internal/ptr"
	"git.wndv.co/LINEMANWongnai/terraform-provider-unleash/internal/unleash"
)

var _ resource.Resource = &FeatureResource{}
var _ resource.ResourceWithImportState = &FeatureResource{}

func NewFeatureResource() resource.Resource {
	return &FeatureResource{}
}

type FeatureResource struct {
	client unleash.ClientWithResponsesInterface
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

	c, ok := req.ProviderData.(unleash.ClientWithResponsesInterface)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = c
}

func (r *FeatureResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FeatureResourceModel

	// Read Terraform plan data into the model
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

	createResp, err := r.client.CreateFeatureWithResponse(ctx, data.Project.ValueString(), body)
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

	// Write logs using the 	tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FeatureResource) updateEnvironments(ctx context.Context, projectID string, featureName string, environments map[string]EnvironmentModel, existingEnvironmentByName map[string]EnvironmentModel) error {
	for name, env := range environments {
		existingEnv, ok := existingEnvironmentByName[name]
		if !ok {
			existingEnv = toEnvironmentModel(unleash.FetchedEnvironment{})
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
			values := make([]string, 0, len(constraint.Values))
			for _, value := range constraint.Values {
				values = append(values, value.ValueString())
			}
			constraintBody.Values = &values

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
				Name: variant.Name.ValueString(),
				Payload: &struct {
					Type  unleash.CreateStrategyVariantSchemaPayloadType `json:"type"`
					Value string                                         `json:"value"`
				}{
					Type:  unleash.CreateStrategyVariantSchemaPayloadType(variant.PayloadType.ValueString()),
					Value: variant.Payload.ValueString(),
				},
				Stickiness: variant.Stickiness.ValueString(),
				WeightType: unleash.CreateStrategyVariantSchemaWeightType(variant.WeightType.ValueString()),
			}
			if variantBody.WeightType == unleash.CreateStrategyVariantSchemaWeightTypeFix {
				variantBody.Weight = int(variant.Weight.ValueInt64())
			}

			variants = append(variants, variantBody)
		}
		strategyBody.Variants = &variants
	}
	resp, err := r.client.AddFeatureStrategyWithResponse(ctx, projectID, featureName, environmentID, strategyBody)
	if err != nil {
		return "", err
	}
	if resp.StatusCode() > 299 {
		return "", fmt.Errorf("failed to add strategy for %s %s %s %s with status %d %s", projectID, featureName, environmentID, strategy.Name.ValueString(), resp.StatusCode(), string(resp.Body))
	}

	return *resp.JSON200.Id, nil
}

func (r *FeatureResource) updateStrategy(ctx context.Context, projectID string, featureName string, environmentID string, strategy StrategyModel, existingStrategy StrategyModel) error {
	body := toUpdateStrategyBody(strategy)
	existingBody := toUpdateStrategyBody(existingStrategy)
	if !cmp.Equal(body, existingBody) {
		resp, err := r.client.UpdateFeatureStrategyWithResponse(ctx, projectID, featureName, environmentID, existingStrategy.Id.ValueString(), body)
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
		resp, err := r.client.SetStrategySortOrderWithResponse(ctx, projectID, featureName, environmentID, []struct {
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
		resp, err := r.client.UpdateFeatureStrategySegmentsWithResponse(ctx, updateStrategySegmentBody)
		if err != nil {
			return err
		}
		if resp.StatusCode() > 299 {
			return fmt.Errorf("failed to update strategy segments for %s %s %s %s with status %d %s", projectID, featureName, environmentID, strategy.Name.ValueString(), resp.StatusCode(), string(resp.Body))
		}
	}

	return nil
}

func toUpdateStrategyBody(strategy StrategyModel) unleash.UpdateFeatureStrategyJSONRequestBody {
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
	if len(strategy.Parameters) > 0 {
		parameters := make(unleash.ParametersSchema)
		for parameterK, parameterV := range strategy.Parameters {
			parameters[parameterK] = parameterV.ValueString()
		}
		body.Parameters = &parameters
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
			if len(constraint.Values) > 0 {
				values := make([]string, 0, len(constraint.Values))
				for _, value := range constraint.Values {
					values = append(values, value.ValueString())
				}
				constraintBody.Values = &values
			}

			constraints = append(constraints, constraintBody)
		}
		body.Constraints = &constraints
	}
	if len(strategy.Variants) > 0 {
		variants := make([]unleash.CreateStrategyVariantSchema, 0, len(strategy.Variants))
		for _, variant := range strategy.Variants {
			variantBody := unleash.CreateStrategyVariantSchema{
				Name: variant.Name.ValueString(),
				Payload: &struct {
					Type  unleash.CreateStrategyVariantSchemaPayloadType `json:"type"`
					Value string                                         `json:"value"`
				}{
					Type:  unleash.CreateStrategyVariantSchemaPayloadType(variant.PayloadType.ValueString()),
					Value: variant.Payload.ValueString(),
				},
				Stickiness: variant.Stickiness.ValueString(),
				WeightType: unleash.CreateStrategyVariantSchemaWeightType(variant.WeightType.ValueString()),
			}
			if variantBody.WeightType == unleash.CreateStrategyVariantSchemaWeightTypeFix {
				variantBody.Weight = int(variant.Weight.ValueInt64())
			}

			variants = append(variants, variantBody)
		}
		body.Variants = &variants
	}

	return body
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
	resp, err := r.client.DeleteFeatureStrategyWithResponse(ctx, projectID, featureName, environmentID, strategy.Id.ValueString())
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
	if r.isVariantsEquals(environment.Variants, existingEnv.Variants) {
		return environment, nil
	}
	variantsBody := unleash.OverwriteFeatureVariantsOnEnvironmentsJSONRequestBody{
		Environments: &[]string{environmentID},
	}
	variants := make([]unleash.VariantSchema, 0, len(environment.Variants))
	for _, variant := range environment.Variants {
		variantBody := unleash.VariantSchema{
			Name: variant.Name.ValueString(),
			Payload: &struct {
				Type  unleash.VariantSchemaPayloadType `json:"type"`
				Value string                           `json:"value"`
			}{
				Type:  unleash.VariantSchemaPayloadType(variant.PayloadType.ValueString()),
				Value: variant.Payload.ValueString(),
			},
			Stickiness: variant.Stickiness.ValueStringPointer(),
		}
		if !variant.WeightType.IsNull() {
			t := unleash.VariantSchemaWeightType(variant.WeightType.ValueString())
			variantBody.WeightType = &t
			if *variantBody.WeightType == unleash.Fix {
				variantBody.Weight = variant.Weight.ValueFloat32()
			}
		}
		if len(variant.Overrides) > 0 {
			overrides := make([]unleash.OverrideSchema, len(variant.Overrides))
			for i, override := range variant.Overrides {
				overrideBody := unleash.OverrideSchema{
					ContextName: override.ContextName.ValueString(),
				}
				if len(override.Values) > 0 {
					for _, v := range override.Values {
						overrideBody.Values = append(overrideBody.Values, v.ValueString())
					}
				}
				overrides[i] = overrideBody
			}
			variantBody.Overrides = &overrides
		}
		variants = append(variants, variantBody)
	}
	variantsBody.Variants = &variants
	resp, err := r.client.OverwriteFeatureVariantsOnEnvironmentsWithResponse(ctx, projectID, featureName, variantsBody)
	if err != nil {
		return environment, err
	}
	if resp.StatusCode() > 299 {
		return environment, fmt.Errorf("failed to overwrite variants for %s %s %s with status %d %s", projectID, featureName, environmentID, resp.StatusCode(), string(resp.Body))
	}

	return environment, nil
}

func (r *FeatureResource) isVariantsEquals(variants []VariantModel, variants2 []VariantModel) bool {
	if len(variants) != len(variants2) {
		return false
	}
	variantModelByName := toVariantModelByName(variants)
	variantModelByName2 := toVariantModelByName(variants2)

	for name, variant := range variantModelByName {
		variant2, ok := variantModelByName2[name]
		if !ok {
			return false
		}
		if cmp.Equal(variant, variant2) {
			return false
		}
	}

	return true
}

func toVariantModelByName(variants []VariantModel) map[string]VariantModel {
	variantModelByName := make(map[string]VariantModel)
	for _, variant := range variants {
		variantModelByName[variant.Name.ValueString()] = variant
	}

	return variantModelByName
}

func (r *FeatureResource) updateEnvironmentStatus(ctx context.Context, projectID string, featureName string, environmentID string, environment EnvironmentModel, existingEnv EnvironmentModel) (EnvironmentModel, error) {
	if environment.Enabled == existingEnv.Enabled {
		return environment, nil
	}

	if environment.Enabled.ValueBool() {
		resp, err := r.client.ToggleFeatureEnvironmentOnWithResponse(ctx, projectID, featureName, environmentID)
		if err != nil {
			return environment, err
		}
		if resp.StatusCode() > 299 {
			return environment, fmt.Errorf("failed to enable environment for %s %s %s with status %d %s", projectID, featureName, environmentID, resp.StatusCode(), string(resp.Body))
		}
	} else {
		resp, err := r.client.ToggleFeatureEnvironmentOffWithResponse(ctx, projectID, featureName, environmentID)
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

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID, featureName := extractProjectAndFeatureName(data)

	fetchedFeature, err := unleash.GetFeature(ctx, r.client, projectID, featureName)
	if err != nil {
		resp.Diagnostics.AddError("failed to get features", err.Error())
		return
	}

	data.FeatureModel = toFeatureModel(fetchedFeature)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read resource")

	// Save updated data into Terraform state
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

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &existingData)...)

	if resp.Diagnostics.HasError() {
		return
	}

	featureBody := toFeatureBody(data)
	existingFeatureBody := toFeatureBody(existingData)
	if !cmp.Equal(featureBody, existingFeatureBody) {
		updateResp, err := r.client.UpdateFeatureWithResponse(ctx, data.Project.ValueString(), data.Name.ValueString(), featureBody)
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

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func toFeatureBody(data FeatureResourceModel) unleash.UpdateFeatureJSONRequestBody {
	return unleash.UpdateFeatureJSONRequestBody{
		Type:           data.Type.ValueStringPointer(),
		Description:    data.Description.ValueStringPointer(),
		ImpressionData: data.ImpressionData.ValueBoolPointer(),
	}
}

func (r *FeatureResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FeatureResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider unleash data and make a call using it.
	// httpResp, err := r.unleash.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
	//     return
	// }

	data.ID = types.StringValue(resolveID(data))

	archiveResp, err := r.client.ArchiveFeatureWithResponse(ctx, data.Project.ValueString(), data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to archive feature "+data.ID.String(), err.Error())
		return
	}
	if archiveResp.StatusCode() > 299 && archiveResp.StatusCode() != 404 {
		resp.Diagnostics.AddError("failed to archive feature "+data.ID.String(), fmt.Sprintf(" with status %d %s", archiveResp.StatusCode(), string(archiveResp.Body)))
		return
	}
	deleteResp, err := r.client.DeleteFeatureWithResponse(ctx, data.Name.ValueString())
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
