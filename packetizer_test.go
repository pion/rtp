package rtp

import (
	"testing"
	"time"

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
