package store

import (
	"github.com/phamviet/xiaozhi-hub/xiaozhi/types"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

const AgentCollectionName = "ai_agent"

func (m *Manager) GetAgentByID(id string) (*types.AIAgent, error) {
	var agent types.AIAgent
	err := m.App.RecordQuery(AgentCollectionName).Where(dbx.NewExp("id = {:p}", dbx.Params{"p": id})).One(&agent)
	if err != nil {
		return nil, err
	}

	return &agent, nil
}

func (m *Manager) CreateNewAgent(userID, agentName string) (*types.AIAgent, error) {
	collection, err := m.App.FindCollectionByNameOrId(AgentCollectionName)
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

func (m *Manager) UpdateAgentMemory(agentID string, summary string) error {
	agentRecord, err := m.App.FindRecordById(AgentCollectionName, agentID)
	if err != nil {
		return err
	}

	agentRecord.Set("summary_memory", summary)
	return m.App.Save(agentRecord)
}
