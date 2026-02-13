package types

import "encoding/json"

// MessageType defines the type of the message
type MessageType string

const (
	MessageTypeHello  MessageType = "hello"
	MessageTypeIoT    MessageType = "iot"
	MessageTypeListen MessageType = "listen"
	MessageTypeAbort  MessageType = "abort"
	MessageTypeSTT    MessageType = "stt"
	MessageTypeLLM    MessageType = "llm"
	MessageTypeTTS    MessageType = "tts"
	MessageTypeMCP    MessageType = "mcp"
	MessageTypeSystem MessageType = "system"
	MessageTypeCustom MessageType = "custom"
)

// BaseMessage contains common fields for routing
type BaseMessage struct {
	Type      MessageType `json:"type"`
	SessionID string      `json:"session_id,omitempty"`
}

// AudioParams defines audio configuration
type AudioParams struct {
	Format        string `json:"format"`
	SampleRate    int    `json:"sample_rate"`
	Channels      int    `json:"channels"`
	FrameDuration int    `json:"frame_duration,omitempty"`
}

// HelloMessage (Client -> Server & Server -> Client)
type HelloMessage struct {
	BaseMessage
	Version     int         `json:"version,omitempty"`
	Features    *Features   `json:"features,omitempty"`
	Transport   string      `json:"transport,omitempty"`
	AudioParams AudioParams `json:"audio_params"`
}

type Features struct {
	MCP bool `json:"mcp,omitempty"`
}

// IoTDescriptor defines an IoT device capability
type IoTDescriptor struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	Methods     map[string]interface{} `json:"methods,omitempty"`
}

// IoTState defines an IoT device state
type IoTState struct {
	Name  string                 `json:"name"`
	State map[string]interface{} `json:"state"`
}

// IoTMessage (Client -> Server)
type IoTMessage struct {
	BaseMessage
	Update      bool            `json:"update"`
	Descriptors []IoTDescriptor `json:"descriptors,omitempty"`
	States      []IoTState      `json:"states,omitempty"`
}

// ListenMessage (Client -> Server)
type ListenMessage struct {
	BaseMessage
	State string `json:"state"` // "start", "stop", "detect"
	Mode  string `json:"mode,omitempty"`
	Text  string `json:"text,omitempty"` // For "detect" state (wake word)
}

// AbortMessage (Client -> Server)
type AbortMessage struct {
	BaseMessage
	Reason string `json:"reason"`
}

// STTMessage (Server -> Client)
type STTMessage struct {
	BaseMessage
	Text string `json:"text"`
}

// LLMMessage (Server -> Client)
type LLMMessage struct {
	BaseMessage
	Text    string `json:"text"`
	Emotion string `json:"emotion,omitempty"`
}

// TTSMessage (Server -> Client)
type TTSMessage struct {
	BaseMessage
	State      string `json:"state"` // "start", "stop", "sentence_start", "sentence_end"
	Text       string `json:"text,omitempty"`
	SampleRate int    `json:"sample_rate,omitempty"`
}

// MCPMessage (Bidirectional)
type MCPMessage struct {
	BaseMessage
	Payload json.RawMessage `json:"payload"`
}

// SystemMessage (Server -> Client)
type SystemMessage struct {
	BaseMessage
	Command string `json:"command"`
}
