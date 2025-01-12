// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package vp9

import (
	"reflect"
	"testing"
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
	}

	for _, ca := range cases {
		t.Run(ca.name, func(t *testing.T) {
			var sh Header
			err := sh.Unmarshal(ca.byts)
			if err != nil {
				t.Fatal("unexpected error")
			}

			if !reflect.DeepEqual(ca.sh, sh) {
				t.Fatalf("expected %#+v, got %#+v", ca.sh, sh)
			}
			if ca.width != sh.Width() {
				t.Fatalf("unexpected width")
			}
			if ca.height != sh.Height() {
				t.Fatalf("unexpected height")
			}
		})
	}
}
