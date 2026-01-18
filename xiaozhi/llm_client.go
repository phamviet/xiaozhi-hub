package xiaozhi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LLMRequest struct {
	Model    string       `json:"model"`
	Messages []LLMMessage `json:"messages"`
}

type LLMResponse struct {
	Choices []struct {
		Message LLMMessage `json:"message"`
	} `json:"choices"`
}

type LLMClient interface {
	Chat(messages []LLMMessage) (string, error)
}

type OpenAIClient struct {
	APIKey  string
	BaseURL string
	Model   string
}

func (c *OpenAIClient) Chat(messages []LLMMessage) (string, error) {
	reqBody := LLMRequest{
		Model:    c.Model,
		Messages: messages,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := c.BaseURL + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai error: status code %d", resp.StatusCode)
	}

	var llmResp LLMResponse
	if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
		return "", err
	}

	if len(llmResp.Choices) > 0 {
		return llmResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response from openai")
}

func (m *Manager) getLLMClient(modelConfig *ModelConfigJson) (LLMClient, error) {
	switch modelConfig.Type {
	case "openai":
		return &OpenAIClient{
			APIKey:  modelConfig.Param["api_key"],
			BaseURL: modelConfig.Param["base_url"],
			Model:   modelConfig.Param["model_name"],
		}, nil
	default:
		return nil, fmt.Errorf("unsupported llm provider: %s", modelConfig.Type)
	}
}
