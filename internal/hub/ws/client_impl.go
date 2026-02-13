package ws

import (
	"log/slog"

	"github.com/phamviet/xiaozhi-hub/internal/hub/services"
)

func (c *Client) Services() *services.ServiceContainer {
	return c.services
}

func (c *Client) Logger() *slog.Logger {
	return c.logger
}

func (c *Client) DeviceID() string {
	return c.deviceID
}

func (c *Client) SessionID() string {
	return c.sessionID
}

func (c *Client) ProtocolVersion() int {
	return c.ClientVersion
}
