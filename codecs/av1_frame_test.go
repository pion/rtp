package codecs

import (
	"reflect"
	"testing"
)

// First is Fragment (and no buffer)
// Self contained OBU
// OBU spread across 3 packets
func TestAV1_ReadFrames(t *testing.T) {
	// First is Fragment of OBU, but no OBU Elements is cached
	f := &AV1Frame{}
	frames, err := f.ReadFrames(&AV1Packet{Z: true, OBUElements: [][]byte{{0x01}}})
	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(frames, [][]byte{}) {
		t.Fatalf("No frames should be generated, %v", frames)
	}

	f = &AV1Frame{}
	frames, err = f.ReadFrames(&AV1Packet{OBUElements: [][]byte{{0x01}}})
	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(frames, [][]byte{{0x01}}) {
		t.Fatalf("One frame should be generated, %v", frames)
	}

	f = &AV1Frame{}
	frames, err = f.ReadFrames(&AV1Packet{Y: true, OBUElements: [][]byte{{0x00}}})
	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(frames, [][]byte{}) {
		t.Fatalf("No frames should be generated, %v", frames)
	}

	frames, err = f.ReadFrames(&AV1Packet{Z: true, OBUElements: [][]byte{{0x01}}})
	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(frames, [][]byte{{0x00, 0x01}}) {
		t.Fatalf("One frame should be generated, %v", frames)
	}
}

// Marshal some AV1 Frames to RTP, assert that AV1Frame can get them back in the original format
func TestAV1_ReadFrames_E2E(t *testing.T) {
	const mtu = 1500
	frames := [][]byte{
		{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A},
		{0x00, 0x01},
		{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A},
		{0x00, 0x01},
	}

	frames = append(frames, []byte{})
	for i := 0; i <= 5; i++ {
		frames[len(frames)-1] = append(frames[len(frames)-1], []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}...)
	}

	frames = append(frames, []byte{})
	for i := 0; i <= 500; i++ {
		frames[len(frames)-1] = append(frames[len(frames)-1], []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}...)
	}

	payloader := &AV1Payloader{}
	f := &AV1Frame{}
	for _, originalFrame := range frames {
		for _, payload := range payloader.Payload(mtu, originalFrame) {
			rtpPacket := &AV1Packet{}
			if _, err := rtpPacket.Unmarshal(payload); err != nil {
				t.Fatal(err)
			}
			decodedFrame, err := f.ReadFrames(rtpPacket)
			if err != nil {
				t.Fatal(err)
			} else if len(decodedFrame) != 0 && !reflect.DeepEqual(originalFrame, decodedFrame[0]) {
				t.Fatalf("Decode(%02x) and Original(%02x) are not equal", decodedFrame[0], originalFrame)
			}
		}
	}
}
