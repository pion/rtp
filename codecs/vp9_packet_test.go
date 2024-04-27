// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package codecs

import (
	"errors"
	"math/rand"
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
		"NonFlexible": {
			b: []byte{0x00, 0xAA},
			pkt: VP9Packet{
				Payload: []byte{0xAA},
			},
		},
		"NonFlexiblePictureID": {
			b: []byte{0x80, 0x02, 0xAA},
			pkt: VP9Packet{
				I:         true,
				PictureID: 0x02,
				Payload:   []byte{0xAA},
			},
		},
		"NonFlexiblePictureIDExt": {
			b: []byte{0x80, 0x81, 0xFF, 0xAA},
			pkt: VP9Packet{
				I:         true,
				PictureID: 0x01FF,
				Payload:   []byte{0xAA},
			},
		},
		"NonFlexiblePictureIDExt_ShortPacket0": {
			b:   []byte{0x80, 0x81},
			err: errShortPacket,
		},
		"NonFlexiblePictureIDExt_ShortPacket1": {
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
		"FlexibleLayerIndicePictureID": {
			b: []byte{0xB0, 0x02, 0x23, 0x01, 0xAA},
			pkt: VP9Packet{
				F:         true,
				I:         true,
				L:         true,
				PictureID: 0x02,
				TID:       0x01,
				SID:       0x01,
				D:         true,
				Payload:   []byte{0x01, 0xAA},
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
		"ScalabilityStructureResolutionsNoPayload": {
			b: []byte{
				0x0A,
				(1 << 5) | (1 << 4), // NS:1 Y:1 G:0
				640 >> 8, 640 & 0xff,
				360 >> 8, 360 & 0xff,
				1280 >> 8, 1280 & 0xff,
				720 >> 8, 720 & 0xff,
			},
			pkt: VP9Packet{
				B:       true,
				V:       true,
				NS:      1,
				Y:       true,
				G:       false,
				NG:      0,
				Width:   []uint16{640, 1280},
				Height:  []uint16{360, 720},
				Payload: []byte{},
			},
		},
		"ScalabilityStructureNoPayload": {
			b: []byte{
				0x0A,
				(1 << 5) | (0 << 4) | (1 << 3), // NS:1 Y:0 G:1
				2,
				(0 << 5) | (1 << 4) | (0 << 2), // T:0 U:1 R:0 -
				(2 << 5) | (0 << 4) | (1 << 2), // T:2 U:0 R:1 -
				33,
			},
			pkt: VP9Packet{
				B:       true,
				V:       true,
				NS:      1,
				Y:       false,
				G:       true,
				NG:      2,
				PGTID:   []uint8{0, 2},
				PGU:     []bool{true, false},
				PGPDiff: [][]uint8{{}, {33}},
				Payload: []byte{},
			},
		},
		"ScalabilityMissingWidth": {
			b:   []byte("200"),
			err: errShortPacket,
		},
		"ScalabilityMissingNG": {
			b:   []byte("b00200000000"),
			err: errShortPacket,
		},
		"ScalabilityMissingTemporalLayerIDs": {
			b:   []byte("20B0"),
			err: errShortPacket,
		},
		"ScalabilityMissingReferenceIndices": {
			b:   []byte("20B007"),
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
				if !errors.Is(err, c.err) {
					t.Errorf("Error should be '%v', got '%v'", c.err, err)
				}
			}
		})
	}
}

func TestVP9Payloader_Payload(t *testing.T) {
	r0 := int(rand.New(rand.NewSource(0)).Int31n(0x7FFF)) //nolint:gosec
	var rands [][2]byte
	for i := 0; i < 10; i++ {
		rands = append(rands, [2]byte{byte(r0>>8) | 0x80, byte(r0 & 0xFF)})
		r0++
	}

	cases := map[string]struct {
		b        [][]byte
		flexible bool
		mtu      uint16
		res      [][]byte
	}{
		"flexible NilPayload": {
			b:        [][]byte{nil},
			flexible: true,
			mtu:      100,
			res:      [][]byte{},
		},
		"flexible SmallMTU": {
			b:        [][]byte{{0x00, 0x00}},
			flexible: true,
			mtu:      1,
			res:      [][]byte{},
		},
		"flexible OnePacket": {
			b:        [][]byte{{0x01, 0x02}},
			flexible: true,
			mtu:      10,
			res: [][]byte{
				{0x9C, rands[0][0], rands[0][1], 0x01, 0x02},
			},
		},
		"flexible TwoPackets": {
			b:        [][]byte{{0x01, 0x02}},
			flexible: true,
			mtu:      4,
			res: [][]byte{
				{0x98, rands[0][0], rands[0][1], 0x01},
				{0x94, rands[0][0], rands[0][1], 0x02},
			},
		},
		"flexible ThreePackets": {
			b:        [][]byte{{0x01, 0x02, 0x03}},
			flexible: true,
			mtu:      4,
			res: [][]byte{
				{0x98, rands[0][0], rands[0][1], 0x01},
				{0x90, rands[0][0], rands[0][1], 0x02},
				{0x94, rands[0][0], rands[0][1], 0x03},
			},
		},
		"flexible TwoFramesFourPackets": {
			b:        [][]byte{{0x01, 0x02, 0x03}, {0x04}},
			flexible: true,
			mtu:      5,
			res: [][]byte{
				{0x98, rands[0][0], rands[0][1], 0x01, 0x02},
				{0x94, rands[0][0], rands[0][1], 0x03},
				{0x9C, rands[1][0], rands[1][1], 0x04},
			},
		},
		"non-flexible NilPayload": {
			b:   [][]byte{nil},
			mtu: 100,
			res: [][]byte{},
		},
		"non-flexible SmallMTU": {
			b:   [][]byte{{0x82, 0x49, 0x83, 0x42, 0x0, 0x77, 0xf0, 0x32, 0x34}},
			mtu: 1,
			res: [][]byte{},
		},
		"non-flexible OnePacket key frame": {
			b:   [][]byte{{0x82, 0x49, 0x83, 0x42, 0x0, 0x77, 0xf0, 0x32, 0x34}},
			mtu: 20,
			res: [][]byte{{
				0x8f, 0xa1, 0xf4, 0x18, 0x07, 0x80, 0x03, 0x24,
				0x01, 0x14, 0x01, 0x82, 0x49, 0x83, 0x42, 0x00,
				0x77, 0xf0, 0x32, 0x34,
			}},
		},
		"non-flexible TwoPackets key frame": {
			b:   [][]byte{{0x82, 0x49, 0x83, 0x42, 0x0, 0x77, 0xf0, 0x32, 0x34}},
			mtu: 12,
			res: [][]byte{
				{
					0x8b, 0xa1, 0xf4, 0x18, 0x07, 0x80, 0x03, 0x24,
					0x01, 0x14, 0x01, 0x82,
				},
				{
					0x85, 0xa1, 0xf4, 0x49, 0x83, 0x42, 0x00, 0x77,
					0xf0, 0x32, 0x34,
				},
			},
		},
		"non-flexible ThreePackets key frame": {
			b: [][]byte{{
				0x82, 0x49, 0x83, 0x42, 0x00, 0x77, 0xf0, 0x32,
				0x34, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
				0x08,
			}},
			mtu: 12,
			res: [][]byte{
				{
					0x8b, 0xa1, 0xf4, 0x18, 0x07, 0x80, 0x03, 0x24,
					0x01, 0x14, 0x01, 0x82,
				},
				{
					0x81, 0xa1, 0xf4, 0x49, 0x83, 0x42, 0x00, 0x77,
					0xf0, 0x32, 0x34, 0x01,
				},
				{
					0x85, 0xa1, 0xf4, 0x02, 0x03, 0x04, 0x05, 0x06,
					0x07, 0x08,
				},
			},
		},
		"non-flexible OnePacket non key frame": {
			b:   [][]byte{{0x86, 0x0, 0x40, 0x92, 0xe1, 0x31, 0x42, 0x8c, 0xc0, 0x40}},
			mtu: 20,
			res: [][]byte{{
				0xcd, 0xa1, 0xf4, 0x86, 0x00, 0x40, 0x92, 0xe1,
				0x31, 0x42, 0x8c, 0xc0, 0x40,
			}},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			pck := VP9Payloader{
				FlexibleMode: c.flexible,
				InitialPictureIDFn: func() uint16 {
					return uint16(rand.New(rand.NewSource(0)).Int31n(0x7FFF)) //nolint:gosec
				},
			}

			res := [][]byte{}
			for _, b := range c.b {
				res = append(res, pck.Payload(c.mtu, b)...)
			}
			if !reflect.DeepEqual(c.res, res) {
				t.Errorf("Payloaded packet expected to be:\n %v\ngot:\n %v", c.res, res)
			}
		})
	}

	t.Run("PictureIDOverflow", func(t *testing.T) {
		pck := VP9Payloader{
			FlexibleMode: true,
			InitialPictureIDFn: func() uint16 {
				return uint16(rand.New(rand.NewSource(0)).Int31n(0x7FFF)) //nolint:gosec
			},
		}
		pPrev := VP9Packet{}
		for i := 0; i < 0x8000; i++ {
			res := pck.Payload(4, []byte{0x01})
			p := VP9Packet{}
			_, err := p.Unmarshal(res[0])
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if i > 0 {
				if pPrev.PictureID == 0x7FFF {
					if p.PictureID != 0 {
						t.Errorf("Picture ID next to 0x7FFF must be 0, got %d", p.PictureID)
					}
				} else if pPrev.PictureID+1 != p.PictureID {
					t.Errorf("Picture ID next must be incremented by 1: %d -> %d", pPrev.PictureID, p.PictureID)
				}
			}

			pPrev = p
		}
	})
}

func TestVP9IsPartitionHead(t *testing.T) {
	vp9 := &VP9Packet{}
	t.Run("SmallPacket", func(t *testing.T) {
		if vp9.IsPartitionHead([]byte{}) {
			t.Fatal("Small packet should not be the head of a new partition")
		}
	})
	t.Run("NormalPacket", func(t *testing.T) {
		if !vp9.IsPartitionHead([]byte{0x18, 0x00, 0x00}) {
			t.Error("VP9 RTP packet with B flag should be head of a new partition")
		}
		if vp9.IsPartitionHead([]byte{0x10, 0x00, 0x00}) {
			t.Error("VP9 RTP packet without B flag should not be head of a new partition")
		}
	})
}
