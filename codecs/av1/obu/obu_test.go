// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package obu

import (
	"bytes"
	"errors"
	"testing"
)

func TestOBUType(t *testing.T) {
	for _, test := range []struct {
		Type      Type
		TypeValue uint8
		Str       string
	}{
		{OBUSequenceHeader, 1, "OBU_SEQUENCE_HEADER"},
		{OBUTemporalDelimiter, 2, "OBU_TEMPORAL_DELIMITER"},
		{OBUFrameHeader, 3, "OBU_FRAME_HEADER"},
		{OBUTileGroup, 4, "OBU_TILE_GROUP"},
		{OBUMetadata, 5, "OBU_METADATA"},
		{OBUFrame, 6, "OBU_FRAME"},
		{OBURedundantFrameHeader, 7, "OBU_REDUNDANT_FRAME_HEADER"},
		{OBUTileList, 8, "OBU_TILE_LIST"},
		{OBUPadding, 15, "OBU_PADDING"},
		{Type(0), 0, "OBU_RESERVED"},
		{Type(9), 9, "OBU_RESERVED"},
	} {
		test := test

		if test.Type.String() != test.Str {
			t.Errorf("Expected %s, got %s", test.Str, test.Type.String())
		}

		if uint8(test.Type) != test.TypeValue {
			t.Errorf("Expected %d, got %d", test.TypeValue, uint8(test.Type))
		}
	}
}

func TestOBUHeader_NoExtension(t *testing.T) {
	tests := []struct {
		Value  byte
		Header Header
	}{
		{0b0_0000_0_0_0, Header{Type: Type(0), HasSizeField: false, Reserved1Bit: false}},
		{0b0_0001_0_0_0, Header{Type: OBUSequenceHeader, HasSizeField: false, Reserved1Bit: false}},
		{0b0_0010_0_0_0, Header{Type: OBUTemporalDelimiter, HasSizeField: false, Reserved1Bit: false}},
		{0b0_0011_0_0_0, Header{Type: OBUFrameHeader, HasSizeField: false, Reserved1Bit: false}},
		{0b0_0100_0_0_0, Header{Type: OBUTileGroup, HasSizeField: false, Reserved1Bit: false}},
		{0b0_0101_0_0_0, Header{Type: OBUMetadata, HasSizeField: false, Reserved1Bit: false}},
		{0b0_0110_0_0_0, Header{Type: OBUFrame, HasSizeField: false, Reserved1Bit: false}},
		{0b0_0111_0_0_0, Header{Type: OBURedundantFrameHeader, HasSizeField: false, Reserved1Bit: false}},
		{0b0_1000_0_0_0, Header{Type: OBUTileList, HasSizeField: false, Reserved1Bit: false}},
		{0b0_1111_0_0_0, Header{Type: OBUPadding, HasSizeField: false, Reserved1Bit: false}},
		{0b0_1001_0_0_0, Header{Type: Type(9), HasSizeField: false, Reserved1Bit: false}},
		{0b0_1001_0_1_0, Header{Type: Type(9), HasSizeField: true, Reserved1Bit: false}},
		{0b0_1001_0_1_1, Header{Type: Type(9), HasSizeField: true, Reserved1Bit: true}},
		{0b0_1001_0_0_1, Header{Type: Type(9), HasSizeField: false, Reserved1Bit: true}},
	}

	for _, test := range tests {
		test := test

		buff := []byte{test.Value}
		header, err := ParseOBUHeader(buff)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if *header != test.Header {
			t.Errorf("Expected %v, got %v", test.Header, *header)
		}

		if test.Header.Size() != 1 {
			t.Errorf("Expected size 1 for header without extension, got %d", test.Header.Size())
		}

		value := test.Header.Marshal()

		if len(value) != 1 {
			t.Errorf("Expected size 1 for header without extension, got %d", len(value))
		}

		if value[0] != test.Value {
			t.Errorf("Expected %d for header value, got %d", test.Value, value[0])
		}
	}
}

func TestOBUHeader_Extension(t *testing.T) {
	tests := []struct {
		HeaderValue          byte
		Header               Header
		ExtensionHeaderValue byte
		ExtensionHeader      ExtensionHeader
	}{
		{
			HeaderValue:          0b0_1001_1_0_0,
			Header:               Header{Type: Type(9), HasSizeField: false, Reserved1Bit: false},
			ExtensionHeaderValue: 0b001_01_000,
			ExtensionHeader:      ExtensionHeader{TemporalID: 1, SpatialID: 1},
		},
		{
			HeaderValue:          0b0_1001_1_1_1,
			Header:               Header{Type: Type(9), HasSizeField: true, Reserved1Bit: true},
			ExtensionHeaderValue: 0b010_01_000,
			ExtensionHeader:      ExtensionHeader{TemporalID: 2, SpatialID: 1},
		},
		{
			HeaderValue:          0b0_1001_1_1_1,
			Header:               Header{Type: Type(9), HasSizeField: true, Reserved1Bit: true},
			ExtensionHeaderValue: 0b011_01_000,
			ExtensionHeader:      ExtensionHeader{TemporalID: 3, SpatialID: 1},
		},
		{
			HeaderValue:          0b0_1001_1_1_1,
			Header:               Header{Type: Type(9), HasSizeField: true, Reserved1Bit: true},
			ExtensionHeaderValue: 0b111_10_000,
			ExtensionHeader:      ExtensionHeader{TemporalID: 7, SpatialID: 2},
		},
		{
			HeaderValue:          0b0_1001_1_1_1,
			Header:               Header{Type: Type(9), HasSizeField: true, Reserved1Bit: true},
			ExtensionHeaderValue: 0b111_11_001,
			ExtensionHeader:      ExtensionHeader{TemporalID: 7, SpatialID: 3, Reserved3Bits: 1},
		},
		{
			HeaderValue:          0b0_1001_1_1_1,
			Header:               Header{Type: Type(9), HasSizeField: true, Reserved1Bit: true},
			ExtensionHeaderValue: 0b111_11_111,
			ExtensionHeader:      ExtensionHeader{TemporalID: 7, SpatialID: 3, Reserved3Bits: 7},
		},
	}

	for _, test := range tests {
		test := test

		buff := []byte{test.HeaderValue, test.ExtensionHeaderValue}
		header, err := ParseOBUHeader(buff)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expected := Header{
			Type:         test.Header.Type,
			HasSizeField: test.Header.HasSizeField,
			Reserved1Bit: test.Header.Reserved1Bit,
		}
		if expected != test.Header {
			t.Errorf("Expected %v, got %v", test.Header, *header)
		}

		if header.Size() != 2 {
			t.Errorf("Expected size 2 for header with extension, got %d", test.Header.Size())
		}

		extension := header.ExtensionHeader
		if extension == nil {
			t.Fatalf("Expected extension header to be present")
		}

		if *extension != test.ExtensionHeader {
			t.Errorf("Expected %v, got %v", test.ExtensionHeader, *extension)
		}

		if extension.Marshal() != test.ExtensionHeaderValue {
			t.Errorf("Expected %d for extension header value, got %d", test.ExtensionHeaderValue, extension.Marshal())
		}

		value := header.Marshal()
		if len(value) != 2 {
			t.Errorf("Expected size 2 for header with extension, got %d", len(value))
		}

		if !bytes.Equal(value, buff) {
			t.Errorf("Expected %v for header value, got %v", buff, value)
		}
	}
}

func TestOBUHeader_Short(t *testing.T) {
	_, err := ParseOBUHeader([]byte{})
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
	if !errors.Is(err, ErrShortHeader) {
		t.Errorf("Expected ErrShortHeader, got %v", err)
	}

	// Missing extension header
	_, err = ParseOBUHeader([]byte{0b0_0000_1_0_0})
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	if !errors.Is(err, ErrShortHeader) {
		t.Errorf("Expected ErrShortHeader, got %v", err)
	}
}

func TestOBUHeader_Invalid(t *testing.T) {
	_, err := ParseOBUHeader([]byte{0b1_0010_0_0_1})
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidOBUHeader) {
		t.Errorf("Expected ErrInvalidOBUHeader, got %v", err)
	}
}

func TestOBUHeader_MarshalOutbound(t *testing.T) {
	// Marshal should turnicate the extension header values.
	header := Header{Type: Type(255)}
	if header.Marshal()[0] != 0b0_1111_000 {
		t.Errorf("Expected 0b0_1111_000, got %b", header.Marshal()[0])
	}

	extentionHeader := ExtensionHeader{TemporalID: 255}

	if extentionHeader.Marshal() != 0b111_00_000 {
		t.Errorf("Expected 0b111_00_000, got %b", extentionHeader.Marshal())
	}

	extensionHeader := ExtensionHeader{SpatialID: 255}
	if extensionHeader.Marshal() != 0b000_11_000 {
		t.Errorf("Expected 0b000_11_000, got %b", extensionHeader.Marshal())
	}

	extensionHeader = ExtensionHeader{Reserved3Bits: 255}
	if extensionHeader.Marshal() != 0b000_00_111 {
		t.Errorf("Expected 0b000_00_111, got %b", extensionHeader.Marshal())
	}
}

func TestOBUMarshal(t *testing.T) {
	obu := OBU{
		Header: Header{
			Type:         OBUFrame,
			HasSizeField: false,
			Reserved1Bit: false,
		},
		Payload: []byte{0x01, 0x02, 0x03},
	}

	data := obu.Marshal()

	if len(data) != 4 {
		t.Fatalf("Expected 4 bytes, got %d", len(data))
	}

	if data[0] != obu.Header.Marshal()[0] {
		t.Errorf("Expected header to be %v, got %v", obu.Header.Marshal(), data[0])
	}

	if !bytes.Equal(data[1:], obu.Payload) {
		t.Errorf("Expected payload to be %v, got %v", obu.Payload, data[1:])
	}
}

func TestOBUMarshal_ExtensionHeader(t *testing.T) {
	obu := OBU{
		Header: Header{
			Type:         OBUFrame,
			HasSizeField: false,
			Reserved1Bit: false,
			ExtensionHeader: &ExtensionHeader{
				TemporalID: 1,
				SpatialID:  2,
			},
		},
		Payload: []byte{0x01, 0x02, 0x03},
	}

	data := obu.Marshal()

	if len(data) != 5 {
		t.Fatalf("Expected 5 bytes, got %d", len(data))
	}

	if data[0] != obu.Header.Marshal()[0] {
		t.Errorf("Expected header to be %v, got %v", obu.Header.Marshal(), data[0])
	}

	if data[1] != obu.Header.ExtensionHeader.Marshal() {
		t.Errorf("Expected extension header to be %v, got %v", obu.Header.ExtensionHeader.Marshal(), data[1])
	}

	if !bytes.Equal(data[2:], obu.Payload) {
		t.Errorf("Expected payload to be %v, got %v", obu.Payload, data[1:])
	}
}

func TestOBUMarshal_HasOBUSize(t *testing.T) {
	const payloadSize = 128
	payload := make([]byte, payloadSize)

	for i := 0; i < payloadSize; i++ {
		payload[i] = byte(i)
	}

	obu := OBU{
		Header: Header{
			Type:         OBUFrame,
			HasSizeField: true,
			Reserved1Bit: false,
		},
		Payload: payload,
	}
	expected := append(
		obu.Header.Marshal(),
		append(
			// obu_size leb128 (128)
			[]byte{0x80, 0x01},
			obu.Payload...,
		)...,
	)

	data := obu.Marshal()

	if len(data) != payloadSize+3 {
		t.Fatalf("Expected 4 bytes, got %d", len(data))
	}

	if data[0] != obu.Header.Marshal()[0] {
		t.Errorf("Expected header to be %v, got %v", obu.Header.Marshal(), data[0])
	}

	if !bytes.Equal(data, expected) {
		t.Errorf("Expected payload to be %v, got %v", expected, data)
	}
}

func TestOBUMarshal_ZeroPayload(t *testing.T) {
	obu := OBU{
		Header: Header{
			Type:         OBUTemporalDelimiter,
			HasSizeField: false,
		},
	}

	data := obu.Marshal()

	if len(data) != 1 {
		t.Fatalf("Expected 1 byte, got %d", len(data))
	}

	obu = OBU{
		Header: Header{
			Type:         OBUTemporalDelimiter,
			HasSizeField: true,
		},
	}

	data = obu.Marshal()

	if len(data) != 2 {
		t.Fatalf("Expected two bytes, got %d", len(data))
	}

	if data[1] != 0 {
		t.Errorf("Expected 0 for size, got %d", data[1])
	}
}
