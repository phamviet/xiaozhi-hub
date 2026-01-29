package xiaozhi

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/phamviet/xiaozhi-hub/xiaozhi/store"
	"github.com/pocketbase/pocketbase/core"
)

type ReportChatRequest struct {
	AudioBase64 string `json:"audioBase64"`
	ChatType    int    `json:"chatType"`
	Content     string `json:"content"`
	MacAddress  string `json:"macAddress"`
	ReportTime  int64  `json:"reportTime"`
	SessionId   string `json:"sessionId"`
}

type ChatContent struct {
	Content  string `json:"content"`
	Language string `json:"language"`
	Speaker  string `json:"speaker"`
}

// reportChat /xiaozhi/agent/chat-history/report
func (m *Manager) reportChat(e *core.RequestEvent) error {
	var req ReportChatRequest
	if err := e.BindBody(&req); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.MacAddress == "" {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "macAddress is required"})
	}

	device, err := m.Store.GetDeviceByMacAddress(req.MacAddress)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]string{"error": "Device not found"})
	}

	if device.AgentId == "" {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "unbound device"})
	}

	chat, err := m.Store.LoadChatSession(req.SessionId, device.AgentId)
	if err != nil {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	content := req.Content
	if len(content) > 0 && content[0] == '{' {
		var parsedContent ChatContent
		if err := json.Unmarshal([]byte(content), &parsedContent); err == nil {
			content = parsedContent.Content
		}
	}

	chatHistoryParams := store.ChatHistoryParams{
		ChatID:      chat.ID,
		DeviceID:    device.Id,
		Content:     content,
		ChatType:    fmt.Sprintf("%d", req.ChatType),
		ReportTime:  req.ReportTime,
		AudioFormat: "mp3",
	}

	if req.AudioBase64 != "" {
		audioData, err := base64.StdEncoding.DecodeString(req.AudioBase64)
		if err != nil {
			return e.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid audioBase64"})
		}
		chatHistoryParams.AudioBytes = audioData
	}

	if err := m.Store.SaveChatHistory(chatHistoryParams); err != nil {
		e.App.Logger().Error("Failed to save chat history", "error", err, "mac", req.MacAddress)
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save chat history"})
	}

	return e.JSON(http.StatusOK, successResponse(true))
}
