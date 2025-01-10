// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package rtp

import (
	"testing"
	"time"
)

func TestAbsCaptureTimeExtension_Roundtrip(t *testing.T) { // nolint: funlen,cyclop
	t.Run("positive captureClockOffset", func(t *testing.T) {
		t0 := time.Now()
		e1 := NewAbsCaptureTimeExtension(t0)
		b1, err1 := e1.Marshal()
		if err1 != nil {
			t.Fatal(err1)
		}
		var o1 AbsCaptureTimeExtension
		if err := o1.Unmarshal(b1); err != nil {
			t.Fatal(err)
		}
		dt1 := o1.CaptureTime().Sub(t0).Seconds()
		if dt1 < -0.001 || dt1 > 0.001 {
			t.Fatalf("timestamp differs, want %v got %v (dt=%f)", t0, o1.CaptureTime(), dt1)
		}
		if o1.EstimatedCaptureClockOffsetDuration() != nil {
			t.Fatalf("duration differs, want nil got %d", o1.EstimatedCaptureClockOffsetDuration())
		}

		e2 := NewAbsCaptureTimeExtensionWithCaptureClockOffset(t0, 1250*time.Millisecond)
		b2, err2 := e2.Marshal()
		if err2 != nil {
			t.Fatal(err2)
		}
		var o2 AbsCaptureTimeExtension
		if err := o2.Unmarshal(b2); err != nil {
			t.Fatal(err)
		}
		dt2 := o1.CaptureTime().Sub(t0).Seconds()
		if dt2 < -0.001 || dt2 > 0.001 {
			t.Fatalf("timestamp differs, want %v got %v (dt=%f)", t0, o2.CaptureTime(), dt2)
		}
		if *o2.EstimatedCaptureClockOffsetDuration() != 1250*time.Millisecond {
			t.Fatalf("duration differs, want 250ms got %d", *o2.EstimatedCaptureClockOffsetDuration())
		}
	})

	// This test can verify the for for the issue 247
	t.Run("negative captureClockOffset", func(t *testing.T) {
		t0 := time.Now()
		e1 := NewAbsCaptureTimeExtension(t0)
		b1, err1 := e1.Marshal()
		if err1 != nil {
			t.Fatal(err1)
		}
		var o1 AbsCaptureTimeExtension
		if err := o1.Unmarshal(b1); err != nil {
			t.Fatal(err)
		}
		dt1 := o1.CaptureTime().Sub(t0).Seconds()
		if dt1 < -0.001 || dt1 > 0.001 {
			t.Fatalf("timestamp differs, want %v got %v (dt=%f)", t0, o1.CaptureTime(), dt1)
		}
		if o1.EstimatedCaptureClockOffsetDuration() != nil {
			t.Fatalf("duration differs, want nil got %d", o1.EstimatedCaptureClockOffsetDuration())
		}

		e2 := NewAbsCaptureTimeExtensionWithCaptureClockOffset(t0, -250*time.Millisecond)
		b2, err2 := e2.Marshal()
		if err2 != nil {
			t.Fatal(err2)
		}
		var o2 AbsCaptureTimeExtension
		if err := o2.Unmarshal(b2); err != nil {
			t.Fatal(err)
		}
		dt2 := o1.CaptureTime().Sub(t0).Seconds()
		if dt2 < -0.001 || dt2 > 0.001 {
			t.Fatalf("timestamp differs, want %v got %v (dt=%f)", t0, o2.CaptureTime(), dt2)
		}
		if *o2.EstimatedCaptureClockOffsetDuration() != -250*time.Millisecond {
			t.Fatalf("duration differs, want -250ms got %v", *o2.EstimatedCaptureClockOffsetDuration())
		}
	})
}
