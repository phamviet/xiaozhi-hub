package ws

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/rand/v2"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/x/session"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/firebase/genkit/go/plugins/mcp"
	"github.com/phamviet/xiaozhi-hub/internal/audio"
	"github.com/phamviet/xiaozhi-hub/internal/tts"
	"github.com/phamviet/xiaozhi-hub/internal/wav"
	"google.golang.org/genai"
)

type ChatState struct {
	Device  string   `json:"device"`
	History []string `json:"history"`
}

type AgentConfig struct {
	SystemPrompt string   `json:"system_prompt"`
	LLMModel     string   `json:"llm_model"`
	TTSModel     string   `json:"tts_model"`
	WakeWords    []string `json:"wake_words"`
	QuickReplies []string `json:"quick_replies"`
}

type AgentOption func(*AgentConfig)

func WithSystemPrompt(prompt string) AgentOption {
	return func(c *AgentConfig) {
		c.SystemPrompt = prompt
	}
}

func WithLlmModel(model string) AgentOption {
	return func(c *AgentConfig) {
		c.LLMModel = model
	}
}

func WithTtsModel(model string) AgentOption {
	return func(c *AgentConfig) {
		c.TTSModel = model
	}
}

func NewAgentConfig() *AgentConfig {
	cfg := &AgentConfig{
		SystemPrompt: "You are a helpful assistant. Use the appropriate tool based on user intent",
		LLMModel:     "googleai/gemini-2.5-flash", // gemini-2.5-flash-lite
		TTSModel:     "googleai/gemini-2.5-flash-preview-tts",
		WakeWords:    []string{"hi", "test", "genkit", "go"},
		QuickReplies: []string{"hi", "Hello! How can I assist you today?"},
	}

	return cfg
}

const sampleText = "Genkit is the best Gen AI library!"

func (c *Client) initializeAgent(cfg *AgentConfig) {
	if cfg == nil {
		cfg = NewAgentConfig()
	}

	c.g = genkit.Init(c.ctx, genkit.WithDefaultModel(cfg.LLMModel), genkit.WithPlugins(&googlegenai.GoogleAI{}))
	c.initInternalTools()

	// connect to the device's mcp server
	client, err := mcp.NewClient(c.ctx, mcp.MCPClientOptions{
		Name:      "xiaozhi-hub-client",
		Transport: c.mcpTransport,
	})
	if err != nil {
		c.logger.Error("Failed to create MCP client", "error", err)
		return
	}

	// Get tools
	tools, _ := client.GetActiveTools(c.ctx, c.g)
	c.logger.Info("Found MCP time tools", "count", len(tools))
	for _, tool := range tools {
		c.tools = append(c.tools, tool)
	}

	c.ttsFlow = genkit.DefineFlow(c.g, "tts", func(ctx context.Context, input string) (*tts.Output, error) {
		if strings.HasPrefix(input, "Genkit") {
			return tts.NewOutputFromFile("sample/lt30.wav", 24000, 1, 16)
		}

		resp, err := genkit.Generate(ctx, c.g,
			ai.WithModelName(cfg.TTSModel),
			ai.WithConfig(&genai.GenerateContentConfig{
				Temperature:        genai.Ptr[float32](1.0),
				ResponseModalities: []string{"AUDIO"},
				SpeechConfig: &genai.SpeechConfig{
					VoiceConfig: &genai.VoiceConfig{
						PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
							VoiceName: "Algenib",
						},
					},
				},
			}),
			ai.WithPrompt(fmt.Sprintf("Say: %s", input)))
		if err != nil {
			return nil, err
		}

		part := resp.Message.Content[0]
		prefix := fmt.Sprintf("data:%s;base64,", part.ContentType)
		base64Encoded := part.Text[len(prefix):]
		rawBytes, err := base64.StdEncoding.DecodeString(base64Encoded)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 audio: %w", err)
		}

		return tts.NewOutput(rawBytes, 24000, 1, 16)
	})

	c.chatFlow = genkit.DefineFlow(c.g, "chat", func(ctx context.Context, input string) (string, error) {
		if input == "genkit" || input == "go" {
			return sampleText, nil
		}

		if slices.Contains(cfg.WakeWords, input) {
			randomIndex := rand.IntN(len(cfg.WakeWords))
			return cfg.WakeWords[randomIndex], nil
		}

		sessionID := c.SessionID()

		// Load existing session or create new one
		sess, err := session.Load(ctx, c.agentStore, sessionID)
		if err != nil {
			sess, err = session.New(ctx,
				session.WithID[ChatState](sessionID),
				session.WithStore(c.agentStore),
				session.WithInitialState(ChatState{Device: c.deviceID}),
			)
			if err != nil {
				return "", err
			}
		}

		// Attach session to context for use in tools and prompts
		ctx = session.NewContext(ctx, sess)

		// Generate with the session-aware context
		resp, err := genkit.Generate(ctx, c.g,
			ai.WithModelName(cfg.LLMModel),
			//ai.WithModel(googlegenai.ModelRef("googleai/gemini-2.5-flash", &genai.GenerateContentConfig{
			//	ThinkingConfig: &genai.ThinkingConfig{ThinkingBudget: genai.Ptr[int32](0)},
			//})),
			ai.WithTools(c.tools...),
			ai.WithToolChoice(ai.ToolChoiceAuto),
			ai.WithSystem(cfg.SystemPrompt),
			ai.WithPrompt(input),
		)

		if err != nil {
			return "", err
		}

		return resp.Text(), nil
	})

	c.mu.Lock()
	c.ready = true
	c.mu.Unlock()
}

func (c *Client) Chat(ctx context.Context, text string) {
	if ctx.Err() != nil {
		return
	}

	result, err := c.chatFlow.Run(ctx, text)
	if err != nil {
		c.logger.Error("chat.flow", "error", err)
		return
	}

	_ = c.SendTtsStart(c.sampleRate)
	_ = c.SendSttMessage(text)
	// send llm intent

	lines := strings.SplitSeq(result, "\n")
	var nonEmptyLines []string
	for line := range lines {
		if line != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}

	if len(nonEmptyLines) == 0 {
		time.Sleep(time.Duration(500) * time.Millisecond)
		_ = c.SendTtsStop()
		return
	}

	type ttsResult struct {
		text     string
		filename string
		err      error
	}

	// Process TTS in parallel but stream in order as soon as each is ready
	results := make([]ttsResult, len(nonEmptyLines))
	readyChan := make(chan int, len(nonEmptyLines))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i, line := range nonEmptyLines {
		wg.Add(1)
		go func(idx int, l string) {
			defer wg.Done()
			output, err := c.ttsFlow.Run(c.ctx, l)
			if err != nil {
				c.logger.Error("ttsFlow.Run", "error", err, "text", l)
				mu.Lock()
				results[idx] = ttsResult{err: err}
				mu.Unlock()
				readyChan <- idx
				return
			}
			filename, err := wav.EncodeTts(output)
			if err != nil {
				c.logger.Error("wav.EncodeTts", "error", err, "text", l)
				mu.Lock()
				results[idx] = ttsResult{err: err}
				mu.Unlock()
				readyChan <- idx
				return
			}
			mu.Lock()
			results[idx] = ttsResult{text: l, filename: filename}
			mu.Unlock()
			readyChan <- idx
		}(i, line)
	}

	go func() {
		wg.Wait()
		close(readyChan)
	}()

	// Stream in order as results become available
	nextIdx := 0
	completed := make([]bool, len(nonEmptyLines))

	for nextIdx < len(nonEmptyLines) {
		// Check if next index is already ready
		mu.Lock()
		isReady := completed[nextIdx]
		mu.Unlock()

		if isReady {
			res := results[nextIdx]
			if res.err == nil {
				c.streamTtsFile(res.text, res.filename)
			}
			nextIdx++
			continue
		}

		// Wait for next completion
		idx, ok := <-readyChan
		if !ok {
			break // Channel closed, all done
		}

		mu.Lock()
		completed[idx] = true
		mu.Unlock()

		// If this is the next expected index, stream it
		if idx == nextIdx {
			res := results[nextIdx]
			if res.err == nil {
				c.streamTtsFile(res.text, res.filename)
			}
			nextIdx++
		}
	}

	time.Sleep(time.Duration(500) * time.Millisecond)
	_ = c.SendTtsStop()

	// Close chat
	if c.exitIntentCalled {
		_ = c.conn.WriteClose(1000, nil)
	}
}

func (c *Client) streamTtsFile(text string, filename string) {
	audioFile, err := os.Open(filename)
	if err != nil {
		c.logger.Error("os.Open", "filename", filename, "error", err)
		return
	}

	_ = c.SendTtsMessage("sentence_start", text)
	decoder := wav.NewDecoder(audioFile)
	duration, _ := decoder.Duration()
	if _, err = audioFile.Seek(0, 0); err != nil {
		c.logger.Error("audioFile.Seek", "error", err)
		_ = audioFile.Close()
		_ = os.Remove(filename)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), duration+5*time.Second)
	if err := audio.StreamOpus(ctx, audioFile, c.conn); err != nil {
		c.logger.Error("audio.StreamOpus", "error", err)
	}
	cancel()
	_ = audioFile.Close()
	_ = os.Remove(filename)
}
