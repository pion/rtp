// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package codecs

import (
	"bytes"
	"errors"
	"testing"

	"github.com/pion/rtp/codecs/av1/obu"
)

// Make an AV1 OBU for testing.
func createAV1OBU(obuType uint8, payload []byte) []byte {
	buf := make([]byte, 1+len(payload))

	buf[0] = (obuType << obuFrameTypeBitshift) | obuFameTypeSequenceHeader
	copy(buf[1:], payload)

	return buf
}

func TestDepacketizer_invalidPackets(t *testing.T) {
	depacketizer := AV1Depacketizer{}
	_, err := depacketizer.Unmarshal([]byte{})
	if !errors.Is(err, errShortPacket) {
		t.Fatalf("Unexpected error: %v", err)
	}

	_, err = depacketizer.Unmarshal([]byte{0x00})
	if !errors.Is(err, errShortPacket) {
		t.Fatalf("Unexpected error: %v", err)
	}

	_, err = depacketizer.Unmarshal([]byte{0b11000000, 0xFF})
	if !errors.Is(err, obu.ErrFailedToReadLEB128) {
		t.Fatalf("Unexpected error: %v", err)
	}

	_, err = depacketizer.Unmarshal(append([]byte{0b00000000}, obu.WriteToLeb128(0x99)...))
	if !errors.Is(err, errShortPacket) {
		t.Fatalf("Unexpected error: %v", err)
	}

	_, err = depacketizer.Unmarshal(append([]byte{0b00000000}, obu.WriteToLeb128(0)...))
	if !errors.Is(err, errShortPacket) {
		t.Fatalf("Unexpected error: %v", err)
	}

	_, err = depacketizer.Unmarshal(
		append(
			[]byte{0b00110000},
			append(
				obu.WriteToLeb128(1),
				[]byte{0x01}...,
			)...,
		),
	)
	if !errors.Is(err, errShortPacket) {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestDepacketizer_singleOBU(t *testing.T) {
	payload := []byte{0x01, 0x02, 0x03}
	obuData := createAV1OBU(4, payload)

	packet := make([]byte, 0)

	packet = append(packet, []byte{0b00000000}...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obuData)))...)
	packet = append(packet, obuData...)

	d := AV1Depacketizer{}
	obu, err := d.Unmarshal(packet)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !bytes.Equal(obu, obuData) {
		t.Fatalf("OBU data mismatch, expected %v, got %v", obuData, obu)
	}
}

func TestDepacketizer_singleOBUWithW(t *testing.T) {
	payload := []byte{0x01, 0x02, 0x03}
	obuData := createAV1OBU(4, payload)

	packet := append([]byte{0b00010000}, obuData...)

	d := AV1Depacketizer{}
	obu, err := d.Unmarshal(packet)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !bytes.Equal(obu, obuData) {
		t.Fatalf("OBU data mismatch, expected %v, got %v", obuData, obu)
	}
}

func TestDepacketizer_multipleFullOBUs(t *testing.T) {
	obu1 := createAV1OBU(4, []byte{0x01, 0x02, 0x03})
	obu2 := createAV1OBU(4, []byte{0x04, 0x05, 0x06})
	obu3 := createAV1OBU(4, []byte{0x07, 0x08, 0x09})

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
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := append(append(obu1, obu2...), obu3...)
	if !bytes.Equal(obus, expected) {
		t.Fatalf("OBU data mismatch, expected %v, got %v", expected, obus)
	}
}

func TestDepacketizer_multipleFullOBUsWithW(t *testing.T) {
	obu1 := createAV1OBU(4, []byte{0x01, 0x02, 0x03})
	obu2 := createAV1OBU(4, []byte{0x04, 0x05, 0x06})
	obu3 := createAV1OBU(4, []byte{0x07, 0x08, 0x09})

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
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := append(append(obu1, obu2...), obu3...)
	if !bytes.Equal(obus, expected) {
		t.Fatalf("OBU data mismatch, expected %v, got %v", expected, obus)
	}
}

func TestDepacketizer_fragmentedOBUS(t *testing.T) {
	// Not up to spec AV1 stream but it should be depacketized.
	// [ SH MD ] Frag(MD(0,0)) [ FH(0,0) TG(0,0) ] Frag(MD(0,1)) [ FH(0,1) ] [ TG(0,1) ]
	obu1 := createAV1OBU(1, []byte{0x01, 0x02, 0x03})
	obu2 := createAV1OBU(7, []byte{0x04, 0x05, 0x06})
	obu3 := createAV1OBU(7, []byte{0x07, 0x08, 0x09})
	obu3f1 := obu3[:2]
	obu3f2 := obu3[2:]
	obu4 := createAV1OBU(3, []byte{0x0A, 0x0B, 0x0C})
	obu5 := createAV1OBU(6, []byte{0x0D, 0x0E, 0x0F})
	obu7 := createAV1OBU(7, []byte{0x10, 0x11, 0x12})
	obu7f1 := obu7[:2]
	obu7f2 := obu7[2:]
	obu8 := createAV1OBU(3, []byte{0x13, 0x14, 0x15})
	obu9 := createAV1OBU(6, []byte{0x16, 0x17, 0x18})

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
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := make([]byte, 0)
	expected = append(expected, obu1...)
	expected = append(expected, obu2...)

	if !bytes.Equal(obus, expected) {
		t.Fatalf("OBU data mismatch, expected %v, got %v", expected, obus)
	}

	packet = make([]byte, 0)
	packet = append(packet, []byte{0b11000000}...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obu3f2)))...)
	packet = append(packet, obu3f2...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obu4)))...)
	packet = append(packet, obu4...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obu5)))...)
	packet = append(packet, obu5...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obu7f1)))...)
	packet = append(packet, obu7f1...)

	obus, err = depacketizer.Unmarshal(packet)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected = append(append(obu3, obu4...), obu5...)
	if !bytes.Equal(obus, expected) {
		t.Fatalf("OBU data mismatch, expected %v, got %v", expected, obus)
	}

	packet = make([]byte, 0)
	packet = append(packet, []byte{0b10100000}...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obu7f2)))...)
	packet = append(packet, obu7f2...)
	// W is defined as 2, so the last OBU MUST NOT have a length field
	packet = append(packet, obu8...)

	obus, err = depacketizer.Unmarshal(packet)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected = make([]byte, 0)
	expected = append(expected, obu7...)
	expected = append(expected, obu8...)

	if !bytes.Equal(obus, expected) {
		t.Fatalf("OBU data mismatch, expected %v, got %v", expected, obus)
	}

	packet = make([]byte, 0)
	packet = append(packet, []byte{0b00000000}...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obu9)))...)
	packet = append(packet, obu9...)

	obus, err = depacketizer.Unmarshal(packet)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected = obu9
	if !bytes.Equal(obus, expected) {
		t.Fatalf("OBU data mismatch, expected %v, got %v", expected, obus)
	}
}

func TestDepacketizer_dropLostFragment(t *testing.T) {
	depacketizer := AV1Depacketizer{}

	obus, err := depacketizer.Unmarshal(
		append(
			append([]byte{0b01000000}, obu.WriteToLeb128(3)...),
			[]byte{0x01, 0x02, 0x03}...,
		),
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(obus) != 0 {
		t.Fatalf("Expected empty OBU for fragmented OBU")
	}

	newOBU := createAV1OBU(4, []byte{0x04, 0x05, 0x06})
	obus, err = depacketizer.Unmarshal(
		append(
			append([]byte{0b00000000}, obu.WriteToLeb128(uint(len(newOBU)))...),
			newOBU...,
		),
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !bytes.Equal(obus, newOBU) {
		t.Fatalf("Expected OBU data to be %v, got %v", newOBU, obus)
	}
}

func TestDepacketizer_dropIfLostFragment(t *testing.T) {
	depacketizer := AV1Depacketizer{}

	obus, err := depacketizer.Unmarshal(
		append(
			append([]byte{0b10000000}, obu.WriteToLeb128(3)...),
			[]byte{0x01, 0x02, 0x03}...,
		),
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(obus) != 0 {
		t.Fatalf("Expected empty OBU for fragmented OBU")
	}

	newOBU := createAV1OBU(4, []byte{0x04, 0x05, 0x06})
	obus, err = depacketizer.Unmarshal(
		append(
			append([]byte{0b00000000}, obu.WriteToLeb128(uint(len(newOBU)))...),
			newOBU...,
		),
	)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !bytes.Equal(obus, newOBU) {
		t.Fatalf("Expected OBU data to be %v, got %v", newOBU, obus)
	}

	packet := make([]byte, 0)
	packet = append(packet, []byte{0b10000000}...)
	packet = append(packet, obu.WriteToLeb128(3)...)
	packet = append(packet, []byte{0x01, 0x02, 0x03}...)
	packet = append(packet, obu.WriteToLeb128(uint(len(newOBU)))...)
	packet = append(packet, newOBU...)

	obus, err = depacketizer.Unmarshal(packet)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !bytes.Equal(obus, newOBU) {
		t.Fatalf("Expected OBU data to be %v, got %v", newOBU, obus)
	}
}

func TestDepacketizer_IsPartitionTail(t *testing.T) {
	depacketizer := &AV1Depacketizer{
		buffer: []byte{1, 2},
	}

	if depacketizer.IsPartitionTail(false, []byte{1, 2}) {
		t.Fatalf("Expected false")
	}

	if !bytes.Equal(depacketizer.buffer, []byte{1, 2}) {
		t.Fatalf("Buffer was modified")
	}

	if !depacketizer.IsPartitionTail(true, []byte{1, 2}) {
		t.Fatalf("Expected true")
	}

	if len(depacketizer.buffer) != 0 {
		t.Fatalf("Buffer was not cleared")
	}
}

func TestDepacketizer_IsPartitionHead(t *testing.T) {
	depacketizer := &AV1Depacketizer{}

	if depacketizer.IsPartitionHead(nil) {
		t.Fatalf("Expected false")
	}

	if depacketizer.IsPartitionHead([]byte{}) {
		t.Fatalf("Expected false")
	}

	if depacketizer.IsPartitionHead([]byte{0b11000000}) {
		t.Fatalf("Expected false")
	}

	if !depacketizer.IsPartitionHead([]byte{0b00000000}) {
		t.Fatalf("Expected true")
	}
}

func TestDepacketizer_ignoreBadOBUs(t *testing.T) {
	shouldIgnore := []uint8{
		2, // OBU_TEMPORAL_DELIMITER
		8, // OBU_TILE_LIST
	}

	for _, obuType := range shouldIgnore {
		payload := []byte{0x01, 0x02, 0x03}
		obuData := createAV1OBU(obuType, payload)

		packet := make([]byte, 0)
		packet = append(packet, []byte{0b00000000}...)
		packet = append(packet, obu.WriteToLeb128(uint(len(obuData)))...)
		packet = append(packet, obuData...)

		depacketizer := AV1Depacketizer{}
		obu, err := depacketizer.Unmarshal(packet)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(obu) != 0 {
			t.Fatalf("Expected empty OBU for OBU type %d", obuType)
		}
	}
}

func TestDepacketizer_fragmentedOverMultiple(t *testing.T) {
	fullOBU := createAV1OBU(4, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08})
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
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(obus) != 0 {
		t.Fatalf("Expected empty OBU for fragmented OBU")
	}

	packet = make([]byte, 0)
	packet = append(packet, []byte{0b11000000}...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obuf2)))...)
	packet = append(packet, obuf2...)

	obus, err = depacketizer.Unmarshal(packet)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(obus) != 0 {
		t.Fatalf("Expected empty OBU for fragmented OBU")
	}

	packet = make([]byte, 0)
	packet = append(packet, []byte{0b11000000}...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obuf3)))...)
	packet = append(packet, obuf3...)

	obus, err = depacketizer.Unmarshal(packet)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(obus) != 0 {
		t.Fatalf("Expected empty OBU for fragmented OBU")
	}

	packet = make([]byte, 0)
	packet = append(packet, []byte{0b10000000}...)
	packet = append(packet, obu.WriteToLeb128(uint(len(obuf4)))...)
	packet = append(packet, obuf4...)

	obus, err = depacketizer.Unmarshal(packet)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !bytes.Equal(obus, fullOBU) {
		t.Fatalf("Expected OBU data to be %v, got %v", fullOBU, obus)
	}
}
