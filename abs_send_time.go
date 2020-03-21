package rtp

import (
	"time"
)

const NTPOffset = 0x83AA7E80

func toNtpTime(t time.Time) uint64 {
	u := uint64(t.UnixNano())
	s := u / 1e9
	s += NTPOffset //offset in seconds between unix epoch and ntp epoch
	s <<= 32
	f := ((u % 1e9) << 32) / 1e9
	return s | f
}

// TimeToAbsSendTime ...
func TimeToAbsSendTime(setTime time.Time) uint32 {
	t := toNtpTime(setTime)
	return uint32((t >> 14) & 0xFFFFFF)
}

// AbsSendTimeSeconds ...
func AbsSendTimeSeconds(abs uint32) uint32 {
	return abs >> 18
}

// AbsSendTimeFractions ...
func AbsSendTimeFractions(abs uint32) uint32 {
	return (abs & 0x03ffff)
}

func AbsSendTimeFractionsToRoughMillis(abs uint32) uint64 {
	// convert this to a rough ntp format and get the lower 32 bits
	// which represent the fractional component
	x := uint64((abs << 14) & 0xFFFFFFFF)
	x = (x * 1e9) >> 32
	return x / 1e6
}

// AbsSendTimeCompareMS ...
func AbsSendTimeCompareMS(now uint32, incomingPacketTime uint32) uint64 {
	delta := AbsSendTimeDelta(now, incomingPacketTime)
	return uint64(uint64(AbsSendTimeSeconds(delta))*1000 + AbsSendTimeFractionsToRoughMillis(delta))
}

func AbsSendTimeDelta(now uint32, prev uint32) uint32 {
	delta := now - prev
	if now < prev {
		delta = 0xFFFFFF - prev + now
	}
	return delta
}
