// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package rtp

import (
	"encoding/hex"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVLAMarshal(t *testing.T) {
	t.Run("3 streams no resolution and framerate", func(t *testing.T) {
		vla := &VLA{
			RTPStreamID:    0,
			RTPStreamCount: 3,
			ActiveSpatialLayer: []SpatialLayer{
				{
					RTPStreamID:    0,
					SpatialID:      0,
					TargetBitrates: []int{150},
				},
				{
					RTPStreamID:    1,
					SpatialID:      0,
					TargetBitrates: []int{240, 400},
				},
				{
					RTPStreamID:    2,
					SpatialID:      0,
					TargetBitrates: []int{720, 1200},
				},
			},
		}

		bytesActual, err := vla.Marshal()
		assert.NoError(t, err)
		bytesExpected, err := hex.DecodeString("21149601f0019003d005b009")
		assert.NoError(t, err)
		assert.Equal(
			t, bytesExpected, bytesActual,
			"expected %s, actual %s", hex.EncodeToString(bytesExpected), hex.EncodeToString(bytesActual),
		)
	})

	t.Run("3 streams with resolution and framerate", func(t *testing.T) {
		vla := &VLA{
			RTPStreamID:    2,
			RTPStreamCount: 3,
			ActiveSpatialLayer: []SpatialLayer{
				{
					RTPStreamID:    0,
					SpatialID:      0,
					TargetBitrates: []int{150},
					Width:          320,
					Height:         180,
					Framerate:      30,
				},
				{
					RTPStreamID:    1,
					SpatialID:      0,
					TargetBitrates: []int{240, 400},
					Width:          640,
					Height:         360,
					Framerate:      30,
				},
				{
					RTPStreamID:    2,
					SpatialID:      0,
					TargetBitrates: []int{720, 1200},
					Width:          1280,
					Height:         720,
					Framerate:      30,
				},
			},
			HasResolutionAndFramerate: true,
		}

		bytesActual, err := vla.Marshal()
		assert.NoError(t, err)
		bytesExpected, err := hex.DecodeString("a1149601f0019003d005b009013f00b31e027f01671e04ff02cf1e")
		assert.NoError(t, err)
		assert.Equal(
			t, bytesExpected, bytesActual,
			"expected %s, actual %s", hex.EncodeToString(bytesExpected), hex.EncodeToString(bytesActual),
		)
	})

	t.Run("Negative RTPStreamCount", func(t *testing.T) {
		vla := &VLA{
			RTPStreamID:        0,
			RTPStreamCount:     -1,
			ActiveSpatialLayer: []SpatialLayer{},
		}
		_, err := vla.Marshal()
		assert.ErrorIs(t, err, ErrVLAInvalidStreamCount)
	})

	t.Run("RTPStreamCount too large", func(t *testing.T) {
		vla := &VLA{
			RTPStreamID:        0,
			RTPStreamCount:     5,
			ActiveSpatialLayer: []SpatialLayer{{}, {}, {}, {}, {}},
		}
		_, err := vla.Marshal()
		assert.ErrorIs(t, err, ErrVLAInvalidStreamCount)
	})

	t.Run("Negative RTPStreamID", func(t *testing.T) {
		vla := &VLA{
			RTPStreamID:        -1,
			RTPStreamCount:     1,
			ActiveSpatialLayer: []SpatialLayer{{}},
		}
		_, err := vla.Marshal()
		assert.ErrorIs(t, err, ErrVLAInvalidStreamID)
	})

	t.Run("RTPStreamID to large", func(t *testing.T) {
		vla := &VLA{
			RTPStreamID:        1,
			RTPStreamCount:     1,
			ActiveSpatialLayer: []SpatialLayer{{}},
		}
		_, err := vla.Marshal()
		assert.ErrorIs(t, err, ErrVLAInvalidStreamID)
	})

	t.Run("Invalid stream ID in the spatial layer", func(t *testing.T) {
		vla := &VLA{
			RTPStreamID:    0,
			RTPStreamCount: 1,
			ActiveSpatialLayer: []SpatialLayer{{
				RTPStreamID: -1,
			}},
		}
		_, err := vla.Marshal()
		assert.ErrorIs(t, err, ErrVLAInvalidStreamID)
		vla = &VLA{
			RTPStreamID:    0,
			RTPStreamCount: 1,
			ActiveSpatialLayer: []SpatialLayer{{
				RTPStreamID: 1,
			}},
		}
		_, err = vla.Marshal()
		assert.ErrorIs(t, err, ErrVLAInvalidStreamID)
	})

	t.Run("Invalid spatial ID in the spatial layer", func(t *testing.T) {
		vla := &VLA{
			RTPStreamID:    0,
			RTPStreamCount: 1,
			ActiveSpatialLayer: []SpatialLayer{{
				RTPStreamID: 0,
				SpatialID:   -1,
			}},
		}
		_, err := vla.Marshal()
		assert.ErrorIs(t, err, ErrVLAInvalidSpatialID)
		vla = &VLA{
			RTPStreamID:    0,
			RTPStreamCount: 1,
			ActiveSpatialLayer: []SpatialLayer{{
				RTPStreamID: 0,
				SpatialID:   5,
			}},
		}
		_, err = vla.Marshal()
		assert.ErrorIs(t, err, ErrVLAInvalidSpatialID)
	})

	t.Run("Invalid temporal layer in the spatial layer", func(t *testing.T) {
		vla := &VLA{
			RTPStreamID:    0,
			RTPStreamCount: 1,
			ActiveSpatialLayer: []SpatialLayer{{
				RTPStreamID:    0,
				SpatialID:      0,
				TargetBitrates: []int{},
			}},
		}
		_, err := vla.Marshal()
		assert.ErrorIs(t, err, ErrVLAInvalidTemporalLayer)
		vla = &VLA{
			RTPStreamID:    0,
			RTPStreamCount: 1,
			ActiveSpatialLayer: []SpatialLayer{{
				RTPStreamID:    0,
				SpatialID:      0,
				TargetBitrates: []int{100, 200, 300, 400, 500},
			}},
		}
		_, err = vla.Marshal()
		assert.ErrorIs(t, err, ErrVLAInvalidTemporalLayer)
	})

	t.Run("Duplicate spatial ID in the spatial layer", func(t *testing.T) {
		vla := &VLA{
			RTPStreamID:    0,
			RTPStreamCount: 1,
			ActiveSpatialLayer: []SpatialLayer{{
				RTPStreamID:    0,
				SpatialID:      0,
				TargetBitrates: []int{100},
			}, {
				RTPStreamID:    0,
				SpatialID:      0,
				TargetBitrates: []int{200},
			}},
		}
		_, err := vla.Marshal()
		assert.ErrorIs(t, err, ErrVLADuplicateSpatialID)
	})
}

func TestVLAUnmarshal(t *testing.T) {
	t.Run("3 streams no resolution and framerate", func(t *testing.T) {
		// two layer ("low", "high")
		b, err := hex.DecodeString("21149601f0019003d005b009")
		assert.NoError(t, err)

		vla := &VLA{}
		n, err := vla.Unmarshal(b)
		assert.NoError(t, err)
		assert.Equal(t, len(b), n)

		assert.Equal(t, 0, vla.RTPStreamID)
		assert.Equal(t, 3, vla.RTPStreamCount)
		assert.Equal(t, 3, len(vla.ActiveSpatialLayer))

		assert.Equal(t, 0, vla.ActiveSpatialLayer[0].RTPStreamID)
		assert.Equal(t, 0, vla.ActiveSpatialLayer[0].SpatialID)
		assert.Equal(t, 1, len(vla.ActiveSpatialLayer[0].TargetBitrates))
		assert.Equal(t, 150, vla.ActiveSpatialLayer[0].TargetBitrates[0])

		assert.Equal(t, 1, vla.ActiveSpatialLayer[1].RTPStreamID)
		assert.Equal(t, 0, vla.ActiveSpatialLayer[1].SpatialID)
		assert.Equal(t, 2, len(vla.ActiveSpatialLayer[1].TargetBitrates))
		assert.Equal(t, 240, vla.ActiveSpatialLayer[1].TargetBitrates[0])
		assert.Equal(t, 400, vla.ActiveSpatialLayer[1].TargetBitrates[1])

		assert.False(t, vla.HasResolutionAndFramerate)

		assert.Equal(t, 2, vla.ActiveSpatialLayer[2].RTPStreamID)
		assert.Equal(t, 0, vla.ActiveSpatialLayer[2].SpatialID)
		assert.Equal(t, 2, len(vla.ActiveSpatialLayer[2].TargetBitrates))
		assert.Equal(t, 720, vla.ActiveSpatialLayer[2].TargetBitrates[0])
		assert.Equal(t, 1200, vla.ActiveSpatialLayer[2].TargetBitrates[1])
	})

	t.Run("3 streams with resolution and framerate", func(t *testing.T) {
		b, err := hex.DecodeString("a1149601f0019003d005b009013f00b31e027f01671e04ff02cf1e")
		assert.NoError(t, err)

		vla := &VLA{}
		n, err := vla.Unmarshal(b)
		assert.NoError(t, err)
		assert.Equal(t, len(b), n)

		assert.Equal(t, 2, vla.RTPStreamID)
		assert.Equal(t, 3, vla.RTPStreamCount)

		assert.Equal(t, 0, vla.ActiveSpatialLayer[0].RTPStreamID)
		assert.Equal(t, 0, vla.ActiveSpatialLayer[0].SpatialID)
		assert.Equal(t, 1, len(vla.ActiveSpatialLayer[0].TargetBitrates))
		assert.Equal(t, 150, vla.ActiveSpatialLayer[0].TargetBitrates[0])

		assert.Equal(t, 1, vla.ActiveSpatialLayer[1].RTPStreamID)
		assert.Equal(t, 0, vla.ActiveSpatialLayer[1].SpatialID)
		assert.Equal(t, 2, len(vla.ActiveSpatialLayer[1].TargetBitrates))
		assert.Equal(t, 240, vla.ActiveSpatialLayer[1].TargetBitrates[0])
		assert.Equal(t, 400, vla.ActiveSpatialLayer[1].TargetBitrates[1])

		assert.Equal(t, 2, vla.ActiveSpatialLayer[2].RTPStreamID)
		assert.Equal(t, 0, vla.ActiveSpatialLayer[2].SpatialID)
		assert.Equal(t, 2, len(vla.ActiveSpatialLayer[2].TargetBitrates))
		assert.Equal(t, 720, vla.ActiveSpatialLayer[2].TargetBitrates[0])
		assert.Equal(t, 1200, vla.ActiveSpatialLayer[2].TargetBitrates[1])

		assert.True(t, vla.HasResolutionAndFramerate)

		assert.Equal(t, 320, vla.ActiveSpatialLayer[0].Width)
		assert.Equal(t, 180, vla.ActiveSpatialLayer[0].Height)
		assert.Equal(t, 30, vla.ActiveSpatialLayer[0].Framerate)
		assert.Equal(t, 640, vla.ActiveSpatialLayer[1].Width)
		assert.Equal(t, 360, vla.ActiveSpatialLayer[1].Height)
		assert.Equal(t, 30, vla.ActiveSpatialLayer[1].Framerate)
		assert.Equal(t, 1280, vla.ActiveSpatialLayer[2].Width)
		assert.Equal(t, 720, vla.ActiveSpatialLayer[2].Height)
		assert.Equal(t, 30, vla.ActiveSpatialLayer[2].Framerate)
	})

	t.Run("2 streams", func(t *testing.T) {
		// two layer ("low", "high")
		b, err := hex.DecodeString("1110c801d005b009")
		assert.NoError(t, err)

		vla := &VLA{}
		n, err := vla.Unmarshal(b)
		assert.NoError(t, err)
		assert.Equal(t, len(b), n)

		assert.Equal(t, 0, vla.RTPStreamID)
		assert.Equal(t, 2, vla.RTPStreamCount)
		assert.Equal(t, 2, len(vla.ActiveSpatialLayer))

		assert.Equal(t, 0, vla.ActiveSpatialLayer[0].RTPStreamID)
		assert.Equal(t, 0, vla.ActiveSpatialLayer[0].SpatialID)
		assert.Equal(t, 1, len(vla.ActiveSpatialLayer[0].TargetBitrates))
		assert.Equal(t, 200, vla.ActiveSpatialLayer[0].TargetBitrates[0])

		assert.Equal(t, 1, vla.ActiveSpatialLayer[1].RTPStreamID)
		assert.Equal(t, 0, vla.ActiveSpatialLayer[1].SpatialID)
		assert.Equal(t, 2, len(vla.ActiveSpatialLayer[1].TargetBitrates))
		assert.Equal(t, 720, vla.ActiveSpatialLayer[1].TargetBitrates[0])
		assert.Equal(t, 1200, vla.ActiveSpatialLayer[1].TargetBitrates[1])

		assert.False(t, vla.HasResolutionAndFramerate)
	})

	t.Run("3 streams mid paused with resolution and framerate", func(t *testing.T) {
		b, err := hex.DecodeString("601010109601d005b009013f00b31e04ff02cf1e")
		assert.NoError(t, err)

		vla := &VLA{}
		n, err := vla.Unmarshal(b)
		assert.NoError(t, err)
		assert.Equal(t, len(b), n)

		assert.Equal(t, 1, vla.RTPStreamID)
		assert.Equal(t, 3, vla.RTPStreamCount)

		assert.Equal(t, 0, vla.ActiveSpatialLayer[0].RTPStreamID)
		assert.Equal(t, 0, vla.ActiveSpatialLayer[0].SpatialID)
		assert.Equal(t, 1, len(vla.ActiveSpatialLayer[0].TargetBitrates))
		assert.Equal(t, 150, vla.ActiveSpatialLayer[0].TargetBitrates[0])

		assert.Equal(t, 2, vla.ActiveSpatialLayer[1].RTPStreamID)
		assert.Equal(t, 0, vla.ActiveSpatialLayer[1].SpatialID)
		assert.Equal(t, 2, len(vla.ActiveSpatialLayer[1].TargetBitrates))
		assert.Equal(t, 720, vla.ActiveSpatialLayer[1].TargetBitrates[0])
		assert.Equal(t, 1200, vla.ActiveSpatialLayer[1].TargetBitrates[1])

		assert.True(t, vla.HasResolutionAndFramerate)

		assert.Equal(t, 320, vla.ActiveSpatialLayer[0].Width)
		assert.Equal(t, 180, vla.ActiveSpatialLayer[0].Height)
		assert.Equal(t, 30, vla.ActiveSpatialLayer[0].Framerate)
		assert.Equal(t, 1280, vla.ActiveSpatialLayer[1].Width)
		assert.Equal(t, 720, vla.ActiveSpatialLayer[1].Height)
		assert.Equal(t, 30, vla.ActiveSpatialLayer[1].Framerate)
	})

	t.Run("extra 1", func(t *testing.T) {
		b, err := hex.DecodeString("a0001040ac02f403")
		assert.NoError(t, err)

		vla := &VLA{}
		n, err := vla.Unmarshal(b)
		assert.NoError(t, err)
		assert.Equal(t, len(b), n)
	})

	t.Run("extra 2", func(t *testing.T) {
		b, err := hex.DecodeString("a00010409405cc08")
		assert.NoError(t, err)

		vla := &VLA{}
		n, err := vla.Unmarshal(b)
		assert.NoError(t, err)
		assert.Equal(t, len(b), n)
	})
}

func TestVLAMarshalThenUnmarshal(t *testing.T) {
	t.Run("multiple spatial layers", func(t *testing.T) {
		var spatialLayers []SpatialLayer
		for streamID := 0; streamID < 3; streamID++ {
			for spatialID := 0; spatialID < 4; spatialID++ {
				spatialLayers = append(spatialLayers, SpatialLayer{
					RTPStreamID:    streamID,
					SpatialID:      spatialID,
					TargetBitrates: []int{150, 200},
					Width:          320,
					Height:         180,
					Framerate:      30,
				})
			}
		}

		vla0 := &VLA{
			RTPStreamID:               2,
			RTPStreamCount:            3,
			ActiveSpatialLayer:        spatialLayers,
			HasResolutionAndFramerate: true,
		}

		b, err := vla0.Marshal()
		assert.NoError(t, err)

		vla1 := &VLA{}
		n, err := vla1.Unmarshal(b)
		assert.NoError(t, err)
		assert.Equal(t, len(b), n)
		assert.Equal(t, vla0, vla1)
	})

	t.Run("different spatial layer bitmasks", func(t *testing.T) {
		var spatialLayers []SpatialLayer
		for streamID := 0; streamID < 4; streamID++ {
			for spatialID := 0; spatialID < streamID+1; spatialID++ {
				spatialLayers = append(spatialLayers, SpatialLayer{
					RTPStreamID:    streamID,
					SpatialID:      spatialID,
					TargetBitrates: []int{150, 200},
					Width:          320,
					Height:         180,
					Framerate:      30,
				})
			}
		}

		vla0 := &VLA{
			RTPStreamID:               0,
			RTPStreamCount:            4,
			ActiveSpatialLayer:        spatialLayers,
			HasResolutionAndFramerate: true,
		}

		b, err := vla0.Marshal()
		assert.NoError(t, err)
		assert.Equal(t, byte(0x00), b[0]&0x0f, "expects sl_bm to be 0")
		assert.Equal(t, byte(0x13), b[1], "expects sl0_bm,sl1_bm to be b0001,b0011")
		assert.Equal(t, byte(0x7f), b[2], "expects sl1_bm,sl2_bm to be b0111,b1111")
		t.Logf("b: %s", hex.EncodeToString(b))

		vla1 := &VLA{}
		n, err := vla1.Unmarshal(b)
		assert.NoError(t, err)
		assert.Equal(t, len(b), n)

		assert.Equal(t, vla0, vla1)
	})
}

func FuzzVLAUnmarshal(f *testing.F) {
	f.Add([]byte{0})
	f.Add([]byte("70"))

	f.Fuzz(func(t *testing.T, data []byte) {
		vla := &VLA{}
		_, err := vla.Unmarshal(data)
		if err != nil {
			t.Skip() // If the function returns an error, we skip the test case
		}
	})
}

func TestVLAMarshalTo(t *testing.T) {
	vla := &VLA{
		RTPStreamID:    0,
		RTPStreamCount: 3,
		ActiveSpatialLayer: []SpatialLayer{
			{RTPStreamID: 0, SpatialID: 0, TargetBitrates: []int{150}},
			{RTPStreamID: 1, SpatialID: 0, TargetBitrates: []int{240, 400}},
			{RTPStreamID: 2, SpatialID: 0, TargetBitrates: []int{720, 1200}},
		},
	}

	size, err := vla.MarshalSize()
	assert.NoError(t, err)

	buf := make([]byte, size)
	n, err := vla.MarshalTo(buf)
	assert.NoError(t, err)
	assert.Equal(t, size, n)

	expected, _ := vla.Marshal()
	assert.Equal(t, expected, buf)

	_, err = vla.MarshalTo(nil)
	assert.ErrorIs(t, err, io.ErrShortBuffer)
}

//nolint:gochecknoglobals
var (
	vlaSink    []byte
	vlaBuf     = make([]byte, 256)
	vlaSinkInt int
)

func BenchmarkVLA_Marshal(b *testing.B) {
	vla := &VLA{
		RTPStreamID:    0,
		RTPStreamCount: 3,
		ActiveSpatialLayer: []SpatialLayer{
			{RTPStreamID: 0, SpatialID: 0, TargetBitrates: []int{150}},
			{RTPStreamID: 1, SpatialID: 0, TargetBitrates: []int{240, 400}},
			{RTPStreamID: 2, SpatialID: 0, TargetBitrates: []int{720, 1200}},
		},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vlaSink, _ = vla.Marshal()
	}
}

func BenchmarkVLA_MarshalTo(b *testing.B) {
	vla := &VLA{
		RTPStreamID:    0,
		RTPStreamCount: 3,
		ActiveSpatialLayer: []SpatialLayer{
			{RTPStreamID: 0, SpatialID: 0, TargetBitrates: []int{150}},
			{RTPStreamID: 1, SpatialID: 0, TargetBitrates: []int{240, 400}},
			{RTPStreamID: 2, SpatialID: 0, TargetBitrates: []int{720, 1200}},
		},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vlaSinkInt, _ = vla.MarshalTo(vlaBuf)
	}
}
