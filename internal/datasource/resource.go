package datasource

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"

	"github.com/labd/terraform-provider-folge/internal/folge"
	"github.com/labd/terraform-provider-folge/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &dataSourceResource{}
	_ resource.ResourceWithConfigure   = &dataSourceResource{}
	_ resource.ResourceWithImportState = &dataSourceResource{}
)

// NewDataSourceResource is a helper function to simplify the provider implementation.
func NewDataSourceResource() resource.Resource {
	return &dataSourceResource{}
}

// dataSourceResource is the resource implementation.
type dataSourceResource struct {
	client folge.ClientWithResponsesInterface
}

// Metadata returns the data source type name.
func (r *dataSourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datasource"
}

// Schema defines the schema for the data source.
func (r *dataSourceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The datasource",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the datasource",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"application_id": schema.Int64Attribute{
				Description: "The ID of the applicaton",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name.",
				Required:    true,
			},
			"url": schema.StringAttribute{
				Description: "The URL.",
				Required:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"basic_auth": schema.SingleNestedBlock{
				Description: "Basic auth credentials",
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						Description: "The username.",
						Required:    true,
					},
					"password": schema.StringAttribute{
						Description: "The password.",
						Required:    true,
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (r *dataSourceResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = utils.GetClient(req.ProviderData)
}

// Create creates the resource and sets the initial Terraform state.
func (r *dataSourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan DataSourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	input := plan.toCreateInput()
	appId := utils.AsInt(plan.ApplicationID)

	content, err := r.client.ApplicationsDataSourcesCreateWithResponse(ctx, appId, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating datasource",
			"Could not create datasource, unexpected error: "+err.Error(),
		)
		return
	}
	if content.StatusCode() != http.StatusCreated {
		resp.Diagnostics.AddError(
			"Error creating datasource",
			fmt.Sprintf(
				"Could not create datasource, status code %d error: %s",
				content.StatusCode(), string(content.Body)),
		)
		return
	}

	datasource := content.JSON201

	// Map response body to schema and populate Computed attribute values
	if err := plan.fromRemote(*datasource, appId); err != nil {
		resp.Diagnostics.AddError(
			"Error creating datasource",
			"Could not create datasource, unexpected error: "+err.Error(),
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
func (r *dataSourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state DataSourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())
	appId := utils.AsInt(state.ApplicationID)
	content, err := r.client.ApplicationsDataSourcesRetrieveWithResponse(ctx, appId, id)
	if d := utils.CheckGetError("datasource", id, content, err); d != nil {
		resp.Diagnostics.Append(d)
		return
	}

	datasource := *content.JSON200

	// Overwrite items with refreshed state
	if err := state.fromRemote(datasource, appId); err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Application",
			fmt.Sprintf("Could not read Application ID %d: %s", state.ID.ValueInt64(), err.Error()),
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
func (r *dataSourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan DataSourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	input := plan.toUpdateInput()
	planId := utils.AsInt(plan.ID)
	appId := utils.AsInt(plan.ApplicationID)

	content, err := r.client.ApplicationsDataSourcesUpdateWithResponse(ctx, appId, planId, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating datasource",
			"Could not update datasource, unexpected error: "+err.Error(),
		)
		return
	}
	if content.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Error updating datasource",
			fmt.Sprintf(
				"Could not update datasource, status code %d error: %s",
				content.StatusCode(), string(content.Body)),
		)
		return
	}

	datasource := *content.JSON200

	// Map response body to schema and populate Computed attribute values
	if err := plan.fromRemote(datasource, appId); err != nil {
		resp.Diagnostics.AddError(
			"Error updating datasource",
			"Could not update datasource, unexpected error: "+err.Error(),
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
func (r *dataSourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state DataSourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := utils.AsInt(state.ID)
	content, err := r.client.ApplicationsDestroyWithResponse(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting datasource",
			"Could not delete datasource, unexpected error: "+err.Error(),
		)
		return
	}
	if content.StatusCode() != http.StatusNoContent {
		resp.Diagnostics.AddError(
			"Error deleting datasource",
			fmt.Sprintf(
				"Could not delete datasource, status code %d error: %s",
				content.StatusCode(), string(content.Body)),
		)
		return
	}

}

func (r *dataSourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
