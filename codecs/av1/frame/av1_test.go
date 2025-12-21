// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package frame

import (
	"testing"

	"github.com/pion/rtp/codecs"
	"github.com/stretchr/testify/assert"
)

// First is Fragment (and no buffer)
// Self contained OBU
// OBU spread across 3 packets.
func TestAV1_ReadFrames(t *testing.T) {
	// First is Fragment of OBU, but no OBU Elements is cached
	fragm := &AV1{}
	frames, err := fragm.ReadFrames(&codecs.AV1Packet{Z: true, OBUElements: [][]byte{{0x01}}}) // nolint:staticcheck
	assert.NoError(t, err)
	assert.Equal(t, [][]byte{}, frames, "No frames should be generated")

	fragm = &AV1{}
	frames, err = fragm.ReadFrames(&codecs.AV1Packet{OBUElements: [][]byte{{0x01}}}) // nolint:staticcheck
	assert.NoError(t, err)
	assert.Equal(t, [][]byte{{0x01}}, frames, "One frame should be generated")

	fragm = &AV1{}
	frames, err = fragm.ReadFrames(&codecs.AV1Packet{Y: true, OBUElements: [][]byte{{0x00}}}) // nolint:staticcheck
	assert.NoError(t, err)
	assert.Equal(t, [][]byte{}, frames, "No frames should be generated")

	frames, err = fragm.ReadFrames(&codecs.AV1Packet{Z: true, OBUElements: [][]byte{{0x01}}}) // nolint:staticcheck
	assert.NoError(t, err)
	assert.Equal(t, [][]byte{{0x00, 0x01}}, frames, "One frame should be generated")
}

// Marshal some AV1 Frames to RTP, assert that AV1 can get them back in the original format.
func TestAV1_ReadFrames_E2E(t *testing.T) {
	const mtu = 1500
	frames := [][]byte{
		{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A},
		{0x00, 0x01},
		{
			0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A,
			0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A,
		},
		{0x00, 0x01},
	}

	frames = append(frames, []byte{})
	for i := 0; i <= 5; i++ {
		frames[len(frames)-1] = append(
			frames[len(frames)-1],
			[]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}...,
		)
	}

	frames = append(frames, []byte{})
	for i := 0; i <= 500; i++ {
		frames[len(frames)-1] = append(
			frames[len(frames)-1],
			[]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}...,
		)
	}

	payloader := &codecs.AV1Payloader{}
	f := &AV1{}
	for _, originalFrame := range frames {
		for _, payload := range payloader.Payload(mtu, originalFrame) {
			rtpPacket := &codecs.AV1Packet{} // nolint:staticcheck
			_, err := rtpPacket.Unmarshal(payload)
			assert.NoError(t, err)

			decodedFrame, err := f.ReadFrames(rtpPacket)
			assert.NoError(t, err)

			if len(decodedFrame) != 0 {
				assert.Equal(t, originalFrame, decodedFrame[0])
			}
		}
	}
}
