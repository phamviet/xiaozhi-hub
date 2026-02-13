package ws

import (
	"fmt"
	"log"
	"time"
	"weak"

	"github.com/lxzan/gws"
	"github.com/phamviet/xiaozhi-hub/internal/hub/ws/handlers"
	"github.com/phamviet/xiaozhi-hub/internal/hub/ws/types"
)

const (
	ReadTimeout          = 30 * time.Second
	PingInterval         = 5 * time.Second
	HeartbeatWaitTimeout = 10 * time.Second
)

// Handler implements the WebSocket event handler for agent connections.
type Handler struct {
	gws.BuiltinEventHandler
}

// WsConn represents a WebSocket connection to an agent.
type WsConn struct {
	conn     *gws.Conn
	Client   *Client
	DownChan chan struct{}

	deviceID string
}

var (
	upgrader   *gws.Upgrader
	dispatcher *Dispatcher
)

func init() {
	dispatcher = NewDispatcher()
	// Register handlers
	dispatcher.Register(types.MessageTypeHello, &handlers.HelloHandler{})
	dispatcher.Register(types.MessageTypeIoT, &handlers.IoTHandler{})
	dispatcher.Register(types.MessageTypeListen, &handlers.ListenHandler{})
}

// GetUpgrader returns a singleton WebSocket upgrader instance.
func GetUpgrader() *gws.Upgrader {
	if upgrader != nil {
		return upgrader
	}
	handler := &Handler{}
	upgrader = gws.NewUpgrader(handler, &gws.ServerOption{
		ParallelEnabled: true,
	})

	return upgrader
}

// NewWsConnection creates a new WebSocket connection wrapper.
func NewWsConnection(conn *gws.Conn, client *Client) *WsConn {
	return &WsConn{
		conn:     conn,
		Client:   client,
		DownChan: make(chan struct{}, 1),
	}
}

func (c *Handler) OnOpen(conn *gws.Conn) {
	conn.SetDeadline(time.Now().Add(300 * time.Second))
}

func (c *Handler) OnPing(conn *gws.Conn, payload []byte) {
	log.Println("ws.OnPing")
	conn.SetDeadline(time.Now().Add(300 * time.Second))
	_ = conn.WritePong(payload)
}

// OnMessage routes incoming WebSocket messages to the client.
func (h *Handler) OnMessage(conn *gws.Conn, message *gws.Message) {
	defer message.Close()

	if message.Data.Len() == 0 {
		return
	}

	wsConn, ok := conn.Session().Load("wsConn")
	if !ok {
		_ = conn.WriteClose(1000, nil)
		return
	}
	client := wsConn.(*WsConn).Client

	defer message.Close()
	if message.Opcode == gws.OpcodeText {
		client.OnTextMessage(message)
		return
	}

	if message.Opcode == gws.OpcodeBinary {
		if err := client.OnBinaryMessage(message.Bytes()); err != nil {
			// Handle error
		}
	}
}

// OnClose handles WebSocket connection closures and triggers system down status after delay.
func (h *Handler) OnClose(conn *gws.Conn, err error) {
	log.Println(fmt.Sprintf("WebSocket closing with error: %v", err))
	wsConn, ok := conn.Session().Load("wsConn")
	if !ok {
		return
	}
	connWrapper := wsConn.(*WsConn)
	connWrapper.conn = nil
	if connWrapper.Client != nil {
		connWrapper.Client.Close()
	}

	// wait 5 seconds to allow reconnection before setting system down
	// use a weak pointer to avoid keeping references if the system is removed
	go func(downChan weak.Pointer[chan struct{}]) {
		time.Sleep(5 * time.Second)
		downChanValue := downChan.Value()
		if downChanValue != nil {
			// Check if channel is closed or full before sending to avoid panic/blocking
			select {
			case *downChanValue <- struct{}{}:
			default:
			}
		}
	}(weak.Make(&connWrapper.DownChan))
}

// Close terminates the WebSocket connection gracefully.
func (ws *WsConn) Close(msg []byte) {
	if ws.IsConnected() {
		_ = ws.conn.WriteClose(1000, msg)
	}

	if ws.Client != nil {
		ws.Client.Close()
	}
}

// IsConnected returns true if the WebSocket connection is active.
func (ws *WsConn) IsConnected() bool {
	return ws.conn != nil
}
