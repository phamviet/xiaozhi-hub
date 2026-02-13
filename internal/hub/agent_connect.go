package hub

import (
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/lxzan/gws"
	"github.com/phamviet/xiaozhi-hub/internal/hub/ws"
	"github.com/pocketbase/pocketbase/core"
)

// agentConnectRequest
type agentConnectRequest struct {
	hub        *Hub
	req        *http.Request
	res        http.ResponseWriter
	token      string
	macAddress string
	clientId   string
}

// handleAgentConnect is the HTTP handler for an agent's connection request.
func (h *Hub) handleAgentConnect(e *core.RequestEvent) error {
	agentRequest := agentConnectRequest{req: e.Request, res: e.Response, hub: h}
	_ = agentRequest.agentConnect()
	return nil
}

// agentConnect validates agent credentials and upgrades the connection to a WebSocket.
func (acr *agentConnectRequest) agentConnect() (err error) {
	header, err := acr.validateAgentHeaders(acr.req.Header)
	if err != nil {
		return acr.sendResponseError(acr.res, http.StatusBadRequest, "missing credentials")
	}

	// Validate device
	deviceID, err := acr.hub.services.Device.ValidateDevice(header.MacAddress)
	if err != nil {
		acr.hub.Logger().Warn("device validation failed", "error", err, "mac", header.MacAddress)
		return acr.sendResponseError(acr.res, http.StatusUnauthorized, "Invalid credentials or device not bound")
	}

	if header.ProtocolVersion != "1" {
		acr.hub.Logger().Warn("received unsupported protocol version", "protocolVersion", header.ProtocolVersion)
	}

	// Upgrade connection to WebSocket
	conn, err := ws.GetUpgrader().Upgrade(acr.res, acr.req)
	if err != nil {
		return acr.sendResponseError(acr.res, http.StatusInternalServerError, "WebSocket upgrade failed")
	}

	go acr.initWsConn(conn, deviceID)

	return nil
}

// initWsConn initiates the WebSocket connection
func (acr *agentConnectRequest) initWsConn(conn *gws.Conn, deviceID string) (err error) {
	sessionID, err := acr.hub.services.Session.CreateSession(deviceID)
	if err != nil {
		acr.hub.Logger().Error("failed to create session", "error", err)
		return err
	}

	logger := acr.hub.Logger().With("device", deviceID).With("sessionId", sessionID)
	client := ws.NewClient(conn, deviceID, sessionID, acr.hub.services, logger)
	wsConn := ws.NewWsConnection(conn, client)

	// must set wsConn in connection store before the read loop
	conn.Session().Store("wsConn", wsConn)

	// make sure connection is closed if there is an error
	defer func() {
		if err != nil {
			wsConn.Close([]byte(err.Error()))
		}
	}()

	// Configure TCP keepalive to prevent connection drops during idle periods
	if tcpConn, ok := conn.NetConn().(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
		tcpConn.SetNoDelay(true)
	}

	go conn.ReadLoop()

	return nil
}

type deviceInfo struct {
	Token           string
	MacAddress      string
	ClientID        string
	ProtocolVersion string
}

// validateAgentHeaders extracts and validates the token and agent version from HTTP headers.
func (acr *agentConnectRequest) validateAgentHeaders(headers http.Header) (deviceInfo, error) {
	token := headers.Get("Authorization")
	macAddress := headers.Get("Device-Id")
	clientID := headers.Get("Client-Id")
	protocolVersion := headers.Get("Protocol-Version")
	if protocolVersion == "" {
		protocolVersion = "1"
	}

	// todo validate mac address
	if macAddress == "" || clientID == "" || token == "" {
		return deviceInfo{}, errors.New("")
	}

	return deviceInfo{
		ClientID:        clientID,
		MacAddress:      macAddress,
		Token:           token,
		ProtocolVersion: protocolVersion,
	}, nil
}

// sendResponseError writes an HTTP error response.
func (acr *agentConnectRequest) sendResponseError(res http.ResponseWriter, code int, message string) error {
	res.WriteHeader(code)
	if message != "" {
		res.Write([]byte(message))
	}

	return nil
}
