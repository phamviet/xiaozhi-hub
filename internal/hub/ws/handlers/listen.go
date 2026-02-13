package handlers

import (
	"github.com/phamviet/xiaozhi-hub/internal/hub/ws/types"
)

type ListenHandler struct {
	BaseHandler
}

func (h *ListenHandler) Handle(ctx Context, msg []byte) error {
	var listenMsg types.ListenMessage
	if err := h.Parse(msg, &listenMsg); err != nil {
		return err
	}

	ctx.Logger().Info("Listen state change", "state", listenMsg.State, "mode", listenMsg.Mode, "session_id", ctx.SessionID())

	// If text is provided (wake word detected), we might want to trigger STT/LLM pipeline immediately
	if listenMsg.Text != "" {
		ctx.Logger().Info("Wake word detected", "text", listenMsg.Text)

		// Save user message to history
		if err := ctx.Services().History.SaveMessage(ctx.SessionID(), "user", listenMsg.Text); err != nil {
			ctx.Logger().Error("Failed to save message history", "error", err)
		}

	}

	return nil
}
