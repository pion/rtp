package rtp

import "time"

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
