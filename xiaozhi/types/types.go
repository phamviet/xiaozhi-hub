package types

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/tools/types"
)

type AIAgent struct {
	ID                 string `db:"id"`
	UserID             string `db:"user"`
	Name               string `db:"agent_name"`
	RolePrompt         string `db:"role_prompt"`
	SummaryMemory      string `db:"summary_memory"`
	LangCode           string `db:"lang_code"`
	ASRModelID         string `db:"asr_model_id"`
	VADModelID         string `db:"vad_model_id"`
	LLMModelID         string `db:"llm_model_id"`
	TTSModelID         string `db:"tts_model_id"`
	TTSVoiceID         string `db:"tts_voice_id"`
	MemModelID         string `db:"mem_model_id"`
	IntentModelID      string `db:"intent_model_id"`
	ChatHistoryEnabled bool   `db:"chat_history_enabled"`
}
type ChatType string

const TypeUser ChatType = "1"
const ChatTypeAssistant ChatType = "2"

type ChatMessage struct {
	ID        string   `db:"id"`
	SessionID string   `db:"chat"`
	Content   string   `db:"content"`
	ChatType  ChatType `db:"chat_type"`
}

type ChatSession struct {
	ID       string         `db:"id"`
	AgentID  string         `db:"agent"`
	Summary  string         `db:"summary"`
	Created  types.DateTime `db:"created"`
	Ended    types.DateTime `db:"ended"`
	Messages []ChatMessage
}

type ModelConfig struct {
	ID         string                `db:"id"`
	ModelName  string                `db:"model_name"`
	ModelType  string                `db:"model_type"`
	IsDefault  bool                  `db:"is_default"`
	IsEnabled  bool                  `db:"is_enabled"`
	ConfigJson types.JSONMap[string] `db:"config_json"`
	ProviderID string                `db:"provider_id"`
}

func (c *ModelConfig) ToModelConfigJson(providerCode string) *ModelConfigJson {
	param := make(map[string]string)
	if c.ConfigJson != nil {
		for k, v := range c.ConfigJson {
			param[k] = v
		}
	}

	param["name"] = c.ModelName

	return &ModelConfigJson{ID: c.ID, Type: providerCode, Param: param}
}

type ModelConfigJson struct {
	ID    string
	Type  string
	Param map[string]string
}

func (c *ModelConfigJson) MarshalJSON() ([]byte, error) {
	// Create a flat map with all Param fields plus type
	flat := make(map[string]string)

	// Copy all Param fields
	for key, value := range c.Param {
		flat[key] = value
	}

	// Add type field
	flat["type"] = c.Type

	return json.Marshal(flat)
}

func (c *ModelConfigJson) IsLLMReference() bool {
	_, isLLM := c.Param["llm"]

	return isLLM
}

type SysParam struct {
	Name  string `db:"name"`
	Value string `db:"value"`
}
