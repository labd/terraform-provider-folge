package checkhttpstatus

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-folge/internal/folge"
)

type CheckHttpStatusModel struct {
	ID            types.Int64  `tfsdk:"id"`
	ApplicationID types.Int64  `tfsdk:"application_id"`
	DataSourceID  types.Int64  `tfsdk:"datasource_id"`
	Name          types.String `tfsdk:"name"`
	StatusCode    types.Int64  `tfsdk:"status_code"`
	Crontab       types.String `tfsdk:"crontab"`
	Enabled       types.Bool   `tfsdk:"enabled"`
}

func (m *CheckHttpStatusModel) toCreateInput() folge.ApplicationsDataSourcesChecksCreateJSONRequestBody {
	req := m.createRequest()
	return folge.ApplicationsDataSourcesChecksCreateJSONRequestBody(req)
}

func (m *CheckHttpStatusModel) toUpdateInput() folge.ApplicationsDataSourcesChecksUpdateJSONRequestBody {
	req := m.createRequest()
	return folge.ApplicationsDataSourcesChecksUpdateJSONRequestBody(req)
}

func (m *CheckHttpStatusModel) createRequest() folge.Check {
	data := folge.HttpStatusCheckTyped{
		Enabled:    m.Enabled.ValueBoolPointer(),
		Label:      m.Name.ValueString(),
		StatusCode: int(m.StatusCode.ValueInt64()),
	}

	req := folge.Check{}
	if err := req.FromHttpStatusCheckTyped(data); err != nil {
		panic(err)
	}
	return req
}

func (m *CheckHttpStatusModel) fromRemote(i folge.Check, applicationId int, datasourceId int) error {
	t, err := i.ValueByDiscriminator()
	if err != nil {
		return err
	}

	switch d := t.(type) {
	case folge.HttpStatusCheckTyped:
		m.ID = types.Int64Value(int64(*d.Id))
		m.Name = types.StringValue(d.Label)
		m.ApplicationID = types.Int64Value(int64(applicationId))
		m.DataSourceID = types.Int64Value(int64(datasourceId))
		m.Enabled = types.BoolPointerValue(d.Enabled)
		m.StatusCode = types.Int64Value(int64(d.StatusCode))

	default:
		return errors.New("unknown data source type")
	}
	return nil
}
