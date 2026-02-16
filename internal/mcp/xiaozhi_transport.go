package mcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/phamviet/xiaozhi-hub/internal/hub/ws/types"
)

type XiaozhiTransport struct {
	sessionID string
	incoming  chan []byte
	sender    func(v interface{}) error
	done      chan struct{}
}

func NewXiaozhiTransport(sessionID string, sender func(v interface{}) error) *XiaozhiTransport {
	return &XiaozhiTransport{
		sessionID: sessionID,
		sender:    sender,
		incoming:  make(chan []byte, 100),
	}
}

type connection struct {
	t      *XiaozhiTransport
	mu     sync.Mutex
	closed bool // set when the stream is closed
}

func (c *connection) Read(ctx context.Context) (jsonrpc.Message, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case <-c.t.done:
		return nil, io.EOF

	case data := <-c.t.incoming:
		msg, err := jsonrpc.DecodeMessage(data)
		if err != nil {
			return nil, err
		}

		return msg, nil
	}
}

func (c *connection) Write(ctx context.Context, msg jsonrpc.Message) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.t.done:
		return io.EOF
	default:
	}

	data, err := jsonrpc.EncodeMessage(msg)
	if err != nil {
		return fmt.Errorf("marshaling message: %v", err)
	}
	mcpMessage := types.MCPMessage{
		BaseMessage: types.BaseMessage{
			Type:      types.MessageTypeMCP,
			SessionID: c.t.sessionID,
		},
		Payload: data,
	}

	return c.t.sender(mcpMessage)
}

func (c *connection) SessionID() string {
	return c.t.sessionID
}

func (c *connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed {
		c.closed = true
		return c.t.Close()
	}

	return nil
}

func (c *connection) isDone() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}

func (t *XiaozhiTransport) Connect(ctx context.Context) (mcp.Connection, error) {
	t.done = make(chan struct{}, 1)
	return &connection{t: t}, nil
}

func (t *XiaozhiTransport) Receive(msg []byte) error {
	select {
	case <-t.done:
		return io.EOF
	case t.incoming <- msg:
		return nil
	default:
		return errors.New("incoming channel is full")
	}
}

func (t *XiaozhiTransport) Close() error {
	close(t.done)
	return nil
}
