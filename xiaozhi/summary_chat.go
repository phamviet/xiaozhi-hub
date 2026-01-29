package xiaozhi

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/phamviet/xiaozhi-hub/xiaozhi/types"
	"github.com/pocketbase/pocketbase/core"
)

// summaryChat /xiaozhi/agent/chat-summary/{sessionId}/save
func (m *Manager) summaryChat(e *core.RequestEvent) error {
	sessionId := e.Request.PathValue("sessionId")
	if sessionId == "" {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "sessionId is required"})
	}

	chat, err := m.Store.LoadChatSession(sessionId, "")
	if err != nil {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "Chat session not found"})
	}

	chatHistory, err := m.Store.FetchChatHistory(chat.ID)
	if err != nil {
		e.App.Logger().Error("fetch chat history error", "error", err.Error())
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if len(chatHistory) == 0 {
		return e.JSON(http.StatusOK, successResponse("No chat history"))
	}

	agent, err := m.Store.GetAgentByID(chat.AgentID)
	if err != nil {
		return e.JSON(http.StatusNotFound, map[string]string{"error": "Agent not found"})
	}

	if !agent.ChatHistoryEnabled {
		return e.JSON(http.StatusOK, successResponse("Agent chat history is disabled"))
	}

	llmConfigJson, err := m.resolveMemoryLLMConfig(agent)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// 1. Generate summary for the specific chat session
	sessionSummary, err := m.generateSessionSummary(llmConfigJson, chatHistory)
	if err != nil {
		e.App.Logger().Error("Failed to generate session summary", "error", err)
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate session summary"})
	}

	// 2. Update the chat session record with the summary
	if err := m.Store.UpdateChatSessionSummary(chat.ID, sessionSummary); err != nil {
		e.App.Logger().Error("Failed to save chat session summary", "error", err)
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save chat session summary"})
	}

	// 3. Generate and update agent's long-term memory (using the new session summary + existing memory)
	agentMemorySummary, err := m.generateAgentMemory(llmConfigJson, agent.SummaryMemory, sessionSummary)
	if err != nil {
		e.App.Logger().Error("Failed to generate agent memory", "error", err)
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate agent memory"})
	}

	if err := m.Store.UpdateAgentMemory(agent.ID, agentMemorySummary); err != nil {
		e.App.Logger().Error("Failed to save agent summary", "error", err)
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save summary"})
	}

	return e.JSON(http.StatusOK, successResponse(true))
}

func (m *Manager) resolveMemoryLLMConfig(agent *types.AIAgent) (*types.ModelConfigJson, error) {
	modelConfig, err := m.Store.GetModelConfigByIDOrDefault(agent.MemModelID, "Memory")
	if err != nil {
		return nil, fmt.Errorf("failed to get memory model config: %w", err)
	}

	providerCode, err := m.Store.GetProviderCodeByID(modelConfig.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("memory provider not found: %w", err)
	}

	modelConfigJson := modelConfig.ToModelConfigJson(providerCode)
	m.Store.ResolveSecretReference(modelConfigJson)

	if modelConfigJson.IsLLMReference() {
		llmID := modelConfigJson.Param["llm"]
		llmConfig, err := m.Store.GetModelConfigByIDOrDefault(llmID, "LLM")
		if err != nil {
			return nil, fmt.Errorf("failed to get referenced LLM config: %w", err)
		}
		llmProviderCode, err := m.Store.GetProviderCodeByID(llmConfig.ProviderID)
		if err != nil {
			return nil, fmt.Errorf("LLM provider not found: %w", err)
		}
		llmConfigJson := llmConfig.ToModelConfigJson(llmProviderCode)
		m.Store.ResolveSecretReference(llmConfigJson)
		return llmConfigJson, nil
	}

	return modelConfigJson, nil
}

func (m *Manager) generateSessionSummary(llmConfig *types.ModelConfigJson, chatHistory []types.ChatMessage) (string, error) {
	sysPrompt := "You are a helpful assistant. Summarize the following conversation in a concise manner in no more than two sentences. Response must in user's language."

	convText := m.formatConversation(chatHistory)

	userMsg := fmt.Sprintf(`Summarize this conversation session briefly.
--
CONVERSATION LOG:
%s
--
`, convText)

	messages := []LLMMessage{
		{Role: "system", Content: sysPrompt},
		{Role: "user", Content: userMsg},
	}

	llmClient, err := m.getLLMClient(llmConfig)
	if err != nil {
		return "", err
	}

	return llmClient.Chat(messages)
}

func (m *Manager) generateAgentMemory(llmConfig *types.ModelConfigJson, existingMemory string, sessionSummary string) (string, error) {
	sysPrompt := "Please update the long-term memory based on the new session summary."
	if val, err := m.Store.GetSysParam("memory.system_prompt"); err == nil {
		sysPrompt = val
	}

	if existingMemory == "" {
		existingMemory = "No prior memory available."
	}

	userMsg := fmt.Sprintf(`Update the user's long-term memory by integrating the new session summary.
--
EXISTING MEMORY:
%s
--
NEW SESSION SUMMARY:
%s
--
`, existingMemory, sessionSummary)

	messages := []LLMMessage{
		{Role: "system", Content: sysPrompt},
		{Role: "user", Content: userMsg},
	}

	llmClient, err := m.getLLMClient(llmConfig)
	if err != nil {
		return "", err
	}

	return llmClient.Chat(messages)
}

func (m *Manager) formatConversation(chatHistory []types.ChatMessage) string {
	var convBuilder strings.Builder
	for _, msg := range chatHistory {
		role := "User"
		if msg.ChatType == types.ChatTypeAssistant {
			role = "Assistant"
		}

		content := msg.Content
		if len(content) > 250 {
			prefix := content[:100]
			suffix := content[len(content)-100:]
			content = fmt.Sprintf("%s ... [CONTENT OMITTED] ... %s", prefix, suffix)
		}
		convBuilder.WriteString(fmt.Sprintf("[%s]: %s\n", role, content))
	}
	return convBuilder.String()
}
