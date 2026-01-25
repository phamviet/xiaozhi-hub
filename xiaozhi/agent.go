package xiaozhi

import (
	"database/sql"
	"errors"

	"github.com/phamviet/xiaozhi-hub/xiaozhi/types"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func (m *Manager) getAgentByID(id string) (*types.AIAgent, error) {
	var agent types.AIAgent
	err := m.App.RecordQuery("ai_agent").Where(dbx.NewExp("id = {:p}", dbx.Params{"p": id})).One(&agent)
	if err != nil {
		return nil, err
	}

	return &agent, nil
}

func (m *Manager) createNewAgent(userID, agentName string) (*types.AIAgent, error) {
	collection, err := m.App.FindCollectionByNameOrId("ai_agent")
	if err != nil {
		return nil, err
	}

	record := core.NewRecord(collection)
	record.Set("user", userID)
	record.Set("agent_name", agentName)
	record.Set("role_prompt", "You are a helpful AI assistant.")
	record.Set("lang_code", "en")
	record.Set("chat_history_enabled", true)

	if err := m.App.Save(record); err != nil {
		return nil, err
	}

	return &types.AIAgent{
		ID:                 record.Id,
		UserID:             userID,
		Name:               agentName,
		RolePrompt:         "You are a helpful AI assistant.",
		LangCode:           "en",
		ChatHistoryEnabled: true,
	}, nil
}

func (m *Manager) loadChatSession(sessionID, agentID string) (*types.ChatSession, error) {
	var session types.ChatSession
	err := m.App.RecordQuery("ai_agent_chat").Where(dbx.NewExp("id = {:id}", dbx.Params{"id": sessionID})).One(&session)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) && agentID != "" {
			collection, err := m.App.FindCollectionByNameOrId("ai_agent_chat")
			if err != nil {
				return nil, err
			}

			// Create a new session
			record := core.NewRecord(collection)
			record.Set("id", sessionID)
			record.Set("agent", agentID)

			if err = m.App.Save(record); err != nil {
				return nil, err
			}

			return &types.ChatSession{
				ID:      sessionID,
				AgentID: agentID,
			}, nil
		}

		return nil, err
	}

	return &session, nil
}

func (m *Manager) getDeviceByMacAddress(mac string) (*types.Device, error) {
	var row types.Device
	err := m.App.RecordQuery("ai_device").Where(dbx.NewExp("mac_address = {:p}", dbx.Params{"p": mac})).One(&row)
	if err != nil {
		return nil, err
	}

	return &row, nil
}

func (m *Manager) getModelConfigByIDOrDefault(id string, modelType string) (*types.ModelConfig, error) {
	if id == "" {
		return m.getDefaultModeConfig(modelType)
	}

	var row types.ModelConfig
	err := m.App.RecordQuery("model_config").
		Where(dbx.NewExp("id = {:id}", dbx.Params{"id": id})).
		One(&row)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return m.getDefaultModeConfig(modelType)
		}
		return nil, err
	}

	if row.IsEnabled {
		return &row, nil
	}

	m.App.Logger().Warn("This model_config is disabled. Falling back to default one", "id", id, "name", row.ModelName)

	return m.getDefaultModeConfig(modelType)
}

func (m *Manager) getDefaultModeConfig(modelType string) (*types.ModelConfig, error) {
	var row types.ModelConfig
	err := m.App.RecordQuery("model_config").
		AndWhere(dbx.NewExp("model_type = {:t}", dbx.Params{"t": modelType})).
		AndWhere(dbx.NewExp("is_enabled = {:b}", dbx.Params{"b": true})).
		AndWhere(dbx.NewExp("is_default = {:2}", dbx.Params{"2": true})).
		One(&row)

	if err != nil {
		return nil, err
	}

	return &row, nil
}
