package xiaozhi

import (
	"database/sql"
	"errors"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func (m *Manager) getAgentByID(id string) (*AIAgent, error) {
	var agent AIAgent
	err := m.App.RecordQuery("ai_agent").Where(dbx.NewExp("id = {:p}", dbx.Params{"p": id})).One(&agent)
	if err != nil {
		return nil, err
	}

	return &agent, nil
}

func (m *Manager) loadChatSession(sessionID, agentID string) (*ChatSession, error) {
	collection, err := m.App.FindCollectionByNameOrId("ai_agent_chat")
	if err != nil {
		return nil, err
	}

	var session ChatSession
	err = m.App.RecordQuery("ai_agent_chat").Where(dbx.NewExp("id = {:id}", dbx.Params{"id": sessionID})).One(&session)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) && agentID != "" {
			// Create a new session
			record := core.NewRecord(collection)
			record.Set("id", sessionID)
			record.Set("agent", agentID)

			if err = m.App.Save(record); err != nil {
				return nil, err
			}

			return &ChatSession{
				ID:      sessionID,
				AgentID: agentID,
			}, nil
		}

		return nil, err
	}

	return &session, nil
}

func (m *Manager) getDeviceByMacAddress(mac string) (*Device, error) {
	var row Device
	err := m.App.RecordQuery("ai_device").Where(dbx.NewExp("mac_address = {:p}", dbx.Params{"p": mac})).One(&row)
	if err != nil {
		return nil, err
	}

	return &row, nil
}

func (m *Manager) getModelConfigByIDOrDefault(id string, modelType string) (*ModelConfig, error) {
	if id == "" {
		return m.getDefaultModeConfig(modelType)
	}

	var row ModelConfig
	err := m.App.RecordQuery("model_config").
		Where(dbx.NewExp("id = {:id}", dbx.Params{"id": id})).
		One(&row)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return m.getDefaultModeConfig(modelType)
		}
		return nil, err
	}

	if &row != nil {
		if row.IsEnabled {
			return &row, nil
		}

		m.App.Logger().Warn("This model_config is disabled. Falling back to default one", "id", id, "name", row.ModelName)
	}

	return m.getDefaultModeConfig(modelType)
}

func (m *Manager) getDefaultModeConfig(modelType string) (*ModelConfig, error) {
	var row ModelConfig
	err := m.App.RecordQuery("model_config").
		AndWhere(dbx.NewExp("model_type = {:t}", dbx.Params{"t": modelType})).
		AndWhere(dbx.NewExp("is_enabled = {:b}", dbx.Params{"b": true})).
		AndWhere(dbx.NewExp("is_default = {:2}", dbx.Params{"2": true})).
		One(&row)

	if &row != nil {
		return &row, nil
	}

	return nil, err
}
