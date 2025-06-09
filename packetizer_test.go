// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package rtp

import (
	"fmt"
	"testing"
	"time"

	"github.com/pion/rtp/codecs"
	"github.com/stretchr/testify/assert"
)

func TestPacketizer(t *testing.T) {
	multiplepayload := make([]byte, 128)
	// use the G722 payloader here, because it's very simple and all 0s is valid G722 data.
	packetizer := NewPacketizer(100, 98, 0x1234ABCD, &codecs.G722Payloader{}, NewRandomSequencer(), 90000)
	packets := packetizer.Packetize(multiplepayload, 2000)

	expectedLen := 2
	if len(packets) != expectedLen {
		packetlengths := ""
		for i := 0; i < len(packets); i++ {
			packetlengths += fmt.Sprintf("Packet %d length %d\n", i, len(packets[i].Payload))
		}
		assert.Failf(
			t, "Packetize failed", "Generated %d packets instead of %d\n%s",
			len(packets), expectedLen, packetlengths,
		)
	}
}

func TestPacketizer_AbsSendTime(t *testing.T) {
	// use the G722 payloader here, because it's very simple and all 0s is valid G722 data.
	pktizer := NewPacketizer(100, 98, 0x1234ABCD, &codecs.G722Payloader{}, NewFixedSequencer(1234), 90000)
	p, ok := pktizer.(*packetizer)
	assert.True(t, ok, "Failed to cast to *packetizer")

	p.Timestamp = 45678
	p.timegen = func() time.Time {
		return time.Date(1985, time.June, 23, 4, 0, 0, 0, time.FixedZone("UTC-5", -5*60*60))
		// (0xa0c65b1000000000>>14) & 0xFFFFFF  = 0x400000
	}
	pktizer.EnableAbsSendTime(1)

	payload := []byte{0x11, 0x12, 0x13, 0x14}
	packets := pktizer.Packetize(payload, 2000)

	expected := &Packet{
		Header: Header{
			Version:          2,
			Padding:          false,
			Extension:        true,
			Marker:           true,
			PayloadType:      98,
			SequenceNumber:   1234,
			Timestamp:        45678,
			SSRC:             0x1234ABCD,
			CSRC:             []uint32{},
			ExtensionProfile: 0xBEDE,
			Extensions: []Extension{
				{
					id:      1,
					payload: []byte{0x40, 0, 0},
				},
			},
		},
		Payload: []byte{0x11, 0x12, 0x13, 0x14},
	}

	assert.Lenf(t, packets, 1, "Generated %d packets instead of 1", len(packets))
	assert.Equal(t, expected, packets[0], "Packetize failed")
}

func TestPacketizer_Roundtrip(t *testing.T) {
	multiplepayload := make([]byte, 128)
	packetizer := NewPacketizer(100, 98, 0x1234ABCD, &codecs.G722Payloader{}, NewRandomSequencer(), 90000)
	packets := packetizer.Packetize(multiplepayload, 1000)

	rawPkts := make([][]byte, 0, 1400)
	for _, pkt := range packets {
		raw, err := pkt.Marshal()
		assert.NoError(t, err)

		rawPkts = append(rawPkts, raw)
	}

	for ndx, raw := range rawPkts {
		expectedPkt := packets[ndx]
		pkt := &Packet{}

		assert.NoError(t, pkt.Unmarshal(raw))
		assert.Equal(t, len(raw), pkt.MarshalSize())
		assert.Equal(t, expectedPkt.MarshalSize(), pkt.MarshalSize())
		assert.Equal(t, expectedPkt.Version, pkt.Version)
		assert.Equal(t, expectedPkt.Padding, pkt.Padding)
		assert.Equal(t, expectedPkt.Extension, pkt.Extension)
		assert.Equal(t, expectedPkt.Marker, pkt.Marker)
		assert.Equal(t, expectedPkt.PayloadType, pkt.PayloadType)
		assert.Equal(t, expectedPkt.SequenceNumber, pkt.SequenceNumber)
		assert.Equal(t, expectedPkt.Timestamp, pkt.Timestamp)
		assert.Equal(t, expectedPkt.SSRC, pkt.SSRC)
		assert.Equal(t, expectedPkt.CSRC, pkt.CSRC)
		assert.Equal(t, expectedPkt.ExtensionProfile, pkt.ExtensionProfile)
		assert.Equal(t, expectedPkt.Extensions, pkt.Extensions)
		assert.Equal(t, expectedPkt.Payload, pkt.Payload)

		pkt.PaddingSize = 0
		pkt.Header.PaddingSize = 0
		assert.Equal(t, expectedPkt, pkt)
	}
}

func TestPacketizer_GeneratePadding(t *testing.T) {
	pktizer := NewPacketizer(100, 98, 0x1234ABCD, &codecs.G722Payloader{}, NewFixedSequencer(1234), 90000)

	packets := pktizer.GeneratePadding(5)

	assert.Len(t, packets, 5, "Should generate exactly 5 padding packets")
	for i, pkt := range packets {
		assert.Equal(t, true, pkt.Header.Padding, "Packet %d should have Padding set to true", i)
		assert.Equal(t, byte(255), pkt.Header.PaddingSize, "Packet %d should have PaddingSize set to 255", i)
		assert.Equal(t, byte(0), pkt.PaddingSize, "Packet %d should have PaddingSize set to 0", i)
		assert.Nil(t, pkt.Payload, "Packet %d should have no Payload", i)
	}
}

func TestNewPacketizerWithOptions_DefaultValues(t *testing.T) {
	pack := NewPacketizerWithOptions(100, &codecs.G722Payloader{}, NewRandomSequencer(), 90000)
	p, ok := pack.(*packetizer)
	assert.True(t, ok, "Failed to cast to *packetizer")

	assert.Equal(t, uint16(100), p.MTU)
	assert.Equal(t, uint8(0), p.PayloadType)
	assert.Equal(t, uint32(0), p.SSRC)
	assert.NotZero(t, p.Timestamp)
	assert.Equal(t, uint32(90000), p.ClockRate)
}

func TestNewPacketizerWithOptions_WithOptions(t *testing.T) {
	pack := NewPacketizerWithOptions(
		100,
		&codecs.G722Payloader{},
		NewRandomSequencer(),
		90000,
		WithSSRC(0x1234ABCD),
		WithPayloadType(98),
		WithTimestamp(45678),
	)
	p, ok := pack.(*packetizer)
	assert.True(t, ok, "Failed to cast to *packetizer")

	assert.Equal(t, uint16(100), p.MTU)
	assert.Equal(t, uint8(98), p.PayloadType)
	assert.Equal(t, uint32(0x1234ABCD), p.SSRC)
	assert.Equal(t, uint32(45678), p.Timestamp)
	assert.Equal(t, uint32(90000), p.ClockRate)

	payload := []byte{0x11, 0x12, 0x13, 0x14}
	packets := pack.Packetize(payload, 2000)

	assert.Len(t, packets, 1, "Should generate exactly one packet")
	assert.Equal(t, uint8(98), packets[0].PayloadType)
	assert.Equal(t, uint32(0x1234ABCD), packets[0].SSRC)
	assert.Equal(t, uint32(45678), packets[0].Timestamp)
}

func TestNewPacketizerWithOptions_PartialOptions(t *testing.T) {
	pack := NewPacketizerWithOptions(
		100,
		&codecs.G722Payloader{},
		NewRandomSequencer(),
		90000,
		WithPayloadType(98),
	)
	p, ok := pack.(*packetizer)
	assert.True(t, ok, "Failed to cast to *packetizer")

	assert.Equal(t, uint16(100), p.MTU)
	assert.Equal(t, uint8(98), p.PayloadType)
	assert.Equal(t, uint32(0), p.SSRC)
	assert.NotZero(t, p.Timestamp)
	assert.Equal(t, uint32(90000), p.ClockRate)
}
