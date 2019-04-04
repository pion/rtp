package rtp

import (
	"fmt"
	"testing"
	"time"

	"github.com/pion/rtp/codecs"
	"github.com/stretchr/testify/assert"
)

func TestNtpConversion(t *testing.T) {
	loc := time.FixedZone("UTC-5", -5*60*60)

	tests := []struct {
		t time.Time
		n uint64
	}{
		{t: time.Date(1985, time.June, 23, 4, 0, 0, 0, loc), n: 0xa0c65b1000000000},
		{t: time.Date(1999, time.December, 31, 23, 59, 59, 500000, loc), n: 0xbc18084f0020c49b},
		{t: time.Date(2019, time.March, 27, 13, 39, 30, 8675309, loc), n: 0xe04641e202388b88},
	}

	for _, in := range tests {
		out := toNtpTime(in.t)
		assert.Equal(t, in.n, out)
	}
}

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
