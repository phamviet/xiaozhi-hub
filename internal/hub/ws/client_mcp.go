package ws

import (
	"encoding/json"

	"github.com/phamviet/xiaozhi-hub/internal/hub/ws/types"
)

func (c *Client) handleMcpMessage(msg []byte) error {
	var mcpMsg types.MCPMessage
	if err := json.Unmarshal(msg, &mcpMsg); err != nil {
		return err
	}

	return c.mcpTransport.Receive(mcpMsg.Payload)
}
