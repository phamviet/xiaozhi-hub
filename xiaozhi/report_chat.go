package xiaozhi

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
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

	device, err := m.getDeviceByMacAddress(req.MacAddress)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]string{"error": "Device not found"})
	}

	if device.AgentId == "" {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "unbound device"})
	}

	chat, err := m.loadChatSession(req.SessionId, device.AgentId)
	if err != nil {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	collection, err := e.App.FindCollectionByNameOrId("ai_agent_chat_history")
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "ai_agent_chat_history collection not found"})
	}

	content := req.Content
	if len(content) > 0 && content[0] == '{' {
		var parsedContent ChatContent
		if err := json.Unmarshal([]byte(content), &parsedContent); err == nil {
			content = parsedContent.Content
		}
	}

	record := core.NewRecord(collection)
	record.Set("chat", chat.ID)
	record.Set("device", device.Id)
	record.Set("content", content)
	record.Set("chat_type", fmt.Sprintf("%d", req.ChatType))

	if req.AudioBase64 != "" {
		audioData, err := base64.StdEncoding.DecodeString(req.AudioBase64)
		if err != nil {
			return e.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid audioBase64"})
		}

		filename := fmt.Sprintf("audio_%d_%d.mp3", time.Now().UnixNano(), req.ReportTime)
		file, err := filesystem.NewFileFromBytes(audioData, filename)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create file from bytes"})
		}

		record.Set("chat_audio", file)
	}

	if err := e.App.Save(record); err != nil {
		e.App.Logger().Error("Failed to save chat history", "error", err, "mac", req.MacAddress)
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save chat history"})
	}

	return e.JSON(http.StatusOK, successResponse(true))
}
