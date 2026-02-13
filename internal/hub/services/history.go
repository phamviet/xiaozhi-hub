package services

import (
	"github.com/pocketbase/pocketbase/core"
)

type HistoryService interface {
	SaveMessage(sessionID string, role string, content string) error
}

type historyService struct {
	app core.App
}

func NewHistoryService(app core.App) HistoryService {
	return &historyService{app: app}
}

// SaveMessage saves a chat message to history
func (s *historyService) SaveMessage(sessionID string, role string, content string) error {
	collection, err := s.app.FindCollectionByNameOrId("ai_agent_chat_history")
	if err != nil {
		return err
	}

	record := core.NewRecord(collection)
	record.Set("chat", sessionID)
	record.Set("role", role)
	record.Set("content", content)

	return s.app.Save(record)
}
