package xiaozhi

import (
	"github.com/pocketbase/dbx"
)

func (m *Manager) getSysParams(names ...string) (map[string]*string, error) {
	var params []SysParam
	interfaceSlice := make([]interface{}, len(names))
	for i, v := range names {
		interfaceSlice[i] = v
	}
	err := m.App.RecordQuery("sys_params").Where(dbx.In("name", interfaceSlice...)).All(&params)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*string, len(names))
	for _, param := range params {
		result[param.Name] = &param.Value
	}

	for _, key := range names {
		if _, ok := result[key]; !ok {
			result[key] = nil
		}
	}

	return result, nil
}
