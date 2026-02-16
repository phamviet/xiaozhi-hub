package ws

import (
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/phamviet/xiaozhi-hub/internal/hub/ws/types"
)

func (c *Client) handleMcpMessage(msg []byte) error {
	var mcpMsg types.MCPMessage
	if err := json.Unmarshal(msg, &mcpMsg); err != nil {
		return err
	}

	c.Logger().Debug("MCP message received", "payload", mcpMsg.Payload)

	if c.mcpTransport != nil {
		err := c.mcpTransport.Receive(mcpMsg.Payload)
		c.logger.Error("clientMcpWrite", "error", err)
		return err
	}

	return nil
}

func (c *Client) initMcpClientSession() *mcp.ClientSession {
	client := mcp.NewClient(&mcp.Implementation{Name: "mcp-client", Version: "v0.0.1"}, nil)
	cs, err := client.Connect(c.ctx, c.mcpTransport, nil)
	if err != nil {
		c.logger.Error("initMcpClientSession", "error", err)
		return nil
	}

	return cs
}
