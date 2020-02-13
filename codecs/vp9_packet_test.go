package codecs

import (
	"reflect"
	"testing"
)

func TestVP9Packet_Unmarshal(t *testing.T) {
	cases := map[string]struct {
		b   []byte
		pkt VP9Packet
		err error
	}{
		"Nil": {
			b:   nil,
			err: errNilPacket,
		},
		"Empty": {
			b:   []byte{},
			err: errShortPacket,
		},
		"Flexible": {
			b: []byte{0x00, 0x01, 0xAA},
			pkt: VP9Packet{
				TL0PICIDX: 0x01,
				Payload:   []byte{0xAA},
			},
		},
		"NonFlexiblePictureID": {
			b: []byte{0x80, 0x02, 0x01, 0xAA},
			pkt: VP9Packet{
				I:         true,
				PictureID: 0x02,
				TL0PICIDX: 0x01,
				Payload:   []byte{0xAA},
			},
		},
		"NonFlexiblePictureIDExt": {
			b: []byte{0x80, 0x81, 0xFF, 0x01, 0xAA},
			pkt: VP9Packet{
				I:         true,
				PictureID: 0x01FF,
				TL0PICIDX: 0x01,
				Payload:   []byte{0xAA},
			},
		},
		"NonFlexiblePictureIDExt_ShortPacket0": {
			b:   []byte{0x80, 0x81, 0xFF},
			err: errShortPacket,
		},
		"NonFlexiblePictureIDExt_ShortPacket1": {
			b:   []byte{0x80, 0x81},
			err: errShortPacket,
		},
		"NonFlexiblePictureIDExt_ShortPacket2": {
			b:   []byte{0x80},
			err: errShortPacket,
		},
		"NonFlexibleLayerIndicePictureID": {
			b: []byte{0xA0, 0x02, 0x23, 0x01, 0xAA},
			pkt: VP9Packet{
				I:         true,
				L:         true,
				PictureID: 0x02,
				TID:       0x01,
				SID:       0x01,
				D:         true,
				TL0PICIDX: 0x01,
				Payload:   []byte{0xAA},
			},
		},
		"NonFlexibleLayerIndicePictureID_ShortPacket0": {
			b:   []byte{0xA0, 0x02, 0x23},
			err: errShortPacket,
		},
		"NonFlexibleLayerIndicePictureID_ShortPacket1": {
			b:   []byte{0xA0, 0x02},
			err: errShortPacket,
		},
		"FlexiblePictureIDRefIndex": {
			b: []byte{0xD0, 0x02, 0x03, 0x04, 0xAA},
			pkt: VP9Packet{
				I:         true,
				P:         true,
				F:         true,
				PictureID: 0x02,
				PDiff:     []uint8{0x01, 0x02},
				Payload:   []byte{0xAA},
			},
		},
		"FlexiblePictureIDRefIndex_TooManyPDiff": {
			b:   []byte{0xD0, 0x02, 0x03, 0x05, 0x07, 0x09, 0x10, 0xAA},
			err: errTooManyPDiff,
		},
		"FlexiblePictureIDRefIndexNoPayload": {
			b: []byte{0xD0, 0x02, 0x03, 0x04},
			pkt: VP9Packet{
				I:         true,
				P:         true,
				F:         true,
				PictureID: 0x02,
				PDiff:     []uint8{0x01, 0x02},
				Payload:   []byte{},
			},
		},
		"FlexiblePictureIDRefIndex_ShortPacket0": {
			b:   []byte{0xD0, 0x02, 0x03},
			err: errShortPacket,
		},
		"FlexiblePictureIDRefIndex_ShortPacket1": {
			b:   []byte{0xD0, 0x02},
			err: errShortPacket,
		},
		"FlexiblePictureIDRefIndex_ShortPacket2": {
			b:   []byte{0xD0},
			err: errShortPacket,
		},
	}
	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			p := VP9Packet{}
			raw, err := p.Unmarshal(c.b)
			if c.err == nil {
				if raw == nil {
					t.Error("Result shouldn't be nil in case of success")
				}
				if err != nil {
					t.Error("Error should be nil in case of success")
				}
				if !reflect.DeepEqual(c.pkt, p) {
					t.Errorf("Unmarshalled packet expected to be:\n %v\ngot:\n %v", c.pkt, p)
				}
			} else {
				if raw != nil {
					t.Error("Result should be nil in case of error")
				}
				if err != c.err {
					t.Errorf("Error should be '%v', got '%v'", c.err, err)
				}
			}
		})
	}
}

func TestVP9Payloader_Payload(t *testing.T) {
	pck := VP9Payloader{}
	payload := []byte{0x90, 0x90, 0x90}

	// Positive MTU, nil payload
	res := pck.Payload(1, nil)
	if len(res) != 0 {
		t.Fatal("Generated payload should be empty")
	}

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

func TestVP9PartitionHeadChecker_IsPartitionHead(t *testing.T) {
	checker := &VP9PartitionHeadChecker{}
	t.Run("SmallPacket", func(t *testing.T) {
		if checker.IsPartitionHead([]byte{}) {
			t.Fatal("Small packet should not be the head of a new partition")
		}
	})
	t.Run("NormalPacket", func(t *testing.T) {
		if !checker.IsPartitionHead([]byte{0x18, 0x00, 0x00}) {
			t.Error("VP9 RTP packet with B flag should be head of a new partition")
		}
		if checker.IsPartitionHead([]byte{0x10, 0x00, 0x00}) {
			t.Error("VP9 RTP packet without B flag should not be head of a new partition")
		}
	})
}
