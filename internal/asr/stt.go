package asr

type SpeechToText interface {
	Transcribe(audio []byte) (string, error)
}

type SttConfig struct {
	OpenAi struct {
		ApiKey  string
		Model   string
		BaseURL string
	}
}
