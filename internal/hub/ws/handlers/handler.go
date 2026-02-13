package handlers

import (
	"encoding/json"
	"log/slog"

	"github.com/phamviet/xiaozhi-hub/internal/hub/services"
)

// MessageHandler is the interface that all message handlers must implement
type MessageHandler interface {
	Handle(ctx Context, msg []byte) error
}

// Context defines the methods available to handlers
type Context interface {
	SendJSON(v interface{}) error
	DeviceID() string
	SessionID() string
	ProtocolVersion() int
	Services() *services.ServiceContainer
	Logger() *slog.Logger
}

// BaseHandler provides common functionality
type BaseHandler struct{}

func (h *BaseHandler) Parse(msg []byte, v interface{}) error {
	return json.Unmarshal(msg, v)
}
