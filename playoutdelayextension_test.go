package rtp

import (
	"bytes"
	"errors"
	"testing"
)

func TestPlayoutDelayExtensionTooSmall(t *testing.T) {
	t1 := PlayoutDelayExtension{}

	var rawData []byte

	if err := t1.Unmarshal(rawData); !errors.Is(err, errTooSmall) {
		t.Fatal("err != errTooSmall")
	}
}

func TestPlayoutDelayExtensionTooLarge(t *testing.T) {
	t1 := PlayoutDelayExtension{minDelay: 1 << 12, maxDelay: 1 << 12}

	if _, err := t1.Marshal(); !errors.Is(err, errPlayoutDelayInvalidValue) {
		t.Fatal("err != errPlayoutDelayInvalidValue")
	}
}

func TestPlayoutDelayExtension(t *testing.T) {
	t1 := PlayoutDelayExtension{}

	rawData := []byte{
		0x01, 0x01, 0x00,
	}

	if err := t1.Unmarshal(rawData); err != nil {
		t.Fatal("Unmarshal error on extension data")
	}

	t2 := PlayoutDelayExtension{
		minDelay: 1 << 4, maxDelay: 1 << 8,
	}

	if t1 != t2 {
		t.Error("Unmarshal failed")
	}

	dstData, _ := t2.Marshal()
	if !bytes.Equal(dstData, rawData) {
		t.Error("Marshal failed")
	}
}

func TestPlayoutDelayExtensionExtraBytes(t *testing.T) {
	t1 := PlayoutDelayExtension{}

	rawData := []byte{
		0x01, 0x01, 0x00, 0xff, 0xff,
	}

	if err := t1.Unmarshal(rawData); err != nil {
		t.Fatal("Unmarshal error on extension data")
	}

	t2 := PlayoutDelayExtension{
		minDelay: 1 << 4, maxDelay: 1 << 8,
	}

	if t1 != t2 {
		t.Error("Unmarshal failed")
	}
}
