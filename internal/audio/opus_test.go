package audio

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-audio/wav"
	"github.com/lxzan/gws"
)

type mockSocket struct {
	packets [][]byte
}

func (m *mockSocket) WriteMessage(opcode gws.Opcode, payload []byte) error {
	// Copy payload because the caller might reuse the buffer (encPool)
	buf := make([]byte, len(payload))
	copy(buf, payload)
	m.packets = append(m.packets, buf)
	return nil
}

func (m *mockSocket) SetDeadline(t time.Time) error {
	return nil
}

func TestStreamOpus(t *testing.T) {
	// Absolute path to the sample file
	projectRoot := "/Users/viet/www/GolandProjects/xiaozhi-hub"
	samplePath := filepath.Join(projectRoot, "sample/genkit.wav")

	f, err := os.Open(samplePath)
	if err != nil {
		t.Skipf("sample file not found: %v", err)
	}
	defer f.Close()

	decoder := wav.NewDecoder(f)
	if !decoder.IsValidFile() {
		t.Fatal("invalid wav file")
	}
	t.Logf("Sample Rate: %d, Channels: %d, Depth: %d",
		decoder.SampleRate, decoder.NumChans, decoder.BitDepth)

	// Reset file for StreamOpus
	f.Seek(0, 0)

	// Give it enough time
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mock := &mockSocket{}

	err = StreamOpus(ctx, f, mock)
	if err != nil {
		t.Fatalf("StreamOpus failed: %v", err)
	}

	t.Logf("Received %d packets", len(mock.packets))

	if len(mock.packets) == 0 {
		t.Error("Expected packets, got 0")
	}
}
