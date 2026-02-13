package asr

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/hraban/opus"
	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
	"github.com/phamviet/xiaozhi-hub/internal/audio"
)

const SampleRate = 16000
const FrameSizeMs = 60
const FrameSize = int(float32(SampleRate) * float32(FrameSizeMs) / 1000)

type Asr struct {
	result     chan string
	speechChan chan *sherpa.GeneratedAudio
	sampleRate int
	decoder    *opus.Decoder
	vad        *sherpa.VoiceActivityDetector
	buffer     *sherpa.CircularBuffer
	stt        SpeechToText
	stopChan   chan struct{}
	started    bool

	logger *slog.Logger
}

type Option func(asr *Asr)

func WithLogger(logger *slog.Logger) Option {
	return func(asr *Asr) {
		asr.logger = logger
	}
}

func NewAsr(opts ...Option) (*Asr, error) {
	decoder, err := opus.NewDecoder(16000, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to create opus decoder: %w", err)
	}

	vad := NewVad()
	if vad == nil {
		return nil, fmt.Errorf("failed to create VAD")
	}
	buffer := sherpa.NewCircularBuffer(10 * SampleRate)
	a := &Asr{
		decoder:    decoder,
		vad:        vad,
		buffer:     buffer,
		stt:        NewOpenAi(""),
		sampleRate: SampleRate,
		speechChan: make(chan *sherpa.GeneratedAudio, 100),
		result:     make(chan string),
		stopChan:   make(chan struct{}),
		logger:     slog.Default(),
	}

	for _, opt := range opts {
		opt(a)
	}

	return a, nil
}

func (a *Asr) Write(data []byte) error {
	if !a.started {
		return fmt.Errorf("ASR not started")
	}

	select {
	case <-a.stopChan:
		return nil
	default:
		if len(data) == 0 {
			return fmt.Errorf("empty audio data")
		}
		a.decodeAudio(data)
	}

	return nil
}

func (a *Asr) decodeAudio(data []byte) {
	buf := make([]float32, FrameSize)
	n, err := a.decoder.DecodeFloat32(data, buf)

	if err != nil {
		a.logger.Error("Failed to decode opus data: %v", err)
		return
	}

	if n == 0 {
		a.logger.Warn("Empty audio data in frame, skipping")
		return
	}

	samples := buf[:n]
	a.buffer.Push(samples)
	windowSize := 512

	for a.buffer.Size() >= windowSize {
		head := a.buffer.Head()
		s := a.buffer.Get(head, windowSize)
		a.buffer.Pop(windowSize)

		a.vad.AcceptWaveform(s)

		for !a.vad.IsEmpty() {
			speechSegment := a.vad.Front()
			a.vad.Pop()

			duration := float32(len(speechSegment.Samples)) / float32(a.sampleRate)
			a.logger.Debug(fmt.Sprintf("Detected speech. Duration: %.2f seconds", duration))

			speech := sherpa.GeneratedAudio{
				Samples:    speechSegment.Samples,
				SampleRate: a.sampleRate,
			}

			select {
			case a.speechChan <- &speech:
			default:
				a.logger.Warn("Speech channel full, dropping speech segment")
			}
		}
	}

}

func (a *Asr) processSpeechChan() {
	for speech := range a.speechChan {
		select {
		case <-a.stopChan:
			return
		default:
			wavBytes, err := audio.Float32ToWavBytes(speech.Samples, speech.SampleRate)
			if err != nil {
				a.logger.Error("Failed to read temp WAV file", "error", err)
				continue
			}

			text, err := a.stt.Transcribe(wavBytes)
			if err != nil {
				a.logger.Error("Failed to transcribe speech", "error", err)
				continue
			}
			a.logger.Info(fmt.Sprintf("Transcribed speech: %s", text))
			a.result <- strings.TrimSpace(text)
		}
	}
}

func (a *Asr) Start() {
	if a.started {
		return
	}

	a.stopChan = make(chan struct{})
	a.speechChan = make(chan *sherpa.GeneratedAudio, 100)
	a.started = true
	go a.processSpeechChan()
}

func (a *Asr) Stop() {
	if !a.started {
		return
	}

	a.started = false
	close(a.stopChan)
	close(a.speechChan)

	//for range a.result {
	//}
	a.logger.Debug("asr.Stop")
	a.vad.Clear()
	// sherpa.DeleteVoiceActivityDetector(vad)
}

func (a *Asr) Result() <-chan string {
	return a.result
}
