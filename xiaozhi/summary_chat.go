package xiaozhi

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// summaryChat /xiaozhi/agent/chat-summary/{sessionId}/save
func (m *Manager) summaryChat(e *core.RequestEvent) error {
	sessionId := e.Request.PathValue("sessionId")
	if sessionId == "" {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "sessionId is required"})
	}

	// 1. Find the agent from sessionId (by looking up chat history)
	var chatHistory []struct {
		AgentID  string `db:"agent_id"`
		Content  string `db:"content"`
		ChatType string `db:"chat_type"`
	}

	err := e.App.DB().Select("agent_id", "content", "chat_type").
		From("ai_agent_chat_history").
		Where(dbx.HashExp{"conversation_id": sessionId}).
		OrderBy("created ASC").
		All(&chatHistory)

	if err != nil || len(chatHistory) == 0 {
		return e.JSON(http.StatusNotFound, map[string]string{"error": "Session not found or no chat history"})
	}

	agentID := chatHistory[0].AgentID
	agent, err := m.getAgentByID(agentID)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]string{"error": "Agent not found"})
	}

	// 2. Take action if chat_history_enabled is true
	if !agent.ChatHistoryEnabled {
		return e.JSON(http.StatusOK, successResponse("Chat history disabled for this agent"))
	}

	// 3. Look at mem_model_id
	if agent.MemModelID == "" {
		return e.JSON(http.StatusOK, successResponse("No memory model configured"))
	}

	// 4. Fetch memory model config
	modelConfig, err := m.getModelConfigByIDOrDefault(agent.MemModelID, "Memory")
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get memory model config"})
	}

	providerRecord, err := e.App.FindRecordById("model_providers", modelConfig.ProviderID)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Memory provider not found"})
	}

	modelConfigJson := modelConfig.ToModelConfigJson(providerRecord.GetString("provider_code"))
	m.resolveSecretReference(e.App, modelConfigJson)

	// If it's a memory module that references an LLM
	var llmConfigJson *ModelConfigJson
	if modelConfigJson.isLLMReference() {
		llmID := modelConfigJson.Param["llm"]
		llmConfig, err := m.getModelConfigByIDOrDefault(llmID, "LLM")
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get referenced LLM config"})
		}
		llmProviderRecord, err := e.App.FindRecordById("model_providers", llmConfig.ProviderID)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{"error": "LLM provider not found"})
		}
		llmConfigJson = llmConfig.ToModelConfigJson(llmProviderRecord.GetString("provider_code"))
		m.resolveSecretReference(e.App, llmConfigJson)
	} else {
		// If it's directly an LLM (though usually Memory type is different)
		llmConfigJson = modelConfigJson
	}

	// 5. Get system prompt for summary
	sysPrompt := ""
	promptParam, err := e.App.FindFirstRecordByData("sys_params", "name", "memory.system_prompt")
	if err == nil {
		sysPrompt = promptParam.GetString("value")
	}

	if sysPrompt == "" {
		sysPrompt = "Please summarize the following conversation briefly."
	}

	// 6. Build conversation text
	var convBuilder strings.Builder
	for _, msg := range chatHistory {
		role := "User"
		if msg.ChatType == "2" { // 1 is User, 2 is Assistant
			role = "Assistant"
		}
		convBuilder.WriteString(fmt.Sprintf("%s: %s\n", role, msg.Content))
	}

	// 7. Call LLM
	llmClient, err := m.getLLMClient(llmConfigJson)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	userMsg := fmt.Sprintf("Current Conversation:\n%s", convBuilder.String())
	if agent.SummaryMemory != "" {
		userMsg = fmt.Sprintf("Previous Memories:\n%s\n\n%s", agent.SummaryMemory, userMsg)
	}

	messages := []LLMMessage{
		{Role: "system", Content: sysPrompt},
		{Role: "user", Content: userMsg},
	}

	summary, err := llmClient.Chat(messages)
	if err != nil {
		e.App.Logger().Error("Failed to generate summary", "error", err)
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate summary"})
	}

	// 8. Update agent summary_memory
	agentRecord, err := e.App.FindRecordById("ai_agent", agent.ID)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find agent record for update"})
	}

	agentRecord.Set("summary_memory", summary)
	if err := e.App.Save(agentRecord); err != nil {
		e.App.Logger().Error("Failed to save agent summary", "error", err)
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save summary"})
	}

	return e.JSON(http.StatusOK, successResponse(true))
}
