package store

import (
	"database/sql"
	"errors"

	"github.com/phamviet/xiaozhi-hub/xiaozhi/types"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

const (
	ModelConfigCollectionName   = "model_config"
	ModelProviderCollectionName = "model_providers"
	SysParamsCollectionName     = "sys_params"
)

func (m *Manager) GetModelConfigByIDOrDefault(id string, modelType string) (*types.ModelConfig, error) {
	if id == "" {
		return m.GetDefaultModelConfig(modelType)
	}

	var row types.ModelConfig
	err := m.App.RecordQuery(ModelConfigCollectionName).
		Where(dbx.NewExp("id = {:id}", dbx.Params{"id": id})).
		One(&row)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return m.GetDefaultModelConfig(modelType)
		}
		return nil, err
	}

	if row.IsEnabled {
		return &row, nil
	}

	m.App.Logger().Warn("This model_config is disabled. Falling back to default one", "id", id, "name", row.ModelName)

	return m.GetDefaultModelConfig(modelType)
}

func (m *Manager) GetDefaultModelConfig(modelType string) (*types.ModelConfig, error) {
	var row types.ModelConfig
	err := m.App.RecordQuery(ModelConfigCollectionName).
		AndWhere(dbx.NewExp("model_type = {:t}", dbx.Params{"t": modelType})).
		AndWhere(dbx.NewExp("is_enabled = {:b}", dbx.Params{"b": true})).
		AndWhere(dbx.NewExp("is_default = {:2}", dbx.Params{"2": true})).
		One(&row)

	if err != nil {
		return nil, err
	}

	return &row, nil
}

func (m *Manager) GetProviderCodeByID(providerID string) (string, error) {
	record, err := m.App.FindRecordById(ModelProviderCollectionName, providerID)
	if err != nil {
		return "", err
	}
	return record.GetString("provider_code"), nil
}

func (m *Manager) GetSysParam(name string) (string, error) {
	record, err := m.App.FindFirstRecordByData(SysParamsCollectionName, "name", name)
	if err != nil {
		return "", err
	}
	return record.GetString("value"), nil
}

func (m *Manager) GetSysParams(names ...string) (map[string]*string, error) {
	var params []types.SysParam
	interfaceSlice := make([]interface{}, len(names))
	for i, v := range names {
		interfaceSlice[i] = v
	}
	err := m.App.RecordQuery(SysParamsCollectionName).Where(dbx.In("name", interfaceSlice...)).All(&params)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*string, len(names))
	for _, param := range params {
		v := param.Value
		result[param.Name] = &v
	}

	for _, key := range names {
		if _, ok := result[key]; !ok {
			result[key] = nil
		}
	}

	return result, nil
}

func (m *Manager) GetActiveSysConfig() (*core.Record, error) {
	return m.App.FindFirstRecordByData("sys_config", "disabled", false)
}

func (m *Manager) ResolveSecretReference(modelConfig *types.ModelConfigJson) error {
	credentialID, ok := modelConfig.Param["secret_ref"]
	if !ok {
		return nil
	}

	cred, err := m.App.FindRecordById("user_credentials", credentialID)
	if err != nil {
		return err
	}

	modelConfig.Param["api_key"] = cred.GetString("api_key")
	delete(modelConfig.Param, "secret_ref")
	return nil
}
