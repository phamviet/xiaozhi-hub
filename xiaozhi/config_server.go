package xiaozhi

import (
	"encoding/json"
	"log"
	"net/http"

	"dario.cat/mergo"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

type BaseConfig struct {
	Server struct {
		Secret            string `json:"secret"`
		Websocket         string `json:"websocket"`
		OTA               string `json:"ota"`
		MCPEndpoint       string `json:"mcp_endpoint"`
		AllowUserRegister bool   `json:"allow_user_register"`
		Auth              struct {
			Enabled bool `json:"enabled"`
		} `json:"auth"`
		PrivateKey string `json:"private_key"`
		PublicKey  string `json:"public_key"`
	} `json:"server"`
	TTSTimeout     int               `json:"tts_timeout"`
	WakeupWords    []string          `json:"wakeup_words"`
	ExitCommands   []string          `json:"exit_commands"`
	SelectedModule map[string]string `json:"selected_module"`
	Prompt         *string           `json:"prompt"`
	EndPrompt      struct {
		Enabled bool    `json:"enable"`
		Prompt  *string `json:"prompt"`
	} `json:"end_prompt"`
	SummaryMemory *string `json:"summary_memory"`
}

// serverBaseConfig /xiaozhi/config/server-base
func (m *Manager) serverBaseConfig(e *core.RequestEvent) error {
	// 1. get a first enabled config from sys_config collection
	config, err := e.App.FindFirstRecordByData("sys_config", "disabled", false)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]string{"error": "No enabled config found"})
	}

	// 2. Merge sys_params to set secret
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

	baseConfig.EndPrompt.Enabled = false
	baseConfig.EndPrompt.Prompt = nil

	baseConfigBytes, err := json.Marshal(baseConfig)
	if err != nil {
		return err
	}

	var srcMap map[string]interface{}
	if err = json.Unmarshal(baseConfigBytes, &srcMap); err != nil {
		return err
	}

	if err := mergo.Merge(&destMap, srcMap, mergo.WithOverride); err != nil {
		log.Fatal(err)
	}

	e.App.Logger().Info("Returning config", "baseConfig", baseConfig)

	return e.JSON(http.StatusOK, success(destMap))
}

type Response struct {
	Code int    `json:"code"`
	Data any    `json:"data"`
	Msg  string `json:"msg"`
}

func success(data any) any {
	return &Response{Code: 0, Data: data, Msg: "success"}
}
