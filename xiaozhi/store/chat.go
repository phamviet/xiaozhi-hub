package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/phamviet/xiaozhi-hub/xiaozhi/types"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

const (
	ChatCollectionName        = "ai_agent_chat"
	ChatHistoryCollectionName = "ai_agent_chat_history"
)

func (m *Manager) LoadChatSession(sessionID, agentID string) (*types.ChatSession, error) {
	var session types.ChatSession
	err := m.App.RecordQuery(ChatCollectionName).Where(dbx.NewExp("id = {:id}", dbx.Params{"id": sessionID})).One(&session)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) && agentID != "" {
			collection, err := m.App.FindCollectionByNameOrId(ChatCollectionName)
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

func (m *Manager) UpdateChatSessionSummary(sessionID string, summary string) error {
	record, err := m.App.FindRecordById(ChatCollectionName, sessionID)
	if err != nil {
		return err
	}
	record.Set("summary", summary)
	return m.App.Save(record)
}

func (m *Manager) FetchChatHistory(chatID string) ([]types.ChatMessage, error) {
	var chatHistory []types.ChatMessage
	err := m.App.DB().Select("*").
		From(ChatHistoryCollectionName).
		Where(dbx.HashExp{"chat": chatID}).
		OrderBy("created ASC").
		All(&chatHistory)
	return chatHistory, err
}

type ChatHistoryParams struct {
	ChatID      string
	DeviceID    string
	Content     string
	ChatType    string
	AudioBytes  []byte
	ReportTime  int64
	AudioFormat string // "mp3"
}

func (m *Manager) SaveChatHistory(params ChatHistoryParams) error {
	collection, err := m.App.FindCollectionByNameOrId(ChatHistoryCollectionName)
	if err != nil {
		return err
	}

	record := core.NewRecord(collection)
	record.Set("chat", params.ChatID)
	record.Set("device", params.DeviceID)
	record.Set("content", params.Content)
	record.Set("chat_type", params.ChatType)

	if len(params.AudioBytes) > 0 {
		filename := fmt.Sprintf("audio_%d_%d.%s", time.Now().UnixNano(), params.ReportTime, params.AudioFormat)
		file, err := filesystem.NewFileFromBytes(params.AudioBytes, filename)
		if err != nil {
			return err
		}
		record.Set("chat_audio", file)
	}

	return m.App.Save(record)
}
