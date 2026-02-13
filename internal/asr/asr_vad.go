package asr

import sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"

const WindowSizeVad = 512

type VadConfig struct {
	model               string
	bufferSizeInSeconds float32
}

type VadOption func(*VadConfig)

func WithVadModel(model string) VadOption {
	return func(cfg *VadConfig) {
		cfg.model = model
	}
}

func NewVad(opts ...VadOption) *sherpa.VoiceActivityDetector {
	cfg := &VadConfig{
		model:               "models/silero_vad.onnx",
		bufferSizeInSeconds: 5,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	vadCfg := sherpa.VadModelConfig{
		SileroVad: sherpa.SileroVadModelConfig{
			Model:      cfg.model,
			WindowSize: WindowSizeVad,
		},
		SampleRate: 16000,
		NumThreads: 1,
		Provider:   "cpu",
	}

	vad := sherpa.NewVoiceActivityDetector(&vadCfg, cfg.bufferSizeInSeconds)

	if vad == nil {
		return nil
	}
	return vad
}
