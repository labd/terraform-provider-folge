package checkjsonproperty

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-folge/internal/folge"
)

type CheckJsonPropertyModel struct {
	ID            types.Int64  `tfsdk:"id"`
	ApplicationID types.Int64  `tfsdk:"application_id"`
	DataSourceID  types.Int64  `tfsdk:"datasource_id"`
	Name          types.String `tfsdk:"name"`
	Crontab       types.String `tfsdk:"crontab"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Path          types.String `tfsdk:"path"`
	DataType      types.String `tfsdk:"datatype"`
	Operator      types.String `tfsdk:"operator"`
	ValueBoolean  types.Bool   `tfsdk:"value_bool"`
	ValueString   types.String `tfsdk:"value_string"`
	ValueInt      types.Int64  `tfsdk:"value_int"`
	ValueDateTime types.String `tfsdk:"value_datetime"`
}

func (m *CheckJsonPropertyModel) toCreateInput() folge.ApplicationsDataSourcesChecksCreateJSONRequestBody {
	req := m.createRequest()
	return folge.ApplicationsDataSourcesChecksCreateJSONRequestBody(req)
}

func (m *CheckJsonPropertyModel) toUpdateInput() folge.ApplicationsDataSourcesChecksUpdateJSONRequestBody {
	req := m.createRequest()
	return folge.ApplicationsDataSourcesChecksUpdateJSONRequestBody(req)
}

func (m *CheckJsonPropertyModel) createRequest() folge.Check {
	data := folge.JsonDataCheckTyped{
		Enabled:  m.Enabled.ValueBoolPointer(),
		Label:    m.Name.ValueString(),
		Datatype: folge.DatatypeEnum(m.DataType.ValueString()),
		Operator: folge.OperatorEnum(m.Operator.ValueString()),
		Path:     m.Path.ValueString(),
	}

	switch data.Datatype {
	case folge.Bool:
		val, err := json.Marshal(m.ValueBoolean.ValueBool())
		if err != nil {
			panic(err)
		}
		data.Value = string(val)
	case folge.Int:
		val, err := json.Marshal(m.ValueInt.ValueInt64())
		if err != nil {
			panic(err)
		}
		data.Value = string(val)
	case folge.Str:
		val, err := json.Marshal(m.ValueString.ValueString())
		if err != nil {
			panic(err)
		}
		data.Value = string(val)
	case folge.Datetime:
		val, err := json.Marshal(m.ValueDateTime.ValueString())
		if err != nil {
			panic(err)
		}
		data.Value = string(val)
	}

	req := folge.Check{}
	if err := req.FromJsonDataCheckTyped(data); err != nil {
		panic(err)
	}
	return req
}

func (m *CheckJsonPropertyModel) fromRemote(i folge.Check, applicationId int, datasourceId int) error {
	t, err := i.ValueByDiscriminator()
	if err != nil {
		return err
	}

	switch d := t.(type) {
	case folge.JsonDataCheckTyped:
		m.ID = types.Int64Value(int64(*d.Id))
		m.Name = types.StringValue(d.Label)
		m.ApplicationID = types.Int64Value(int64(applicationId))
		m.DataSourceID = types.Int64Value(int64(datasourceId))
		m.Enabled = types.BoolPointerValue(d.Enabled)

		m.Operator = types.StringValue(string(d.Operator))
		m.Path = types.StringValue(d.Path)
		m.DataType = types.StringValue(string(d.Datatype))

		switch d.Datatype {
		case folge.Bool:
			var dst bool
			err := json.Unmarshal([]byte(d.Value), &dst)
			if err != nil {
				return fmt.Errorf("Invalid bool value: %s", err)
			}
			m.ValueBoolean = types.BoolValue(dst)

		case folge.Int:
			var dst int64
			err := json.Unmarshal([]byte(d.Value), &dst)
			if err != nil {
				return fmt.Errorf("Invalid int value: %s", err)
			}
			m.ValueInt = types.Int64Value(dst)

		case folge.Str:
			var dst string
			err := json.Unmarshal([]byte(d.Value), &dst)
			if err != nil {
				return fmt.Errorf("Invalid datetime value: %s", err)
			}
			m.ValueString = types.StringValue(dst)

		case folge.Datetime:
			var dst string
			err := json.Unmarshal([]byte(d.Value), &dst)
			if err != nil {
				return fmt.Errorf("Invalid datetime value: %s", err)
			}
			m.ValueDateTime = types.StringValue(dst)

		default:
			return fmt.Errorf("unknown datatype: %s", d.Datatype)
		}

	default:
		return errors.New("unknown data source type")
	}
	return nil
}
