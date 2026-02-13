package audio

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"sync"
	"time"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/hraban/opus"
	"github.com/lxzan/gws"
	"github.com/zaf/resample"
)

type MessageWriter interface {
	SetDeadline(t time.Time) error
	WriteMessage(opcode gws.Opcode, payload []byte) error
}

const (
	TargetSampleRate = 16000
	TargetChannels   = 1
	FrameDurationMs  = 60
	// 16000 samples/sec * 0.060 sec = 960 samples
	SamplesPerFrame = (TargetSampleRate * FrameDurationMs) / 1000
	BurstCount      = 3
)

// Reusable pools for high-frequency allocations
var (
	// Pool for the raw PCM samples (960 samples * 1 channel)
	pcmPool = sync.Pool{
		New: func() interface{} { return make([]int16, SamplesPerFrame*TargetChannels) },
	}
	// Pool for the compressed Opus output
	encPool = sync.Pool{
		New: func() interface{} { return make([]byte, 1024) },
	}
)

func StreamOpus(ctx context.Context, r io.ReadSeeker, socket MessageWriter) error {
	decoder := wav.NewDecoder(r)
	if !decoder.IsValidFile() {
		return log.Output(2, "invalid wav header")
	}

	duration, _ := decoder.Duration()
	// 1. Setup the Resampling Pipe
	// We read from the pipeReader, the resampler writes to the pipeWriter
	pr, pw := io.Pipe()

	log.Printf("Resampling from %dHz to %dHz, %d channels", decoder.SampleRate, TargetSampleRate, TargetChannels)
	res, err := resample.New(pw, float64(decoder.SampleRate), float64(TargetSampleRate), TargetChannels, resample.I16, resample.HighQ)
	if err != nil {
		pw.Close()
		return err
	}

	packetChan := make(chan []byte, 15) // Decoupling buffer
	errChan := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(2)

	// --- GOROUTINE A: The Feeder (WAV -> Downmix -> Resampler) ---
	go func() {
		defer wg.Done()
		defer pw.Close()  // 2. Closes the pipe, sending EOF to Encoder
		defer res.Close() // 1. Flushes resampler buffer to pw

		// Buffer for reading raw WAV samples (ints)
		// Read 1024 frames at a time
		numChans := decoder.Format().NumChannels
		buf := &audio.IntBuffer{Data: make([]int, 1024*numChans), Format: decoder.Format()}

		// Buffer for the converted bytes to send to Resampler
		// We are outputting TargetChannels (1), 2 bytes per sample
		rawBytes := make([]byte, 1024*TargetChannels*2)

		for {
			n, err := decoder.PCMBuffer(buf)
			if n == 0 {
				if err == nil {
					log.Println("Feeder: Read 0 samples with nil error (forcing EOF)")
					break
				}
			}
			if n > 0 {
				// log.Printf("Feeder: Read %d samples", n)
				// Calculate how many full frames we have
				frames := n / numChans

				// Downmix and Convert to Bytes
				for i := 0; i < frames; i++ {
					// Simple Downmix: Average all channels to get 1 Mono sample
					sum := 0
					for ch := 0; ch < numChans; ch++ {
						sum += buf.Data[i*numChans+ch]
					}
					avgSample := sum / numChans // This is now a Mono sample

					// Write 2 bytes (Int16 Little Endian) to our byte buffer
					// i*2 because each sample is 2 bytes
					binary.LittleEndian.PutUint16(rawBytes[i*2:], uint16(int16(avgSample)))
				}

				// WRITE ONCE PER CHUNK (Performance Fix)
				// We only write the bytes we actually filled (frames * 2 bytes)
				if _, writeErr := res.Write(rawBytes[:frames*2]); writeErr != nil {
					// Can't continue if we can't write to the pipe
					log.Printf("resampler write error: %v", writeErr)
					return
				}
			}
			if err != nil {
				if err != io.EOF {
					log.Printf("PCM buffer read error: %v", err)
				} else {
					log.Println("Feeder: EOF reached")
				}
				return
			}
		}
	}()

	// --- GOROUTINE B: The Encoder (Pipe -> Opus) ---
	go func() {
		defer wg.Done()
		defer pr.Close()

		byteBuf := make([]byte, SamplesPerFrame*2*TargetChannels) // 960 * 2 = 1920 bytes
		enc, _ := opus.NewEncoder(TargetSampleRate, TargetChannels, opus.AppVoIP)

		for {
			// Read exactly one 60ms frame from the resampler pipe
			n, err := io.ReadFull(pr, byteBuf)
			isLastFrame := false

			if err != nil {
				if err == io.EOF {
					log.Println("Encoder: EOF reached")
					return
				}
				if errors.Is(err, io.ErrUnexpectedEOF) {
					log.Printf("Encoder: Partial frame at end: %d bytes (padding to full frame)", n)
					// Pad with silence (0)
					for i := n; i < len(byteBuf); i++ {
						byteBuf[i] = 0
					}
					isLastFrame = true
				} else {
					log.Printf("Encoder: Read error: %v", err)
					errChan <- err
					return
				}
			}

			pcm := pcmPool.Get().([]int16)
			// Convert bytes back to Int16 for Opus
			for i := 0; i < SamplesPerFrame; i++ {
				pcm[i] = int16(binary.LittleEndian.Uint16(byteBuf[i*2 : i*2+2]))
			}

			out := encPool.Get().([]byte)
			nBytes, err := enc.Encode(pcm, out)
			pcmPool.Put(pcm)

			if err != nil {
				encPool.Put(out)
				errChan <- err
				return
			}

			select {
			case packetChan <- out[:nBytes]:
			case <-ctx.Done():
				encPool.Put(out)
				return
			}

			if isLastFrame {
				log.Println("Encoder: Sent last partial frame")
				return
			}
		}
	}()

	// Goroutine C: The closer
	// Waits for both the feeder and encoder to finish, then closes the packetChan
	go func() {
		wg.Wait()
		close(packetChan)
	}()

	// --- MAIN LOOP: Pacing & Sending ---
	start := time.Now()

	defer func() {
		elapsed := time.Since(start)
		sleepTime := duration - elapsed
		if sleepTime > 0 {
			// Wait for the remaining audio duration to ensure client receives/plays everything
			log.Printf("Sender: sleeping for %v to allow playback completion", sleepTime)
			time.Sleep(sleepTime)
		}
	}()

	return runSender(ctx, socket, packetChan, errChan)
}

func runSender(ctx context.Context, socket MessageWriter, packetChan chan []byte, errChan chan error) error {
	var ticker *time.Ticker
	count := 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errChan:
			return err
		case packet, ok := <-packetChan:
			if !ok {
				//log.Println("Sender: packetChan closed, done")
				time.Sleep(time.Duration(BurstCount*FrameDurationMs) * time.Millisecond)
				return nil
			}

			if count >= BurstCount {
				if ticker == nil {
					ticker = time.NewTicker(FrameDurationMs * time.Millisecond)
					defer ticker.Stop()
				}
				select {
				case <-ticker.C:
				case <-ctx.Done():
					return ctx.Err()
				}
			}

			// log.Println("send tts...")
			if err := socket.SetDeadline(time.Now().Add(30 * time.Second)); err != nil {
				// Check if this is a "use of closed network connection" error
				// If so, the connection was already closed by peer
				log.Printf("Sender: deadline error (connection may be closed): %v\n", err)
				return err
			}

			if err := socket.WriteMessage(gws.OpcodeBinary, packet); err != nil {
				log.Printf("Sender: failed to write message: %v\n", err)
				return err
			}
			encPool.Put(packet[:cap(packet)])
			count++
		}
	}
}
