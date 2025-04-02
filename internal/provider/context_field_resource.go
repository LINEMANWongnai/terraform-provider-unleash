// Copyright (c) HashiCorp, Inc.

package provider

import (
	"context"
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/LINEMANWongnai/terraform-provider-unleash/internal/unleash"
)

var _ resource.Resource = &ContextFieldResource{}
var _ resource.ResourceWithImportState = &ContextFieldResource{}

func NewContextFieldResource() resource.Resource {
	return &ContextFieldResource{}
}

type ContextFieldResource struct {
	providerData UnleashProviderData
}

type ContextFieldResourceModel struct {
	ContextFieldModel
}

func (r *ContextFieldResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_context_field"
}

func (r *ContextFieldResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Context field resource",

		Attributes: createContextFieldResourceSchemaAttr(),
	}
}

func (r *ContextFieldResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ContextFieldResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ContextFieldResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	requestBody := toCreateContextFieldJSONRequestBody(data)

	tflog.Debug(ctx, "Creating context field", map[string]interface{}{"body": requestBody})
	createResp, err := r.providerData.Client.CreateContextFieldWithResponse(ctx, requestBody)
	if err != nil {
		resp.Diagnostics.AddError("failed to create context field "+data.Name.String(), err.Error())
		return
	}
	if createResp.StatusCode() > 299 {
		resp.Diagnostics.AddError("failed to create context field "+data.Name.String(), fmt.Sprintf(" with status %d %s", createResp.StatusCode(), string(createResp.Body)))
		return
	}
	data.ID = types.StringValue(createResp.JSON201.Name)

	tflog.Trace(ctx, "created a resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContextFieldResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ContextFieldResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading context field", map[string]interface{}{"id": data.ID.ValueString()})
	readResp, err := r.providerData.Client.GetContextFieldWithResponse(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get context field", err.Error())
		return
	}
	if readResp.StatusCode() > 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if readResp.StatusCode() > 299 {
		resp.Diagnostics.AddError("failed to read context field "+data.Name.String(), fmt.Sprintf(" with status %d %s", readResp.StatusCode(), string(readResp.Body)))
		return
	}
	contextModel := toContextFieldModel(readResp.JSON200)
	ensureContextModelNullAndEmptyConsistency(&contextModel, data.ContextFieldModel)
	data.ContextFieldModel = contextModel

	tflog.Trace(ctx, "read resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContextFieldResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ContextFieldResourceModel
	var existingData ContextFieldResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &existingData)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID != data.Name {
		resp.Diagnostics.AddError("name of context field cannot be changed "+data.Name.String(), data.ID.String())
		return
	}

	requestBody := toUpdateContextFieldJSONRequestBody(data)
	existingRequestBody := toUpdateContextFieldJSONRequestBody(existingData)
	data.ID = existingData.ID
	if !cmp.Equal(requestBody, existingRequestBody) {
		tflog.Debug(ctx, "Updating context field", map[string]interface{}{
			"id":   data.ID.ValueString(),
			"name": data.Name.ValueString(),
			"body": requestBody,
		})
		updateResp, err := r.providerData.Client.UpdateContextFieldWithResponse(ctx, data.ID.ValueString(), requestBody)
		if err != nil {
			resp.Diagnostics.AddError("failed to update context field "+data.ID.String(), err.Error())
			return
		}
		if updateResp.StatusCode() > 299 {
			resp.Diagnostics.AddError("failed to update context field "+data.ID.String(), fmt.Sprintf(" with status %d %s", updateResp.StatusCode(), string(updateResp.Body)))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func toCreateContextFieldJSONRequestBody(contextFieldModel ContextFieldResourceModel) unleash.CreateContextFieldJSONRequestBody {
	body := unleash.CreateContextFieldJSONRequestBody{
		Name: contextFieldModel.Name.ValueString(),
	}
	if !contextFieldModel.Description.IsNull() {
		body.Description = contextFieldModel.Description.ValueStringPointer()
	}
	if !contextFieldModel.SortOrder.IsNull() {
		i := int(contextFieldModel.SortOrder.ValueInt32())
		body.SortOrder = &i
	}
	if !contextFieldModel.Stickiness.IsNull() {
		body.Stickiness = contextFieldModel.Stickiness.ValueBoolPointer()
	}
	if len(contextFieldModel.LegalValues) > 0 {
		legalValues := toLegalValuesBody(contextFieldModel.LegalValues)
		body.LegalValues = &legalValues
	}
	return body
}

func toLegalValuesBody(legalValueModels []LegalValueModel) []unleash.LegalValueSchema {
	body := make([]unleash.LegalValueSchema, len(legalValueModels))
	for i, legalValueModel := range legalValueModels {
		body[i] = toLegalValueBody(legalValueModel)
	}

	return body
}

func toLegalValueBody(legalValueModel LegalValueModel) unleash.LegalValueSchema {
	body := unleash.LegalValueSchema{
		Value: legalValueModel.Value.ValueString(),
	}
	if !legalValueModel.Description.IsNull() {
		body.Description = legalValueModel.Description.ValueStringPointer()
	}

	return body
}

func toUpdateContextFieldJSONRequestBody(contextFieldModel ContextFieldResourceModel) unleash.UpdateContextFieldJSONRequestBody {
	body := unleash.UpdateContextFieldJSONRequestBody{}
	if !contextFieldModel.Description.IsNull() {
		body.Description = contextFieldModel.Description.ValueStringPointer()
	}
	if !contextFieldModel.SortOrder.IsNull() {
		i := int(contextFieldModel.SortOrder.ValueInt32())
		body.SortOrder = &i
	}
	if !contextFieldModel.Stickiness.IsNull() {
		body.Stickiness = contextFieldModel.Stickiness.ValueBoolPointer()
	}
	if len(contextFieldModel.LegalValues) > 0 {
		legalValues := toLegalValuesBody(contextFieldModel.LegalValues)
		body.LegalValues = &legalValues
	}
	return body
}

func (r *ContextFieldResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ContextFieldResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting context field", map[string]interface{}{
		"id":   data.ID.ValueString(),
		"name": data.Name.ValueString(),
	})
	removeResp, err := r.providerData.Client.DeleteContextFieldWithResponse(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete context field "+data.ID.String(), err.Error())
		return
	}
	if removeResp.StatusCode() > 299 && removeResp.StatusCode() != 404 {
		resp.Diagnostics.AddError("failed to delete context field "+data.ID.String(), fmt.Sprintf(" with status %d %s", removeResp.StatusCode(), string(removeResp.Body)))
		return
	}
}

func (r *ContextFieldResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
