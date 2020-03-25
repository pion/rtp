package rtp

import (
	"time"
)

const (
	absSendTimeExtensionSize = 4
)

// AbsSendTimeExtension is a extension payload format in
// http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time
//  0 1 2 3 4 5 6 7
// +-+-+-+-+-+-+-+-+
// |  ID   |  len  |
// +-+-+-+-+-+-+-+-+
// per RFC 5285
// Len is the number of bytes in the extension - 1.
type AbsSendTimeExtension struct {
	ID        uint8
	Timestamp uint64
}

// Marshal serializes the members to buffer.
func (t *AbsSendTimeExtension) Marshal() ([]byte, error) {
	return []byte{
		(t.ID << 4) | 2,
		byte(t.Timestamp & 0xFF0000 >> 16),
		byte(t.Timestamp & 0xFF00 >> 8),
		byte(t.Timestamp & 0xFF),
	}, nil
}

// Unmarshal parses the passed byte slice and stores the result in the members.
func (t *AbsSendTimeExtension) Unmarshal(rawData []byte) error {
	if len(rawData) < absSendTimeExtensionSize {
		return errTooSmall
	}
	t.ID = rawData[0] >> 4
	t.Timestamp = uint64(rawData[1])<<16 | uint64(rawData[2])<<8 | uint64(rawData[3])
	return nil
}

// Estimate absolute send time according to the receive time.
// Note that if the transmission delay is larger than 64 seconds, estimated time will be wrong.
func (t *AbsSendTimeExtension) Estimate(receive time.Time) time.Time {
	receiveNTP := toNtpTime(receive)
	ntp := receiveNTP&0xFFFFFFC000000000 | (t.Timestamp&0xFFFFFF)<<14
	if receiveNTP < ntp {
		// Receive time must be always later than send time
		ntp -= 0x1000000 << 14
	}

	return toTime(ntp)
}

func toNtpTime(t time.Time) uint64 {
	var s uint64
	var f uint64
	u := uint64(t.UnixNano())
	s = u / 1e9
	s += 0x83AA7E80 //offset in seconds between unix epoch and ntp epoch
	f = u % 1e9
	f <<= 32
	f /= 1e9
	s <<= 32

	return s | f
}

func toTime(t uint64) time.Time {
	s := t >> 32
	f := t & 0xFFFFFFFF
	f *= 1e9
	f >>= 32
	s -= 0x83AA7E80
	u := s*1e9 + f

	return time.Unix(0, int64(u))
}
