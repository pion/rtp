package rtp

import (
	"errors"
)

const (
	// AudioLevelOneByteExtensionSize One byte header size
	AudioLevelOneByteExtensionSize = 2
	// AudioLevelTwoByteExtensionSize Two byte header size
	AudioLevelTwoByteExtensionSize = 4
)

var (
	errInvalidSize           = errors.New("invalid buffer size")
	errInvalidExtensonLength = errors.New("invalid extension length")
	errAudioLevelOverflow    = errors.New("audio level overflow")
)

// AudioLevelExtension is a extension payload format described in
// https://tools.ietf.org/html/rfc6464
//
// Implementation based on:
// https://chromium.googlesource.com/external/webrtc/+/e2a017725570ead5946a4ca8235af27470ca0df9/webrtc/modules/rtp_rtcp/source/rtp_header_extensions.cc#49
//
// One byte format:
// 0                   1
// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |  ID   | len=0 |V| level       |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//
// Two byte format:
// 0                   1                   2                   3
// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |      ID       |     len=1     |V|    level    |    0 (pad)    |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
type AudioLevelExtension struct {
	ID    uint8
	Level uint8
	Voice bool
}

// Marshal serializes the members to buffer
func (a *AudioLevelExtension) Marshal() ([]byte, error) {
	if a.Level > 127 {
		return nil, errAudioLevelOverflow
	}
	voice := uint8(0x00)
	if a.Voice {
		voice = 0x80
	}
	buf := make([]byte, AudioLevelOneByteExtensionSize)
	buf[0] = a.ID << 4 & 0xf0
	buf[1] = voice | a.Level
	return buf, nil
}

// Unmarshal parses the passed byte slice and stores the result in the members
func (a *AudioLevelExtension) Unmarshal(rawData []byte) error {
	// one byte format
	switch len(rawData) {
	case AudioLevelOneByteExtensionSize:
		if rawData[0]&^0xF0 != 0 {
			return errInvalidExtensonLength
		}
		a.ID = rawData[0] >> 4
		a.Level = rawData[1] & 0x7F
		a.Voice = rawData[1]&0x80 != 0
		return nil
	case AudioLevelTwoByteExtensionSize:
		if rawData[1] != 1 {
			return errInvalidExtensonLength
		}
		a.ID = rawData[0]
		a.Level = rawData[2] & 0x7F
		a.Voice = rawData[2]&0x80 != 0
		return nil
	}
	return errInvalidSize
}
