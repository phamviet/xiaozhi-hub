package ws

import (
	"context"
	"encoding/json"
	"time"

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
		c.listenChan <- listenMsg.Text
		//c.Chat(listenMsg.Text)
	}

	return nil
}

func (c *Client) processListenChan() {
	defer c.workerWg.Done()

	// Wait until client is ready before processing
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-time.After(100 * time.Millisecond):
			c.mu.RLock()
			ready := c.ready
			c.mu.RUnlock()
			if ready {
				goto StartProcessing
			}
		}
	}

StartProcessing:

	for text := range c.listenChan {
		ctx, cancel := context.WithTimeout(c.ctx, 3*time.Minute)
		done := make(chan struct{})

		go func() {
			defer close(done)
			c.Chat(ctx, text)
		}()

		select {
		case <-done:
			cancel()
		case <-ctx.Done():
			cancel()
			c.logger.Warn("Work timeout or cancelled")
		}
	}
}
