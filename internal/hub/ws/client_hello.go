package ws

import (
	"encoding/json"
	"log/slog"

	"github.com/phamviet/xiaozhi-hub/internal/hub/ws/types"
)

func (c *Client) handleHelloMessage(data []byte) error {
	var helloMsg types.HelloMessage
	if err := json.Unmarshal(data, &helloMsg); err != nil {
		return err
	}

	c.mu.Lock()
	c.ClientAudioFormat = helloMsg.AudioParams.Format
	c.ClientVersion = helloMsg.Version
	c.ClientSampleRate = helloMsg.AudioParams.SampleRate
	c.ClientChannels = helloMsg.AudioParams.Channels
	c.ClientFrameDuration = helloMsg.AudioParams.FrameDuration
	c.mu.Unlock()

	c.logger.Info("Client hello", slog.Any("message", helloMsg))
	if c.ClientAudioFormat != "opus" {
		c.logger.Warn("unsupported audio format", "audioFormat", helloMsg.AudioParams.Format)
	}

	response := types.HelloMessage{
		BaseMessage: types.BaseMessage{
			Type:      types.MessageTypeHello,
			SessionID: c.SessionID(),
		},
		Transport: "websocket",
		AudioParams: types.AudioParams{
			Format:        "opus",
			SampleRate:    16000,
			Channels:      1,
			FrameDuration: 60,
		},
	}

	if err := c.SendJSON(response); err != nil {
		return err
	}

	go c.initializeAgent(nil)

	return nil
}
