// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package rtp

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestBasic(t *testing.T) { // nolint:maintidx,cyclop
	packet := &Packet{}

	if err := packet.Unmarshal([]byte{}); err == nil {
		t.Fatal("Unmarshal did not error on zero length packet")
	}

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
			CSRC:           []uint32{},
		},
		Payload:     rawPkt[20:],
		PaddingSize: 0,
	}

	// Unmarshal to the used Packet should work as well.
	for i := 0; i < 2; i++ {
		t.Run(fmt.Sprintf("Run%d", i+1), func(t *testing.T) {
			if err := packet.Unmarshal(rawPkt); err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(packet, parsedPacket) {
				t.Errorf("TestBasic unmarshal: got %#v, want %#v", packet, parsedPacket)
			}

			if parsedPacket.Header.MarshalSize() != 20 {
				t.Errorf("wrong computed header marshal size")
			} else if parsedPacket.MarshalSize() != len(rawPkt) {
				t.Errorf("wrong computed marshal size")
			}

			raw, err := packet.Marshal()
			if err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(raw, rawPkt) {
				t.Errorf("TestBasic marshal: got %#v, want %#v", raw, rawPkt)
			}
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
			CSRC:           []uint32{},
		},
		Payload:     rawPkt[20:21],
		PaddingSize: 4,
	}
	if err := packet.Unmarshal(rawPkt); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(packet, parsedPacket) {
		t.Errorf("TestBasic padding unmarshal: got %#v, want %#v", packet, parsedPacket)
	}

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
			CSRC:           []uint32{},
		},
		Payload:     rawPkt[20:],
		PaddingSize: 0,
	}
	if err := packet.Unmarshal(rawPkt); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(packet, parsedPacket) {
		t.Errorf("TestBasic zero padding unmarshal: got %#v, want %#v", packet, parsedPacket)
	}

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
			CSRC:           []uint32{},
		},
		Payload:     []byte{},
		PaddingSize: 5,
	}
	if err := packet.Unmarshal(rawPkt); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(packet, parsedPacket) {
		t.Errorf("TestBasic padding only unmarshal: got %#v, want %#v", packet, parsedPacket)
	}
	if len(packet.Payload) != 0 {
		t.Errorf("Unmarshal of padding only packet has payload of non-zero length: %d", len(packet.Payload))
	}

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
			CSRC:           []uint32{},
		},
		Payload:     []byte{},
		PaddingSize: 0,
	}
	err := packet.Unmarshal(rawPkt)
	if err == nil {
		t.Fatal("Unmarshal did not error on packet with excessive padding")
	}
	if !errors.Is(err, errTooSmall) {
		t.Errorf("Expected error: %v, got: %v", errTooSmall, err)
	}

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
			CSRC:           []uint32{},
		},
		Payload:     rawPkt[20:21],
		PaddingSize: 4,
	}
	buf, err := parsedPacket.Marshal()
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(buf, rawPkt) {
		t.Errorf("TestBasic padding marshal: got %#v, want %#v", buf, rawPkt)
	}

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
			CSRC:           []uint32{},
		},
		Payload:     []byte{},
		PaddingSize: 5,
	}
	buf, err = parsedPacket.Marshal()
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(buf, rawPkt) {
		t.Errorf("TestBasic padding marshal: got %#v, want %#v", buf, rawPkt)
	}

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
			CSRC:           []uint32{},
		},
		Payload:     []byte{},
		PaddingSize: 5,
	}
	buf, err = parsedPacket.Marshal()
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(buf, rawPkt) {
		t.Errorf("TestBasic padding marshal: got %#v, want %#v", buf, rawPkt)
	}
}

func TestExtension(t *testing.T) {
	packet := &Packet{}

	missingExtensionPkt := []byte{
		0x90, 0x60, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82,
	}
	if err := packet.Unmarshal(missingExtensionPkt); err == nil {
		t.Fatal("Unmarshal did not error on packet with missing extension data")
	}

	invalidExtensionLengthPkt := []byte{
		0x90, 0x60, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0x99, 0x99, 0x99, 0x99,
	}
	if err := packet.Unmarshal(invalidExtensionLengthPkt); err == nil {
		t.Fatal("Unmarshal did not error on packet with invalid extension length")
	}

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
	if _, err := packet.Marshal(); err == nil {
		t.Fatal("Marshal did not error on packet with invalid extension length")
	}
}

func TestRFC8285OneByteExtension(t *testing.T) {
	packet := &Packet{}

	rawPkt := []byte{
		0x90, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0xBE, 0xDE, 0x00, 0x01, 0x50, 0xAA, 0x00, 0x00,
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	if err := packet.Unmarshal(rawPkt); err != nil {
		t.Fatal("Unmarshal err for valid extension")
	}

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
			CSRC:           []uint32{},
		},
		Payload: rawPkt[20:],
	}

	dstData, _ := packet.Marshal()
	if !bytes.Equal(dstData, rawPkt) {
		t.Errorf("Marshal failed raw \nMarshaled:\n%s\nrawPkt:\n%s", hex.Dump(dstData), hex.Dump(rawPkt))
	}
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
	if err := packet.Unmarshal(rawPkt); err != nil {
		t.Fatal("Unmarshal err for valid extension")
	}

	ext1 := packet.GetExtension(1)
	ext1Expect := []byte{0xAA}
	if !bytes.Equal(ext1, ext1Expect) {
		t.Errorf("Extension has incorrect data. Got: %+v, Expected: %+v", ext1, ext1Expect)
	}

	ext2 := packet.GetExtension(2)
	ext2Expect := []byte{0xBB}
	if !bytes.Equal(ext2, ext2Expect) {
		t.Errorf("Extension has incorrect data. Got: %+v, Expected: %+v", ext2, ext2Expect)
	}

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
			CSRC:           []uint32{},
		},
		Payload: rawPkt[20:],
	}

	dstData, _ := packet.Marshal()
	if !bytes.Equal(dstData, rawPkt) {
		t.Errorf("Marshal failed raw \nMarshaled:\n%s\nrawPkt:\n%s", hex.Dump(dstData), hex.Dump(rawPkt))
	}
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
	if err := packet.Unmarshal(rawPkt); err != nil {
		t.Fatal("Unmarshal err for valid extension")
	}

	ext1 := packet.GetExtension(1)
	ext1Expect := []byte{0xAA}
	if !bytes.Equal(ext1, ext1Expect) {
		t.Errorf("Extension has incorrect data. Got: %v+, Expected: %v+", ext1, ext1Expect)
	}

	ext2 := packet.GetExtension(2)
	ext2Expect := []byte{0xBB, 0xBB}
	if !bytes.Equal(ext2, ext2Expect) {
		t.Errorf("Extension has incorrect data. Got: %v+, Expected: %v+", ext2, ext2Expect)
	}

	ext3 := packet.GetExtension(3)
	ext3Expect := []byte{0xCC, 0xCC, 0xCC, 0xCC}
	if !bytes.Equal(ext3, ext3Expect) {
		t.Errorf("Extension has incorrect data. Got: %v+, Expected: %v+", ext3, ext3Expect)
	}

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
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(buf[:n], rawPktReMarshal) {
				t.Errorf("Marshal failed raw \nMarshaled:\n%s\nrawPkt:\n%s", hex.Dump(buf[:n]), hex.Dump(rawPktReMarshal))
			}
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
			CSRC:           []uint32{},
		},
		Payload: rawPkt[28:],
	}

	dstData, _ := packet.Marshal()
	if !bytes.Equal(dstData, rawPkt) {
		t.Errorf("Marshal failed raw \nMarshaled:\n%s\nrawPkt:\n%s", hex.Dump(dstData), hex.Dump(rawPkt))
	}
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
	if err := packet.Unmarshal(rawPkt); err != nil {
		t.Fatal("Unmarshal err for valid extension")
	}

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
			CSRC:           []uint32{},
		},
		Payload: rawPkt[44:],
	}

	dstData, _ := packet.Marshal()
	if !bytes.Equal(dstData, rawPkt) {
		t.Errorf("Marshal failed raw \nMarshaled:\n%s\nrawPkt:\n%s", hex.Dump(dstData), hex.Dump(rawPkt))
	}
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
	if err := packet.Unmarshal(rawPkt); err != nil {
		t.Fatal("Unmarshal err for valid extension")
	}

	ext1 := packet.GetExtension(1)
	ext1Expect := []byte{}
	if !bytes.Equal(ext1, ext1Expect) {
		t.Errorf("Extension has incorrect data. Got: %v+, Expected: %v+", ext1, ext1Expect)
	}

	ext2 := packet.GetExtension(2)
	ext2Expect := []byte{0xBB}
	if !bytes.Equal(ext2, ext2Expect) {
		t.Errorf("Extension has incorrect data. Got: %v+, Expected: %v+", ext2, ext2Expect)
	}

	ext3 := packet.GetExtension(3)
	ext3Expect := []byte{0xCC, 0xCC, 0xCC, 0xCC}
	if !bytes.Equal(ext3, ext3Expect) {
		t.Errorf("Extension has incorrect data. Got: %v+, Expected: %v+", ext3, ext3Expect)
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
			CSRC:           []uint32{},
		},
		Payload: rawPkt[40:],
	}

	dstData, _ := packet.Marshal()
	if !bytes.Equal(dstData, rawPkt) {
		t.Errorf("Marshal failed raw \nMarshaled: %+v,\nrawPkt:    %+v", dstData, rawPkt)
	}
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
			CSRC:           []uint32{},
		},
		Payload: payload,
	}

	err := packet.GetExtension(1)
	if err != nil {
		t.Error("Should return nil on GetExtension when h.Extension: false")
	}
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
			CSRC:           []uint32{},
		},
		Payload: payload,
	}

	ext := packet.GetExtension(1)
	if ext == nil {
		t.Error("Extension should exist")
	}

	err := packet.DelExtension(1)
	if err != nil {
		t.Error("Should successfully delete extension")
	}

	ext = packet.GetExtension(1)
	if ext != nil {
		t.Error("Extension should not exist")
	}

	err = packet.DelExtension(1)
	if err == nil {
		t.Error("Should return error when deleting extension that doesnt exist")
	}
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
			CSRC:           []uint32{},
		},
		Payload: payload,
	}

	ids := packet.GetExtensionIDs()
	if ids == nil {
		t.Error("Extension should exist")
	}
	if len(ids) != len(packet.Extensions) {
		t.Errorf(
			"The number of IDs should be equal to the number of extensions,want=%d,have=%d",
			len(packet.Extensions),
			len(ids),
		)
	}

	for _, id := range ids {
		ext := packet.GetExtension(id)
		if ext == nil {
			t.Error("Extension should exist")
		}
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
			CSRC:           []uint32{},
		},
		Payload: payload,
	}

	ids := packet.GetExtensionIDs()
	if ids != nil {
		t.Error("Should return nil on GetExtensionIDs when h.Extensions is nil")
	}
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
			CSRC:           []uint32{},
		},
		Payload: payload,
	}

	err := packet.DelExtension(1)
	if err == nil {
		t.Error("Should return error on DelExtension when h.Extension: false")
	}
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
			CSRC:           []uint32{},
		},
		Payload: payload,
	}

	extension := []byte{0xAA, 0xAA}
	err := packet.SetExtension(1, extension)
	if err != nil {
		t.Error("Error setting extension")
	}

	if packet.Extension != true {
		t.Error("Extension should be set to true")
	}

	if packet.ExtensionProfile != 0xBEDE {
		t.Error("Extension profile should be set to 0xBEDE")
	}

	if len(packet.Extensions) != 1 {
		t.Error("Extensions should be set to 1")
	}

	if !bytes.Equal(packet.GetExtension(1), extension) {
		t.Error("Extension value is not set")
	}
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
			CSRC:           []uint32{},
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
	if err != nil {
		t.Error("Error setting extension")
	}

	if packet.ExtensionProfile != 0xBEDE {
		t.Error("Extension profile should be set to 0xBEDE")
	}
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
			CSRC:           []uint32{},
		},
		Payload: payload,
	}

	if !bytes.Equal(packet.GetExtension(1), []byte{0xAA}) {
		t.Error("Extension value not initialize properly")
	}

	extension := []byte{0xBB}
	err := packet.SetExtension(1, extension)
	if err != nil {
		t.Error("Error setting extension")
	}

	if !bytes.Equal(packet.GetExtension(1), extension) {
		t.Error("Extension value was not set")
	}
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
			CSRC:           []uint32{},
		},
		Payload: payload,
	}

	if packet.SetExtension(0, []byte{0xBB}) == nil {
		t.Error("SetExtension did not error on invalid id")
	}

	if packet.SetExtension(15, []byte{0xBB}) == nil {
		t.Error("SetExtension did not error on invalid id")
	}
}

func TestRFC8285OneByteExtensionTermianteProcessingWhenReservedIDEncountered(t *testing.T) {
	packet := &Packet{}

	reservedIDPkt := []byte{
		0x90, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0xBE, 0xDE, 0x00, 0x01, 0xF0, 0xAA, 0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	if err := packet.Unmarshal(reservedIDPkt); err != nil {
		t.Error("Unmarshal error on packet with reserved extension id")
	}

	if len(packet.Extensions) != 0 {
		t.Error("Extensions should be empty for invalid id")
	}

	payload := reservedIDPkt[17:]
	if !bytes.Equal(packet.Payload, payload) {
		t.Errorf("p.Payload must be same as payload.\n  p.Payload: %+v,\n payload: %+v",
			packet.Payload, payload,
		)
	}
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
			CSRC:           []uint32{},
		},
		Payload: payload,
	}

	if packet.SetExtension(1, []byte{
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
	}) == nil {
		t.Error("SetExtension did not error on too large payload")
	}
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
			CSRC:           []uint32{},
		},
		Payload: payload,
	}

	extension := []byte{
		0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
		0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
	}
	err := packet.SetExtension(1, extension)
	if err != nil {
		t.Error("Error setting extension")
	}

	if packet.Extension != true {
		t.Error("Extension should be set to true")
	}

	if packet.ExtensionProfile != 0x1000 {
		t.Error("Extension profile should be set to 0xBEDE")
	}

	if len(packet.Extensions) != 1 {
		t.Error("Extensions should be set to 1")
	}

	if !bytes.Equal(packet.GetExtension(1), extension) {
		t.Error("Extension value is not set")
	}
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
			CSRC:           []uint32{},
		},
		Payload: payload,
	}

	if !bytes.Equal(packet.GetExtension(1), []byte{0xAA}) {
		t.Error("Extension value not initialize properly")
	}

	extension := []byte{
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
		0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB,
	}
	err := packet.SetExtension(1, extension)
	if err != nil {
		t.Error("Error setting extension")
	}

	if !bytes.Equal(packet.GetExtension(1), extension) {
		t.Error("Extension value was not set")
	}
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
			CSRC:           []uint32{},
		},
		Payload: payload,
	}

	if packet.SetExtension(1, []byte{
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
	}) == nil {
		t.Error("SetExtension did not error on too large payload")
	}
}

func TestRFC8285Padding(t *testing.T) {
	header := &Header{}

	for _, payload := range [][]byte{
		{
			0b00010000,                      // header.Extension = true
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // SequenceNumber, Timestamp, SSRC
			0xBE, 0xDE, // header.ExtensionProfile = extensionProfileOneByte
			0, 1, // extensionLength
			0, 0, 0, // padding
			1, // extid
		},
		{
			0b00010000,                      // header.Extension = true
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // SequenceNumber, Timestamp, SSRC
			0x10, 0x00, // header.ExtensionProfile = extensionProfileOneByte
			0, 1, // extensionLength
			0, 0, 0, // padding
			1, // extid
		},
	} {
		_, err := header.Unmarshal(payload)
		if !errors.Is(err, errHeaderSizeInsufficientForExtension) {
			t.Fatal("Expected errHeaderSizeInsufficientForExtension")
		}
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
			CSRC:           []uint32{},
		},
		Payload: payload,
	}

	expect := []byte{0xBB}
	if packet.SetExtension(0, expect) != nil {
		t.Error("SetExtension should not error on valid id")
	}

	actual := packet.GetExtension(0)
	if !bytes.Equal(actual, expect) {
		t.Error("p.GetExtension returned incorrect value.")
	}
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
			CSRC:             []uint32{},
		},
		Payload: payload,
	}

	if packet.SetExtension(1, []byte{0xBB}) == nil {
		t.Error("SetExtension did not error on invalid id")
	}
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
			if !errors.Is(err, testCase.err) {
				t.Errorf("Expected error: %v, got: %v", testCase.err, err)
			}
		})
	}
}

func TestRoundtrip(t *testing.T) {
	rawPkt := []byte{
		0x00, 0x10, 0x23, 0x45, 0x12, 0x34, 0x45, 0x67, 0xCC, 0xDD, 0xEE, 0xFF,
		0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77,
	}
	payload := rawPkt[12:]

	packet := &Packet{}
	if err := packet.Unmarshal(rawPkt); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(payload, packet.Payload) {
		t.Errorf("p.Payload must be same as payload.\n  payload: %+v,\np.Payload: %+v",
			payload, packet.Payload,
		)
	}

	buf, err := packet.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rawPkt, buf) {
		t.Errorf("buf must be same as rawPkt.\n   buf: %+v,\nrawPkt: %+v", buf, rawPkt)
	}
	if !bytes.Equal(payload, packet.Payload) {
		t.Errorf("p.Payload must be same as payload.\n  payload: %+v,\np.Payload: %+v",
			payload, packet.Payload,
		)
	}
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
		CSRC:           []uint32{},
	}
	clone := header.Clone()
	if !reflect.DeepEqual(header, clone) {
		t.Errorf("Cloned clone does not match the original")
	}

	header.CSRC = append(header.CSRC, 1)
	if len(clone.CSRC) == len(header.CSRC) {
		t.Errorf("Expected CSRC to be unchanged")
	}
	header.Extensions[0].payload[0] = 0x1F
	if clone.Extensions[0].payload[0] == 0x1F {
		t.Errorf("Expected Extensions to be unchanged")
	}
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
	if !reflect.DeepEqual(packet, clone) {
		t.Errorf("Cloned Packet does not match the original")
	}

	packet.Payload[0] = 0x1F
	if clone.Payload[0] == 0x1F {
		t.Errorf("Expected Payload to be unchanged")
	}
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
			ExtensionProfile: extensionProfileTwoByte,
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
