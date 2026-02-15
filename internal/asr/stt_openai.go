package asr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type OpenAi struct {
	apiKey   string
	model    string
	language string
	baseURL  string
}

func NewOpenAi(apiKey string) SpeechToText {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}

	return &OpenAi{
		apiKey:   apiKey,
		model:    "whisper-large-v3",
		language: "vi",
		baseURL:  "https://api.groq.com/openai/v1",
	}
}

func (o *OpenAi) Transcribe(audio []byte) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "speech.wav")
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, bytes.NewReader(audio)); err != nil {
		return "", fmt.Errorf("failed to copy audio data: %w", err)
	}

	_ = writer.WriteField("model", o.model)
	if o.language != "" {
		_ = writer.WriteField("language", o.language)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/audio/transcriptions", o.baseURL), body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai api error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Text, nil
}

var _ SpeechToText = (*OpenAi)(nil)
