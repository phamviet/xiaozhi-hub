package xiaozhi

import (
	"encoding/json"
	"net/http"

	"dario.cat/mergo"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

type BaseConfig struct {
	Server struct {
		Secret            string `json:"secret"`
		Websocket         string `json:"websocket"`
		MCPEndpoint       string `json:"mcp_endpoint"`
		AllowUserRegister bool   `json:"allow_user_register"`
		Auth              struct {
			Enabled bool `json:"enabled"`
		} `json:"auth"`
	} `json:"server"`
	TTSTimeout      int               `json:"tts_timeout"`
	WakeupWords     []string          `json:"wakeup_words"`
	ExitCommands    []string          `json:"exit_commands"`
	SelectedModule  map[string]string `json:"selected_module"`
	AgentBasePrompt *string           `json:"agent_base_prompt"`
	EndPrompt       struct {
		Enabled bool    `json:"enable"`
		Prompt  *string `json:"prompt"`
	} `json:"end_prompt"`
}

// serverBaseConfig /xiaozhi/config/server-base
func (m *Manager) serverBaseConfig(e *core.RequestEvent) error {
	// 1. get a first enabled config from sys_config collection
	config, err := m.Store.GetActiveSysConfig()
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]string{"error": "No enabled config found"})
	}

	// Get the config value as map
	sysValue := config.GetRaw("value").(types.JSONRaw)
	var destMap map[string]interface{}
	if err = json.Unmarshal(sysValue, &destMap); err != nil {
		return err
	}

	var baseConfig BaseConfig
	err = json.Unmarshal(sysValue, &baseConfig)
	if err != nil {
		e.App.Logger().Error("Unmarshal baseConfig", "error", err)
		return err
	}

	logError := func(err error) error {
		e.App.Logger().Error(err.Error())
		return e.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
	}

	// 2. Merge sys_params into baseConfig
	sysParams, err := m.Store.GetSysParams("server.secret", "agent.base_prompt", "server.websocket")
	if err != nil {
		return logError(err)
	}

	baseConfig.Server.Secret = *sysParams["server.secret"]
	baseConfig.Server.Websocket = *sysParams["server.websocket"]
	baseConfig.AgentBasePrompt = sysParams["agent.base_prompt"]

	baseConfigBytes, err := json.Marshal(baseConfig)
	if err != nil {
		return err
	}

	var srcMap map[string]interface{}
	if err = json.Unmarshal(baseConfigBytes, &srcMap); err != nil {
		return err
	}

	if err := mergo.Merge(&destMap, srcMap, mergo.WithOverride); err != nil {
		e.App.Logger().Error("Failed to merge baseConfig", "error", err)
		return e.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
	}

	return e.JSON(http.StatusOK, successResponse(destMap))
}

type XiaozhiResponse struct {
	Code int    `json:"code"`
	Data any    `json:"data"`
	Msg  string `json:"msg"`
}

func successResponse(data any) any {
	return &XiaozhiResponse{Code: 0, Data: data, Msg: "success"}
}

func errorResponse(message string) any {
	return &XiaozhiResponse{Code: 1, Msg: message}
}
