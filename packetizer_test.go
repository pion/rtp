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

func TestPacketizer_Empty_Payload(t *testing.T) {
	pktizer := NewPacketizer(100, 98, 0x1234ABCD, &codecs.G722Payloader{}, NewFixedSequencer(1234), 90000)
	const expectedSamples = uint32(4000)

	prevTimestamp := uint32(0)
	for i := 0; i < 10; i++ {
		payload := []byte{0x11, 0x12, 0x13, 0x14}
		isEmptyPayload := i%2 == 0
		if isEmptyPayload {
			payload = nil
		}

		packets := pktizer.Packetize(payload, 2000)

		if isEmptyPayload {
			assert.Len(t, packets, 0)
		} else {
			assert.Len(t, packets, 1)

			if prevTimestamp != 0 {
				assert.Equal(t, packets[0].Timestamp-prevTimestamp, expectedSamples)
			}
			prevTimestamp = packets[0].Timestamp
		}
	}
}

func FuzzPacketizer_Packetize_G722(f *testing.F) {
	// mixed seeds.
	f.Add(uint16(100), uint8(98), uint32(0x1234ABCD), uint32(960), false, []byte{0})
	f.Add(uint16(100), uint8(98), uint32(0x00000001), uint32(0), false, []byte{})
	f.Add(uint16(12), uint8(0), uint32(0), uint32(1), true, []byte{1, 2, 3, 4})
	f.Add(uint16(1500), uint8(120), uint32(0xCAFEBABE), uint32(480), true, make([]byte, 4096))
	f.Add(uint16(1), uint8(34), uint32(7), uint32(160), false, make([]byte, 32))

	f.Fuzz(func(t *testing.T, mtu uint16, pt uint8, ssrc uint32, samples uint32, enableAST bool, payload []byte) {
		if len(payload) > 1<<16 {
			payload = payload[:1<<16]
		}

		packetizer := NewPacketizerWithOptions(
			mtu,
			&codecs.G722Payloader{},
			NewFixedSequencer(0),
			90000,
			WithPayloadType(pt),
			WithSSRC(ssrc),
			WithTimestamp(0xAABBCCDD),
		)

		if enableAST {
			packetizer.EnableAbsSendTime(1)
		}

		packets := packetizer.Packetize(payload, samples)

		if len(payload) == 0 {
			assert.Nil(t, packets)

			return
		}

		eff := int(mtu - 12)
		if eff == 0 {
			assert.Equal(t, 0, len(packets))

			return
		}

		assert.GreaterOrEqual(t, len(packets), 1)

		for i, packet := range packets {
			assert.Equal(t, uint8(2), packet.Version)
			assert.Equal(t, pt, packet.PayloadType)
			assert.Equal(t, ssrc, packet.SSRC)

			if i == len(packets)-1 {
				assert.True(t, packet.Marker)

				if enableAST {
					assert.True(t, packet.Extension)
					raw, err := packet.Marshal()
					assert.NoError(t, err)

					var back Packet
					assert.NoError(t, back.Unmarshal(raw))
				}
			} else {
				assert.False(t, packet.Marker)
			}

			if mtu >= 12 && !enableAST {
				raw, err := packet.Marshal()
				assert.NoError(t, err)
				assert.LessOrEqual(t, len(raw), int(mtu))
			} else {
				raw, err := packet.Marshal()
				assert.NoError(t, err)

				var back Packet
				assert.NoError(t, back.Unmarshal(raw))
			}
		}
	})
}

func FuzzPacketizer_SkipSamples_And_Timestamps(f *testing.F) {
	// mixed seeds.
	f.Add(uint16(1200), uint32(0), uint32(480), uint32(960), uint16(100), uint16(200), false)
	f.Add(uint16(200), uint32(160), uint32(160), uint32(160), uint16(10), uint16(20), true)
	f.Add(uint16(20), uint32(32000), uint32(0), uint32(1), uint16(0), uint16(1), false)

	f.Fuzz(func(
		t *testing.T,
		mtu uint16,
		skip uint32,
		samples1 uint32,
		samples2 uint32,
		len1 uint16,
		len2 uint16,
		enableAST bool,
	) {
		p1 := make([]byte, int(len1))
		for i := range p1 {
			p1[i] = byte(i)
		}

		p2 := make([]byte, int(len2))
		for i := range p2 {
			p2[i] = byte(255 - i)
		}

		const startTS = uint32(0x10203040)
		packetizer := NewPacketizerWithOptions(
			mtu,
			&codecs.G722Payloader{},
			NewFixedSequencer(1000),
			90000,
			WithPayloadType(111),
			WithSSRC(0xFEEDBEEF),
			WithTimestamp(startTS),
		)

		if enableAST {
			packetizer.EnableAbsSendTime(1)
		}

		packetizer.SkipSamples(skip)

		pkts1 := packetizer.Packetize(p1, samples1)
		if len(p1) == 0 {
			assert.Nil(t, pkts1)
		} else {
			assert.GreaterOrEqual(t, len(pkts1), 1)
			assert.Equal(t, startTS+skip, pkts1[0].Timestamp)
		}

		pkts2 := packetizer.Packetize(p2, samples2)
		if len(p2) == 0 {
			assert.Nil(t, pkts2)
		} else {
			assert.GreaterOrEqual(t, len(pkts2), 1)
			expectedTS2 := startTS + skip + samples1
			assert.Equal(t, expectedTS2, pkts2[0].Timestamp)
		}

		for _, p := range append(pkts1, pkts2...) {
			raw, err := p.Marshal()
			assert.NoError(t, err)
			var back Packet
			assert.NoError(t, back.Unmarshal(raw))
		}
	})
}

func FuzzPacketizer_GeneratePadding(f *testing.F) {
	// mixed seeds.
	f.Add(uint32(0))
	f.Add(uint32(1))
	f.Add(uint32(5))
	f.Add(uint32(16))

	f.Fuzz(func(t *testing.T, samples uint32) {
		samples %= 64

		packetizer := NewPacketizerWithOptions(
			1200,
			&codecs.G722Payloader{},
			NewFixedSequencer(0),
			90000,
		)

		pads := packetizer.GeneratePadding(samples)
		if samples == 0 {
			assert.Nil(t, pads)

			return
		}

		assert.Len(t, pads, int(samples))
		for _, p := range pads {
			assert.True(t, p.Header.Padding)
			assert.Equal(t, byte(255), p.Header.PaddingSize)
			assert.Nil(t, p.Payload)

			raw, err := p.Marshal()
			assert.NoError(t, err)

			var back Packet
			assert.NoError(t, back.Unmarshal(raw))
			assert.True(t, back.Padding)
		}
	})
}
