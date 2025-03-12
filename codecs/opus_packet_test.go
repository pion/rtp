// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package codecs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpusPacket_Unmarshal(t *testing.T) {
	pck := OpusPacket{}

	// Nil packet
	raw, err := pck.Unmarshal(nil)
	assert.ErrorIs(t, err, errNilPacket)
	assert.Nil(t, raw, "Result should be nil in case of error")

	// Empty packet
	raw, err = pck.Unmarshal([]byte{})
	assert.ErrorIs(t, err, errShortPacket)
	assert.Nil(t, raw, "Result should be nil in case of error")

	// Normal packet
	raw, err = pck.Unmarshal([]byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x90})
	assert.NoError(t, err)
	assert.NotNil(t, raw, "Result shouldn't be nil in case of success")
}

func TestOpusPayloader_Payload(t *testing.T) {
	pck := OpusPayloader{}
	payload := []byte{0x90, 0x90, 0x90}

	// Positive MTU, nil payload
	res := pck.Payload(1, nil)
	assert.Len(t, res, 0, "Generated payload should be empty")

	// Positive MTU, small payload
	res = pck.Payload(1, payload)
	assert.Len(t, res, 1, "Generated payload should be the 1")

	// Positive MTU, small payload
	res = pck.Payload(2, payload)
	assert.Len(t, res, 1, "Generated payload should be the 1")
}

func TestOpusIsPartitionHead(t *testing.T) {
	opus := &OpusPacket{}
	t.Run("NormalPacket", func(t *testing.T) {
		assert.True(
			t, opus.IsPartitionHead([]byte{0x00, 0x00}),
			"All OPUS RTP packet should be the head of a new partition",
		)
	})
}
