// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package codecs

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/pion/rtp/codecs/av1/obu"
)

func TestAV1_Marshal(t *testing.T) { // nolint:funlen,cyclop
	payloader := &AV1Payloader{}

	t.Run("Unfragmented OBU", func(t *testing.T) {
		OBU := []byte{0x00, 0x01, 0x2, 0x3, 0x4, 0x5}
		payloads := payloader.Payload(100, OBU)

		if len(payloads) != 1 || len(payloads[0]) != 7 {
			t.Fatal("Expected one unfragmented Payload")
		}

		if payloads[0][0] != 0x10 {
			t.Fatal("Only W bit should be set")
		}

		if !bytes.Equal(OBU, payloads[0][1:]) {
			t.Fatal("OBU modified during packetization")
		}
	})

	t.Run("Fragmented OBU", func(t *testing.T) {
		OBU := []byte{0x00, 0x01, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8}
		payloads := payloader.Payload(4, OBU)

		if len(payloads) != 3 || len(payloads[0]) != 4 || len(payloads[1]) != 4 || len(payloads[2]) != 4 {
			t.Fatal("Expected three fragmented Payload")
		}

		if payloads[0][0] != 0x10|yMask {
			t.Fatal("W and Y bit should be set")
		}

		if payloads[1][0] != 0x10|yMask|zMask {
			t.Fatal("W, Y and Z bit should be set")
		}

		if payloads[2][0] != 0x10|zMask {
			t.Fatal("W and Z bit should be set")
		}

		if !bytes.Equal(OBU[0:3], payloads[0][1:]) ||
			!bytes.Equal(OBU[3:6], payloads[1][1:]) ||
			!bytes.Equal(OBU[6:9], payloads[2][1:]) {
			t.Fatal("OBU modified during packetization")
		}
	})

	t.Run("Sequence Header Caching", func(t *testing.T) {
		sequenceHeaderFrame := []byte{0xb, 0xA, 0xB, 0xC}
		normalFrame := []byte{0x0, 0x1, 0x2, 0x3}

		payloads := payloader.Payload(100, sequenceHeaderFrame)
		if len(payloads) != 0 {
			t.Fatal("Sequence Header was not properly cached")
		}

		payloads = payloader.Payload(100, normalFrame)
		if len(payloads) != 1 {
			t.Fatal("Expected one payload")
		}

		if payloads[0][0] != 0x20|nMask {
			t.Fatal("W and N bit should be set")
		}

		if !bytes.Equal(sequenceHeaderFrame, payloads[0][2:6]) || !bytes.Equal(normalFrame, payloads[0][6:10]) {
			t.Fatal("OBU modified during packetization")
		}
	})
}

func TestAV1_Unmarshal_Error(t *testing.T) {
	for _, test := range []struct {
		expectedError error
		input         []byte
	}{
		{errNilPacket, nil},
		{errShortPacket, []byte{0x00}},
		{errIsKeyframeAndFragment, []byte{byte(0b10001000), 0x00}},
		{obu.ErrFailedToReadLEB128, []byte{byte(0b10000000), 0xFF, 0xFF}},
		{errShortPacket, []byte{byte(0b10000000), 0xFF, 0x0F, 0x00, 0x00}},
	} {
		test := test
		av1Pkt := &AV1Packet{}

		if _, err := av1Pkt.Unmarshal(test.input); !errors.Is(err, test.expectedError) {
			t.Fatalf("Expected error(%s) but got (%s)", test.expectedError, err)
		}
	}
}

func TestAV1_Unmarshal(t *testing.T) { // nolint: funlen
	// nolint: dupl
	av1Payload := []byte{
		0x68, 0x0c, 0x08, 0x00, 0x00, 0x00, 0x2c,
		0xd6, 0xd3, 0x0c, 0xd5, 0x02, 0x00, 0x80,
		0x30, 0x10, 0xc3, 0xc0, 0x07, 0xff, 0xff,
		0xf8, 0xb7, 0x30, 0xc0, 0x00, 0x00, 0x88,
		0x17, 0xf9, 0x0c, 0xcf, 0xc6, 0x7b, 0x9c,
		0x0d, 0xda, 0x55, 0x82, 0x82, 0x67, 0x2f,
		0xf0, 0x07, 0x26, 0x5d, 0xf6, 0xc6, 0xe3,
		0x12, 0xdd, 0xf9, 0x71, 0x77, 0x43, 0xe6,
		0xba, 0xf2, 0xce, 0x36, 0x08, 0x63, 0x92,
		0xac, 0xbb, 0xbd, 0x26, 0x4c, 0x05, 0x52,
		0x91, 0x09, 0xf5, 0x37, 0xb5, 0x18, 0xbe,
		0x5c, 0x95, 0xb1, 0x2c, 0x13, 0x27, 0x81,
		0xc2, 0x52, 0x8c, 0xaf, 0x27, 0xca, 0xf2,
		0x93, 0xd6, 0x2e, 0x46, 0x32, 0xed, 0x71,
		0x87, 0x90, 0x1d, 0x0b, 0x84, 0x46, 0x7f,
		0xd1, 0x57, 0xc1, 0x0d, 0xc7, 0x5b, 0x41,
		0xbb, 0x8a, 0x7d, 0xe9, 0x2c, 0xae, 0x36,
		0x98, 0x13, 0x39, 0xb9, 0x0c, 0x66, 0x47,
		0x05, 0xa2, 0xdf, 0x55, 0xc4, 0x09, 0xab,
		0xe4, 0xfb, 0x11, 0x52, 0x36, 0x27, 0x88,
		0x86, 0xf3, 0x4a, 0xbb, 0xef, 0x40, 0xa7,
		0x85, 0x2a, 0xfe, 0x92, 0x28, 0xe4, 0xce,
		0xce, 0xdc, 0x4b, 0xd0, 0xaa, 0x3c, 0xd5,
		0x16, 0x76, 0x74, 0xe2, 0xfa, 0x34, 0x91,
		0x4f, 0xdc, 0x2b, 0xea, 0xae, 0x71, 0x36,
		0x74, 0xe1, 0x2a, 0xf3, 0xd3, 0x53, 0xe8,
		0xec, 0xd6, 0x63, 0xf6, 0x6a, 0x75, 0x95,
		0x68, 0xcc, 0x99, 0xbe, 0x17, 0xd8, 0x3b,
		0x87, 0x5b, 0x94, 0xdc, 0xec, 0x32, 0x09,
		0x18, 0x4b, 0x37, 0x58, 0xb5, 0x67, 0xfb,
		0xdf, 0x66, 0x6c, 0x16, 0x9e, 0xba, 0x72,
		0xc6, 0x21, 0xac, 0x02, 0x6d, 0x6b, 0x17,
		0xf9, 0x68, 0x22, 0x2e, 0x10, 0xd7, 0xdf,
		0xfb, 0x24, 0x69, 0x7c, 0xaf, 0x11, 0x64,
		0x80, 0x7a, 0x9d, 0x09, 0xc4, 0x1f, 0xf1,
		0xd7, 0x3c, 0x5a, 0xc2, 0x2c, 0x8e, 0xf5,
		0xff, 0xee, 0xc2, 0x7c, 0xa1, 0xe4, 0xcb,
		0x1c, 0x6d, 0xd8, 0x15, 0x0e, 0x40, 0x36,
		0x85, 0xe7, 0x04, 0xbb, 0x64, 0xca, 0x6a,
		0xd9, 0x21, 0x8e, 0x95, 0xa0, 0x83, 0x95,
		0x10, 0x48, 0xfa, 0x00, 0x54, 0x90, 0xe9,
		0x81, 0x86, 0xa0, 0x4a, 0x6e, 0xbe, 0x9b,
		0xf0, 0x73, 0x0a, 0x17, 0xbb, 0x57, 0x81,
		0x17, 0xaf, 0xd6, 0x70, 0x1f, 0xe8, 0x6d,
		0x32, 0x59, 0x14, 0x39, 0xd8, 0x1d, 0xec,
		0x59, 0xe4, 0x98, 0x4d, 0x44, 0xf3, 0x4f,
		0x7b, 0x47, 0xd9, 0x92, 0x3b, 0xd9, 0x5c,
		0x98, 0xd5, 0xf1, 0xc9, 0x8b, 0x9d, 0xb1,
		0x65, 0xb3, 0xe1, 0x87, 0xa4, 0x6a, 0xcc,
		0x42, 0x96, 0x66, 0xdb, 0x5f, 0xf9, 0xe1,
		0xa1, 0x72, 0xb6, 0x05, 0x02, 0x1f, 0xa3,
		0x14, 0x3e, 0xfe, 0x99, 0x7f, 0xeb, 0x42,
		0xcf, 0x76, 0x09, 0x19, 0xd2, 0xd2, 0x99,
		0x75, 0x1c, 0x67, 0xda, 0x4d, 0xf4, 0x87,
		0xe5, 0x55, 0x8b, 0xed, 0x01, 0x82, 0xf6,
		0xd6, 0x1c, 0x5c, 0x05, 0x96, 0x96, 0x79,
		0xc1, 0x61, 0x87, 0x74, 0xcd, 0x29, 0x83,
		0x27, 0xae, 0x47, 0x87, 0x36, 0x34, 0xab,
		0xc4, 0x73, 0x76, 0x58, 0x1b, 0x4a, 0xec,
		0x0e, 0x4c, 0x2f, 0xb1, 0x76, 0x08, 0x7f,
		0xaf, 0xfa, 0x6d, 0x8c, 0xde, 0xe4, 0xae,
		0x58, 0x87, 0xe7, 0xa0, 0x27, 0x05, 0x0d,
		0xf5, 0xa7, 0xfb, 0x2a, 0x75, 0x33, 0xd9,
		0x3b, 0x65, 0x60, 0xa4, 0x13, 0x27, 0xa5,
		0xe5, 0x1b, 0x83, 0x78, 0x7a, 0xd7, 0xec,
		0x0c, 0xed, 0x8b, 0xe6, 0x4e, 0x8f, 0xfe,
		0x6b, 0x5d, 0xbb, 0xa8, 0xee, 0x38, 0x81,
		0x6f, 0x09, 0x23, 0x08, 0x8f, 0x07, 0x21,
		0x09, 0x39, 0xf0, 0xf8, 0x03, 0x17, 0x24,
		0x2a, 0x22, 0x44, 0x84, 0xe1, 0x5c, 0xf3,
		0x4f, 0x20, 0xdc, 0xc1, 0xe7, 0xeb, 0xbc,
		0x0b, 0xfb, 0x7b, 0x20, 0x66, 0xa4, 0x27,
		0xe2, 0x01, 0xb3, 0x5f, 0xb7, 0x47, 0xa1,
		0x88, 0x4b, 0x8c, 0x47, 0xda, 0x36, 0x98,
		0x60, 0xd7, 0x46, 0x92, 0x0b, 0x7e, 0x5b,
		0x4e, 0x34, 0x50, 0x12, 0x67, 0x50, 0x8d,
		0xe7, 0xc9, 0xe4, 0x96, 0xef, 0xae, 0x2b,
		0xc7, 0xfa, 0x36, 0x29, 0x05, 0xf5, 0x92,
		0xbd, 0x62, 0xb7, 0xbb, 0x90, 0x66, 0xe0,
		0xad, 0x14, 0x3e, 0xe7, 0xb4, 0x24, 0xf3,
		0x04, 0xcf, 0x22, 0x14, 0x86, 0xa4, 0xb8,
		0xfb, 0x83, 0x56, 0xce, 0xaa, 0xb4, 0x87,
		0x5a, 0x9e, 0xf2, 0x0b, 0xaf, 0xad, 0x40,
		0xe1, 0xb5, 0x5c, 0x6b, 0xa7, 0xee, 0x9f,
		0xbb, 0x1a, 0x68, 0x4d, 0xc3, 0xbf, 0x22,
		0x4d, 0xbe, 0x58, 0x52, 0xc9, 0xcc, 0x0d,
		0x88, 0x04, 0xf1, 0xf8, 0xd4, 0xfb, 0xd6,
		0xad, 0xcf, 0x13, 0x84, 0xd6, 0x2f, 0x90,
		0x0c, 0x5f, 0xb4, 0xe2, 0xd8, 0x29, 0x26,
		0x8d, 0x7c, 0x6b, 0xab, 0x91, 0x91, 0x3c,
		0x25, 0x39, 0x9c, 0x86, 0x08, 0x39, 0x54,
		0x59, 0x0d, 0xa4, 0xa8, 0x31, 0x9f, 0xa3,
		0xbc, 0xc2, 0xcb, 0xf9, 0x30, 0x49, 0xc3,
		0x68, 0x0e, 0xfc, 0x2b, 0x9f, 0xce, 0x59,
		0x02, 0xfa, 0xd4, 0x4e, 0x11, 0x49, 0x0d,
		0x93, 0x0c, 0xae, 0x57, 0xd7, 0x74, 0xdd,
		0x13, 0x1a, 0x15, 0x79, 0x10, 0xcc, 0x99,
		0x32, 0x9b, 0x57, 0x6d, 0x53, 0x75, 0x1f,
		0x6d, 0xbb, 0xe4, 0xbc, 0xa9, 0xd4, 0xdb,
		0x06, 0xe7, 0x09, 0xb0, 0x6f, 0xca, 0xb3,
		0xb1, 0xed, 0xc5, 0x0b, 0x8d, 0x8e, 0x70,
		0xb0, 0xbf, 0x8b, 0xad, 0x2f, 0x29, 0x92,
		0xdd, 0x5a, 0x19, 0x3d, 0xca, 0xca, 0xed,
		0x05, 0x26, 0x25, 0xee, 0xee, 0xa9, 0xdd,
		0xa0, 0xe3, 0x78, 0xe0, 0x56, 0x99, 0x2f,
		0xa1, 0x3f, 0x07, 0x5e, 0x91, 0xfb, 0xc4,
		0xb3, 0xac, 0xee, 0x07, 0xa4, 0x6a, 0xcb,
		0x42, 0xae, 0xdf, 0x09, 0xe7, 0xd0, 0xbb,
		0xc6, 0xd4, 0x38, 0x58, 0x7d, 0xb4, 0x45,
		0x98, 0x38, 0x21, 0xc8, 0xc1, 0x3c, 0x81,
		0x12, 0x7e, 0x37, 0x03, 0xa8, 0xcc, 0xf3,
		0xf9, 0xd9, 0x9d, 0x8f, 0xc1, 0xa1, 0xcc,
		0xc1, 0x1b, 0xe3, 0xa8, 0x93, 0x91, 0x2c,
		0x0a, 0xe8, 0x1f, 0x28, 0x13, 0x44, 0x07,
		0x68, 0x5a, 0x8f, 0x27, 0x41, 0x18, 0xc9,
		0x31, 0xc4, 0xc1, 0x71, 0xe2, 0xf0, 0xc4,
		0xf4, 0x1e, 0xac, 0x29, 0x49, 0x2f, 0xd0,
		0xc0, 0x98, 0x13, 0xa6, 0xbc, 0x5e, 0x34,
		0x28, 0xa7, 0x30, 0x13, 0x8d, 0xb4, 0xca,
		0x91, 0x26, 0x6c, 0xda, 0x35, 0xb5, 0xf1,
		0xbf, 0x3f, 0x35, 0x3b, 0x87, 0x37, 0x63,
		0x40, 0x59, 0x73, 0x49, 0x06, 0x59, 0x04,
		0xe0, 0x84, 0x16, 0x3a, 0xe8, 0xc4, 0x28,
		0xd1, 0xf5, 0x11, 0x9c, 0x34, 0xf4, 0x5a,
		0xc0, 0xf8, 0x67, 0x47, 0x1c, 0x90, 0x63,
		0xbc, 0x06, 0x39, 0x2e, 0x8a, 0xa5, 0xa0,
		0xf1, 0x6b, 0x41, 0xb1, 0x16, 0xbd, 0xb9,
		0x50, 0x78, 0x72, 0x91, 0x8e, 0x8c, 0x99,
		0x0f, 0x7d, 0x99, 0x7e, 0x77, 0x36, 0x85,
		0x87, 0x1f, 0x2e, 0x47, 0x13, 0x55, 0xf8,
		0x07, 0xba, 0x7b, 0x1c, 0xaa, 0xbf, 0x20,
		0xd0, 0xfa, 0xc4, 0xe1, 0xd0, 0xb3, 0xe4,
		0xf4, 0xf9, 0x57, 0x8d, 0x56, 0x19, 0x4a,
		0xdc, 0x4c, 0x83, 0xc8, 0xf1, 0x30, 0xc0,
		0xb5, 0xdf, 0x67, 0x25, 0x58, 0xd8, 0x09,
		0x41, 0x37, 0x2e, 0x0b, 0x47, 0x2b, 0x86,
		0x4b, 0x73, 0x38, 0xf0, 0xa0, 0x6b, 0x83,
		0x30, 0x80, 0x3e, 0x46, 0xb5, 0x09, 0xc8,
		0x6d, 0x3e, 0x97, 0xaa, 0x70, 0x4e, 0x8c,
		0x75, 0x29, 0xec, 0x8a, 0x37, 0x4a, 0x81,
		0xfd, 0x92, 0xf1, 0x29, 0xf0, 0xe8, 0x9d,
		0x8c, 0xb4, 0x39, 0x2d, 0x67, 0x06, 0xcd,
		0x5f, 0x25, 0x02, 0x30, 0xbb, 0x6b, 0x41,
		0x93, 0x55, 0x1e, 0x0c, 0xc9, 0x6e, 0xb5,
		0xd5, 0x9f, 0x80, 0xf4, 0x7d, 0x9d, 0x8a,
		0x0d, 0x8d, 0x3b, 0x15, 0x14, 0xc9, 0xdf,
		0x03, 0x9c, 0x78, 0x39, 0x4e, 0xa0, 0xdc,
		0x3a, 0x1b, 0x8c, 0xdf, 0xaa, 0xed, 0x25,
		0xda, 0x60, 0xdd, 0x30, 0x64, 0x09, 0xcc,
		0x94, 0x53, 0xa1, 0xad, 0xfd, 0x9e, 0xe7,
		0x65, 0x15, 0xb8, 0xb1, 0xda, 0x9a, 0x28,
		0x80, 0x51, 0x88, 0x93, 0x92, 0xe3, 0x03,
		0xdf, 0x70, 0xba, 0x1b, 0x59, 0x3b, 0xb4,
		0x8a, 0xb6, 0x0b, 0x0a, 0xa8, 0x48, 0xdf,
		0xcc, 0x74, 0x4c, 0x71, 0x80, 0x08, 0xec,
		0xc8, 0x8a, 0x73, 0xf5, 0x0e, 0x3d, 0xec,
		0x16, 0xf6, 0x32, 0xfd, 0xf3, 0x6b, 0xba,
		0xa9, 0x65, 0xd1, 0x87, 0xe2, 0x56, 0xcd,
		0xde, 0x2c, 0xa4, 0x1b, 0x25, 0x81, 0xb2,
		0xed, 0xea, 0xe9, 0x11, 0x07, 0xf5, 0x17,
		0xd0, 0xca, 0x5d, 0x07, 0xb9, 0xb2, 0xa9,
		0xa9, 0xee, 0x42, 0x33, 0x93, 0x21, 0x30,
		0x5e, 0xd2, 0x58, 0xfd, 0xdd, 0x73, 0x0d,
		0xb2, 0x93, 0x58, 0x77, 0x78, 0x40, 0x69,
		0xba, 0x3c, 0x95, 0x1c, 0x61, 0xc6, 0xc6,
		0x97, 0x1c, 0xef, 0x4d, 0x91, 0x0a, 0x42,
		0x91, 0x1d, 0x14, 0x93, 0xf5, 0x78, 0x41,
		0x32, 0x8a, 0x0a, 0x43, 0xd4, 0x3e, 0x6b,
		0xb0, 0xd8, 0x0e, 0x04,
	}

	av1Pkt := &AV1Packet{}
	if _, err := av1Pkt.Unmarshal(av1Payload); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(av1Pkt, &AV1Packet{
		Z: false,
		Y: true,
		W: 2,
		N: true,
		OBUElements: [][]byte{
			av1Payload[2:14],
			av1Payload[14:],
		},
	}) {
		t.Fatal("AV1 Unmarshal didn't store the expected results in the packet")
	}
}
