package audio

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"math"

	"github.com/hraban/opus"
	"github.com/lxzan/gws"
)

const DefaultSampleRate = 16000
const SampleRate = DefaultSampleRate
const FrameSizeMs = 60
const FrameSize = int(float32(SampleRate) * float32(FrameSizeMs) / 1000)

func StreamPcm(ctx context.Context, source io.ReadSeeker, socket *gws.Conn) error {
	// 3. Encoder initialization
	enc, err := opus.NewEncoder(TargetSampleRate, TargetChannels, opus.AppVoIP)
	if err != nil {
		return err
	}

	packetChan := make(chan []byte, 15) // Decoupling buffer
	errChan := make(chan error, 1)

	// GOROUTINE: The Encoding Engine
	go func() {
		defer close(packetChan)
		byteBuf := make([]byte, SamplesPerFrame*2*TargetChannels)

		for {
			select {
			case <-ctx.Done(): // Stop encoding if context is cancelled
				return
			default:
				_, err := io.ReadFull(source, byteBuf)
				if err != nil {
					if err != io.EOF && !errors.Is(err, io.ErrUnexpectedEOF) {
						errChan <- err
					}
					return
				}

				pcm := pcmPool.Get().([]int16)
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

				// Block until channel is ready or context is done
				select {
				case packetChan <- out[:nBytes]:
				case <-ctx.Done():
					encPool.Put(out)
					return
				}
			}
		}
	}()

	// MAIN LOOP: Bursted & Paced Transmission
	//var ticker *time.Ticker
	count := 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err() // Return manual cancellation or timeout
		case err := <-errChan:
			return err
		case packet, ok := <-packetChan:
			if !ok {
				log.Println("sending done.")
				// Encoder finished and channel drained
				return nil
			}

			// Logic: Burst the first 3 frames (180ms) immediately, then pace the rest
			//if count >= BurstCount {
			//	if ticker == nil {
			//		ticker = time.NewTicker(FrameDurationMs * time.Millisecond)
			//		defer ticker.Stop()
			//	}
			//	<-ticker.C
			//}

			log.Println("sending tts...")
			if err := socket.WriteMessage(gws.OpcodeBinary, packet); err != nil {
				return err
			}

			// Return encoded buffer to pool (gws.WriteMessage copies the data)
			encPool.Put(packet[:cap(packet)])
			count++
		}
	}
}

func BytesToInt16(b []byte) ([]int16, error) {
	// PCM data is typically little-endian. Check your source's documentation if unsure.
	buf := bytes.NewReader(b)
	var result []int16

	for i := 0; i < len(b); i += 2 {
		var sample int16
		// Read a single int16 from the buffer, respecting little-endian byte order
		err := binary.Read(buf, binary.LittleEndian, &sample)
		if err != nil {
			return nil, err
		}
		result = append(result, sample)
	}
	return result, nil
}

// Float32sToBytes converts a slice of float32 to a slice of bytes
// using little-endian byte order.
func Float32sToBytes(floats []float32) []byte {
	// A float32 is 4 b.
	b := make([]byte, len(floats)*4)
	for i, f := range floats {
		// Get the IEEE 754 binary representation of the float as a uint32
		bits := math.Float32bits(f)
		// Put the uint32 into the byte slice with the desired endianness
		binary.LittleEndian.PutUint32(b[i*4:(i+1)*4], bits)
	}

	return b
}
