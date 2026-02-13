package wav

import (
	"io"

	"github.com/go-audio/wav"
)

func NewDecoder(r io.ReadSeeker) *wav.Decoder {
	return wav.NewDecoder(r)
}
