package codecs

import (
	"reflect"
	"testing"
)

func TestH264Payloader_Payload(t *testing.T) {
	pck := H264Payloader{}
	smallpayload := []byte{0x90, 0x90, 0x90}
	multiplepayload := []byte{0x00, 0x00, 0x01, 0x90, 0x00, 0x00, 0x01, 0x90}

	largepayload := []byte{0x00, 0x00, 0x01, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15}
	largePayloadPacketized := [][]byte{
		{0x1c, 0x80, 0x01, 0x02, 0x03},
		{0x1c, 0x00, 0x04, 0x05, 0x06},
		{0x1c, 0x00, 0x07, 0x08, 0x09},
		{0x1c, 0x00, 0x10, 0x11, 0x12},
		{0x1c, 0x40, 0x13, 0x14, 0x15},
	}

	// Positive MTU, nil payload
	res := pck.Payload(1, nil)
	if len(res) != 0 {
		t.Fatal("Generated payload should be empty")
	}

	// Negative MTU, small payload
	res = pck.Payload(0, smallpayload)
	if len(res) != 0 {
		t.Fatal("Generated payload should be empty")
	}

	// 0 MTU, small payload
	res = pck.Payload(0, smallpayload)
	if len(res) != 0 {
		t.Fatal("Generated payload should be empty")
	}

	// Positive MTU, small payload
	res = pck.Payload(1, smallpayload)
	if len(res) != 0 {
		t.Fatal("Generated payload should be empty")
	}

	// Positive MTU, small payload
	res = pck.Payload(5, smallpayload)
	if len(res) != 1 {
		t.Fatal("Generated payload shouldn't be empty")
	}
	if len(res[0]) != len(smallpayload) {
		t.Fatal("Generated payload should be the same size as original payload size")
	}

	// Multiple NALU in a single payload
	res = pck.Payload(5, multiplepayload)
	if len(res) != 2 {
		t.Fatal("2 nal units should be broken out")
	}
	for i := 0; i < 2; i++ {
		if len(res[i]) != 1 {
			t.Fatalf("Payload %d of 2 is packed incorrectly", i+1)
		}
	}

	// Large Payload split across multiple RTP Packets
	res = pck.Payload(5, largepayload)
	if !reflect.DeepEqual(res, largePayloadPacketized) {
		t.Fatal("FU-A packetization failed")
	}

	// Nalu type 9 or 12
	res = pck.Payload(5, []byte{0x09, 0x00, 0x00})
	if len(res) != 0 {
		t.Fatal("Generated payload should be empty")
	}
}
