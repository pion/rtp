package rtp

import (
	"testing"
	"time"
)

const absSendTimeResolution = 3800 * time.Nanosecond

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

	for i, in := range tests {
		out := toNtpTime(in.t)
		if out != in.n {
			t.Errorf("[%d] Converted NTP time from time.Time differs, expected: %d, got: %d",
				i, in.n, out,
			)
		}
	}
	for i, in := range tests {
		out := toTime(in.n)
		diff := in.t.Sub(out)
		if diff < -absSendTimeResolution || absSendTimeResolution < diff {
			t.Errorf("[%d] Converted time.Time from NTP time differs, expected: %v, got: %v",
				i, in.t.UTC(), out.UTC(),
			)
		}
	}
}

func TestAbsSendTimeExtension_Roundtrip(t *testing.T) {
	tests := []AbsSendTimeExtension{
		{
			Timestamp: 123456,
		},
		{
			Timestamp: 654321,
		},
	}
	for i, in := range tests {
		b, err := in.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		var out AbsSendTimeExtension
		if err = out.Unmarshal(b); err != nil {
			t.Fatal(err)
		}
		if in.Timestamp != out.Timestamp {
			t.Errorf("[%d] Timestamp differs, expected: %d, got: %d", i, in.Timestamp, out.Timestamp)
		}
	}
}

func TestAbsSendTimeExtension_Estimate(t *testing.T) {
	tests := []struct {
		sendNTP    uint64
		receiveNTP uint64
	}{ // FFFFFFC000000000 mask of second
		{0xa0c65b1000100000, 0xa0c65b1001000000}, // not carried
		{0xa0c65b3f00000000, 0xa0c65b4001000000}, // carried during transmission
	}
	for i, in := range tests {
		inTime := toTime(in.sendNTP)
		send := &AbsSendTimeExtension{in.sendNTP >> 14}
		b, err := send.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		var received AbsSendTimeExtension
		if err = received.Unmarshal(b); err != nil {
			t.Fatal(err)
		}

		estimated := received.Estimate(toTime(in.receiveNTP))
		diff := estimated.Sub(inTime)
		if diff < -absSendTimeResolution || absSendTimeResolution < diff {
			t.Errorf("[%d] Estimated time differs, expected: %v, estimated: %v (receive time: %v)",
				i, inTime.UTC(), estimated.UTC(), toTime(in.receiveNTP).UTC(),
			)
		}
	}
}
