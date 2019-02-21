package codecs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestH264Payloader_Payload(t *testing.T) {
	assert := assert.New(t)

	pck := H264Payloader{}
	payload := []byte{0x90, 0x90, 0x90}

	// Positive MTU, nil payload
	res := pck.Payload(1, nil)
	assert.Len(res, 0, "Generated payload should be empty")

	// Negative MTU, small payload
	res = pck.Payload(0, payload)
	assert.Len(res, 0, "Generated payload should be empty")

	// 0 MTU, small payload
	res = pck.Payload(0, payload)
	assert.Len(res, 0, "Generated payload should be empty")

	// Positive MTU, small payload
	res = pck.Payload(1, payload)
	assert.Len(res, 0, "Generated payload should be empty")

	// Positive MTU, small payload
	res = pck.Payload(5, payload)
	assert.Len(res, 1, "Generated payload shouldn't be empty")
	assert.Len(res[0], len(payload), "Generated payload should be the same size as original payload size")

	// Nalu type 9 or 12
	res = pck.Payload(5, []byte{0x09, 0x00, 0x00})
	assert.Len(res, 0, "Generated payload should be empty")
}

func TestNextNALU(t *testing.T) {
	assert := assert.New(t)

	input := []byte{
		0x00, 0x00, 0x00, 0x01, 0x67, 0x42, 0x00, 0x1f, 0x01,
		0x00, 0x00, 0x01, 0x80, 0x00, 0x00, 0x03, 0x00, 0x7a,
		0x00, 0x00, 0x00, 0x01, 0xab, 0xbc, 0xef,
	}

	// The first NALU is empty to trim off the start code.
	nalu, input := nextNALU(input)
	assert.Len(nalu, 0)

	nalu, input = nextNALU(input)
	assert.Equal([]byte{0x67, 0x42, 0x00, 0x1f, 0x01}, nalu)

	nalu, input = nextNALU(input)
	assert.Equal([]byte{0x80, 0x00, 0x00, 0x03, 0x00, 0x7a}, nalu)

	nalu, input = nextNALU(input)
	assert.Equal([]byte{0xab, 0xbc, 0xef}, nalu)

	assert.Len(input, 0)
}

func BenchmarkNextNALU(b *testing.B) {
	input := []byte{
		0x00, 0x00, 0x00, 0x01, 0x67, 0x42, 0x00, 0x1f, 0x01,
		0x00, 0x00, 0x01, 0x80, 0x00, 0x00, 0x03, 0x00, 0x7a,
		0x00, 0x00, 0x00, 0x01, 0xab, 0xbc, 0xef,
	}

	for i := 0; i < b.N; i++ {
		remaining := input

		for len(remaining) > 0 {
			_, remaining = nextNALU(remaining)
		}
	}
}
