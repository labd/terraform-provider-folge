package checkhttpstatus

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"

	"github.com/labd/terraform-provider-folge/internal/folge"
	"github.com/labd/terraform-provider-folge/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &checkHttpStatusResource{}
	_ resource.ResourceWithConfigure   = &checkHttpStatusResource{}
	_ resource.ResourceWithImportState = &checkHttpStatusResource{}
)

// NewCheckHttpStatusResource is a helper function to simplify the provider implementation.
func NewCheckHttpStatusResource() resource.Resource {
	return &checkHttpStatusResource{}
}

// checkHttpStatusResource is the resource implementation.
type checkHttpStatusResource struct {
	client folge.ClientWithResponsesInterface
}

// Metadata returns the data source type name.
func (r *checkHttpStatusResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_check_http_status"
}

// Schema defines the schema for the data source.
func (r *checkHttpStatusResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The check_http_status",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the check_http_status",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"application_id": schema.Int64Attribute{
				Description: "The ID of the applicaton",
				Required:    true,
			},
			"datasource_id": schema.Int64Attribute{
				Description: "The ID of the datasource",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the check is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"crontab": schema.StringAttribute{
				Description: "The crontab.",
				Required:    true,
			},
			"status_code": schema.Int64Attribute{
				Description: "The expected HTTP status code.",
				Required:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (r *checkHttpStatusResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = utils.GetClient(req.ProviderData)
}

// Create creates the resource and sets the initial Terraform state.
func (r *checkHttpStatusResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan CheckHttpStatusModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	input := plan.toCreateInput()
	appId := utils.AsInt(plan.ApplicationID)
	dsId := utils.AsInt(plan.DataSourceID)

	content, err := r.client.ApplicationsDataSourcesChecksCreateWithResponse(ctx, appId, dsId, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating check_http_status",
			"Could not create check_http_status, unexpected error: "+err.Error(),
		)
		return
	}
	if content.StatusCode() != http.StatusCreated {
		resp.Diagnostics.AddError(
			"Error creating check_http_status",
			fmt.Sprintf(
				"Could not create check_http_status, status code %d error: %s",
				content.StatusCode(), string(content.Body)),
		)
		return
	}

	check_http_status := content.JSON201

	// Map response body to schema and populate Computed attribute values
	if err := plan.fromRemote(*check_http_status, appId, dsId); err != nil {
		resp.Diagnostics.AddError(
			"Error creating check_http_status",
			"Could not create check_http_status, unexpected error: "+err.Error(),
		)
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *checkHttpStatusResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state CheckHttpStatusModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())
	appId := utils.AsInt(state.ApplicationID)
	dsId := utils.AsInt(state.DataSourceID)
	content, err := r.client.ApplicationsDataSourcesChecksRetrieveWithResponse(ctx, appId, dsId, id)
	if d := utils.CheckGetError("check_http_status", id, content, err); d != nil {
		resp.Diagnostics.Append(d)
		return
	}

	check := *content.JSON200

	// Overwrite items with refreshed state
	if err := state.fromRemote(check, appId, dsId); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Check",
			fmt.Sprintf("Could not read Check ID %d: %s", state.ID.ValueInt64(), err.Error()),
		)
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *checkHttpStatusResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan CheckHttpStatusModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	input := plan.toUpdateInput()
	planId := utils.AsInt(plan.ID)
	appId := utils.AsInt(plan.ApplicationID)
	dsId := utils.AsInt(plan.DataSourceID)

	content, err := r.client.ApplicationsDataSourcesChecksUpdateWithResponse(ctx, appId, dsId, planId, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating check_http_status",
			"Could not update check_http_status, unexpected error: "+err.Error(),
		)
		return
	}
	if content.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Error updating check_http_status",
			fmt.Sprintf(
				"Could not update check_http_status, status code %d error: %s",
				content.StatusCode(), string(content.Body)),
		)
		return
	}

	check_http_status := *content.JSON200

	// Map response body to schema and populate Computed attribute values
	if err := plan.fromRemote(check_http_status, appId, dsId); err != nil {
		resp.Diagnostics.AddError(
			"Error updating check_http_status",
			"Could not update check_http_status, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *checkHttpStatusResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state CheckHttpStatusModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := utils.AsInt(state.ID)
	content, err := r.client.ApplicationsDestroyWithResponse(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting check_http_status",
			"Could not delete check_http_status, unexpected error: "+err.Error(),
		)
		return
	}
	if content.StatusCode() != http.StatusNoContent {
		resp.Diagnostics.AddError(
			"Error deleting check_http_status",
			fmt.Sprintf(
				"Could not delete check_http_status, status code %d error: %s",
				content.StatusCode(), string(content.Body)),
		)
		return
	}

}

func (r *checkHttpStatusResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
