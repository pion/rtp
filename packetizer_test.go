package rtp

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/pion/rtp/codecs"
)

func TestPacketizer(t *testing.T) {
	multiplepayload := make([]byte, 128)
	//use the G722 payloader here, because it's very simple and all 0s is valid G722 data.
	packetizer := NewPacketizer(100, 98, 0x1234ABCD, &codecs.G722Payloader{}, NewRandomSequencer(), 90000)
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
	//use the G722 payloader here, because it's very simple and all 0s is valid G722 data.
	pktizer := NewPacketizer(100, 98, 0x1234ABCD, &codecs.G722Payloader{}, NewFixedSequencer(1234), 90000)
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
			PayloadOffset:    0, // not set by Packetize() at now
			PayloadType:      98,
			SequenceNumber:   1234,
			Timestamp:        45678,
			SSRC:             0x1234ABCD,
			CSRC:             nil,
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
