package rtp

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/pion/rtp/v2/codecs"
)

func TestPacketizer(t *testing.T) {
	multiplepayload := make([]byte, 128)
	// use the G722 payloader here, because it's very simple and all 0s is valid G722 data.
	packetizer := NewPacketizer(100, 98, 0x1234ABCD, &codecs.G722Payloader{}, NewRandomSequencer())
	packets := packetizer.Packetize(multiplepayload, 2000)

	if len(packets) != 2 {
		packetlengths := ""
		for i := 0; i < len(packets); i++ {
			packetlengths += fmt.Sprintf("Packet %d length %d\n", i, len(packets[i].Payload))
		}
		t.Fatalf("Generated %d packets instead of 2\n%s", len(packets), packetlengths)
	}
}

func TestPacketizer_AbsSendTime(t *testing.T) {
	// use the G722 payloader here, because it's very simple and all 0s is valid G722 data.
	pktizer := NewPacketizer(100, 98, 0x1234ABCD, &codecs.G722Payloader{}, NewFixedSequencer(1234))
	pktizer.(*packetizer).Timestamp = 45678
	pktizer.(*packetizer).timegen = func() time.Time {
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

	if len(packets) != 1 {
		t.Fatalf("Generated %d packets instead of 1", len(packets))
	}
	if !reflect.DeepEqual(expected, packets[0]) {
		t.Errorf("Packetize failed\nexpected: %v\n     got: %v", expected, packets[0])
	}
}

func TestPacketizer_Roundtrip(t *testing.T) {
	multiplepayload := make([]byte, 128)
	packetizer := NewPacketizer(100, 98, 0x1234ABCD, &codecs.G722Payloader{}, NewRandomSequencer())
	packets := packetizer.Packetize(multiplepayload, 1000)

	rawPkts := make([][]byte, 0, 1400)
	for _, pkt := range packets {
		raw, err := pkt.Marshal()
		if err != nil {
			t.Errorf("Packet Marshal failed: %v", err)
		}

		rawPkts = append(rawPkts, raw)
	}

	for ndx, raw := range rawPkts {
		expectedPkt := packets[ndx]
		pkt := &Packet{}

		err := pkt.Unmarshal(raw)
		if err != nil {
			t.Errorf("Packet Unmarshal failed: %v", err)
		}

		if len(raw) != pkt.MarshalSize() {
			t.Errorf("Packet sizes don't match, expected %d but got %d", len(raw), pkt.MarshalSize())
		}
		if expectedPkt.MarshalSize() != pkt.MarshalSize() {
			t.Errorf("Packet marshal sizes don't match, expected %d but got %d", expectedPkt.MarshalSize(), pkt.MarshalSize())
		}

		if expectedPkt.Version != pkt.Version {
			t.Errorf("Packet versions don't match, expected %d but got %d", expectedPkt.Version, pkt.Version)
		}
		if expectedPkt.Padding != pkt.Padding {
			t.Errorf("Packet versions don't match, expected %t but got %t", expectedPkt.Padding, pkt.Padding)
		}
		if expectedPkt.Extension != pkt.Extension {
			t.Errorf("Packet versions don't match, expected %v but got %v", expectedPkt.Extension, pkt.Extension)
		}
		if expectedPkt.Marker != pkt.Marker {
			t.Errorf("Packet versions don't match, expected %v but got %v", expectedPkt.Marker, pkt.Marker)
		}
		if expectedPkt.PayloadType != pkt.PayloadType {
			t.Errorf("Packet versions don't match, expected %d but got %d", expectedPkt.PayloadType, pkt.PayloadType)
		}
		if expectedPkt.SequenceNumber != pkt.SequenceNumber {
			t.Errorf("Packet versions don't match, expected %d but got %d", expectedPkt.SequenceNumber, pkt.SequenceNumber)
		}
		if expectedPkt.Timestamp != pkt.Timestamp {
			t.Errorf("Packet versions don't match, expected %d but got %d", expectedPkt.Timestamp, pkt.Timestamp)
		}
		if expectedPkt.SSRC != pkt.SSRC {
			t.Errorf("Packet versions don't match, expected %d but got %d", expectedPkt.SSRC, pkt.SSRC)
		}
		if !reflect.DeepEqual(expectedPkt.CSRC, pkt.CSRC) {
			t.Errorf("Packet versions don't match, expected %v but got %v", expectedPkt.CSRC, pkt.CSRC)
		}
		if expectedPkt.ExtensionProfile != pkt.ExtensionProfile {
			t.Errorf("Packet versions don't match, expected %d but got %d", expectedPkt.ExtensionProfile, pkt.ExtensionProfile)
		}
		if !reflect.DeepEqual(expectedPkt.Extensions, pkt.Extensions) {
			t.Errorf("Packet versions don't match, expected %v but got %v", expectedPkt.Extensions, pkt.Extensions)
		}
		if !reflect.DeepEqual(expectedPkt.Payload, pkt.Payload) {
			t.Errorf("Packet versions don't match, expected %v but got %v", expectedPkt.Payload, pkt.Payload)
		}

		if !reflect.DeepEqual(expectedPkt, pkt) {
			t.Errorf("Packets don't match, expected %v but got %v", expectedPkt, pkt)
		}
	}
}
