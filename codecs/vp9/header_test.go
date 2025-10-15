// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package vp9

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaderUnmarshal(t *testing.T) {
	cases := []struct {
		name   string
		byts   []byte
		sh     Header
		width  uint16
		height uint16
	}{
		{
			"chrome webrtc",
			[]byte{
				0x82, 0x49, 0x83, 0x42, 0x00, 0x77, 0xf0, 0x32,
				0x34, 0x30, 0x38, 0x24, 0x1c, 0x19, 0x40, 0x18,
				0x03, 0x40, 0x5f, 0xb4,
			},
			Header{
				ShowFrame: true,
				ColorConfig: &HeaderColorConfig{
					BitDepth:     8,
					SubsamplingX: true,
					SubsamplingY: true,
				},
				FrameSize: &HeaderFrameSize{
					FrameWidthMinus1:  1919,
					FrameHeightMinus1: 803,
				},
			},
			1920,
			804,
		},
		{
			"vp9 sample",
			[]byte{
				0x82, 0x49, 0x83, 0x42, 0x40, 0xef, 0xf0, 0x86,
				0xf4, 0x04, 0x21, 0xa0, 0xe0, 0x00, 0x30, 0x70,
				0x00, 0x00, 0x00, 0x01,
			},
			Header{
				ShowFrame: true,
				ColorConfig: &HeaderColorConfig{
					BitDepth:     8,
					ColorSpace:   2,
					SubsamplingX: true,
					SubsamplingY: true,
				},
				FrameSize: &HeaderFrameSize{
					FrameWidthMinus1:  3839,
					FrameHeightMinus1: 2159,
				},
			},
			3840,
			2160,
		},
		{
			"show existing frame",
			[]byte{
				0b10101010, 0x49, 0x83, 0x42, 0x40, 0xef, 0xf0, 0x86,
				0xf4, 0x04, 0x21, 0xa0, 0xe0, 0x00, 0x30, 0x70,
				0x00, 0x00, 0x00, 0x01,
			},
			Header{
				Profile:           1,
				ShowExistingFrame: true,
				FrameToShowMapIdx: 2,
			},
			0,
			0,
		},
		{
			"profile 0",
			[]byte{
				0x92, 0x49, 0x83, 0x42, 0x40, 0xef, 0xf0, 0x86,
				0xf4, 0x04, 0x21, 0xa0, 0xe0, 0x00, 0x30, 0x70,
				0x00, 0x00, 0x00, 0x01,
			},
			Header{
				Profile:   2,
				ShowFrame: true,
				ColorConfig: &HeaderColorConfig{
					BitDepth:     10,
					ColorSpace:   4,
					SubsamplingX: true,
					SubsamplingY: true,
				},
				FrameSize: &HeaderFrameSize{
					FrameWidthMinus1:  7678,
					FrameHeightMinus1: 4318,
				},
			},
			0x1dff,
			0x10df,
		},
	}

	for _, ca := range cases {
		t.Run(ca.name, func(t *testing.T) {
			var sh Header
			assert.NoError(t, sh.Unmarshal(ca.byts))
			assert.Equal(t, ca.sh, sh)
			assert.Equal(t, ca.width, sh.Width())
			assert.Equal(t, ca.height, sh.Height())
		})
	}
}
