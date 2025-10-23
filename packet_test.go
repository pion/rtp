// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package rtp

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) { // nolint:maintidx,cyclop
	packet := &Packet{}

	assert.Error(t, packet.Unmarshal([]byte{}), "Unmarshal did not error on zero length packet")
	assert.ErrorIs(t, packet.Unmarshal([]byte{}), errHeaderSizeInsufficient)

	rawPkt := []byte{
		0x90, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0x00, 0x01, 0x00, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	parsedPacket := &Packet{
		Header: Header{
			Padding:          false,
			Marker:           true,
			Extension:        true,
			ExtensionProfile: 1,
			Extensions: []Extension{
				{0, []byte{
					0xFF, 0xFF, 0xFF, 0xFF,
				}},
			},
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
			PaddingSize:    0,
		},
		Payload: rawPkt[20:],
	}

	// Unmarshal to the used Packet should work as well.
	for i := 0; i < 2; i++ {
		t.Run(fmt.Sprintf("Run%d", i+1), func(t *testing.T) {
			assert.NoError(t, packet.Unmarshal(rawPkt))
			assert.Equal(t, packet, parsedPacket)

			assert.Equal(t, packet.Header.MarshalSize(), 20, "wrong computed header marshal size")
			assert.Equal(t, packet.MarshalSize(), len(rawPkt), "wrong computed marshal size")

			raw, err := packet.Marshal()
			assert.NoError(t, err)
			assert.Equal(t, rawPkt, raw)
		})
	}

	// packet with padding
	rawPkt = []byte{
		0xb0, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0x00, 0x01, 0x00, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0x98, 0x36, 0xbe, 0x88, 0x04,
	}
	parsedPacket = &Packet{
		Header: Header{
			Padding:          true,
			Marker:           true,
			Extension:        true,
			ExtensionProfile: 1,
			Extensions: []Extension{
				{0, []byte{
					0xFF, 0xFF, 0xFF, 0xFF,
				}},
			},
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
			PaddingSize:    4,
		},
		Payload:     rawPkt[20:21],
		PaddingSize: 4,
	}
	assert.NoError(t, packet.Unmarshal(rawPkt))
	assert.Equal(t, packet, parsedPacket)

	// packet with zero padding following packet with non-zero padding
	rawPkt = []byte{
		0x90, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0x00, 0x01, 0x00, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	parsedPacket = &Packet{
		Header: Header{
			Padding:          false,
			Marker:           true,
			Extension:        true,
			ExtensionProfile: 1,
			Extensions: []Extension{
				{0, []byte{
					0xFF, 0xFF, 0xFF, 0xFF,
				}},
			},
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
			PaddingSize:    0,
		},
		Payload: rawPkt[20:],
	}
	assert.NoError(t, packet.Unmarshal(rawPkt))
	assert.Equal(t, packet, parsedPacket)

	// packet with only padding
	rawPkt = []byte{
		0xb0, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0x00, 0x01, 0x00, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0x98, 0x36, 0xbe, 0x88, 0x05,
	}
	parsedPacket = &Packet{
		Header: Header{
			Padding:          true,
			Marker:           true,
			Extension:        true,
			ExtensionProfile: 1,
			Extensions: []Extension{
				{0, []byte{
					0xFF, 0xFF, 0xFF, 0xFF,
				}},
			},
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
			PaddingSize:    5,
		},
		Payload:     []byte{},
		PaddingSize: 5,
	}
	assert.NoError(t, packet.Unmarshal(rawPkt))
	assert.Equal(t, packet, parsedPacket)
	assert.Len(t, packet.Payload, 0, "Unmarshal of padding only packet has payload of non-zero length")

	// packet with excessive padding
	rawPkt = []byte{
		0xb0, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0x00, 0x01, 0x00, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0x98, 0x36, 0xbe, 0x88, 0x06,
	}
	parsedPacket = &Packet{
		Header: Header{
			Padding:          true,
			Marker:           true,
			Extension:        true,
			ExtensionProfile: 1,
			Extensions: []Extension{
				{0, []byte{
					0xFF, 0xFF, 0xFF, 0xFF,
				}},
			},
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
			PaddingSize:    0,
		},
		Payload: []byte{},
	}
	err := packet.Unmarshal(rawPkt)
	assert.Error(t, err, "Unmarshal did not error on packet with excessive padding")
	assert.ErrorIs(t, err, errTooSmall)

	// marshal packet with padding
	rawPkt = []byte{
		0xb0, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0x00, 0x01, 0x00, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0x98, 0x00, 0x00, 0x00, 0x04,
	}
	parsedPacket = &Packet{
		Header: Header{
			Padding:          true,
			Marker:           true,
			Extension:        true,
			ExtensionProfile: 1,
			Extensions: []Extension{
				{0, []byte{
					0xFF, 0xFF, 0xFF, 0xFF,
				}},
			},
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
			PaddingSize:    4,
		},
		Payload: rawPkt[20:21],
	}
	buf, err := parsedPacket.Marshal()
	assert.NoError(t, err)
	assert.Equal(t, rawPkt, buf)

	// marshal packet with padding only
	rawPkt = []byte{
		0xb0, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0x00, 0x01, 0x00, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x05,
	}
	parsedPacket = &Packet{
		Header: Header{
			Padding:          true,
			Marker:           true,
			Extension:        true,
			ExtensionProfile: 1,
			Extensions: []Extension{
				{0, []byte{
					0xFF, 0xFF, 0xFF, 0xFF,
				}},
			},
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
			PaddingSize:    5,
		},
		Payload: []byte{},
	}
	buf, err = parsedPacket.Marshal()
	assert.NoError(t, err)
	assert.Equal(t, rawPkt, buf)

	// marshal packet with padding only without setting Padding explicitly in Header
	rawPkt = []byte{
		0xb0, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0x00, 0x01, 0x00, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x05,
	}
	parsedPacket = &Packet{
		Header: Header{
			Marker:           true,
			Extension:        true,
			ExtensionProfile: 1,
			Extensions: []Extension{
				{0, []byte{
					0xFF, 0xFF, 0xFF, 0xFF,
				}},
			},
			Version:        2,
			Padding:        true,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
			PaddingSize:    5,
		},
		Payload: []byte{},
	}
	buf, err = parsedPacket.Marshal()
	assert.NoError(t, err)
	assert.Equal(t, rawPkt, buf)
}

func TestExtension(t *testing.T) {
	packet := &Packet{}

	missingExtensionPkt := []byte{
		0x90, 0x60, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82,
	}
	assert.Error(
		t, packet.Unmarshal(missingExtensionPkt),
		"Unmarshal did not error on packet with missing extension data",
	)

	invalidExtensionLengthPkt := []byte{
		0x90, 0x60, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0x99, 0x99, 0x99, 0x99,
	}
	assert.Error(
		t, packet.Unmarshal(invalidExtensionLengthPkt),
		"Unmarshal did not error on packet with invalid extension length",
	)

	packet = &Packet{
		Header: Header{
			Extension:        true,
			ExtensionProfile: 3,
			Extensions: []Extension{
				{0, []byte{
					0,
				}},
			},
		},
		Payload: []byte{},
	}
	_, err := packet.Marshal()
	assert.Error(t, err, "Marshal did not error on packet with invalid extension length")
}

func TestRFC8285OneByteExtension(t *testing.T) {
	packet := &Packet{}

	rawPkt := []byte{
		0x90, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0xBE, 0xDE, 0x00, 0x01, 0x50, 0xAA, 0x00, 0x00,
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	assert.NoError(t, packet.Unmarshal(rawPkt))

	packet = &Packet{
		Header: Header{
			Marker:           true,
			Extension:        true,
			ExtensionProfile: 0xBEDE,
			Extensions: []Extension{
				{5, []byte{
					0xAA,
				}},
			},
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
		},
		Payload: rawPkt[20:],
	}

	dstData, _ := packet.Marshal()
	assert.Equal(t, rawPkt, dstData)
}

func TestRFC8285OneByteTwoExtensionOfTwoBytes(t *testing.T) {
	packet := &Packet{}

	//  0                   1                   2                   3
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |       0xBE    |    0xDE       |           length=1            |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |  ID   | L=0   |     data      |  ID   |  L=0  |   data...
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	rawPkt := []byte{
		0x90, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0xBE, 0xDE, 0x00, 0x01, 0x10, 0xAA, 0x20, 0xBB,
		// Payload
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	assert.NoError(t, packet.Unmarshal(rawPkt))

	ext1 := packet.GetExtension(1)
	ext1Expect := []byte{0xAA}
	assert.Equal(t, ext1Expect, ext1, "Extension has incorrect data")

	ext2 := packet.GetExtension(2)
	ext2Expect := []byte{0xBB}
	assert.Equal(t, ext2Expect, ext2, "Extension has incorrect data")

	// Test Marshal
	packet = &Packet{
		Header: Header{
			Marker:           true,
			Extension:        true,
			ExtensionProfile: 0xBEDE,
			Extensions: []Extension{
				{1, []byte{
					0xAA,
				}},
				{2, []byte{
					0xBB,
				}},
			},
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
		},
		Payload: rawPkt[20:],
	}

	dstData, _ := packet.Marshal()
	assert.Equal(t, rawPkt, dstData)
}

func TestRFC8285OneByteMultipleExtensionsWithPadding(t *testing.T) {
	packet := &Packet{}

	//  0                   1                   2                   3
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |       0xBE    |    0xDE       |           length=3            |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |  ID   | L=0   |     data      |  ID   |  L=1  |   data...
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//       ...data   |    0 (pad)    |    0 (pad)    |  ID   | L=3   |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |                          data                                 |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	rawPkt := []byte{
		0x90, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0xBE, 0xDE, 0x00, 0x03, 0x10, 0xAA, 0x21, 0xBB,
		0xBB, 0x00, 0x00, 0x33, 0xCC, 0xCC, 0xCC, 0xCC,
		// Payload
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	assert.NoError(t, packet.Unmarshal(rawPkt))

	ext1 := packet.GetExtension(1)
	ext1Expect := []byte{0xAA}
	assert.Equal(t, ext1Expect, ext1, "Extension has incorrect data")

	ext2 := packet.GetExtension(2)
	ext2Expect := []byte{0xBB, 0xBB}
	assert.Equal(t, ext2Expect, ext2, "Extension has incorrect data")

	ext3 := packet.GetExtension(3)
	ext3Expect := []byte{0xCC, 0xCC, 0xCC, 0xCC}
	assert.Equal(t, ext3Expect, ext3, "Extension has incorrect data")

	rawPktReMarshal := []byte{
		0x90, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0xBE, 0xDE, 0x00, 0x03, 0x10, 0xAA, 0x21, 0xBB,
		0xBB, 0x33, 0xCC, 0xCC, 0xCC, 0xCC, 0x00, 0x00, // padding is moved to the end by re-marshaling
		// Payload
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	dstBuf := map[string][]byte{
		"CleanBuffer": make([]byte, 1000),
		"DirtyBuffer": make([]byte, 1000),
	}
	for i := range dstBuf["DirtyBuffer"] {
		dstBuf["DirtyBuffer"][i] = 0xFF
	}
	for name, buf := range dstBuf {
		buf := buf
		t.Run(name, func(t *testing.T) {
			n, err := packet.MarshalTo(buf)
			assert.NoError(t, err)
			assert.Equal(t, rawPktReMarshal, buf[:n])
		})
	}
}

func TestRFC8285OneByteMultipleExtensions(t *testing.T) {
	//  0                   1                   2                   3
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |       0xBE    |    0xDE       |           length=3            |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |  ID=1 | L=0   |     data      |  ID=2 |  L=1  |   data...
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//       ...data   |  ID=3 | L=3   |           data...
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//             ...data             |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	rawPkt := []byte{
		0x90, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0xBE, 0xDE, 0x00, 0x03, 0x10, 0xAA, 0x21, 0xBB,
		0xBB, 0x33, 0xCC, 0xCC, 0xCC, 0xCC, 0x00, 0x00,
		// Payload
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}

	packet := &Packet{
		Header: Header{
			Marker:           true,
			Extension:        true,
			ExtensionProfile: 0xBEDE,
			Extensions: []Extension{
				{1, []byte{
					0xAA,
				}},
				{2, []byte{
					0xBB, 0xBB,
				}},
				{3, []byte{
					0xCC, 0xCC, 0xCC, 0xCC,
				}},
			},
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
		},
		Payload: rawPkt[28:],
	}

	dstData, _ := packet.Marshal()
	assert.Equal(t, rawPkt, dstData)
}

func TestRFC8285TwoByteExtension(t *testing.T) {
	packet := &Packet{}

	rawPkt := []byte{
		0x90, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0x10, 0x00, 0x00, 0x07, 0x05, 0x18, 0xAA, 0xAA,
		0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
		0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
		0xAA, 0xAA, 0x00, 0x00, 0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	assert.NoError(t, packet.Unmarshal(rawPkt))

	packet = &Packet{
		Header: Header{
			Marker:           true,
			Extension:        true,
			ExtensionProfile: 0x1000,
			Extensions: []Extension{
				{5, []byte{
					0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
					0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
					0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
				}},
			},
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
		},
		Payload: rawPkt[44:],
	}

	dstData, _ := packet.Marshal()
	assert.Equal(t, rawPkt, dstData)
}

func TestRFC8285TwoByteMultipleExtensionsWithPadding(t *testing.T) {
	packet := &Packet{}

	// 0                   1                   2                   3
	// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |       0x10    |    0x00       |           length=3            |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |      ID=1     |     L=0       |     ID=2      |     L=1       |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |       data    |    0 (pad)    |       ID=3    |      L=4      |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |                          data                                 |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	rawPkt := []byte{
		0x90, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0x10, 0x00, 0x00, 0x03, 0x01, 0x00, 0x02, 0x01,
		0xBB, 0x00, 0x03, 0x04, 0xCC, 0xCC, 0xCC, 0xCC, 0x98, 0x36,
		0xbe, 0x88, 0x9e,
	}
	assert.NoError(t, packet.Unmarshal(rawPkt))

	ext1 := packet.GetExtension(1)
	ext1Expect := []byte{}
	assert.Equal(t, ext1Expect, ext1, "Extension has incorrect data")

	ext2 := packet.GetExtension(2)
	ext2Expect := []byte{0xBB}
	assert.Equal(t, ext2Expect, ext2, "Extension has incorrect data")

	ext3 := packet.GetExtension(3)
	ext3Expect := []byte{0xCC, 0xCC, 0xCC, 0xCC}
	assert.Equal(t, ext3Expect, ext3, "Extension has incorrect data")

	rawPktReMarshal := []byte{
		0x90, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0x10, 0x00, 0x00, 0x03, 0x01, 0x00, 0x02, 0x01,
		0xBB, 0x03, 0x04, 0xCC, 0xCC, 0xCC, 0xCC, 0x00, // padding is moved to the end by re-marshaling
		// Payload
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	dstBuf := map[string][]byte{
		"CleanBuffer": make([]byte, 1000),
		"DirtyBuffer": make([]byte, 1000),
	}
	for i := range dstBuf["DirtyBuffer"] {
		dstBuf["DirtyBuffer"][i] = 0xFF
	}
	for name, buf := range dstBuf {
		buf := buf
		t.Run(name, func(t *testing.T) {
			n, err := packet.MarshalTo(buf)
			assert.NoError(t, err)
			assert.Equal(t, rawPktReMarshal, buf[:n])
		})
	}
}

func TestRFC8285TwoByteMultipleExtensionsWithLargeExtension(t *testing.T) {
	// 0                   1                   2                   3
	// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |       0x10    |    0x00       |           length=3            |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |      ID=1     |     L=0       |     ID=2      |     L=1       |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |       data    |       ID=3    |      L=17      |    data...
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//                            ...data...
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//                            ...data...
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//                            ...data...
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//                            ...data...                           |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	rawPkt := []byte{
		0x90, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0x10, 0x00, 0x00, 0x06, 0x01, 0x00, 0x02, 0x01,
		0xBB, 0x03, 0x11, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC,
		0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC,
		// Payload
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}

	packet := &Packet{
		Header: Header{
			Marker:           true,
			Extension:        true,
			ExtensionProfile: 0x1000,
			Extensions: []Extension{
				{1, []byte{}},
				{2, []byte{
					0xBB,
				}},
				{3, []byte{
					0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC,
					0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC,
				}},
			},
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
		},
		Payload: rawPkt[40:],
	}
	dstData, _ := packet.Marshal()
	assert.Equal(t, rawPkt, dstData)
}

func TestRFC8285GetExtensionReturnsNilWhenExtensionsDisabled(t *testing.T) {
	payload := []byte{
		// Payload
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	packet := &Packet{
		Header: Header{
			Marker:         true,
			Extension:      false,
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
		},
		Payload: payload,
	}
	assert.Nil(t, packet.GetExtension(1), "Should return nil on GetExtension when h.Extension: false")
}

func TestRFC8285DelExtension(t *testing.T) {
	payload := []byte{
		// Payload
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	packet := &Packet{
		Header: Header{
			Marker:           true,
			Extension:        true,
			ExtensionProfile: 0xBEDE,
			Extensions: []Extension{
				{1, []byte{
					0xAA,
				}},
			},
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
		},
		Payload: payload,
	}
	assert.NotNil(t, packet.GetExtension(1), "Extension should exist")
	assert.NoError(t, packet.DelExtension(1), "Should successfully delete extension")
	assert.Nil(t, packet.GetExtension(1), "Extension should not exist")
	assert.Error(t, packet.DelExtension(1), "Should return error when deleting extension that doesnt exist")
}

func TestRFC8285GetExtensionIDs(t *testing.T) {
	payload := []byte{
		// Payload
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	packet := &Packet{
		Header: Header{
			Marker:           true,
			Extension:        true,
			ExtensionProfile: 0xBEDE,
			Extensions: []Extension{
				{1, []byte{
					0xAA,
				}},
				{2, []byte{
					0xBB,
				}},
			},
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
		},
		Payload: payload,
	}
	ids := packet.GetExtensionIDs()
	assert.NotNil(t, ids, "Extension should exist")
	assert.Len(t, ids, len(packet.Extensions), "The number of IDs should be equal to the number of extensions")

	for _, id := range ids {
		ext := packet.GetExtension(id)
		assert.NotNil(t, ext, "Extension should exist")
	}
}

func TestRFC8285GetExtensionIDsReturnsErrorWhenExtensionsDisabled(t *testing.T) {
	payload := []byte{
		// Payload
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	packet := &Packet{
		Header: Header{
			Marker:         true,
			Extension:      false,
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
		},
		Payload: payload,
	}
	assert.Nil(t, packet.GetExtensionIDs(), "Should return nil on GetExtensionIDs when h.Extensions is nil")
}

func TestRFC8285DelExtensionReturnsErrorWhenExtensionsDisabled(t *testing.T) {
	payload := []byte{
		// Payload
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	packet := &Packet{
		Header: Header{
			Marker:         true,
			Extension:      false,
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
		},
		Payload: payload,
	}
	assert.Error(
		t, packet.DelExtension(1), "DelExtension did not error on h.Extension: false",
	)
}

func TestRFC8285OneByteSetExtensionShouldEnableExensionsWhenAdding(t *testing.T) {
	payload := []byte{
		// Payload
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	packet := &Packet{
		Header: Header{
			Marker:         true,
			Extension:      false,
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
		},
		Payload: payload,
	}
	extension := []byte{0xAA, 0xAA}
	assert.NoError(t, packet.SetExtension(1, extension))
	assert.True(t, packet.Extension)
	assert.Equal(t, uint16(0xBEDE), packet.ExtensionProfile)
	assert.Len(t, packet.Extensions, 1)
	assert.Equal(t, extension, packet.GetExtension(1))
}

func TestRFC8285OneByteSetExtensionShouldSetCorrectExtensionProfileFor16ByteExtension(t *testing.T) {
	payload := []byte{
		// Payload
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	packet := &Packet{
		Header: Header{
			Marker:         true,
			Extension:      false,
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
		},
		Payload: payload,
	}

	extension := []byte{
		0xAA, 0xAA, 0xAA, 0xAA,
		0xAA, 0xAA, 0xAA, 0xAA,
		0xAA, 0xAA, 0xAA, 0xAA,
		0xAA, 0xAA, 0xAA, 0xAA,
	}
	err := packet.SetExtension(1, extension)
	assert.NoError(t, err, "Error setting extension")
	assert.Equal(t, uint16(0xBEDE), packet.ExtensionProfile)
}

func TestRFC8285OneByteSetExtensionShouldUpdateExistingExension(t *testing.T) {
	payload := []byte{
		// Payload
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	packet := &Packet{
		Header: Header{
			Marker:           true,
			Extension:        true,
			ExtensionProfile: 0xBEDE,
			Extensions: []Extension{
				{1, []byte{
					0xAA,
				}},
			},
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
		},
		Payload: payload,
	}
	assert.Equal(t, []byte{0xAA}, packet.GetExtension(1))

	extension := []byte{0xBB}
	err := packet.SetExtension(1, extension)
	assert.NoError(t, err, "Error setting extension")
	assert.Equal(t, extension, packet.GetExtension(1))
}

func TestRFC8285OneByteSetExtensionShouldErrorWhenInvalidIDProvided(t *testing.T) {
	payload := []byte{
		// Payload
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	packet := &Packet{
		Header: Header{
			Marker:           true,
			Extension:        true,
			ExtensionProfile: 0xBEDE,
			Extensions: []Extension{
				{1, []byte{
					0xAA,
				}},
			},
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
		},
		Payload: payload,
	}
	assert.Error(
		t, packet.SetExtension(0, []byte{0xBB}),
		"SetExtension did not error on invalid id",
	)
	assert.Error(
		t, packet.SetExtension(15, []byte{0xBB}),
		"SetExtension did not error on invalid id",
	)
}

func TestRFC8285OneByteExtensionTerminateProcessingWhenReservedIDEncountered(t *testing.T) {
	packet := &Packet{}

	reservedIDPkt := []byte{
		0x90, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0xBE, 0xDE, 0x00, 0x01, 0xF0, 0xAA, 0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	assert.NoError(
		t, packet.Unmarshal(reservedIDPkt),
		"Unmarshal error on packet with reserved extension id",
	)
	assert.Len(t, packet.Extensions, 0, "Extensions should be empty for invalid id")

	payload := reservedIDPkt[20:]
	assert.Equal(t, payload, packet.Payload)
}

func TestRFC8285OneByteExtensionTerminateProcessingWhenPaddingWithSizeEncountered(t *testing.T) {
	packet := &Packet{}

	reservedIDPkt := []byte{
		0x90, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0xBE, 0xDE, 0x00, 0x01, 0x01, 0xAA, 0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	assert.NoError(
		t, packet.Unmarshal(reservedIDPkt),
		"Unmarshal error on packet with non-zero padding size",
	)
	assert.Len(t, packet.Extensions, 0, "Extensions should be empty for non-zero padding size")

	payload := reservedIDPkt[20:]
	assert.Equal(t, payload, packet.Payload)
}

func TestRFC8285OneByteSetExtensionShouldErrorWhenPayloadTooLarge(t *testing.T) {
	payload := []byte{
		// Payload
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	packet := &Packet{
		Header: Header{
			Marker:           true,
			Extension:        true,
			ExtensionProfile: 0xBEDE,
			Extensions: []Extension{
				{1, []byte{
					0xAA,
				}},
			},
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
		},
		Payload: payload,
	}
	assert.Error(
		t,
		packet.SetExtension(1, []byte{
			0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
			0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		}),
		"SetExtension did not error on too large payload",
	)
}

func TestRFC8285TwoByteSetExtensionShouldEnableExensionsWhenAdding(t *testing.T) {
	payload := []byte{
		// Payload
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	packet := &Packet{
		Header: Header{
			Marker:         true,
			Extension:      false,
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
		},
		Payload: payload,
	}

	extension := []byte{
		0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
		0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
	}
	assert.NoError(t, packet.SetExtension(1, extension))
	assert.True(t, packet.Extension)
	assert.Equal(t, uint16(0x1000), packet.ExtensionProfile)
	assert.Len(t, packet.Extensions, 1)
	assert.Equal(t, extension, packet.GetExtension(1))
}

func TestRFC8285TwoByteSetExtensionShouldUpdateExistingExension(t *testing.T) {
	payload := []byte{
		// Payload
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	packet := &Packet{
		Header: Header{
			Marker:           true,
			Extension:        true,
			ExtensionProfile: 0x1000,
			Extensions: []Extension{
				{1, []byte{
					0xAA,
				}},
			},
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
		},
		Payload: payload,
	}
	assert.Equal(t, []byte{0xAA}, packet.GetExtension(1), "Extension value not initialize properly")

	extension := []byte{
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
	}
	err := packet.SetExtension(1, extension)
	assert.NoError(t, err)
	assert.Equal(t, packet.GetExtension(1), extension)
}

func TestRFC8285TwoByteSetExtensionShouldErrorWhenPayloadTooLarge(t *testing.T) {
	payload := []byte{
		// Payload
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	packet := &Packet{
		Header: Header{
			Marker:           true,
			Extension:        true,
			ExtensionProfile: 0xBEDE,
			Extensions: []Extension{
				{1, []byte{
					0xAA,
				}},
			},
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
		},
		Payload: payload,
	}

	err := packet.SetExtension(1, []byte{
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
	})
	assert.Error(t, err, "SetExtension did not error on too large payload")
}

func TestRFC8285Padding(t *testing.T) {
	header := &Header{}

	for n, payload := range [][]byte{
		{
			0b00010000,                      // header.Extension = true
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // SequenceNumber, Timestamp, SSRC
			0xBE, 0xDE, // header.ExtensionProfile = extensionProfileOneByte
			0, 1, // extensionLength
			0, 0, 0, // padding
			0x10, // extid and length
		},
		{
			0b00010000,                      // header.Extension = true
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // SequenceNumber, Timestamp, SSRC
			0x10, 0x00, // header.ExtensionProfile = extensionProfileTwoByte
			0, 1, // extensionLength
			0, 0, // padding
			0x01, 0x01, // extid and length
		},
	} {
		_, err := header.Unmarshal(payload)
		assert.ErrorIs(t, err, errHeaderSizeInsufficientForExtension, "case %d", n)
	}
}

func TestRFC3550SetExtensionShouldErrorWhenNonZero(t *testing.T) {
	payload := []byte{
		// Payload
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	packet := &Packet{
		Header: Header{
			Marker:           true,
			Extension:        true,
			ExtensionProfile: 0x1111,
			Extensions: []Extension{
				{0, []byte{
					0xAA,
				}},
			},
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
		},
		Payload: payload,
	}

	expect := []byte{0xBB}
	assert.NoError(t, packet.SetExtension(0, expect), "SetExtension should not error on valid id")

	actual := packet.GetExtension(0)
	assert.Equal(t, expect, actual)
}

func TestRFC3550SetExtensionShouldRaiseErrorWhenSettingNonzeroID(t *testing.T) {
	payload := []byte{
		// Payload
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	packet := &Packet{
		Header: Header{
			Marker:           true,
			Extension:        true,
			ExtensionProfile: 0x1111,
			Version:          2,
			PayloadType:      96,
			SequenceNumber:   27023,
			Timestamp:        3653407706,
			SSRC:             476325762,
		},
		Payload: payload,
	}

	assert.Error(t, packet.SetExtension(1, []byte{0xBB}), "SetExtension should error on invalid id")
}

func TestUnmarshal_ErrorHandling(t *testing.T) {
	cases := map[string]struct {
		input []byte
		err   error
	}{
		"ShortHeader": {
			input: []byte{
				0x80, 0xe0, 0x69, 0x8f,
				0xd9, 0xc2, 0x93, 0xda, // timestamp
				0x1c, 0x64, 0x27, // SSRC (one byte missing)
			},
			err: errHeaderSizeInsufficient,
		},
		"MissingCSRC": {
			input: []byte{
				0x81, 0xe0, 0x69, 0x8f,
				0xd9, 0xc2, 0x93, 0xda, // timestamp
				0x1c, 0x64, 0x27, 0x82, // SSRC
			},
			err: errHeaderSizeInsufficient,
		},
		"MissingExtension": {
			input: []byte{
				0x90, 0xe0, 0x69, 0x8f,
				0xd9, 0xc2, 0x93, 0xda, // timestamp
				0x1c, 0x64, 0x27, 0x82, // SSRC
			},
			err: errHeaderSizeInsufficientForExtension,
		},
		"MissingExtensionData": {
			input: []byte{
				0x90, 0xe0, 0x69, 0x8f,
				0xd9, 0xc2, 0x93, 0xda, // timestamp
				0x1c, 0x64, 0x27, 0x82, // SSRC
				0xBE, 0xDE, 0x00, 0x03, // specified to have 3 extensions, but actually not
			},
			err: errHeaderSizeInsufficientForExtension,
		},
		"MissingExtensionDataPayload": {
			input: []byte{
				0x90, 0xe0, 0x69, 0x8f,
				0xd9, 0xc2, 0x93, 0xda, // timestamp
				0x1c, 0x64, 0x27, 0x82, // SSRC
				0xBE, 0xDE, 0x00, 0x01, // have 1 extension
				0x12, 0x00, // length of the payload is expected to be 3, but actually have only 1
			},
			err: errHeaderSizeInsufficientForExtension,
		},
	}

	for name, testCase := range cases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			h := &Header{}
			_, err := h.Unmarshal(testCase.input)
			assert.ErrorIs(t, err, testCase.err)
		})
	}
}

// https://github.com/pion/rtp/issues/275
func TestUnmarshal_OneByteExtensionWithoutRTPPayload(t *testing.T) {
	rawPkt := []byte{
		0x10, 0x64, 0x57, 0x49, 0x00, 0x00, 0x01, 0x90, 0x12, 0x34, 0xAB, 0xCD,
		0xBE, 0xDE, 0x00, 0x01, // One-Byte extension header, 4 bytes
		0x02,             // ID=0, Len=2 (3 bytes data)
		0x01, 0x02, 0x03, // Extension data
	}

	p := &Packet{}
	assert.NoError(t, p.Unmarshal(rawPkt))
}

func TestUnmarshal_TwoByteExtensionWithoutRTPPayload(t *testing.T) {
	rawPkt := []byte{
		0x10, 0x64, 0x57, 0x49, 0x00, 0x00, 0x01, 0x90, 0x12, 0x34, 0xAB, 0xCD,
		0x10, 0x00, 0x00, 0x01, // Two-Byte extension header, 4 bytes
		0x02, 0x02, // ID=0, Len=2 (2 bytes data)
		0x02, 0x03, // Extension data
	}

	p := &Packet{}
	assert.NoError(t, p.Unmarshal(rawPkt))
}

func TestUnmarshal_NonStandardExtensionWithoutRTPPayload(t *testing.T) {
	rawPkt := []byte{
		0x10, 0x64, 0x57, 0x49, 0x00, 0x00, 0x01, 0x90, 0x12, 0x34, 0xAB, 0xCD,
		0xAA, 0xAA, 0x00, 0x01, // Non-standard header extension 0xAAAA, 4 bytes
		0x01, 0x02, 0x03, 0x04, // Extension data
	}

	p := &Packet{}
	assert.NoError(t, p.Unmarshal(rawPkt))
}

func TestUnmarshal_EmptyOneByteExtensionWithoutRTPPayload(t *testing.T) {
	rawPkt := []byte{
		0x10, 0x64, 0x57, 0x49, 0x00, 0x00, 0x01, 0x90, 0x12, 0x34, 0xAB, 0xCD,
		0xBE, 0xDE, 0x00, 0x00, // One-Byte extension header, 0 bytes
	}

	p := &Packet{}
	assert.NoError(t, p.Unmarshal(rawPkt))
}

func TestUnmarshal_EmptyTwoByteExtensionWithoutRTPPayload(t *testing.T) {
	rawPkt := []byte{
		0x10, 0x64, 0x57, 0x49, 0x00, 0x00, 0x01, 0x90, 0x12, 0x34, 0xAB, 0xCD,
		0x10, 0x00, 0x00, 0x00, // Two-Byte extension header, 0 bytes
	}

	p := &Packet{}
	assert.NoError(t, p.Unmarshal(rawPkt))
}

func TestUnmarshal_EmptyNonStandardExtensionWithoutRTPPayload(t *testing.T) {
	rawPkt := []byte{
		0x10, 0x64, 0x57, 0x49, 0x00, 0x00, 0x01, 0x90, 0x12, 0x34, 0xAB, 0xCD,
		0xAA, 0xAA, 0x00, 0x00, // Non-standard header extension 0xAAAA, 0 bytes
	}

	p := &Packet{}
	assert.NoError(t, p.Unmarshal(rawPkt))
}

func TestRoundtrip(t *testing.T) {
	rawPkt := []byte{
		0x00, 0x10, 0x23, 0x45, 0x12, 0x34, 0x45, 0x67, 0xCC, 0xDD, 0xEE, 0xFF,
		0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77,
	}
	payload := rawPkt[12:]

	packet := &Packet{}
	assert.NoError(t, packet.Unmarshal(rawPkt))
	assert.Equal(t, payload, packet.Payload)

	buf, err := packet.Marshal()
	assert.NoError(t, err)
	assert.Equal(t, rawPkt, buf)
	assert.Equal(t, payload, packet.Payload)
}

func TestCloneHeader(t *testing.T) {
	header := Header{
		Marker:           true,
		Extension:        true,
		ExtensionProfile: 1,
		Extensions: []Extension{
			{0, []byte{
				0xFF, 0xFF, 0xFF, 0xFF,
			}},
		},
		Version:        2,
		PayloadType:    96,
		SequenceNumber: 27023,
		Timestamp:      3653407706,
		SSRC:           476325762,
	}
	clone := header.Clone()
	assert.Equal(t, header, clone)

	header.CSRC = append(header.CSRC, 1)
	assert.NotEqual(t, len(clone.CSRC), len(header.CSRC), "Expected CSRC to be unchanged")
	header.Extensions[0].payload[0] = 0x1F
	assert.NotEqual(t, clone.Extensions[0].payload[0], byte(0x1F), "Expected extension to be unchanged")
}

func TestClonePacket(t *testing.T) {
	rawPkt := []byte{
		0x90, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0xBE, 0xDE, 0x00, 0x01, 0x50, 0xAA, 0x00, 0x00,
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	packet := &Packet{
		Payload: rawPkt[20:],
	}

	clone := packet.Clone()
	assert.Equal(t, packet, clone)

	packet.Payload[0] = 0x1F
	assert.NotEqual(t, clone.Payload[0], 0x1F, "Expected payload to be unchanged")
}

func TestMarshalRTPPacketFuncs(t *testing.T) {
	// packet with only padding
	rawPkt := []byte{
		0xb0, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0x00, 0x01, 0x00, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x05,
	}
	parsedPacket := &Packet{
		Header: Header{
			Padding:          true,
			Marker:           true,
			Extension:        true,
			ExtensionProfile: 1,
			Extensions: []Extension{
				{0, []byte{
					0xFF, 0xFF, 0xFF, 0xFF,
				}},
			},
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
			PaddingSize:    5,
		},
		Payload: []byte{},
	}

	buf := make([]byte, 100)
	n, err := MarshalPacketTo(buf, &parsedPacket.Header, parsedPacket.Payload)
	assert.NoError(t, err)
	assert.Equal(t, len(rawPkt), n)
	assert.Equal(t, rawPkt, buf[:n])

	assert.Equal(t, n, PacketMarshalSize(&parsedPacket.Header, parsedPacket.Payload))

	hdrLen, packetLen := HeaderAndPacketMarshalSize(&parsedPacket.Header, parsedPacket.Payload)
	assert.Equal(t, parsedPacket.Header.MarshalSize(), hdrLen)
	assert.Equal(t, n, packetLen)
}

func TestDeprecatedPaddingSizeField(t *testing.T) {
	// packet with only padding
	rawPkt := []byte{
		0xb0, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0x00, 0x01, 0x00, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x05,
	}
	parsedPacket := &Packet{
		Header: Header{
			Padding:          true,
			Marker:           true,
			Extension:        true,
			ExtensionProfile: 1,
			Extensions: []Extension{
				{0, []byte{
					0xFF, 0xFF, 0xFF, 0xFF,
				}},
			},
			Version:        2,
			PayloadType:    96,
			SequenceNumber: 27023,
			Timestamp:      3653407706,
			SSRC:           476325762,
		},
		Payload:     []byte{},
		PaddingSize: 5,
	}

	buf, err := parsedPacket.Marshal()
	assert.NoError(t, err)
	assert.Equal(t, rawPkt, buf)
	assert.EqualValues(t, 0, parsedPacket.Header.PaddingSize)

	assert.Equal(t, len(rawPkt), parsedPacket.MarshalSize())
	assert.EqualValues(t, 0, parsedPacket.Header.PaddingSize)

	parsedPacket2 := parsedPacket.Clone()
	assert.EqualValues(t, 5, parsedPacket2.PaddingSize)
	assert.EqualValues(t, 0, parsedPacket2.Header.PaddingSize)
}

func TestSetExtensionWithProfile(t *testing.T) {
	t.Run("add two-byte extension due to the size > 16", func(t *testing.T) {
		h := Header{}
		assert.NoError(t, h.SetExtension(1, make([]byte, 2)))
		assert.NoError(t, h.SetExtension(2, make([]byte, 3)))

		// Adding another extension that requires two-byte header extension
		assert.NoError(t, h.SetExtensionWithProfile(3, make([]byte, 20), ExtensionProfileTwoByte))
		assert.Equal(t, h.ExtensionProfile, uint16(ExtensionProfileTwoByte))
	})

	t.Run("add two-byte extension due to id > 14", func(t *testing.T) {
		h := Header{}
		assert.NoError(t, h.SetExtension(1, make([]byte, 2)))
		assert.NoError(t, h.SetExtension(2, make([]byte, 3)))

		// Adding another extension that requires two-byte header extension
		// because the extmap ID is greater than 14.
		assert.NoError(t, h.SetExtensionWithProfile(16, make([]byte, 4), ExtensionProfileTwoByte))
		assert.Equal(t, h.ExtensionProfile, uint16(ExtensionProfileTwoByte))
	})

	t.Run("Downgrade 2 byte header Extension", func(t *testing.T) {
		pkt := []byte{
			0x90, 0x60, 0x00, 0x01, // V=2, P=0, X=1, CC=0; M=0, PT=96; sequence=1
			0x00, 0x00, 0x00, 0x01, // timestamp=1
			0x12, 0x34, 0x56, 0x78, // SSRC=0x12345678
			0x10, 0x00, 0x00, 0x01, // profile=0x1000 (two-byte), length=1 (4 bytes)
			0x01, 0x02, 0x00, 0x01, // id=1, len=2, data=0x00,0x01 (padded to 32-bit)
		}
		h := Header{}

		_, err := h.Unmarshal(pkt)
		assert.NoError(t, err)
		assert.Equal(t, h.ExtensionProfile, uint16(ExtensionProfileTwoByte))

		assert.NoError(t, h.SetExtensionWithProfile(1, []byte{0x02, 0x03}, ExtensionProfileOneByte))
		assert.Equal(t, h.ExtensionProfile, uint16(ExtensionProfileOneByte))

		pkt, err = h.Marshal()
		assert.NoError(t, err)

		assert.Equal(t, pkt, []byte{
			0x90, 0x60, 0x00, 0x01,
			0x00, 0x00, 0x00, 0x01,
			0x12, 0x34, 0x56, 0x78,
			0xbe, 0xde, 0x00, 0x01,
			0x11, 0x02, 0x03, 0x00,
		})
	})

	t.Run("Do not mutate packet for invalid extension", func(t *testing.T) {
		h := Header{}
		assert.NoError(t, h.SetExtension(1, make([]byte, 2)))

		assert.Error(t, h.SetExtensionWithProfile(16, make([]byte, 4096), ExtensionProfileTwoByte))

		assert.Equal(t, h.ExtensionProfile, uint16(ExtensionProfileOneByte))
		assert.Len(t, h.Extensions, 1)
	})
}

func BenchmarkMarshal(b *testing.B) {
	rawPkt := []byte{
		0x90, 0x60, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0x00, 0x01, 0x00, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0x98, 0x36, 0xbe, 0x88, 0x9e,
	}

	packet := &Packet{}
	err := packet.Unmarshal(rawPkt)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err = packet.Marshal()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshalTo(b *testing.B) {
	rawPkt := []byte{
		0x90, 0x60, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0x00, 0x01, 0x00, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0x98, 0x36, 0xbe, 0x88, 0x9e,
	}

	packet := &Packet{}

	err := packet.Unmarshal(rawPkt)
	if err != nil {
		b.Fatal(err)
	}

	buf := [100]byte{}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err = packet.MarshalTo(buf[:])
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	pkt := Packet{
		Header: Header{
			Extension:        true,
			CSRC:             []uint32{1, 2},
			ExtensionProfile: ExtensionProfileTwoByte,
			Extensions: []Extension{
				{id: 1, payload: []byte{3, 4}},
				{id: 2, payload: []byte{5, 6}},
			},
		},
		Payload: []byte{
			0x07, 0x08, 0x09, 0x0a,
		},
	}
	rawPkt, errMarshal := pkt.Marshal()
	if errMarshal != nil {
		b.Fatal(errMarshal)
	}

	b.Run("SharedStruct", func(b *testing.B) {
		packet := &Packet{}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := packet.Unmarshal(rawPkt); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("NewStruct", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			packet := &Packet{}
			if err := packet.Unmarshal(rawPkt); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// https://github.com/pion/rtp/issues/315
func TestMarshalSizePanic(t *testing.T) {
	hdr := &Header{
		Extension: true,
	}

	assert.Equal(t, 16, hdr.MarshalSize())
}

// https://github.com/pion/rtp/issues/315
func TestMarshalToPanic(t *testing.T) {
	hdr := &Header{
		Extension: true,
	}

	buf := make([]byte, 16)
	n, err := hdr.MarshalTo(buf)
	assert.NoError(t, err)
	assert.Equal(t, 16, n)
}

func BenchmarkUnmarshalHeader(b *testing.B) {
	rawPkt := []byte{
		0x90, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda,
		0x1c, 0x64, 0x27, 0x82, 0xBE, 0xDE, 0x00, 0x01,
		0x50, 0xAA, 0x00, 0x00, 0x98, 0x36, 0xbe, 0x88,
	}
	b.Run("NewStructWithoutCSRC", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			h := &Header{}
			if _, err := h.Unmarshal(rawPkt); err != nil {
				b.Fatal(err)
			}
		}
	})

	rawPkt = []byte{
		0x92, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda,
		0x1c, 0x64, 0x27, 0x82, 0x00, 0x00, 0x11, 0x11,
		0x00, 0x00, 0x22, 0x22, 0xBE, 0xDE, 0x00, 0x01,
		0x50, 0xAA, 0x00, 0x00, 0x98, 0x36, 0xbe, 0x88,
	}
	b.Run("NewStructWithCSRC", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			h := &Header{}
			if _, err := h.Unmarshal(rawPkt); err != nil {
				b.Fatal(err)
			}
		}
	})
}
