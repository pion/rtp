// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package rtp

import (
	"bytes"
	"encoding/hex"
	"errors"
	"reflect"
	"testing"
)

func TestVLAMarshal(t *testing.T) { // nolint: funlen,cyclop
	requireNoError := func(t *testing.T, err error) {
		t.Helper()

		if err != nil {
			t.Fatal(err)
		}
	}

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
		requireNoError(t, err)
		bytesExpected, err := hex.DecodeString("21149601f0019003d005b009")
		requireNoError(t, err)
		if !bytes.Equal(bytesExpected, bytesActual) {
			t.Fatalf("expected %s, actual %s", hex.EncodeToString(bytesExpected), hex.EncodeToString(bytesActual))
		}
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
		requireNoError(t, err)
		bytesExpected, err := hex.DecodeString("a1149601f0019003d005b009013f00b31e027f01671e04ff02cf1e")
		requireNoError(t, err)
		if !bytes.Equal(bytesExpected, bytesActual) {
			t.Fatalf("expected %s, actual %s", hex.EncodeToString(bytesExpected), hex.EncodeToString(bytesActual))
		}
	})

	t.Run("Negative RTPStreamCount", func(t *testing.T) {
		vla := &VLA{
			RTPStreamID:        0,
			RTPStreamCount:     -1,
			ActiveSpatialLayer: []SpatialLayer{},
		}
		_, err := vla.Marshal()
		if !errors.Is(err, ErrVLAInvalidStreamCount) {
			t.Fatal("expected ErrVLAInvalidRTPStreamCount")
		}
	})

	t.Run("RTPStreamCount too large", func(t *testing.T) {
		vla := &VLA{
			RTPStreamID:        0,
			RTPStreamCount:     5,
			ActiveSpatialLayer: []SpatialLayer{{}, {}, {}, {}, {}},
		}
		_, err := vla.Marshal()
		if !errors.Is(err, ErrVLAInvalidStreamCount) {
			t.Fatal("expected ErrVLAInvalidRTPStreamCount")
		}
	})

	t.Run("Negative RTPStreamID", func(t *testing.T) {
		vla := &VLA{
			RTPStreamID:        -1,
			RTPStreamCount:     1,
			ActiveSpatialLayer: []SpatialLayer{{}},
		}
		_, err := vla.Marshal()
		if !errors.Is(err, ErrVLAInvalidStreamID) {
			t.Fatalf("expected ErrVLAInvalidRTPStreamID, actual %v", err)
		}
	})

	t.Run("RTPStreamID to large", func(t *testing.T) {
		vla := &VLA{
			RTPStreamID:        1,
			RTPStreamCount:     1,
			ActiveSpatialLayer: []SpatialLayer{{}},
		}
		_, err := vla.Marshal()
		if !errors.Is(err, ErrVLAInvalidStreamID) {
			t.Fatalf("expected ErrVLAInvalidRTPStreamID: %v", err)
		}
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
		if !errors.Is(err, ErrVLAInvalidStreamID) {
			t.Fatalf("expected ErrVLAInvalidStreamID: %v", err)
		}
		vla = &VLA{
			RTPStreamID:    0,
			RTPStreamCount: 1,
			ActiveSpatialLayer: []SpatialLayer{{
				RTPStreamID: 1,
			}},
		}
		_, err = vla.Marshal()
		if !errors.Is(err, ErrVLAInvalidStreamID) {
			t.Fatalf("expected ErrVLAInvalidStreamID: %v", err)
		}
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
		if !errors.Is(err, ErrVLAInvalidSpatialID) {
			t.Fatalf("expected ErrVLAInvalidSpatialID: %v", err)
		}
		vla = &VLA{
			RTPStreamID:    0,
			RTPStreamCount: 1,
			ActiveSpatialLayer: []SpatialLayer{{
				RTPStreamID: 0,
				SpatialID:   5,
			}},
		}
		_, err = vla.Marshal()
		if !errors.Is(err, ErrVLAInvalidSpatialID) {
			t.Fatalf("expected ErrVLAInvalidSpatialID: %v", err)
		}
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
		if !errors.Is(err, ErrVLAInvalidTemporalLayer) {
			t.Fatalf("expected ErrVLAInvalidTemporalLayer: %v", err)
		}
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
		if !errors.Is(err, ErrVLAInvalidTemporalLayer) {
			t.Fatalf("expected ErrVLAInvalidTemporalLayer: %v", err)
		}
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
		if !errors.Is(err, ErrVLADuplicateSpatialID) {
			t.Fatalf("expected ErrVLADuplicateSpatialID: %v", err)
		}
	})
}

func TestVLAUnmarshal(t *testing.T) { // nolint: funlen
	requireEqualInt := func(t *testing.T, expected, actual int) {
		t.Helper()

		if expected != actual {
			t.Fatalf("expected %d, actual %d", expected, actual)
		}
	}
	requireNoError := func(t *testing.T, err error) {
		t.Helper()

		if err != nil {
			t.Fatal(err)
		}
	}
	requireTrue := func(t *testing.T, val bool) {
		t.Helper()

		if !val {
			t.Fatal("expected true")
		}
	}
	requireFalse := func(t *testing.T, val bool) {
		t.Helper()

		if val {
			t.Fatal("expected false")
		}
	}

	t.Run("3 streams no resolution and framerate", func(t *testing.T) {
		// two layer ("low", "high")
		b, err := hex.DecodeString("21149601f0019003d005b009")
		requireNoError(t, err)
		if err != nil {
			t.Fatal("failed to decode input data")
		}

		vla := &VLA{}
		n, err := vla.Unmarshal(b)
		requireNoError(t, err)
		requireEqualInt(t, len(b), n)

		requireEqualInt(t, 0, vla.RTPStreamID)
		requireEqualInt(t, 3, vla.RTPStreamCount)
		requireEqualInt(t, 3, len(vla.ActiveSpatialLayer))

		requireEqualInt(t, 0, vla.ActiveSpatialLayer[0].RTPStreamID)
		requireEqualInt(t, 0, vla.ActiveSpatialLayer[0].SpatialID)
		requireEqualInt(t, 1, len(vla.ActiveSpatialLayer[0].TargetBitrates))
		requireEqualInt(t, 150, vla.ActiveSpatialLayer[0].TargetBitrates[0])

		requireEqualInt(t, 1, vla.ActiveSpatialLayer[1].RTPStreamID)
		requireEqualInt(t, 0, vla.ActiveSpatialLayer[1].SpatialID)
		requireEqualInt(t, 2, len(vla.ActiveSpatialLayer[1].TargetBitrates))
		requireEqualInt(t, 240, vla.ActiveSpatialLayer[1].TargetBitrates[0])
		requireEqualInt(t, 400, vla.ActiveSpatialLayer[1].TargetBitrates[1])

		requireFalse(t, vla.HasResolutionAndFramerate)

		requireEqualInt(t, 2, vla.ActiveSpatialLayer[2].RTPStreamID)
		requireEqualInt(t, 0, vla.ActiveSpatialLayer[2].SpatialID)
		requireEqualInt(t, 2, len(vla.ActiveSpatialLayer[2].TargetBitrates))
		requireEqualInt(t, 720, vla.ActiveSpatialLayer[2].TargetBitrates[0])
		requireEqualInt(t, 1200, vla.ActiveSpatialLayer[2].TargetBitrates[1])
	})

	t.Run("3 streams with resolution and framerate", func(t *testing.T) {
		b, err := hex.DecodeString("a1149601f0019003d005b009013f00b31e027f01671e04ff02cf1e")
		requireNoError(t, err)

		vla := &VLA{}
		n, err := vla.Unmarshal(b)
		requireNoError(t, err)
		requireEqualInt(t, len(b), n)

		requireEqualInt(t, 2, vla.RTPStreamID)
		requireEqualInt(t, 3, vla.RTPStreamCount)

		requireEqualInt(t, 0, vla.ActiveSpatialLayer[0].RTPStreamID)
		requireEqualInt(t, 0, vla.ActiveSpatialLayer[0].SpatialID)
		requireEqualInt(t, 1, len(vla.ActiveSpatialLayer[0].TargetBitrates))
		requireEqualInt(t, 150, vla.ActiveSpatialLayer[0].TargetBitrates[0])

		requireEqualInt(t, 1, vla.ActiveSpatialLayer[1].RTPStreamID)
		requireEqualInt(t, 0, vla.ActiveSpatialLayer[1].SpatialID)
		requireEqualInt(t, 2, len(vla.ActiveSpatialLayer[1].TargetBitrates))
		requireEqualInt(t, 240, vla.ActiveSpatialLayer[1].TargetBitrates[0])
		requireEqualInt(t, 400, vla.ActiveSpatialLayer[1].TargetBitrates[1])

		requireEqualInt(t, 2, vla.ActiveSpatialLayer[2].RTPStreamID)
		requireEqualInt(t, 0, vla.ActiveSpatialLayer[2].SpatialID)
		requireEqualInt(t, 2, len(vla.ActiveSpatialLayer[2].TargetBitrates))
		requireEqualInt(t, 720, vla.ActiveSpatialLayer[2].TargetBitrates[0])
		requireEqualInt(t, 1200, vla.ActiveSpatialLayer[2].TargetBitrates[1])

		requireTrue(t, vla.HasResolutionAndFramerate)

		requireEqualInt(t, 320, vla.ActiveSpatialLayer[0].Width)
		requireEqualInt(t, 180, vla.ActiveSpatialLayer[0].Height)
		requireEqualInt(t, 30, vla.ActiveSpatialLayer[0].Framerate)
		requireEqualInt(t, 640, vla.ActiveSpatialLayer[1].Width)
		requireEqualInt(t, 360, vla.ActiveSpatialLayer[1].Height)
		requireEqualInt(t, 30, vla.ActiveSpatialLayer[1].Framerate)
		requireEqualInt(t, 1280, vla.ActiveSpatialLayer[2].Width)
		requireEqualInt(t, 720, vla.ActiveSpatialLayer[2].Height)
		requireEqualInt(t, 30, vla.ActiveSpatialLayer[2].Framerate)
	})

	t.Run("2 streams", func(t *testing.T) {
		// two layer ("low", "high")
		b, err := hex.DecodeString("1110c801d005b009")
		requireNoError(t, err)

		vla := &VLA{}
		n, err := vla.Unmarshal(b)
		requireNoError(t, err)
		requireEqualInt(t, len(b), n)

		requireEqualInt(t, 0, vla.RTPStreamID)
		requireEqualInt(t, 2, vla.RTPStreamCount)
		requireEqualInt(t, 2, len(vla.ActiveSpatialLayer))

		requireEqualInt(t, 0, vla.ActiveSpatialLayer[0].RTPStreamID)
		requireEqualInt(t, 0, vla.ActiveSpatialLayer[0].SpatialID)
		requireEqualInt(t, 1, len(vla.ActiveSpatialLayer[0].TargetBitrates))
		requireEqualInt(t, 200, vla.ActiveSpatialLayer[0].TargetBitrates[0])

		requireEqualInt(t, 1, vla.ActiveSpatialLayer[1].RTPStreamID)
		requireEqualInt(t, 0, vla.ActiveSpatialLayer[1].SpatialID)
		requireEqualInt(t, 2, len(vla.ActiveSpatialLayer[1].TargetBitrates))
		requireEqualInt(t, 720, vla.ActiveSpatialLayer[1].TargetBitrates[0])
		requireEqualInt(t, 1200, vla.ActiveSpatialLayer[1].TargetBitrates[1])

		requireFalse(t, vla.HasResolutionAndFramerate)
	})

	t.Run("3 streams mid paused with resolution and framerate", func(t *testing.T) {
		b, err := hex.DecodeString("601010109601d005b009013f00b31e04ff02cf1e")
		requireNoError(t, err)

		vla := &VLA{}
		n, err := vla.Unmarshal(b)
		requireNoError(t, err)
		requireEqualInt(t, len(b), n)

		requireEqualInt(t, 1, vla.RTPStreamID)
		requireEqualInt(t, 3, vla.RTPStreamCount)

		requireEqualInt(t, 0, vla.ActiveSpatialLayer[0].RTPStreamID)
		requireEqualInt(t, 0, vla.ActiveSpatialLayer[0].SpatialID)
		requireEqualInt(t, 1, len(vla.ActiveSpatialLayer[0].TargetBitrates))
		requireEqualInt(t, 150, vla.ActiveSpatialLayer[0].TargetBitrates[0])

		requireEqualInt(t, 2, vla.ActiveSpatialLayer[1].RTPStreamID)
		requireEqualInt(t, 0, vla.ActiveSpatialLayer[1].SpatialID)
		requireEqualInt(t, 2, len(vla.ActiveSpatialLayer[1].TargetBitrates))
		requireEqualInt(t, 720, vla.ActiveSpatialLayer[1].TargetBitrates[0])
		requireEqualInt(t, 1200, vla.ActiveSpatialLayer[1].TargetBitrates[1])

		requireTrue(t, vla.HasResolutionAndFramerate)

		requireEqualInt(t, 320, vla.ActiveSpatialLayer[0].Width)
		requireEqualInt(t, 180, vla.ActiveSpatialLayer[0].Height)
		requireEqualInt(t, 30, vla.ActiveSpatialLayer[0].Framerate)
		requireEqualInt(t, 1280, vla.ActiveSpatialLayer[1].Width)
		requireEqualInt(t, 720, vla.ActiveSpatialLayer[1].Height)
		requireEqualInt(t, 30, vla.ActiveSpatialLayer[1].Framerate)
	})

	t.Run("extra 1", func(t *testing.T) {
		b, err := hex.DecodeString("a0001040ac02f403")
		requireNoError(t, err)

		vla := &VLA{}
		n, err := vla.Unmarshal(b)
		requireNoError(t, err)
		requireEqualInt(t, len(b), n)
	})

	t.Run("extra 2", func(t *testing.T) {
		b, err := hex.DecodeString("a00010409405cc08")
		requireNoError(t, err)

		vla := &VLA{}
		n, err := vla.Unmarshal(b)
		requireNoError(t, err)
		requireEqualInt(t, len(b), n)
	})
}

func TestVLAMarshalThenUnmarshal(t *testing.T) { // nolint:funlen, cyclop
	requireEqualInt := func(t *testing.T, expected, actual int) {
		t.Helper()

		if expected != actual {
			t.Fatalf("expected %d, actual %d", expected, actual)
		}
	}
	requireNoError := func(t *testing.T, err error) {
		t.Helper()

		if err != nil {
			t.Fatal(err)
		}
	}

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
		requireNoError(t, err)

		vla1 := &VLA{}
		n, err := vla1.Unmarshal(b)
		requireNoError(t, err)
		requireEqualInt(t, len(b), n)

		if !reflect.DeepEqual(vla0, vla1) {
			t.Fatalf("expected %v, actual %v", vla0, vla1)
		}
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
		requireNoError(t, err)
		if b[0]&0x0f != 0 {
			t.Error("expects sl_bm to be 0")
		}
		if b[1] != 0x13 {
			t.Error("expects sl0_bm,sl1_bm to be b0001,b0011")
		}
		if b[2] != 0x7f {
			t.Error("expects sl1_bm,sl2_bm to be b0111,b1111")
		}
		t.Logf("b: %s", hex.EncodeToString(b))

		vla1 := &VLA{}
		n, err := vla1.Unmarshal(b)
		requireNoError(t, err)
		requireEqualInt(t, len(b), n)

		if !reflect.DeepEqual(vla0, vla1) {
			t.Fatalf("expected %v, actual %v", vla0, vla1)
		}
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
