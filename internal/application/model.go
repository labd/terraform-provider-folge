package application

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-folge/internal/folge"
)

type ApplicationModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (m *ApplicationModel) toCreateInput() folge.ApplicationsCreateJSONRequestBody {
	return folge.ApplicationsCreateJSONRequestBody{
		Name: m.Name.ValueString(),
	}
}
func (m *ApplicationModel) toUpdateInput() folge.ApplicationsUpdateJSONRequestBody {
	return folge.ApplicationsUpdateJSONRequestBody{
		Name: m.Name.ValueString(),
	}
}

func (m *ApplicationModel) fromRemote(i folge.Application) error {
	m.ID = types.Int64Value(int64(*i.Id))
	return nil
}
