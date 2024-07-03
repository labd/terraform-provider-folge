package datasource

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-folge/internal/folge"
)

type DataSourceModel struct {
	ID            types.Int64  `tfsdk:"id"`
	ApplicationID types.Int64  `tfsdk:"application_id"`
	Name          types.String `tfsdk:"name"`
	URL           types.String `tfsdk:"url"`

	BasicAuth *BasicAuthModel `tfsdk:"basic_auth"`
}

type BasicAuthModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (m *DataSourceModel) toCreateInput() folge.ApplicationsDataSourcesCreateJSONRequestBody {
	req := m.createRequest()
	return folge.ApplicationsDataSourcesCreateJSONRequestBody(req)
}

func (m *DataSourceModel) toUpdateInput() folge.ApplicationsDataSourcesUpdateJSONRequestBody {
	req := m.createRequest()
	return folge.ApplicationsDataSourcesUpdateJSONRequestBody(req)
}

func (m *DataSourceModel) createRequest() folge.DataSourceRequest {
	data := folge.HttpDataSourceTypedRequest{
		Label: m.Name.ValueStringPointer(),
		Url:   m.URL.ValueString(),
	}

	if m.BasicAuth != nil {
		data.BasicAuthUsername = m.BasicAuth.Username.ValueStringPointer()
		data.BasicAuthPassword = m.BasicAuth.Password.ValueStringPointer()
	}

	req := folge.DataSourceRequest{}
	if err := req.FromHttpDataSourceTypedRequest(data); err != nil {
		panic(err)
	}
	return req
}

func (m *DataSourceModel) fromRemote(i folge.DataSource, applicationId int) error {
	t, err := i.ValueByDiscriminator()
	if err != nil {
		return err
	}

	switch d := t.(type) {
	case folge.HttpDataSourceTyped:
		m.ID = types.Int64Value(int64(*d.Id))
		m.Name = types.StringPointerValue(d.Label)
		m.URL = types.StringValue(d.Url)
		m.ApplicationID = types.Int64Value(int64(applicationId))

		if (d.BasicAuthUsername != nil) || (d.BasicAuthPassword != nil) {
			m.BasicAuth = &BasicAuthModel{
				Username: types.StringPointerValue(d.BasicAuthUsername),
				Password: types.StringPointerValue(d.BasicAuthPassword),
			}
		}
	default:
		return errors.New("unknown data source type")
	}
	return nil
}
