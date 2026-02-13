package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	DefaultRequestTimeout = 30 * time.Second
	DefaultCloseTimeout   = 5 * time.Second
)

type WebsocketTransportConfig struct {
	SessionID     string
	RequestWriter func(v interface{}) error
}

type WebsocketTransport struct {
	sessionID      string
	notifyHandler  func(notification mcp.JSONRPCNotification)
	onCloseHandler func(reason string)

	respChans    map[string]chan *transport.JSONRPCResponse
	respChansMux sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc

	closed    bool
	closedMux sync.RWMutex

	requestTimeout time.Duration
	closeTimeout   time.Duration

	requestWriter func(v interface{}) error
}

func (t *WebsocketTransport) GetSessionId() string {
	return t.sessionID
}

func NewWebsocketTransport(config *WebsocketTransportConfig) *WebsocketTransport {
	ctx, cancel := context.WithCancel(context.Background())
	t := &WebsocketTransport{
		sessionID:      config.SessionID,
		respChans:      make(map[string]chan *transport.JSONRPCResponse),
		ctx:            ctx,
		cancel:         cancel,
		requestWriter:  config.RequestWriter,
		requestTimeout: DefaultRequestTimeout,
		closeTimeout:   DefaultCloseTimeout,
	}

	return t
}

func (t *WebsocketTransport) Start(ctx context.Context) error {
	return nil
}

func (t *WebsocketTransport) SendRequest(ctx context.Context, request transport.JSONRPCRequest) (*transport.JSONRPCResponse, error) {
	idStr := request.ID.String()
	respChan := make(chan *transport.JSONRPCResponse, 1)

	// Respect any caller-provided deadline. If none is set, apply a reasonable default
	// so pending requests don't live forever if the agent never responds.
	reqCtx := ctx
	var cancel context.CancelFunc
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		reqCtx, cancel = context.WithCancel(ctx)
	} else {
		reqCtx, cancel = context.WithTimeout(ctx, 5*time.Second)
	}

	defer cancel()
	t.respChansMux.Lock()
	t.respChans[idStr] = respChan
	t.respChansMux.Unlock()

	deleteRequest := func() {
		t.respChansMux.Lock()
		delete(t.respChans, idStr)
		t.respChansMux.Unlock()
	}

	defer deleteRequest()
	err := t.requestWriter(request)
	if err != nil {
		return nil, err
	}

	// Wait for response
	select {
	case response := <-respChan:
		return response, nil
	case <-reqCtx.Done():
		return nil, ctx.Err()
	}
}

func (t *WebsocketTransport) SendNotification(ctx context.Context, notification mcp.JSONRPCNotification) error {
	return t.requestWriter(notification)
}

func (t *WebsocketTransport) SetNotificationHandler(handler func(notification mcp.JSONRPCNotification)) {
	t.notifyHandler = handler
}

func (t *WebsocketTransport) Close() error {
	t.closedMux.Lock()
	t.closed = true
	t.closedMux.Unlock()

	if t.onCloseHandler != nil {
		t.onCloseHandler("manual_close")
	}

	t.cancel()

	// Clear response channel
	t.respChansMux.Lock()
	for idStr, respChan := range t.respChans {
		close(respChan)
		delete(t.respChans, idStr)
	}
	t.respChansMux.Unlock()

	return nil
}

func (t *WebsocketTransport) OnResponse(response *transport.JSONRPCResponse) error {
	respByte, _ := json.Marshal(response)
	idStr := response.ID.String()

	t.respChansMux.RLock()
	respChan, exists := t.respChans[idStr]
	t.respChansMux.RUnlock()

	if exists {
		select {
		case respChan <- response:
			t.respChansMux.Lock()
			delete(t.respChans, idStr)
			t.respChansMux.Unlock()
		case <-time.After(t.requestTimeout):
			return fmt.Errorf("websocket mcp timeout for ID: %s, response: %+v", idStr, string(respByte))
		}
	}

	return nil
}

func (t *WebsocketTransport) OnNotification(notification mcp.JSONRPCNotification) {
	if t.notifyHandler != nil {
		t.notifyHandler(notification)
	}
}

var _ transport.Interface = (*WebsocketTransport)(nil)
