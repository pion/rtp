// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package obu

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
		assert.Equal(t, test.Str, test.Type.String())
		assert.Equal(t, test.TypeValue, uint8(test.Type))
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
		assert.NoError(t, err)
		assert.Equal(t, test.Header, *header)
		assert.Equal(t, 1, header.Size())

		value := test.Header.Marshal()
		assert.Len(t, value, 1, "Expected size 1 for header without extension")
		assert.Equal(t, test.Value, value[0])
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
		assert.NoError(t, err)

		expected := Header{
			Type:         test.Header.Type,
			HasSizeField: test.Header.HasSizeField,
			Reserved1Bit: test.Header.Reserved1Bit,
		}
		assert.Equal(t, expected, test.Header)
		assert.Equal(t, 2, header.Size())

		extension := header.ExtensionHeader
		assert.NotNil(t, extension)
		assert.Equal(t, test.ExtensionHeader, *extension)
		assert.Equal(t, test.ExtensionHeaderValue, extension.Marshal())

		value := header.Marshal()
		assert.Lenf(
			t, value, 2,
			"Expected size 2 for header with extension, got %d", len(value),
		)
		assert.Equal(t, buff, value)
	}
}

func TestOBUHeader_Short(t *testing.T) {
	_, err := ParseOBUHeader([]byte{})
	assert.ErrorIs(t, err, ErrShortHeader)

	// Missing extension header
	_, err = ParseOBUHeader([]byte{0b0_0000_1_0_0})
	assert.ErrorIs(t, err, ErrShortHeader)
}

func TestOBUHeader_Invalid(t *testing.T) {
	// forbidden bit is set
	_, err := ParseOBUHeader([]byte{0b1_0010_0_0_1})
	assert.ErrorIs(t, err, ErrInvalidOBUHeader)
}

func TestOBUHeader_MarshalOutbound(t *testing.T) {
	// Marshal should turnicate the extension header values.
	header := Header{Type: Type(255)}
	assert.Equal(t, uint8(0b0_1111_000), header.Marshal()[0])

	extentionHeader := ExtensionHeader{TemporalID: 255}
	assert.Equal(t, uint8(0b111_00_000), extentionHeader.Marshal())

	extensionHeader := ExtensionHeader{SpatialID: 255}
	assert.Equal(t, uint8(0b000_11_000), extensionHeader.Marshal())

	extensionHeader = ExtensionHeader{Reserved3Bits: 255}
	assert.Equal(t, uint8(0b000_00_111), extensionHeader.Marshal())
}

func TestOBUMarshal(t *testing.T) {
	testOBU := OBU{
		Header: Header{
			Type:         OBUFrame,
			HasSizeField: false,
			Reserved1Bit: false,
		},
		Payload: []byte{0x01, 0x02, 0x03},
	}

	data := testOBU.Marshal()
	assert.Len(t, data, 4)
	assert.Equal(t, testOBU.Header.Marshal()[0], data[0], "Expected header to be equal")
	assert.Equal(t, testOBU.Payload, data[1:])
}

func TestOBUMarshal_ExtensionHeader(t *testing.T) {
	testOBU := OBU{
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
	data := testOBU.Marshal()
	assert.Len(t, data, 5)
	assert.Equal(t, testOBU.Header.Marshal()[0], data[0], "Expected header to be equal")
	assert.Equal(t, testOBU.Header.ExtensionHeader.Marshal(), data[1], "Expected extension header to equal")
	assert.Equal(t, testOBU.Payload, data[2:])
}

func TestOBUMarshal_HasOBUSize(t *testing.T) {
	const payloadSize = 128
	payload := make([]byte, payloadSize)

	for i := 0; i < payloadSize; i++ {
		payload[i] = byte(i)
	}

	testOBU := OBU{
		Header: Header{
			Type:         OBUFrame,
			HasSizeField: true,
			Reserved1Bit: false,
		},
		Payload: payload,
	}
	expected := append(
		testOBU.Header.Marshal(),
		append(
			// obu_size leb128 (128)
			[]byte{0x80, 0x01},
			testOBU.Payload...,
		)...,
	)

	data := testOBU.Marshal()
	assert.Len(t, data, payloadSize+3)
	assert.Equal(t, testOBU.Header.Marshal()[0], data[0], "Expected header to be equal")
	assert.Equal(t, expected, data)
}

func TestOBUMarshal_ZeroPayload(t *testing.T) {
	testOBU := OBU{
		Header: Header{
			Type:         OBUTemporalDelimiter,
			HasSizeField: false,
		},
	}
	data := testOBU.Marshal()
	assert.Len(t, data, 1)

	testOBU = OBU{
		Header: Header{
			Type:         OBUTemporalDelimiter,
			HasSizeField: true,
		},
	}
	data = testOBU.Marshal()
	assert.Len(t, data, 2)
	assert.Equal(t, uint8(0), data[1], "Expected 0 for size")
}
