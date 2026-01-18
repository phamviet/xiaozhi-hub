package xiaozhi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

type ClientRequest struct {
	ClientId       string            `json:"clientId"`
	MacAddress     string            `json:"macAddress"`
	SelectedModule map[string]string `json:"selectedModule"`
}

type ModelsConfigResponse struct {
	ModelConfigMap      map[string]map[string]*ModelConfigJson `json:"-"`
	SystemPrompt        string                                 `json:"prompt"`
	SummaryMemory       string                                 `json:"summaryMemory"`
	SelectedModule      map[string]string                      `json:"selected_module"`
	ChatHistoryConf     int                                    `json:"chat_history_conf"`
	DeviceMaxOutputSize string                                 `json:"device_max_output_size"`
	Plugins             map[string]string                      `json:"plugins"`
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
	device, err := m.getDeviceByMacAddress(req.MacAddress)
	if err != nil {
		code := m.BindingManager.GetOrGenerateCode(req.MacAddress, req.ClientId)
		return e.JSON(http.StatusOK, map[string]interface{}{
			"code": 10042,
			"data": nil,
			"msg":  code,
		})
	}

	agent, err := m.getAgentByID(device.AgentID)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]string{"error": "Agent not found"})
	}

	response := &ModelsConfigResponse{
		ModelConfigMap:      make(map[string]map[string]*ModelConfigJson),
		SystemPrompt:        agent.SystemPrompt,
		SummaryMemory:       agent.SummaryMemory,
		ChatHistoryConf:     0, // disable by default
		DeviceMaxOutputSize: "0",
		Plugins:             make(map[string]string),
	}

	response.Plugins["get_weather"] = `{"api_key": "test", "api_host": "mj7p3y7naa.re.qweatherapi.com", "default_location": "广州"}`
	response.Plugins["play_music"] = `{}`

	if agent.ChatHistoryEnabled {
		response.ChatHistoryConf = 2 // store voice & message
	}

	getModelConfigJson := func(id string, modelType string) (*ModelConfigJson, error) {
		modelConfig, err := m.getModelConfigByIDOrDefault(id, modelType)
		if err != nil {
			return nil, err
		}

		if modelConfig.ProviderID == "" {
			e.App.Logger().Error("Model provider_id is empty. Hint: config_json may have unexpected value", "id", id, "name", modelConfig.ModelName)
			return nil, errors.New("model provider_id is empty")
		}

		providerRecord, err := e.App.FindRecordById("model_providers", modelConfig.ProviderID)
		if err != nil {
			e.App.Logger().Error("Model provider not found", "id", modelConfig.ProviderID, "error", err)
			return nil, errors.New("model provider not found")
		}

		return modelConfig.ToModelConfigJson(providerRecord.GetString("provider_code")), nil
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
			m.resolveSecretReference(e.App, modelConfig)
			selectedModule[modelType] = modelConfig.ID
			if modelConfig.isLLMReference() {
				llmIDs = append(llmIDs, modelConfig.Param["llm"])
			}

			if response.ModelConfigMap[modelType] == nil {
				response.ModelConfigMap[modelType] = make(map[string]*ModelConfigJson)
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
			m.resolveSecretReference(e.App, modelConfig)
			if response.ModelConfigMap["LLM"] == nil {
				response.ModelConfigMap["LLM"] = make(map[string]*ModelConfigJson)
			}
			response.ModelConfigMap["LLM"][modelConfig.ID] = modelConfig
		}
	}

	response.SelectedModule = selectedModule

	return e.JSON(http.StatusOK, successResponse(response))
}

func (m *Manager) resolveSecretReference(app core.App, modelConfig *ModelConfigJson) {
	credentialID, ok := modelConfig.Param["secret_ref"]
	if !ok {
		return
	}

	cred, err := app.FindRecordById("user_credentials", credentialID)
	if err != nil {
		app.Logger().Error("Failed to resolve user credential", "id", credentialID, "error", err)
		return
	}

	modelConfig.Param["api_key"] = cred.GetString("api_key")
	delete(modelConfig.Param, "secret_ref")
}
