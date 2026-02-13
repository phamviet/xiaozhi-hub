package audio

import (
	"bytes"
	"fmt"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/zaf/resample"
)

func downsample(sample []byte, inputRate float64) ([]byte, error) {
	var buf bytes.Buffer
	resampler, err := resample.New(&buf, inputRate, 16000, 1, resample.I16, resample.MediumQ)
	if err != nil {
		return nil, fmt.Errorf("failed to create resampler: %w", err)
	}
	defer resampler.Close()

	_, err = resampler.Write(sample)
	if err != nil {
		return nil, fmt.Errorf("failed to resample synthesized audio: %w", err)
	}

	return buf.Bytes(), nil
}

func Float32ToWavBytes(samples []float32, sampleRate int) ([]byte, error) {
	buf := &audio.IntBuffer{
		Format: &audio.Format{
			NumChannels: 1,
			SampleRate:  sampleRate,
		},
		Data: make([]int, len(samples)),
	}

	// Convert float32 to int (16-bit PCM)
	for i, sample := range samples {
		buf.Data[i] = int(sample * 32767) // Convert float32 [-1,1] to int16 range
	}

	tempFile, err := os.CreateTemp("", "tts-*.wav")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	enc := wav.NewEncoder(tempFile,
		SampleRate,
		16, // 16-bit PCM
		1,  // mono
		1,
	)

	defer func() {
		_ = os.Remove(tempFile.Name())
	}()

	if err := enc.Write(buf); err != nil {
		_ = enc.Close()
		return nil, err
	}

	_ = enc.Close()
	return os.ReadFile(tempFile.Name())
}

func PCMToWavBytes(samples []byte, sampleRate int) ([]byte, error) {
	buf := &audio.IntBuffer{
		Format: &audio.Format{
			NumChannels: 1,
			SampleRate:  sampleRate,
		},
		Data: make([]int, len(samples)),
	}

	// Convert byte to 16-bit (2 bytes)
	for i, sample := range samples {
		buf.Data[i] = int(sample)
	}

	tempFile, err := os.CreateTemp("", "tts-*.wav")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	enc := wav.NewEncoder(tempFile,
		sampleRate,
		16, // 16-bit PCM
		1,  // mono
		1,
	)

	defer func() {
		_ = os.Remove(tempFile.Name())
	}()

	if err := enc.Write(buf); err != nil {
		_ = enc.Close()
		return nil, err
	}

	_ = enc.Close()
	return os.ReadFile(tempFile.Name())
}
