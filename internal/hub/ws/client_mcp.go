package ws

import (
	"encoding/json"

	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/phamviet/xiaozhi-hub/internal/hub/ws/types"
)

func (c *Client) sendMcpMessage(v interface{}) error {
	payload, err := json.Marshal(v)
	if err != nil {
		return err
	}

	message := types.MCPMessage{
		BaseMessage: types.BaseMessage{
			Type:      types.MessageTypeMCP,
			SessionID: c.sessionID,
		},
		Payload: payload,
	}

	return c.SendJSON(message)
}

func (c *Client) handleMcpMessage(msg []byte) error {
	var mcpMsg types.MCPMessage
	if err := json.Unmarshal(msg, &mcpMsg); err != nil {
		return err
	}

	c.Logger().Debug("MCP message received", "payload", mcpMsg.Payload)

	var response transport.JSONRPCResponse
	if err := json.Unmarshal(mcpMsg.Payload, &response); err == nil {
		return c.handleMcpResponse(&response)
	}

	var notification mcp.JSONRPCNotification
	if err := json.Unmarshal(mcpMsg.Payload, &notification); err == nil && notification.Method != "" {
		c.handleMcpNotification(notification)
	}

	return nil
}

func (c *Client) handleMcpResponse(response *transport.JSONRPCResponse) error {
	return c.clientMcpTransport.OnResponse(response)
}

func (c *Client) handleMcpNotification(notification mcp.JSONRPCNotification) {
	c.clientMcpTransport.OnNotification(notification)
}
