package xiaozhi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/phamviet/xiaozhi-hub/xiaozhi/types"
	"github.com/pocketbase/pocketbase/core"
)

type ClientRequest struct {
	ClientId       string            `json:"clientId"`
	MacAddress     string            `json:"macAddress"`
	SelectedModule map[string]string `json:"selectedModule"`
}

type ModelsConfigResponse struct {
	ModelConfigMap      map[string]map[string]*types.ModelConfigJson `json:"-"`
	RolePrompt          string                                       `json:"prompt"`
	SummaryMemory       string                                       `json:"summaryMemory"`
	SelectedModule      map[string]string                            `json:"selected_module"`
	ChatHistoryConf     int                                          `json:"chat_history_conf"`
	DeviceMaxOutputSize string                                       `json:"device_max_output_size"`
	Plugins             map[string]string                            `json:"plugins"`
}

func (r *ModelsConfigResponse) MarshalJSON() ([]byte, error) {
	type Alias ModelsConfigResponse
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	data, err := json.Marshal(aux)
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	for k, v := range r.ModelConfigMap {
		m[k] = v
	}

	return json.Marshal(m)
}

// getAgentModels /xiaozhi/config/agent-models
func (m *Manager) getAgentModels(e *core.RequestEvent) error {
	var req ClientRequest
	if err := e.BindBody(&req); err != nil {
		return err
	}

	e.App.Logger().Info("Received agent models config request", "req", req, "headers", e.Request.Header)
	device, err := m.Store.GetDeviceByMacAddress(req.MacAddress)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]string{"error": "Device not found"})
	}

	agent, err := m.Store.GetAgentByID(device.AgentId)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]string{"error": "Agent not found"})
	}

	response := &ModelsConfigResponse{
		ModelConfigMap:      make(map[string]map[string]*types.ModelConfigJson),
		RolePrompt:          agent.RolePrompt,
		SummaryMemory:       agent.SummaryMemory,
		ChatHistoryConf:     0, // disable by default
		DeviceMaxOutputSize: "0",
		Plugins:             make(map[string]string),
	}

	response.Plugins["get_weather"] = `{"api_key": "test", "api_host": "mj7p3y7naa.re.qweatherapi.com", "default_location": "Tay Ninh"}`
	response.Plugins["play_music"] = `{}`

	if agent.ChatHistoryEnabled {
		response.ChatHistoryConf = 2 // store voice & message
	}

	getModelConfigJson := func(id string, modelType string) (*types.ModelConfigJson, error) {
		modelConfig, err := m.Store.GetModelConfigByIDOrDefault(id, modelType)
		if err != nil {
			return nil, err
		}

		if modelConfig.ProviderID == "" {
			e.App.Logger().Error("Model provider_id is empty. Hint: config_json may have unexpected value", "id", id, "name", modelConfig.ModelName)
			return nil, errors.New("model provider_id is empty")
		}

		providerCode, err := m.Store.GetProviderCodeByID(modelConfig.ProviderID)
		if err != nil {
			e.App.Logger().Error("Model provider not found", "id", modelConfig.ProviderID, "error", err)
			return nil, errors.New("model provider not found")
		}

		return modelConfig.ToModelConfigJson(providerCode), nil
	}

	var selectedModule = make(map[string]string)
	llmIDs := make([]string, 0)

	loadModelConfig := func(id string, modelType string) error {
		modelConfig, err := getModelConfigJson(id, modelType)
		if err != nil {
			e.App.Logger().Error("Failed to get model config", "model_type", modelType, "error", err)
			return err
		}

		if modelConfig != nil {
			m.Store.ResolveSecretReference(modelConfig)
			selectedModule[modelType] = modelConfig.ID
			if modelConfig.IsLLMReference() {
				llmIDs = append(llmIDs, modelConfig.Param["llm"])
			}

			if response.ModelConfigMap[modelType] == nil {
				response.ModelConfigMap[modelType] = make(map[string]*types.ModelConfigJson)
			}

			response.ModelConfigMap[modelType][modelConfig.ID] = modelConfig
		}

		return nil
	}

	_ = loadModelConfig(agent.ASRModelID, "ASR")
	_ = loadModelConfig(agent.VADModelID, "VAD")
	_ = loadModelConfig(agent.TTSModelID, "TTS")
	_ = loadModelConfig(agent.LLMModelID, "LLM")
	_ = loadModelConfig(agent.MemModelID, "Memory")
	_ = loadModelConfig(agent.IntentModelID, "Intent")

	// Load referenced LLMs
	for _, llmID := range llmIDs {
		if response.ModelConfigMap["LLM"] != nil && response.ModelConfigMap["LLM"][llmID] != nil {
			continue
		}
		modelConfig, err := getModelConfigJson(llmID, "LLM")
		if err == nil && modelConfig != nil {
			m.Store.ResolveSecretReference(modelConfig)
			if response.ModelConfigMap["LLM"] == nil {
				response.ModelConfigMap["LLM"] = make(map[string]*types.ModelConfigJson)
			}
			response.ModelConfigMap["LLM"][modelConfig.ID] = modelConfig
		}
	}

	response.SelectedModule = selectedModule

	return e.JSON(http.StatusOK, successResponse(response))
}
