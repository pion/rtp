package rtp

import (
	"fmt"
	"testing"

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
