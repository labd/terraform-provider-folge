package application

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/labd/terraform-provider-folge/internal/folge"
	"github.com/labd/terraform-provider-folge/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &applicationResource{}
	_ resource.ResourceWithConfigure   = &applicationResource{}
	_ resource.ResourceWithImportState = &applicationResource{}
)

// NewApplicationResource is a helper function to simplify the provider implementation.
func NewApplicationResource() resource.Resource {
	return &applicationResource{}
}

// applicationResource is the resource implementation.
type applicationResource struct {
	client folge.ClientWithResponsesInterface
}

// Metadata returns the data source type name.
func (r *applicationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

// Schema defines the schema for the data source.
func (r *applicationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The application",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the application",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name.",
				Required:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (r *applicationResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = utils.GetClient(req.ProviderData)
}

// Create creates the resource and sets the initial Terraform state.
func (r *applicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan ApplicationModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	input := plan.toCreateInput()

	content, err := r.client.ApplicationsCreateWithResponse(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating application",
			"Could not create application, unexpected error: "+err.Error(),
		)
		return
	}
	if content.StatusCode() != http.StatusCreated {
		resp.Diagnostics.AddError(
			"Error creating application",
			fmt.Sprintf(
				"Could not create application, status code %d error: %s",
				content.StatusCode(), string(content.Body)),
		)
		return
	}

	application := content.JSON201

	// Map response body to schema and populate Computed attribute values
	if err := plan.fromRemote(*application); err != nil {
		resp.Diagnostics.AddError(
			"Error creating application",
			"Could not create application, unexpected error: "+err.Error(),
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
func (r *applicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state ApplicationModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())
	content, err := r.client.ApplicationsRetrieveWithResponse(ctx, id)
	if d := utils.CheckGetError("application", id, content, err); d != nil {
		resp.Diagnostics.Append(d)
		return
	}

	application := *content.JSON200

	// Overwrite items with refreshed state
	if err := state.fromRemote(application); err != nil {
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
func (r *applicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan ApplicationModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	input := plan.toUpdateInput()
	planId := utils.AsInt(plan.ID)

	tflog.Info(ctx, fmt.Sprintf("Updating application %d", planId))
	content, err := r.client.ApplicationsUpdateWithResponse(ctx, planId, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating application",
			"Could not update application, unexpected error: "+err.Error(),
		)
		return
	}
	if content.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Error updating application",
			fmt.Sprintf(
				"Could not update application, status code %d error: %s",
				content.StatusCode(), string(content.Body)),
		)
		return
	}

	application := *content.JSON200

	// Map response body to schema and populate Computed attribute values
	if err := plan.fromRemote(application); err != nil {
		resp.Diagnostics.AddError(
			"Error updating application",
			"Could not update application, unexpected error: "+err.Error(),
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
func (r *applicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state ApplicationModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := utils.AsInt(state.ID)
	content, err := r.client.ApplicationsDestroyWithResponse(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting application",
			"Could not delete application, unexpected error: "+err.Error(),
		)
		return
	}
	if content.StatusCode() != http.StatusNoContent {
		resp.Diagnostics.AddError(
			"Error deleting application",
			fmt.Sprintf(
				"Could not delete application, status code %d error: %s",
				content.StatusCode(), string(content.Body)),
		)
		return
	}

}

func (r *applicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
