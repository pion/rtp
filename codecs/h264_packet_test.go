package codecs

import (
	"testing"
)

func TestH264Payloader_Payload(t *testing.T) {
	pck := H264Payloader{}
	payload := []byte{0x90, 0x90, 0x90}

	// Positive MTU, nil payload
	res := pck.Payload(1, nil)
	if len(res) != 0 {
		t.Fatal("Generated payload should be empty")
	}

	// Negative MTU, small payload
	res = pck.Payload(0, payload)
	if len(res) != 0 {
		t.Fatal("Generated payload should be empty")
	}

	// 0 MTU, small payload
	res = pck.Payload(0, payload)
	if len(res) != 0 {
		t.Fatal("Generated payload should be empty")
	}

	// Positive MTU, small payload
	res = pck.Payload(1, payload)
	if len(res) != 0 {
		t.Fatal("Generated payload should be empty")
	}

	// Positive MTU, small payload
	res = pck.Payload(5, payload)
	if len(res) != 1 {
		t.Fatal("Generated payload shouldn't be empty")
	}
	if len(res[0]) != len(payload) {
		t.Fatal("Generated payload should be the same size as original payload size")
	}

	// Nalu type 9 or 12
	res = pck.Payload(5, []byte{0x09, 0x00, 0x00})
	if len(res) != 0 {
		t.Fatal("Generated payload should be empty")
	}
}
