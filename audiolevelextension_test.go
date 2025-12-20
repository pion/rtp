// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package rtp

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAudioLevelExtensionTooSmall(t *testing.T) {
	a := AudioLevelExtension{}
	rawData := []byte{}
	assert.ErrorIs(t, a.Unmarshal(rawData), errTooSmall)
}

func TestAudioLevelExtensionVoiceTrue(t *testing.T) {
	a1 := AudioLevelExtension{}
	rawData := []byte{
		0x88,
	}
	assert.NoError(t, a1.Unmarshal(rawData))

	a2 := AudioLevelExtension{
		Level: 8,
		Voice: true,
	}
	assert.Equal(t, a2, a1)

	dstData, _ := a2.Marshal()
	assert.Equal(t, rawData, dstData)
}

func TestAudioLevelExtensionVoiceFalse(t *testing.T) {
	a1 := AudioLevelExtension{}
	rawData := []byte{
		0x8,
	}
	assert.NoError(t, a1.Unmarshal(rawData))

	a2 := AudioLevelExtension{
		Level: 8,
		Voice: false,
	}
	assert.Equal(t, a2, a1)

	dstData, _ := a2.Marshal()
	assert.Equal(t, rawData, dstData)
}

func TestAudioLevelExtensionLevelOverflow(t *testing.T) {
	a := AudioLevelExtension{
		Level: 128,
		Voice: false,
	}

	_, err := a.Marshal()
	assert.ErrorIs(t, err, errAudioLevelOverflow)

	_, err = a.MarshalTo(make([]byte, 10))
	assert.ErrorIs(t, err, errAudioLevelOverflow)
}

func TestAudioLevelExtensionMarshalTo(t *testing.T) {
	a := AudioLevelExtension{Level: 8, Voice: true}

	buf := make([]byte, a.MarshalSize())
	n, err := a.MarshalTo(buf)
	assert.NoError(t, err)
	assert.Equal(t, a.MarshalSize(), n)

	expected, _ := a.Marshal()
	assert.Equal(t, expected, buf)

	_, err = a.MarshalTo(nil)
	assert.ErrorIs(t, err, io.ErrShortBuffer)
}

//nolint:gochecknoglobals
var (
	audioLevelSink    []byte
	audioLevelBuf     = make([]byte, audioLevelExtensionSize)
	audioLevelSinkInt int
)

func BenchmarkAudioLevelExtension_Marshal(b *testing.B) {
	ext := AudioLevelExtension{Level: 8, Voice: true}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		audioLevelSink, _ = ext.Marshal()
	}
}

func BenchmarkAudioLevelExtension_MarshalTo(b *testing.B) {
	ext := AudioLevelExtension{Level: 8, Voice: true}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		audioLevelSinkInt, _ = ext.MarshalTo(audioLevelBuf)
	}
}
