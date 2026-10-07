// SPDX-FileCopyrightText: 2026 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package rtp

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testPayloader struct {
	mtu     uint16
	payload []byte
}

func (p *testPayloader) Payload(mtu uint16, payload []byte) [][]byte {
	p.mtu = mtu
	p.payload = payload

	return [][]byte{
		{0x01},
		{0x02},
	}
}

func TestLegacyPayloaderAdapter(t *testing.T) {
	payloader := &testPayloader{}
	adapter := LegacyPayloaderAdapter{Payloader: payloader}
	sample := MediaSample{Payload: []byte{0xAA, 0xBB}}

	var fragments []PayloadFragment
	err := adapter.Packetize(PacketizeContext{MTU: 112}, sample, func(fragment PayloadFragment) error {
		fragments = append(fragments, fragment)

		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, uint16(100), payloader.mtu)
	assert.Equal(t, sample.Payload, payloader.payload)
	require.Len(t, fragments, 2)
	assert.Equal(t, []byte{0x01}, fragments[0].Payload)
	assert.False(t, fragments[0].Marker)
	assert.Equal(t, []byte{0x02}, fragments[1].Payload)
	assert.True(t, fragments[1].Marker)
}

func TestLegacyPayloaderAdapterEmitError(t *testing.T) {
	expectedErr := errors.New("emit failed") // nolint:err113
	adapter := LegacyPayloaderAdapter{Payloader: &testPayloader{}}

	err := adapter.Packetize(PacketizeContext{MTU: 112}, MediaSample{}, func(PayloadFragment) error {
		return expectedErr
	})

	assert.ErrorIs(t, err, expectedErr)
}

type testDepacketizer struct {
	reset bool
}

func (d *testDepacketizer) Unmarshal(packet []byte) ([]byte, error) {
	return append([]byte{0x00}, packet...), nil
}

func (d *testDepacketizer) IsPartitionHead(payload []byte) bool {
	return len(payload) > 0 && payload[0] == 0x01
}

func (d *testDepacketizer) IsPartitionTail(marker bool, _ []byte) bool {
	return marker
}

func (d *testDepacketizer) Reset() {
	d.reset = true
}

func TestLegacyDepacketizerAdapter(t *testing.T) {
	depacketizer := &testDepacketizer{}
	adapter := LegacyDepacketizerAdapter{Depacketizer: depacketizer}
	packet := PacketView{
		Header:  &Header{Marker: true},
		Payload: []byte{0x01, 0x02},
	}

	info, err := adapter.Inspect(packet)
	require.NoError(t, err)
	assert.True(t, info.StartsSample)
	assert.True(t, info.EndsSample)

	sample, err := adapter.AppendToSample([]byte{0xFF}, packet)
	require.NoError(t, err)
	assert.Equal(t, []byte{0xFF, 0x00, 0x01, 0x02}, sample)

	adapter.Reset()
	assert.True(t, depacketizer.reset)
}
