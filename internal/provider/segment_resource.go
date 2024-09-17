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

var _ resource.Resource = &SegmentResource{}
var _ resource.ResourceWithImportState = &SegmentResource{}

func NewSegmentResource() resource.Resource {
	return &SegmentResource{}
}

type SegmentResource struct {
	providerData UnleashProviderData
}

type SegmentResourceModel struct {
	SegmentModel
}

func (r *SegmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_segment"
}

func (r *SegmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Segment resource",

		Attributes: createSegmentResourceSchemaAttr(),
	}
}

func (r *SegmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SegmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SegmentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	body, err := toSegmentBody(data)
	if err != nil {
		resp.Diagnostics.AddError("failed to create segment "+data.Name.String(), err.Error())
		return
	}

	tflog.Debug(ctx, "Creating segment", map[string]interface{}{"body": body})
	createResp, err := r.providerData.Client.CreateSegmentWithResponse(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("failed to create segment "+data.Name.String(), err.Error())
		return
	}
	if createResp.StatusCode() > 299 {
		resp.Diagnostics.AddError("failed to create segment "+data.Name.String(), fmt.Sprintf(" with status %d %s", createResp.StatusCode(), string(createResp.Body)))
		return
	}
	data.ID = types.StringValue(fmt.Sprintf("%d", createResp.JSON201.Id))
	data.IDInt = types.Int64Value(int64(createResp.JSON201.Id))

	tflog.Trace(ctx, "created a resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SegmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SegmentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading segment", map[string]interface{}{"id": data.ID.ValueString()})
	readResp, err := r.providerData.Client.GetSegmentWithResponse(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get segment", err.Error())
		return
	}
	if readResp.StatusCode() > 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if readResp.StatusCode() > 299 {
		resp.Diagnostics.AddError("failed to read segment "+data.Name.String(), fmt.Sprintf(" with status %d %s", readResp.StatusCode(), string(readResp.Body)))
		return
	}
	segmentModel, err := toSegmentModel(readResp.JSON200)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert segment", err.Error())
		return
	}
	ensureSegmentModelNullAndEmptyConsistency(&segmentModel, data.SegmentModel)
	data.SegmentModel = segmentModel

	tflog.Trace(ctx, "read resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SegmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SegmentResourceModel
	var existingData SegmentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &existingData)...)

	if resp.Diagnostics.HasError() {
		return
	}

	segmentBody, err := toSegmentBody(data)
	if err != nil {
		resp.Diagnostics.AddError("failed to build segment request body "+data.Name.String(), err.Error())
		return
	}
	existingSegmentBody, err := toSegmentBody(existingData)
	if err != nil {
		resp.Diagnostics.AddError("failed to build segment request body "+existingData.Name.String(), err.Error())
		return
	}
	data.ID = existingData.ID
	data.IDInt = existingData.IDInt
	if !cmp.Equal(segmentBody, existingSegmentBody) {
		tflog.Debug(ctx, "Updating segment", map[string]interface{}{
			"projectID": data.Project.ValueString(),
			"id":        data.ID.ValueString(),
			"name":      data.Name.ValueString(),
			"body":      segmentBody,
		})
		updateResp, err := r.providerData.Client.UpdateSegmentWithResponse(ctx, data.ID.ValueString(), segmentBody)
		if err != nil {
			resp.Diagnostics.AddError("failed to update segment "+data.ID.String(), err.Error())
			return
		}
		if updateResp.StatusCode() > 299 {
			resp.Diagnostics.AddError("failed to update segment "+data.ID.String(), fmt.Sprintf(" with status %d %s", updateResp.StatusCode(), string(updateResp.Body)))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func toSegmentBody(segmentModel SegmentResourceModel) (unleash.UpsertSegmentSchema, error) {
	body := unleash.CreateSegmentJSONRequestBody{
		Name: segmentModel.Name.ValueString(),
	}
	if !segmentModel.Description.IsNull() {
		body.Description = segmentModel.Description.ValueStringPointer()
	}
	if !segmentModel.Project.IsNull() {
		body.Project = segmentModel.Project.ValueStringPointer()
	}
	if len(segmentModel.Constraints) > 0 {
		var err error
		body.Constraints, err = toConstraintsBody(segmentModel.Constraints)
		if err != nil {
			return body, err
		}
	}
	return body, nil
}

func (r *SegmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SegmentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting segment", map[string]interface{}{
		"projectID": data.Project.ValueString(),
		"id":        data.ID.ValueString(),
		"name":      data.Name.ValueString(),
	})
	removeResp, err := r.providerData.Client.RemoveSegmentWithResponse(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete segment "+data.ID.String(), err.Error())
		return
	}
	if removeResp.StatusCode() > 299 && removeResp.StatusCode() != 404 {
		resp.Diagnostics.AddError("failed to delete segment "+data.ID.String(), fmt.Sprintf(" with status %d %s", removeResp.StatusCode(), string(removeResp.Body)))
		return
	}
}

func (r *SegmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
