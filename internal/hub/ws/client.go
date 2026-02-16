package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/core/x/session"
	"github.com/firebase/genkit/go/genkit"
	"github.com/lxzan/gws"
	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/phamviet/xiaozhi-hub/internal/asr"
	"github.com/phamviet/xiaozhi-hub/internal/audio"
	"github.com/phamviet/xiaozhi-hub/internal/hub/services"
	"github.com/phamviet/xiaozhi-hub/internal/hub/ws/handlers"
	"github.com/phamviet/xiaozhi-hub/internal/hub/ws/types"
	"github.com/phamviet/xiaozhi-hub/internal/mcp"
	"github.com/phamviet/xiaozhi-hub/internal/tts"
)

// Client represents a connected agent/device
type Client struct {
	ctx    context.Context
	cancel context.CancelFunc

	mu         sync.RWMutex
	conn       *gws.Conn
	g          *genkit.Genkit
	tools      []ai.ToolRef
	chatFlow   *core.Flow[string, string, struct{}]
	ttsFlow    *core.Flow[string, *tts.Output, struct{}]
	asr        *asr.Asr
	agentStore session.Store[ChatState]
	deviceID   string
	sessionID  string
	dispatcher *Dispatcher
	services   *services.ServiceContainer
	logger     *slog.Logger

	mcpClient *gomcp.Client

	ClientVersion         int
	ClientAudioFormat     string
	ClientAudioFormatOpus bool
	ClientAudioFormatPcm  bool
	ClientSampleRate      int
	ClientChannels        int
	ClientFrameDuration   int

	mcpTransport     *mcp.XiaozhiTransport
	mcpClientSession *gomcp.ClientSession

	startTime        time.Time
	sampleRate       int
	exitIntentCalled bool

	listenChan chan string
	readyCh    chan struct{}
	workChan   chan func()
	workerWg   sync.WaitGroup
}

// Ensure Client implements handlers.Context
var _ handlers.Context = (*Client)(nil)

// NewClient creates a new client instance
func NewClient(conn *gws.Conn, deviceID string, sessionID string, services *services.ServiceContainer, logger *slog.Logger) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	stt, err := asr.NewAsr(asr.WithLogger(logger))
	if err != nil {
		logger.Error("Failed to create ASR instance: %v", err)
	}

	c := &Client{
		ctx:        ctx,
		cancel:     cancel,
		conn:       conn,
		deviceID:   deviceID,
		sessionID:  sessionID,
		dispatcher: dispatcher,
		services:   services,
		logger:     logger,

		asr:        stt,
		sampleRate: audio.DefaultSampleRate,
		agentStore: session.NewInMemoryStore[ChatState](),

		// Client default values
		ClientVersion:         1,
		ClientAudioFormat:     "opus",
		ClientAudioFormatOpus: true,
		ClientAudioFormatPcm:  false,
		ClientFrameDuration:   60,
		ClientChannels:        1,
		ClientSampleRate:      16000,
		startTime:             time.Now(),
		exitIntentCalled:      false,

		listenChan: make(chan string, 100),
		readyCh:    make(chan struct{}),
	}

	c.mcpTransport = mcp.NewXiaozhiTransport(sessionID, c.SendJSON)

	// Start worker goroutine
	c.workerWg.Add(1)
	go c.processListenChan()

	go c.processAsrResults()

	return c
}

func (c *Client) SendJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return c.SendMessage(gws.OpcodeText, data)
}

func (c *Client) SendMessage(opcode gws.Opcode, payload []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil {
		return gws.ErrConnClosed
	}

	// Reset read deadline when sending to keep connection alive during TTS streaming
	if err := c.conn.SetDeadline(time.Now().Add(ReadTimeout)); err != nil {
		// Connection might be closing, log but don't necessarily fail
		c.logger.Debug("failed to set deadline, connection may be closing", "error", err)
	}

	return c.conn.WriteMessage(opcode, payload)
}

func (c *Client) SetDeadline(t time.Time) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil {
		return gws.ErrConnClosed
	}

	return c.conn.SetDeadline(t)
}

func (c *Client) OnTextMessage(message *gws.Message) {
	msgBytes := make([]byte, len(message.Bytes()))
	copy(msgBytes, message.Bytes())
	if err := c.handleTextMessage(msgBytes); err != nil {
		c.logger.Error("handle message", "error", err)
	}
}

// handleTextMessage processes incoming JSON text messages
func (c *Client) handleTextMessage(msg []byte) error {
	var base types.BaseMessage
	if err := json.Unmarshal(msg, &base); err != nil {
		return fmt.Errorf("error decoding base message: %v", err)
	}

	c.logger.Debug("Received message", "type", base.Type, "message", string(msg))

	if base.Type == types.MessageTypeHello {
		return c.handleHelloMessage(msg)
	}

	if base.Type == types.MessageTypeMCP {
		return c.handleMcpMessage(msg)
	}

	if base.Type == types.MessageTypeListen {
		return c.handleListenMessage(msg)
	}

	if base.Type == types.MessageTypeAbort {
		c.asr.Stop()
		return nil
	}

	handler, err := c.dispatcher.GetHandler(base.Type)
	if err != nil {
		c.logger.Error("No handler found for message", "type", base.Type)
		return err
	}

	return handler.Handle(c, msg)
}

// OnBinaryMessage processes incoming audio data
func (c *Client) OnBinaryMessage(data []byte) {
	if c.ClientAudioFormat == "opus" {
		_ = c.asr.Write(data)
		return
	}

	c.logger.Warn("Not supported audio format")
}

func (c *Client) processAsrResults() {
	select {
	case <-c.ctx.Done():
		return
	default:
		for text := range c.asr.Result() {
			c.Logger().Info("ASR result", "text", text)
			if text != "" {

				// Save user message to history
				//if err := c.Services().History.SaveMessage(c.SessionID(), "user", result); err != nil {
				//	c.Logger().Error("Failed to save message history", "error", err)
				//}

				// Send to chat for processing
				c.asr.Stop()
				c.listenChan <- text
			}
		}
	}
}

func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	duration := time.Since(c.startTime)
	durationInSeconds := fmt.Sprintf("%.2f seconds", duration.Seconds())
	c.logger.Debug("closing connection...", "duration", durationInSeconds)
	c.cancel()
	close(c.listenChan)
	c.workerWg.Wait()
	c.asr.Close()
}
