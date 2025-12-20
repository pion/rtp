// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package rtp

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransportCCExtensionTooSmall(t *testing.T) {
	t1 := TransportCCExtension{}

	rawData := []byte{}

	err := t1.Unmarshal(rawData)
	assert.ErrorIs(t, err, errTooSmall)
}

func TestTransportCCExtension(t *testing.T) {
	t1 := TransportCCExtension{}

	rawData := []byte{
		0x00, 0x02,
	}

	err := t1.Unmarshal(rawData)
	assert.NoError(t, err)

	t2 := TransportCCExtension{
		TransportSequence: 2,
	}

	assert.Equal(t, t1, t2)

	dstData, _ := t2.Marshal()
	assert.Equal(t, dstData, rawData)
}

func TestTransportCCExtensionExtraBytes(t *testing.T) {
	t1 := TransportCCExtension{}

	rawData := []byte{
		0x00, 0x02, 0x00, 0xff, 0xff,
	}

	err := t1.Unmarshal(rawData)
	assert.NoError(t, err)

	t2 := TransportCCExtension{
		TransportSequence: 2,
	}

	assert.Equal(t, t1, t2)
}

func TestTransportCCExtensionMarshalTo(t *testing.T) {
	ext := TransportCCExtension{TransportSequence: 1234}

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
	transportCCSink    []byte
	transportCCBuf     = make([]byte, transportCCExtensionSize)
	transportCCSinkInt int
)

func BenchmarkTransportCCExtension_Marshal(b *testing.B) {
	ext := TransportCCExtension{TransportSequence: 1234}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		transportCCSink, _ = ext.Marshal()
	}
}

func BenchmarkTransportCCExtension_MarshalTo(b *testing.B) {
	ext := TransportCCExtension{TransportSequence: 1234}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		transportCCSinkInt, _ = ext.MarshalTo(transportCCBuf)
	}
}
