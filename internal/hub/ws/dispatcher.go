package ws

import (
	"errors"
	"sync"

	"github.com/phamviet/xiaozhi-hub/internal/hub/ws/handlers"
	"github.com/phamviet/xiaozhi-hub/internal/hub/ws/types"
)

type Dispatcher struct {
	handlers map[types.MessageType]handlers.MessageHandler
	mu       sync.RWMutex
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[types.MessageType]handlers.MessageHandler),
	}
}

func (d *Dispatcher) Register(msgType types.MessageType, handler handlers.MessageHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[msgType] = handler
}

func (d *Dispatcher) GetHandler(msgType types.MessageType) (handlers.MessageHandler, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	handler, ok := d.handlers[msgType]
	if !ok {
		return nil, errors.New("handler not found")
	}
	return handler, nil
}
