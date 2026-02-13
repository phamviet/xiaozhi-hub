package ws

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
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

	ctx := context.Background()
	if err := c.mcpClient.Start(ctx); err != nil {
		return err
	}

	initRequest := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    "mcp-go",
				Version: "0.1.0",
			},
			Capabilities: mcp.ClientCapabilities{},
		},
	}

	initResult, err := c.mcpClient.Initialize(ctx, initRequest)
	if err != nil {
		c.logger.Error("initialize mcp failed", "error", err)
		return err
	}

	c.logger.Info("MCP initialized", "result", initResult)

	return nil
}
