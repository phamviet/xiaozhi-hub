package xiaozhi

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/tools/types"
)

type AIAgent struct {
	Name               string `db:"agent_name"`
	SystemPrompt       string `db:"system_prompt"`
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

type Device struct {
	MacAddress    string         `db:"mac_address"`
	UserID        string         `db:"user"`
	AgentID       string         `db:"agent"`
	Board         string         `db:"board"`
	LastConnected types.DateTime `db:"last_connected"`
	Created       types.DateTime `db:"created"`
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

func (c *ModelConfigJson) isLLMReference() bool {
	_, isLLM := c.Param["llm"]

	return isLLM
}
