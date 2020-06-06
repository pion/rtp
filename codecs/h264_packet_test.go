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

func TestH264Packet_Unmarshal(t *testing.T) {
	singlePayload := []byte{0x90, 0x90, 0x90}
	singlePayloadUnmarshaled := []byte{0x00, 0x00, 0x00, 0x01, 0x90, 0x90, 0x90}

	largepayload := []byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15}
	largePayloadPacketized := [][]byte{
		{0x1c, 0x80, 0x01, 0x02, 0x03},
		{0x1c, 0x00, 0x04, 0x05, 0x06},
		{0x1c, 0x00, 0x07, 0x08, 0x09},
		{0x1c, 0x00, 0x10, 0x11, 0x12},
		{0x1c, 0x40, 0x13, 0x14, 0x15},
	}

	singlePayloadMultiNALU := []byte{0x78, 0x00, 0x0f, 0x67, 0x42, 0xc0, 0x1f, 0x1a, 0x32, 0x35, 0x01, 0x40, 0x7a, 0x40, 0x3c, 0x22, 0x11, 0xa8, 0x00, 0x05, 0x68, 0x1a, 0x34, 0xe3, 0xc8}
	singlePayloadMultiNALUUnmarshaled := []byte{0x00, 0x00, 0x00, 0x01, 0x67, 0x42, 0xc0, 0x1f, 0x1a, 0x32, 0x35, 0x01, 0x40, 0x7a, 0x40, 0x3c, 0x22, 0x11, 0xa8, 0x00, 0x00, 0x00, 0x01, 0x68, 0x1a, 0x34, 0xe3, 0xc8}

	incompleteSinglePayloadMultiNALU := []byte{0x78, 0x00, 0x0f, 0x67, 0x42, 0xc0, 0x1f, 0x1a, 0x32, 0x35, 0x01, 0x40, 0x7a, 0x40, 0x3c, 0x22, 0x11}

	pkt := H264Packet{}
	if _, err := pkt.Unmarshal(nil); err == nil {
		t.Fatal("Unmarshal did not fail on nil payload")
	}

	if _, err := pkt.Unmarshal([]byte{0x00, 0x00}); err == nil {
		t.Fatal("Unmarshal accepted a packet that is too small for a payload and header")
	}

	if _, err := pkt.Unmarshal([]byte{0xFF, 0x00, 0x00}); err == nil {
		t.Fatal("Unmarshal accepted a packet with a NALU Type we don't handle")
	}

	if _, err := pkt.Unmarshal(incompleteSinglePayloadMultiNALU); err == nil {
		t.Fatal("Unmarshal accepted a STAP-A packet with insufficient data")
	}

	res, err := pkt.Unmarshal(singlePayload)
	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(res, singlePayloadUnmarshaled) {
		t.Fatal("Unmarshaling a single payload shouldn't modify the payload")
	}

	largePayloadResult := []byte{}
	for i := range largePayloadPacketized {
		res, err = pkt.Unmarshal(largePayloadPacketized[i])
		if err != nil {
			t.Fatal(err)
		}
		largePayloadResult = append(largePayloadResult, res...)
	}
	if !reflect.DeepEqual(largePayloadResult, largepayload) {
		t.Fatal("Failed to unmarshal a large payload")
	}

	res, err = pkt.Unmarshal(singlePayloadMultiNALU)
	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(res, singlePayloadMultiNALUUnmarshaled) {
		t.Fatal("Failed to unmarshal a single packet with multiple NALUs")
	}
}
