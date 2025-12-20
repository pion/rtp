// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package rtp

import (
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
		assert.Equalf(
			t, in.n, out,
			"[%d] Converted NTP time from time.Time differs", i,
		)
	}
	for i, in := range tests {
		out := toTime(in.n)
		diff := in.t.Sub(out)
		assert.GreaterOrEqualf(
			t, diff, -absSendTimeResolution,
			"[%d] Converted time.Time from NTP time differs", i,
		)
		assert.LessOrEqual(
			t, diff, absSendTimeResolution,
			"[%d] Converted time.Time from NTP time differs", i,
		)
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
		assert.NoError(t, err)

		var out AbsSendTimeExtension
		assert.NoError(t, out.Unmarshal(b))
		assert.Equalf(
			t, in.Timestamp, out.Timestamp,
			"[%d] Timestamp differs", i,
		)
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
		assert.NoError(t, err)
		var received AbsSendTimeExtension
		assert.NoError(t, received.Unmarshal(b))

		estimated := received.Estimate(toTime(in.receiveNTP))
		diff := estimated.Sub(inTime)
		assert.GreaterOrEqualf(
			t, diff, -absSendTimeResolution,
			"[%d] Estimated time differs, expected: %v, estimated: %v (receive time: %v)",
			i, inTime.UTC(), estimated.UTC(), toTime(in.receiveNTP).UTC(),
		)
		assert.LessOrEqual(
			t, diff, absSendTimeResolution,
			"[%d] Estimated time differs, expected: %v, estimated: %v (receive time: %v)",
			i, inTime.UTC(), estimated.UTC(), toTime(in.receiveNTP).UTC(),
		)
	}
}

func TestAbsSendTimeExtensionMarshalTo(t *testing.T) {
	ext := AbsSendTimeExtension{Timestamp: 123456}

	buf := make([]byte, ext.MarshalSize())
	n, err := ext.MarshalTo(buf)
	assert.NoError(t, err)
	assert.Equal(t, ext.MarshalSize(), n)

	expected, _ := ext.Marshal()
	assert.Equal(t, expected, buf)

	_, err = ext.MarshalTo(nil)
	assert.ErrorIs(t, err, io.ErrShortBuffer)
}

//nolint:gochecknoglobals
var (
	absSendTimeSink    []byte
	absSendTimeBuf     = make([]byte, absSendTimeExtensionSize)
	absSendTimeSinkInt int
)

func BenchmarkAbsSendTimeExtension_Marshal(b *testing.B) {
	ext := AbsSendTimeExtension{Timestamp: 123456}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		absSendTimeSink, _ = ext.Marshal()
	}
}

func BenchmarkAbsSendTimeExtension_MarshalTo(b *testing.B) {
	ext := AbsSendTimeExtension{Timestamp: 123456}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		absSendTimeSinkInt, _ = ext.MarshalTo(absSendTimeBuf)
	}
}
