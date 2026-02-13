package wav

import (
	"fmt"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/phamviet/xiaozhi-hub/internal/tts"
)

func EncodeTts(in *tts.Output) (string, error) {
	tempFile, err := os.CreateTemp("", "tts-output-*.wav")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	encoder := wav.NewEncoder(tempFile, in.SampleRate, in.BitDepth, in.Channels, 1)
	defer func() {
		_ = tempFile.Close()
		_ = encoder.Close()
	}()

	audioBuffer := &audio.IntBuffer{
		Format: &audio.Format{
			SampleRate:  in.SampleRate,
			NumChannels: in.Channels,
		},
		Data: make([]int, len(in.Content)/2),
	}

	for i := 0; i < len(in.Content)/2; i++ {
		audioBuffer.Data[i] = int(int16(in.Content[i*2]) | int16(in.Content[i*2+1])<<8)
	}

	if err := encoder.Write(audioBuffer); err != nil {
		return "", fmt.Errorf("failed to write temp WAV file: %w", err)
	}

	if err := encoder.Close(); err != nil {
		return "", fmt.Errorf("failed to close WAV encoder: %w", err)
	}

	return tempFile.Name(), nil
}
