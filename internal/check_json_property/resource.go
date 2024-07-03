package checkjsonproperty

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/labd/terraform-provider-folge/internal/folge"
	"github.com/labd/terraform-provider-folge/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &checkJsonPropertyResource{}
	_ resource.ResourceWithConfigure   = &checkJsonPropertyResource{}
	_ resource.ResourceWithImportState = &checkJsonPropertyResource{}
)

// NewCheckJsonPropertyResource is a helper function to simplify the provider implementation.
func NewCheckJsonPropertyResource() resource.Resource {
	return &checkJsonPropertyResource{}
}

// checkJsonPropertyResource is the resource implementation.
type checkJsonPropertyResource struct {
	client folge.ClientWithResponsesInterface
}

// Metadata returns the data source type name.
func (r *checkJsonPropertyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_check_json_property"
}

// Schema defines the schema for the data source.
func (r *checkJsonPropertyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The check_json_property",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the check_json_property",
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
			"path": schema.StringAttribute{
				Description: "The json path to check.",
				Required:    true,
			},
			"datatype": schema.StringAttribute{
				Description: "The data type of the property.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("str", "int", "bool", "datetime"),
				},
			},
			"operator": schema.StringAttribute{
				Description: "The operator to use for the check.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("eq", "neq", "gt", "lt"),
				},
			},
			"value_bool": schema.BoolAttribute{
				Description: "The value to compare against for a boolean property.",
				Optional:    true,
			},
			"value_int": schema.Int64Attribute{
				Description: "The value to compare against for an integer property.",
				Optional:    true,
			},
			"value_string": schema.StringAttribute{
				Description: "The value to compare against for a string property.",
				Optional:    true,
			},
			"value_datetime": schema.StringAttribute{
				Description: "The value to compare against for a datetime property.",
				Optional:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (r *checkJsonPropertyResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = utils.GetClient(req.ProviderData)
}

// Create creates the resource and sets the initial Terraform state.
func (r *checkJsonPropertyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan CheckJsonPropertyModel
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
			"Error creating check_json_property",
			"Could not create check_json_property, unexpected error: "+err.Error(),
		)
		return
	}
	if content.StatusCode() != http.StatusCreated {
		resp.Diagnostics.AddError(
			"Error creating check_json_property",
			fmt.Sprintf(
				"Could not create check_json_property, status code %d error: %s",
				content.StatusCode(), string(content.Body)),
		)
		return
	}

	check_json_property := content.JSON201

	// Map response body to schema and populate Computed attribute values
	if err := plan.fromRemote(*check_json_property, appId, dsId); err != nil {
		resp.Diagnostics.AddError(
			"Error creating check_json_property",
			"Could not create check_json_property, unexpected error: "+err.Error(),
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
func (r *checkJsonPropertyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state CheckJsonPropertyModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())
	appId := utils.AsInt(state.ApplicationID)
	dsId := utils.AsInt(state.DataSourceID)
	content, err := r.client.ApplicationsDataSourcesChecksRetrieveWithResponse(ctx, appId, dsId, id)
	if d := utils.CheckGetError("check_json_property", id, content, err); d != nil {
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
func (r *checkJsonPropertyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan CheckJsonPropertyModel
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
			"Error updating check_json_property",
			"Could not update check_json_property, unexpected error: "+err.Error(),
		)
		return
	}
	if content.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Error updating check_json_property",
			fmt.Sprintf(
				"Could not update check_json_property, status code %d error: %s",
				content.StatusCode(), string(content.Body)),
		)
		return
	}

	check_json_property := *content.JSON200

	// Map response body to schema and populate Computed attribute values
	if err := plan.fromRemote(check_json_property, appId, dsId); err != nil {
		resp.Diagnostics.AddError(
			"Error updating check_json_property",
			"Could not update check_json_property, unexpected error: "+err.Error(),
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
func (r *checkJsonPropertyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state CheckJsonPropertyModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := utils.AsInt(state.ID)
	content, err := r.client.ApplicationsDestroyWithResponse(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting check_json_property",
			"Could not delete check_json_property, unexpected error: "+err.Error(),
		)
		return
	}
	if content.StatusCode() != http.StatusNoContent {
		resp.Diagnostics.AddError(
			"Error deleting check_json_property",
			fmt.Sprintf(
				"Could not delete check_json_property, status code %d error: %s",
				content.StatusCode(), string(content.Body)),
		)
		return
	}

}

func (r *checkJsonPropertyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
