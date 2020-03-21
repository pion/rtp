package rtp

import "time"

func toNtpTime(t time.Time) uint64 {
	var f uint64
	u := uint64(t.UnixNano())
	s := u / 1e9
	s += 0x83AA7E80 //offset in seconds between unix epoch and ntp epoch
	s <<= 32
	f = ((u % 1e9) << 32) / 1e9
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
	return ((abs & 0x03ffff) >> 8)
}
