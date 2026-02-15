package ws

import (
	"encoding/json"

	"github.com/phamviet/xiaozhi-hub/internal/hub/ws/types"
)

func (c *Client) handleListenMessage(msg []byte) error {
	var listenMsg types.ListenMessage
	if err := json.Unmarshal(msg, &listenMsg); err != nil {
		return err
	}

	c.Logger().Info("Listen state change", "state", listenMsg.State, "mode", listenMsg.Mode)

	if listenMsg.Mode == "auto" {
		if listenMsg.State == "start" {
			c.asr.Start()
			c.logger.Debug("Start audio streaming from client")
		}
	}

	// If text is provided (wake word detected), we might want to trigger STT/LLM pipeline immediately
	if listenMsg.Text != "" {
		c.Logger().Info("Wake word detected", "text", listenMsg.Text)

		// Save user message to history
		//if err := c.Services().History.SaveMessage(c.SessionID(), "user", listenMsg.Text); err != nil {
		//	c.Logger().Error("Failed to save message history", "error", err)
		//}

		c.Chat(listenMsg.Text)
	}

	return nil
}
