package codecs

import (
	"encoding/binary"
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createHeader(pType, layerId, TID uint8, F, Z bool) H266NALUHeader {
	var fVal, zVal uint16
	if F {
		fVal = 1 << 15
	}
	if Z {
		zVal = 1 << 14
	}
	return H266NALUHeader(uint16(TID) | (uint16(pType) << 3) | (uint16(layerId) << 8) | fVal | zVal)
}

func TestH266_AggregationRoundtrip(t *testing.T) {
	simplePacket := H266SingleNALUnitPacket{
		newH266NALUHeader(0b0000000, 0b00001000),
		nil,
		[]byte{0x00, 0x01, 0x02, 0x03},
	}
	diffPacket := H266SingleNALUnitPacket{
		newH266NALUHeader(0b1000000, 0b00010000),
		nil,
		[]byte{0x03, 0x02, 0x01, 0x00, 0x12},
	}
	testAggregation := func(expected []H266SingleNALUnitPacket) {
		created, err := newH266AggregationPacket(expected)
		assert.Nil(t, err)
		packet := created.packetize(make([]byte, 0))
		parsed, err := parseH266Packet(packet, false)
		assert.Nil(t, err)
		aggr := parsed.(*H266AggregationPacket)
		split, err := splitH266AggregationPacket(*aggr)
		assert.Nil(t, err)

		assert.True(t, slices.EqualFunc(split, expected, func(a, b H266SingleNALUnitPacket) bool {
			return slices.Equal(a.payload, b.payload) && a.payloadHeader == b.payloadHeader
		}))

	}
	testAggregation([]H266SingleNALUnitPacket{simplePacket, simplePacket, simplePacket})
	testAggregation([]H266SingleNALUnitPacket{diffPacket, simplePacket, simplePacket})
	testAggregation([]H266SingleNALUnitPacket{diffPacket, diffPacket, simplePacket})
}

func TestH266Packetizer_Single(t *testing.T) {
	initSequence := []byte{0x00, 0x00, 0x00, 0x01, 0x00}

	packetizer := H266Packetizer{}

	// type 1, 8 payload length NALU
	basicPacket := make([]byte, len(initSequence))
	copy(basicPacket, initSequence)
	header := createHeader(1, 0, 0, false, false)
	basicPacket = binary.BigEndian.AppendUint16(basicPacket, uint16(header))
	payload := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	basicPacket = append(basicPacket, payload...)

	packets := packetizer.Payload(100, basicPacket)
	assert.Equal(t, 1, len(packets), "Expected only 1 NALU to be generated")
	assert.Equal(t, uint16(header), binary.BigEndian.Uint16(packets[0][0:2]), "Expected headers to match")
	assert.Equal(t, payload, packets[0][2:], "Expected payloads to match")
}

func TestH266Packetizer_Aggregated(t *testing.T) {
	initSequence := []byte{0x00, 0x00, 0x00, 0x01, 0x00}

	packetizer := H266Packetizer{}
	// type 0, 8 payload length
	basicPacket := make([]byte, len(initSequence))
	copy(basicPacket, initSequence)
	header := createHeader(1, 0, 0, false, false)
	basicPacket = binary.BigEndian.AppendUint16(basicPacket, uint16(header))
	payload := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	basicPacket = append(basicPacket, payload...)

	two := make([]byte, len(basicPacket)*2)
	copy(two, basicPacket)
	copy(two[len(basicPacket):], basicPacket)

	packets := packetizer.Payload(100, two)

	assert.Equal(t, 1, len(packets), "Expected only 1 NALU to be generated")
	aggregated := packets[0]
	parsedHeader := H266NALUHeader(binary.BigEndian.Uint16(aggregated[0:2]))

	assert.Equal(t, h266NaluAggregationPacketType, int(parsedHeader.Type()), "NALU header should be type 28")

	assert.Equal(t, h266NaluHeaderSize+len(payload), int(binary.BigEndian.Uint16(aggregated[2:4])), "Expected length to match")
	assert.Equal(t, uint16(header), binary.BigEndian.Uint16(aggregated[4:6]), "Expected headers to match")
	assert.Equal(t, payload, aggregated[6:14], "Expected payloads to match")

	assert.Equal(t, h266NaluHeaderSize+len(payload), int(binary.BigEndian.Uint16(aggregated[14:16])), "Expected length to match")
	assert.Equal(t, uint16(header), binary.BigEndian.Uint16(aggregated[16:18]), "Expected headers to match")
	assert.Equal(t, payload, aggregated[18:], "Expected payloads to match")
}

func TestH266Packetizer_Fragmented(t *testing.T) {
	initSequence := []byte{0x00, 0x00, 0x00, 0x01, 0x00}

	packetizer := H266Packetizer{}
	// type 0, 50 payload length
	bigPackets := make([]byte, len(initSequence))
	copy(bigPackets, initSequence)
	header := createHeader(1, 0, 0, false, false)

	bigPackets = binary.BigEndian.AppendUint16(bigPackets, uint16(header))

	payload := make([]byte, 0)
	for i := 0; i < 50; i++ {
		payload = append(payload, uint8(i))
	}
	bigPackets = append(bigPackets, payload...)

	packets := packetizer.Payload(50, bigPackets)

	assert.Equal(t, 2, len(packets), "Expected 2 NALUs to be generated")
	parsedHeader := H266NALUHeader(binary.BigEndian.Uint16(packets[0][0:2]))

	assert.Equal(t, h266NaluFragmentationUnitType, int(parsedHeader.Type()), "NALU header should be type 28")
	assert.True(t, H265FragmentationUnitHeader(packets[0][2]).S(), "First FU header should be S")
	assert.True(t, H265FragmentationUnitHeader(packets[1][2]).E(), "Second FU header should be E")
}

func TestH266_Depacketizer(t *testing.T) {
	simplePacket := H266SingleNALUnitPacket{
		newH266NALUHeader(0b0000000, 0b00001000),
		nil,
		[]byte{0x00, 0x01, 0x02, 0x03},
	}
	diffPacket := H266SingleNALUnitPacket{
		newH266NALUHeader(0b1000000, 0b00010000),
		nil,
		[]byte{0x03, 0x02, 0x01, 0x00, 0x12},
	}
	testThang := func(expected []H266SingleNALUnitPacket) {
		depacketizer := H266Depacketizer{}
		annexB, err := depacketizer.Unmarshal(
			[]byte{
				0x00, 0b11100000, 0x00, 0x06, 0x00, 0x08, 0x01, 0x02, 0x03, 0x00, 0x00, 0x06, 0x00, 0x08, 0x01, 0x02, 0x03, 0x00,
			})
		assert.Nil(t, err)
		fmt.Println(annexB)
		stream := make([]byte, 0)
		for i := 0; i < 3; i++ {
			stream = append(stream, naluStartCode...)
			stream = append(stream, 0x00)
			stream = simplePacket.packetize(stream)
		}

		// assert.Equal(t, len(packets), 1)
	}
	testThang([]H266SingleNALUnitPacket{simplePacket, simplePacket, simplePacket})
	testThang([]H266SingleNALUnitPacket{diffPacket, simplePacket, simplePacket})
}
