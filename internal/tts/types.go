package tts

import "os"

type Output struct {
	Content    []byte
	SampleRate int
	Channels   int
	BitDepth   int // Sample bit depth in bits, another name: SampleSize, SampleWidth
}

func NewOutputFromFile(filename string, sampleRate, channels, size int) (*Output, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return NewOutput(content, sampleRate, channels, size)
}

func NewOutput(content []byte, sampleRate, channels, size int) (*Output, error) {
	return &Output{
		Content:    content,
		SampleRate: sampleRate,
		Channels:   channels,
		BitDepth:   size,
	}, nil
}
