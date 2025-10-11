// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package codecs

import (
	"testing"

	"github.com/pion/rtp/codecs/av1/obu"
	"github.com/stretchr/testify/assert"
)

// Create an AV1 OBU for testing. Returns one without the obu_size_field and another with it included.
func createAV1OBU(obuType obu.Type, payload []byte) ([]byte, []byte) {
	header := obu.Header{Type: obuType}
	withoutSize := createTestPayload(header, payload)
	header.HasSizeField = true
	withSize := createTestPayload(header, payload)

	return withoutSize, withSize
}

func createTestPayload(obuHeader obu.Header, payload []byte) []byte {
	buf := make([]byte, 0)
	buf = append(buf, obuHeader.Marshal()...)
	if obuHeader.HasSizeField {
		buf = append(buf, obu.WriteToLeb128(uint(len(payload)))...)
	}

	buf = append(buf, payload...)

	return buf
}

func TestAV1Depacketizer_invalidPackets(t *testing.T) {
	depacketizer := AV1Depacketizer{}
	_, err := depacketizer.Unmarshal([]byte{})
	assert.ErrorIs(t, err, errShortPacket)

	_, err = depacketizer.Unmarshal([]byte{0b11000000, 0xFF})
	assert.ErrorIs(t, err, obu.ErrFailedToReadLEB128)

	_, err = depacketizer.Unmarshal(append([]byte{0b00000000}, obu.WriteToLeb128(0x99)...))
	assert.ErrorIs(t, err, errShortPacket)

	_, err = depacketizer.Unmarshal(append([]byte{0b00000000}, obu.WriteToLeb128(0x01)...))
	assert.ErrorIs(t, err, errShortPacket)

	_, err = depacketizer.Unmarshal(
		append(
			[]byte{0b00110000},
			append(
				obu.WriteToLeb128(1),
				[]byte{0x01}...,
			)...,
		),
	)
	assert.ErrorIs(t, err, errShortPacket)
}

func TestAV1Depacketizer_singleOBU(t *testing.T) {
	payload := []byte{0x01, 0x02, 0x03}
	obuData, expectedOBU := createAV1OBU(4, payload)

	packet := make([]byte, 0)

	packet = append(packet, []byte{0b00000000}...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obuData)))...)
	packet = append(packet, obuData...)

	d := AV1Depacketizer{}
	obu, err := d.Unmarshal(packet)
	assert.NoError(t, err)
	assert.Equal(t, expectedOBU, obu)
}

func TestAV1Depacketizer_singleOBUWithPadding(t *testing.T) {
	payload := []byte{0x01, 0x02, 0x03}
	obuData, expectedOBU := createAV1OBU(4, payload)

	packet := make([]byte, 0)

	packet = append(packet, []byte{0b00000000}...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obuData)))...)
	packet = append(packet, obuData...)
	// padding
	packet = append(packet, []byte{0x00, 0x00, 0x00}...)

	d := AV1Depacketizer{}
	obu, err := d.Unmarshal(packet)
	assert.NoError(t, err)
	assert.Equal(t, expectedOBU, obu)
}

// AV1 OBUs shouldn't include the obu_size_field when packetized in RTP,
// but we still support it since it's encountered in the wild (Including pion old clients).
func TestAV1Depacketizer_withOBUSize(t *testing.T) {
	payload := []byte{0x01, 0x02, 0x03}
	_, obuData := createAV1OBU(4, payload)

	packet := make([]byte, 0)

	packet = append(packet, []byte{0b00000000}...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obuData)))...)
	packet = append(packet, obuData...)

	d := AV1Depacketizer{}
	obu, err := d.Unmarshal(packet)
	assert.NoError(t, err)
	assert.Equal(t, obuData, obu)
}

func TestAV1Depacketizer_validateOBUSize(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		err     error
	}{
		{
			name: "invalid OBU size",
			payload: []byte{
				0,    // Aggregation header
				0x02, // Length field
				0x22, // OBU header (has_size_field = 1)
				0xFF, // Invalid LEB128 size
			},
			err: obu.ErrFailedToReadLEB128,
		},
		{
			name: "OBU size larger than payload",
			payload: []byte{
				0,                // Aggregation header
				0x05,             // Length field
				0x22,             // OBU header (has_size_field = 1)
				0x04,             // LEB128 size
				0x03, 0x01, 0x02, // OBU data
			},
			err: errShortPacket,
		},
		{
			name: "OBU size smaller than length field",
			payload: []byte{
				0,                // Aggregation header
				0x05,             // Length field
				0x22,             // OBU header (has_size_field = 1)
				0x02,             // LEB128 size
				0x03, 0x01, 0x02, // OBU data
			},
			err: errShortPacket,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := AV1Depacketizer{}
			_, err := d.Unmarshal(tt.payload)
			assert.ErrorIs(t, err, tt.err)
		})
	}
}

func TestAV1Depacketizer_dropBuffer(t *testing.T) {
	depacketizer := &AV1Depacketizer{}
	empty, err := depacketizer.Unmarshal([]byte{0x41, 0x02, 0x00, 0x01})
	assert.NoError(t, err)
	assert.Len(t, empty, 0)

	payload := []byte{0x08, 0x02, 0x03}
	obuData, expectedOBU := createAV1OBU(4, payload)

	packet := make([]byte, 0)

	// N=true, should clear buffer
	packet = append(packet, []byte{0b00001000}...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obuData)))...)
	packet = append(packet, obuData...)

	obu, err := depacketizer.Unmarshal(packet)
	assert.NoError(t, err)
	assert.Equal(t, expectedOBU, obu)
}

func TestAV1Depacketizer_singleOBUWithW(t *testing.T) {
	payload := []byte{0x01, 0x02, 0x03}
	obuData, expectedOBU := createAV1OBU(4, payload)

	packet := append([]byte{0b00010000}, obuData...)

	d := AV1Depacketizer{}
	obu, err := d.Unmarshal(packet)
	assert.NoError(t, err)
	assert.Equal(t, expectedOBU, obu)
}

func TestDepacketizer_multipleFullOBUs(t *testing.T) {
	obu1, expectedOBU1 := createAV1OBU(4, []byte{0x01, 0x02, 0x03})
	obu2, expectedOBU2 := createAV1OBU(4, []byte{0x04, 0x05, 0x06})
	obu3, expectedOBU3 := createAV1OBU(4, []byte{0x07, 0x08, 0x09})
	expected := append(append(expectedOBU1, expectedOBU2...), expectedOBU3...)

	packet := make([]byte, 0)

	packet = append(packet, []byte{0b00000000}...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obu1)))...)
	packet = append(packet, obu1...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obu2)))...)
	packet = append(packet, obu2...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obu3)))...)
	packet = append(packet, obu3...)

	d := AV1Depacketizer{}
	obus, err := d.Unmarshal(packet)
	assert.NoError(t, err)
	assert.Equal(t, expected, obus)
}

func TestAV1Depacketizer_multipleFullOBUsWithW(t *testing.T) {
	obu1, expectedOBU1 := createAV1OBU(4, []byte{0x01, 0x02, 0x03})
	obu2, expectedOBU2 := createAV1OBU(4, []byte{0x04, 0x05, 0x06})
	obu3, expectedOBU3 := createAV1OBU(4, []byte{0x07, 0x08, 0x09})
	expected := append(append(expectedOBU1, expectedOBU2...), expectedOBU3...)

	packet := make([]byte, 0)

	packet = append(packet, []byte{0b00110000}...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obu1)))...)
	packet = append(packet, obu1...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obu2)))...)
	packet = append(packet, obu2...)
	// Last MUST NOT be preceded by a length field if W is not 0
	packet = append(packet, obu3...)

	depacketizer := AV1Depacketizer{}
	obus, err := depacketizer.Unmarshal(packet)
	assert.NoError(t, err)
	assert.Equal(t, expected, obus)
}

func TestDepacketizer_fragmentedOBUS(t *testing.T) {
	// Not up to spec AV1 stream but it should be depacketized.
	// [ SH MD ] Frag(MD(0,0)) [ FH(0,0) TG(0,0) ] Frag(MD(0,1)) [ FH(0,1) ] [ TG(0,1) ]
	obu1, expectedOBU1 := createAV1OBU(1, []byte{0x01, 0x02, 0x03})
	obu2, expectedOBU2 := createAV1OBU(7, []byte{0x04, 0x05, 0x06})
	obu3, expectedOBU3 := createAV1OBU(7, []byte{0x07, 0x08, 0x09})
	obu3f1 := obu3[:2]
	obu3f2 := obu3[2:]
	obu4, expectedOBU4 := createAV1OBU(3, []byte{0x0A, 0x0B, 0x0C})
	obu5, expectedOBU5 := createAV1OBU(6, []byte{0x0D, 0x0E, 0x0F})
	obu6, expectedOBU6 := createAV1OBU(7, []byte{0x10, 0x11, 0x12})
	obu6f1 := obu6[:2]
	obu6f2 := obu6[2:]
	obu7, expectedOBU7 := createAV1OBU(3, []byte{0x13, 0x14, 0x15})
	obu8, expectedOBU8 := createAV1OBU(6, []byte{0x16, 0x17, 0x18})

	depacketizer := AV1Depacketizer{}

	packet := make([]byte, 0)
	packet = append(packet, []byte{0b01000000}...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obu1)))...)
	packet = append(packet, obu1...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obu2)))...)
	packet = append(packet, obu2...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obu3f1)))...)
	packet = append(packet, obu3f1...)

	obus, err := depacketizer.Unmarshal(packet)
	assert.NoError(t, err)

	expected := make([]byte, 0)
	expected = append(expected, expectedOBU1...)
	expected = append(expected, expectedOBU2...)
	assert.Equal(t, expected, obus)

	packet = make([]byte, 0)
	packet = append(packet, []byte{0b11000000}...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obu3f2)))...)
	packet = append(packet, obu3f2...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obu4)))...)
	packet = append(packet, obu4...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obu5)))...)
	packet = append(packet, obu5...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obu6f1)))...)
	packet = append(packet, obu6f1...)

	obus, err = depacketizer.Unmarshal(packet)
	assert.NoError(t, err)

	expected = append(append(expectedOBU3, expectedOBU4...), expectedOBU5...)
	assert.Equal(t, expected, obus)

	packet = make([]byte, 0)
	packet = append(packet, []byte{0b10100000}...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obu6f2)))...)
	packet = append(packet, obu6f2...)
	// W is defined as 2, so the last OBU MUST NOT have a length field
	packet = append(packet, obu7...)

	obus, err = depacketizer.Unmarshal(packet)
	assert.NoError(t, err)

	expected = make([]byte, 0)
	expected = append(expected, expectedOBU6...)
	expected = append(expected, expectedOBU7...)
	assert.Equal(t, expected, obus)

	packet = make([]byte, 0)
	packet = append(packet, []byte{0b00000000}...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obu8)))...)
	packet = append(packet, obu8...)

	obus, err = depacketizer.Unmarshal(packet)
	assert.NoError(t, err)
	assert.Equal(t, expectedOBU8, obus)
}

func TestAV1Depacketizer_dropLostFragment(t *testing.T) {
	depacketizer := AV1Depacketizer{}

	obus, err := depacketizer.Unmarshal(
		append(
			append([]byte{0b01000000}, obu.WriteToLeb128(3)...),
			[]byte{0x01, 0x02, 0x03}...,
		),
	)
	assert.NoError(t, err)
	assert.Len(t, obus, 0, "Expected empty OBU for fragmented OBU")

	newOBU, expected := createAV1OBU(obu.OBUTileGroup, []byte{0x04, 0x05, 0x06})
	obus, err = depacketizer.Unmarshal(
		append(
			append([]byte{0b00000000}, obu.WriteToLeb128(uint(len(newOBU)))...),
			newOBU...,
		),
	)
	assert.NoError(t, err)
	assert.Equal(t, expected, obus)
}

func TestAV1Depacketizer_dropIfLostFragment(t *testing.T) {
	depacketizer := AV1Depacketizer{}

	obus, err := depacketizer.Unmarshal(
		append(
			append([]byte{0b10000000}, obu.WriteToLeb128(3)...),
			[]byte{0x01, 0x02, 0x03}...,
		),
	)
	assert.NoError(t, err)
	assert.Len(t, obus, 0, "Expected empty OBU for fragmented OBU")

	newOBU, expected := createAV1OBU(obu.OBUTileGroup, []byte{0x04, 0x05, 0x06})
	obus, err = depacketizer.Unmarshal(
		append(
			append([]byte{0b00000000}, obu.WriteToLeb128(uint(len(newOBU)))...),
			newOBU...,
		),
	)
	assert.NoError(t, err)
	assert.Equal(t, expected, obus)

	packet := make([]byte, 0)
	packet = append(packet, []byte{0b10000000}...)
	packet = append(packet, obu.WriteToLeb128(3)...)
	packet = append(packet, []byte{0x01, 0x02, 0x03}...)
	packet = append(packet, obu.WriteToLeb128(uint(len(newOBU)))...)
	packet = append(packet, newOBU...)

	obus, err = depacketizer.Unmarshal(packet)
	assert.NoError(t, err)
	assert.Equal(t, expected, obus)
}

func TestAV1Depacketizer_IsPartitionTail(t *testing.T) {
	depacketizer := &AV1Depacketizer{
		buffer: []byte{1, 2},
	}

	assert.False(t, depacketizer.IsPartitionTail(false, []byte{1, 2}))
	assert.Equal(t, depacketizer.buffer, []byte{1, 2})
	assert.True(t, depacketizer.IsPartitionTail(true, []byte{1, 2}))
}

func TestAV1Depacketizer_IsPartitionHead(t *testing.T) {
	depacketizer := &AV1Depacketizer{}

	assert.False(t, depacketizer.IsPartitionHead(nil))
	assert.False(t, depacketizer.IsPartitionHead([]byte{}))
	assert.False(t, depacketizer.IsPartitionHead([]byte{0b11000000}))
	assert.True(t, depacketizer.IsPartitionHead([]byte{0b00000000}))
}

func TestAV1Depacketizer_ignoreBadOBUs(t *testing.T) {
	shouldIgnore := []obu.Type{
		obu.OBUTemporalDelimiter,
		obu.OBUTileList,
	}

	for _, obuType := range shouldIgnore {
		payload := []byte{0x01, 0x02, 0x03}
		obuData, _ := createAV1OBU(obuType, payload)

		packet := make([]byte, 0)
		packet = append(packet, []byte{0b00000000}...)
		packet = append(packet, obu.WriteToLeb128(uint(len(obuData)))...)
		packet = append(packet, obuData...)

		depacketizer := AV1Depacketizer{}
		obu, err := depacketizer.Unmarshal(packet)
		assert.NoError(t, err)
		assert.Len(t, obu, 0, "Expected empty payload for OBU type %d", obuType)
	}
}

func TestAV1Depacketizer_fragmentedOverMultiple(t *testing.T) {
	fullOBU, expected := createAV1OBU(
		obu.OBUTileGroup,
		[]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
	)
	obuf1 := fullOBU[:2]
	obuf2 := fullOBU[2:5]
	obuf3 := fullOBU[5:7]
	obuf4 := fullOBU[7:]

	depacketizer := AV1Depacketizer{}

	packet := make([]byte, 0)
	packet = append(packet, []byte{0b01000000}...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obuf1)))...)
	packet = append(packet, obuf1...)

	obus, err := depacketizer.Unmarshal(packet)
	assert.NoError(t, err)
	assert.Len(t, obus, 0, "Expected empty OBU for fragmented OBU")

	packet = make([]byte, 0)
	packet = append(packet, []byte{0b11000000}...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obuf2)))...)
	packet = append(packet, obuf2...)

	obus, err = depacketizer.Unmarshal(packet)
	assert.NoError(t, err)
	assert.Len(t, obus, 0, "Expected empty OBU for fragmented OBU")

	packet = make([]byte, 0)
	packet = append(packet, []byte{0b11000000}...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obuf3)))...)
	packet = append(packet, obuf3...)

	obus, err = depacketizer.Unmarshal(packet)
	assert.NoError(t, err)
	assert.Len(t, obus, 0, "Expected empty OBU for fragmented OBU")

	packet = make([]byte, 0)
	packet = append(packet, []byte{0b10000000}...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obuf4)))...)
	packet = append(packet, obuf4...)

	obus, err = depacketizer.Unmarshal(packet)
	assert.NoError(t, err)
	assert.Equal(t, expected, obus)
}

func TestAV1Depacketizer_shortOBUHeader(t *testing.T) {
	d := AV1Depacketizer{}

	payload, err := d.Unmarshal([]byte{0x00, 0x01, 0x04})
	assert.Error(t, err)
	assert.Len(t, payload, 0, "Expected empty payload for short OBU header")
}

func TestAV1Depacketizer_aggregationHeader(t *testing.T) {
	depacketizer := AV1Depacketizer{}
	tests := []struct {
		name    string
		input   []byte
		payload []byte
		Z, Y, N bool
	}{
		{
			name: "Z=0, Y=0, N=0",
			// aggregation header = 0, length field = 1, obu header = 0x30
			input: []byte{0x00, 0x01, 0x30},
			// obu header = 0x32, obu size = 0
			payload: []byte{0x32, 0x00},
		},
		{
			name: "Z=1, Y=0, N=0",
			// aggregation header = z = 1, length field = 1, obu header = 0x20
			input: []byte{0x80, 0x01, 0x20},
			// packet is fragmented, with missing previous packet, so the result is empty
			payload: []byte{},
			Z:       true,
		},
		{
			name: "Z=0, Y=1, N=0",
			// aggregation header = Y = 1, length field = 1, obu header = 0x20
			input: []byte{0x40, 0x01, 0x04},
			// Packet is fragmented with the next packet.
			payload: []byte{},
			Y:       true,
		},
		{
			name: "Z=0, Y=0, N=1",
			// aggregation header = N = 1, length field = 1, obu header = 0x30
			input: []byte{0x08, 0x01, 0x30},
			// obu header = 0x32, obu size = 0
			payload: []byte{0x32, 0x00},
			N:       true,
		},
		{
			name: "Z=1, Y=1, N=1",
			// aggregation header = N, Y, Z = 1, length field = 1, obu header = 0x30
			input: []byte{0xC8, 0x01, 0x30},
			// Packet is fragmented no payload.
			payload: []byte{},
			Z:       true,
			Y:       true,
			N:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := depacketizer.Unmarshal(tt.input)
			assert.NoError(t, err)

			assert.Equal(t, tt.payload, payload)
			assert.Equal(t, tt.Z, depacketizer.Z)
			assert.Equal(t, tt.Y, depacketizer.Y)
			assert.Equal(t, tt.N, depacketizer.N)
		})
	}
}

func FuzzAV1DepacketizerUnmarshal(f *testing.F) {
	f.Add([]byte{0x10, 0x01, 0x00})
	f.Add([]byte{0x20, 0x01, 0x00, 0x01, 0x00})
	f.Add([]byte{0x00, 0x01, 0x00})
	f.Add([]byte{0x80, 0x01, 0x00})
	f.Add([]byte{0x40, 0x01, 0x00})
	f.Add([]byte{0x08, 0x01, 0x00})
	f.Add([]byte{0xC0, 0x01, 0x00})
	f.Add([]byte{0x30, 0x01, 0x00, 0x01, 0x00, 0x00})

	obuData, _ := createAV1OBU(obu.OBUFrameHeader, []byte{0x01, 0x02, 0x03})
	packet := append([]byte{0x00}, obu.WriteToLeb128(uint(len(obuData)))...)
	packet = append(packet, obuData...)
	f.Add(packet)

	obuData2, _ := createAV1OBU(obu.OBUFrame, []byte{0x04, 0x05})
	packet2 := append([]byte{0x10}, obuData2...)
	f.Add(packet2)

	// just check for crashes :)
	f.Fuzz(func(t *testing.T, data []byte) {
		depacketizer := &AV1Depacketizer{}
		_, err := depacketizer.Unmarshal(data)
		_ = err
	})
}
