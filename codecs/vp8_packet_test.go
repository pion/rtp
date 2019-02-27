package codecs

import (
	"fmt"
	"testing"

	"github.com/pions/rtp"
)

func TestVP8Packet_Unmarshal(t *testing.T) {
	pck := VP8Packet{}

	errNilPacket := fmt.Errorf("invalid nil packet")
	errSmallerThanHeaderLen := fmt.Errorf("Payload is not large enough to container header")
	errPayloadTooSmall := fmt.Errorf("Payload is not large enough")

	// Nil packet
	raw, err := pck.Unmarshal(nil)
	if raw != nil {
		t.Fatal("Result should be nil in case of error")
	}
	if err == nil || err.Error() != errNilPacket.Error() {
		t.Fatal("Error should be:", errNilPacket)
	}

	// Nil payload
	raw, err = pck.Unmarshal(&rtp.Packet{
		Payload: nil,
	})
	if raw != nil {
		t.Fatal("Result should be nil in case of error")
	}
	if err == nil || err.Error() != errSmallerThanHeaderLen.Error() {
		t.Fatal("Error should be:", errSmallerThanHeaderLen)
	}

	// Payload smaller than header size
	raw, err = pck.Unmarshal(&rtp.Packet{
		Payload: []byte{0x00, 0x11, 0x22},
	})
	if raw != nil {
		t.Fatal("Result should be nil in case of error")
	}
	if err == nil || err.Error() != errSmallerThanHeaderLen.Error() {
		t.Fatal("Error should be:", errSmallerThanHeaderLen)
	}

	// Normal payload
	raw, err = pck.Unmarshal(&rtp.Packet{
		Payload: []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x90},
	})
	if raw == nil {
		t.Fatal("Result shouldn't be nil in case of success")
	}
	if err != nil {
		t.Fatal("Error should be nil in case of success")
	}

	// Header size, only X
	raw, err = pck.Unmarshal(&rtp.Packet{
		Payload: []byte{0x80, 0x00, 0x00, 0x00},
	})
	if raw == nil {
		t.Fatal("Result shouldn't be nil in case of success")
	}
	if err != nil {
		t.Fatal("Error should be nil in case of success")
	}

	// Header size, X and I
	raw, err = pck.Unmarshal(&rtp.Packet{
		Payload: []byte{0x80, 0x80, 0x00, 0x00},
	})
	if raw == nil {
		t.Fatal("Result shouldn't be nil in case of success")
	}
	if err != nil {
		t.Fatal("Error should be nil in case of success")
	}

	// Header size, X and I, PID 16bits
	raw, err = pck.Unmarshal(&rtp.Packet{
		Payload: []byte{0x80, 0x80, 0x81, 0x00},
	})
	if raw != nil {
		t.Fatal("Result should be nil in case of error")
	}
	if err == nil || err.Error() != errPayloadTooSmall.Error() {
		t.Fatal("Error should be:", errPayloadTooSmall)
	}

	// Header size, X and L
	raw, err = pck.Unmarshal(&rtp.Packet{
		Payload: []byte{0x80, 0x40, 0x00, 0x00},
	})
	if raw == nil {
		t.Fatal("Result shouldn't be nil in case of success")
	}
	if err != nil {
		t.Fatal("Error should be nil in case of success")
	}

	// Header size, X and T
	raw, err = pck.Unmarshal(&rtp.Packet{
		Payload: []byte{0x80, 0x20, 0x00, 0x00},
	})
	if raw == nil {
		t.Fatal("Result shouldn't be nil in case of success")
	}
	if err != nil {
		t.Fatal("Error should be nil in case of success")
	}

	// Header size, X and K
	raw, err = pck.Unmarshal(&rtp.Packet{
		Payload: []byte{0x80, 0x10, 0x00, 0x00},
	})
	if raw == nil {
		t.Fatal("Result shouldn't be nil in case of success")
	}
	if err != nil {
		t.Fatal("Error should be nil in case of success")
	}

	// Header size, all flags
	raw, err = pck.Unmarshal(&rtp.Packet{
		Payload: []byte{0xff, 0xff, 0x00, 0x00},
	})
	if raw != nil {
		t.Fatal("Result should be nil in case of error")
	}
	if err == nil || err.Error() != errPayloadTooSmall.Error() {
		t.Fatal("Error should be:", errPayloadTooSmall)
	}
}

func TestVP8Payloader_Payload(t *testing.T) {
	pck := VP8Payloader{}
	payload := []byte{0x90, 0x90, 0x90}

	// Positive MTU, nil payload
	res := pck.Payload(1, nil)
	if len(res) != 0 {
		t.Fatal("Generated payload should be empty")
	}

	// Positive MTU, small payload
	// MTU of 1 results in fragment size of 0
	res = pck.Payload(1, payload)
	if len(res) != 0 {
		t.Fatal("Generated payload should be empty")
	}

	// Negative MTU, small payload
	res = pck.Payload(-1, payload)
	if len(res) != 0 {
		t.Fatal("Generated payload should be empty")
	}

	// Positive MTU, small payload
	res = pck.Payload(2, payload)
	if len(res) != len(payload) {
		t.Fatal("Generated payload should be the same size as original payload size")
	}
}
