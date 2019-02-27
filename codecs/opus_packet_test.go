package codecs

import (
	"fmt"
	"testing"

	"github.com/pions/rtp"
)

func TestOpusPacket_Unmarshal(t *testing.T) {
	pck := OpusPacket{}

	errNilPacket := fmt.Errorf("invalid nil packet")
	errPayloadTooSmall := fmt.Errorf("Payload is not large enough")

	// Nil packet
	raw, err := pck.Unmarshal(nil)
	if raw != nil {
		t.Fatal("Result should be nil in case of error")
	}
	if err == nil || err.Error() != errNilPacket.Error() {
		t.Fatal("Error should be:", errNilPacket)
	}

	// Empty packet
	raw, err = pck.Unmarshal(&rtp.Packet{
		Payload: nil,
	})
	if raw != nil {
		t.Fatal("Result should be nil in case of error")
	}
	if err == nil || err.Error() != errPayloadTooSmall.Error() {
		t.Fatal("Error should be:", errPayloadTooSmall)
	}

	// Normal packet
	raw, err = pck.Unmarshal(&rtp.Packet{
		Payload: []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x90},
	})
	if raw == nil {
		t.Fatal("Result shouldn't be nil in case of success")
	}
	if err != nil {
		t.Fatal("Error should be nil in case of success")
	}
}

func TestOpusPayloader_Payload(t *testing.T) {
	pck := OpusPayloader{}
	payload := []byte{0x90, 0x90, 0x90}

	// Positive MTU, nil payload
	res := pck.Payload(1, nil)
	if len(res) != 0 {
		t.Fatal("Generated payload should be empty")
	}

	// Note: MTU has no effect
	// Positive MTU, small payload
	res = pck.Payload(1, payload)
	if len(res) != 1 {
		t.Fatal("Generated payload should be the 1")
	}

	// Negative MTU, small payload
	res = pck.Payload(-1, payload)
	if len(res) != 1 {
		t.Fatal("Generated payload should be the 1")
	}

	// Positive MTU, small payload
	res = pck.Payload(2, payload)
	if len(res) != 1 {
		t.Fatal("Generated payload should be the 1")
	}

}
