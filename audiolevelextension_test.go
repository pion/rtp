package rtp

import (
	"bytes"
	"errors"
	"testing"
)

func TestAudioLevelExtensionTooSmall(t *testing.T) {
	a := AudioLevelExtension{}

	rawData := []byte{}

	if err := a.Unmarshal(rawData); !errors.Is(err, errTooSmall) {
		t.Fatal("err != errTooSmall")
	}
}

func TestAudioLevelExtensionVoiceTrue(t *testing.T) {
	a1 := AudioLevelExtension{}

	rawData := []byte{
		0x88,
	}

	if err := a1.Unmarshal(rawData); err != nil {
		t.Fatal("Unmarshal error on extension data")
	}

	a2 := AudioLevelExtension{
		Level: 8,
		Voice: true,
	}

	if a1 != a2 {
		t.Error("Unmarshal failed")
	}

	dstData, _ := a2.Marshal()
	if !bytes.Equal(dstData, rawData) {
		t.Error("Marshal failed")
	}
}

func TestAudioLevelExtensionVoiceFalse(t *testing.T) {
	a1 := AudioLevelExtension{}

	rawData := []byte{
		0x8,
	}

	if err := a1.Unmarshal(rawData); err != nil {
		t.Fatal("Unmarshal error on extension data")
	}

	a2 := AudioLevelExtension{
		Level: 8,
		Voice: false,
	}

	if a1 != a2 {
		t.Error("Unmarshal failed")
	}

	dstData, _ := a2.Marshal()
	if !bytes.Equal(dstData, rawData) {
		t.Error("Marshal failed")
	}
}

func TestAudioLevelExtensionLevelOverflow(t *testing.T) {
	a := AudioLevelExtension{
		Level: 128,
		Voice: false,
	}

	if _, err := a.Marshal(); !errors.Is(err, errAudioLevelOverflow) {
		t.Fatal("err != errAudioLevelOverflow")
	}
}
