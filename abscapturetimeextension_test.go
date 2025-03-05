// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package rtp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAbsCaptureTimeExtension_Roundtrip(t *testing.T) { //nolint:cyclop
	t.Run("positive captureClockOffset", func(t *testing.T) {
		t0 := time.Now()
		e1 := NewAbsCaptureTimeExtension(t0)
		b1, err := e1.Marshal()
		assert.NoError(t, err)
		var o1 AbsCaptureTimeExtension
		assert.NoError(t, o1.Unmarshal(b1))
		dt1 := o1.CaptureTime().Sub(t0).Seconds()
		assert.GreaterOrEqual(t, dt1, -0.001)
		assert.LessOrEqual(t, dt1, 0.001)
		assert.Nil(t, o1.EstimatedCaptureClockOffsetDuration())

		e2 := NewAbsCaptureTimeExtensionWithCaptureClockOffset(t0, 1250*time.Millisecond)
		b2, err := e2.Marshal()
		assert.NoError(t, err)
		var o2 AbsCaptureTimeExtension
		assert.NoError(t, o2.Unmarshal(b2))
		dt2 := o1.CaptureTime().Sub(t0).Seconds()
		assert.GreaterOrEqual(t, dt2, -0.001)
		assert.LessOrEqual(t, dt2, 0.001)
		assert.Equal(t, 1250*time.Millisecond, *o2.EstimatedCaptureClockOffsetDuration())
	})

	// This test can verify the for for the issue 247
	t.Run("negative captureClockOffset", func(t *testing.T) {
		t0 := time.Now()
		e1 := NewAbsCaptureTimeExtension(t0)
		b1, err := e1.Marshal()
		assert.NoError(t, err)
		var o1 AbsCaptureTimeExtension
		assert.NoError(t, o1.Unmarshal(b1))
		dt1 := o1.CaptureTime().Sub(t0).Seconds()
		assert.GreaterOrEqual(t, dt1, -0.001)
		assert.LessOrEqual(t, dt1, 0.001)
		assert.Nil(t, o1.EstimatedCaptureClockOffsetDuration())

		e2 := NewAbsCaptureTimeExtensionWithCaptureClockOffset(t0, -250*time.Millisecond)
		b2, err := e2.Marshal()
		assert.NoError(t, err)

		var o2 AbsCaptureTimeExtension
		assert.NoError(t, o2.Unmarshal(b2))
		dt2 := o1.CaptureTime().Sub(t0).Seconds()
		assert.GreaterOrEqual(t, dt2, -0.001)
		assert.LessOrEqual(t, dt2, 0.001)
		assert.Equal(t, -250*time.Millisecond, *o2.EstimatedCaptureClockOffsetDuration())
	})
}
