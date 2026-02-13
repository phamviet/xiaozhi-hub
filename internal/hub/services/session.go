package services

import (
	"github.com/pocketbase/pocketbase/core"
)

type SessionService interface {
	CreateSession(deviceID string) (string, error)
}

type sessionService struct {
	app core.App
}

func NewSessionService(app core.App) SessionService {
	return &sessionService{app: app}
}

// CreateSession creates a new chat session record
func (s *sessionService) CreateSession(deviceID string) (string, error) {
	device, err := s.app.FindRecordById("ai_device", deviceID)
	if err != nil {
		return "", err
	}

	// todo: check field `chat_history_enabled`
	collection, err := s.app.FindCollectionByNameOrId("ai_agent_chat")
	if err != nil {
		return "", err
	}

	record := core.NewRecord(collection)
	record.Set("agent", device.GetString("agent"))

	if err := s.app.Save(record); err != nil {
		return "", err
	}

	return record.Id, nil
}
