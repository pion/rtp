package rtp

import (
	"bytes"
	"testing"
)

func TestAudioLevelExtensionTooSmall(t *testing.T) {
	a := AudioLevelExtension{}

	rawData := []byte{}

	if err := a.Unmarshal(rawData); err != errInvalidSize {
		t.Fatal("err != errInvalidSize")
	}
}

func TestAudioLevelExtensionTooBig(t *testing.T) {
	a := AudioLevelExtension{}

	rawData := []byte{
		0x00, 0x00, 0x00, 0x00, 0x00,
	}

	if err := a.Unmarshal(rawData); err != errInvalidSize {
		t.Fatal("err != errInvalidSize")
	}
}

func TestAudioLevelOneByteExtensionInvalidLength(t *testing.T) {
	a := AudioLevelExtension{}

	rawData := []byte{
		0x31, 0x88,
	}

	if err := a.Unmarshal(rawData); err != errInvalidExtensonLength {
		t.Fatal("err != errInvalidExtensonLength")
	}
}

func TestAudioLevelTwoByteExtensionInvalidLength(t *testing.T) {
	a := AudioLevelExtension{}

	rawData := []byte{
		0x30, 0x00, 0x00, 0x00,
	}

	if err := a.Unmarshal(rawData); err != errInvalidExtensonLength {
		t.Fatal("err != errInvalidExtensonLength")
	}
}

func TestAudioLevelOneByteExtensionVoiceTrue(t *testing.T) {
	a1 := AudioLevelExtension{}

	rawData := []byte{
		0x30, 0x88,
	}

	if err := a1.Unmarshal(rawData); err != nil {
		t.Fatal("Unmarshal error on extension data")
	}

	a2 := AudioLevelExtension{
		ID:    3,
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

func TestAudioLevelOneByteExtensionVoiceFalse(t *testing.T) {
	a1 := AudioLevelExtension{}

	rawData := []byte{
		0x30, 0x8,
	}

	if err := a1.Unmarshal(rawData); err != nil {
		t.Fatal("Unmarshal error on extension data")
	}

	a2 := AudioLevelExtension{
		ID:    3,
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

func TestAudioLevelOneByteExtensionLevelOverflow(t *testing.T) {
	a := AudioLevelExtension{
		ID:    3,
		Level: 128,
		Voice: false,
	}

	if _, err := a.Marshal(); err != errAudioLevelOverflow {
		t.Fatal("err != errAudioLevelOverflow")
	}
}

func TestAudioLevelTwoByteExtensionVoiceFalse(t *testing.T) {
	a1 := AudioLevelExtension{}

	oneByteRawData := []byte{
		0x30, 0x8,
	}

	twoByteRawData := []byte{
		0x3, 0x1, 0x8, 0x0,
	}

	if err := a1.Unmarshal(twoByteRawData); err != nil {
		t.Fatal("Unmarshal error on extension data")
	}

	a2 := AudioLevelExtension{
		ID:    3,
		Level: 8,
		Voice: false,
	}

	if a1 != a2 {
		t.Error("Unmarshal failed")
	}

	dstData, _ := a2.Marshal()
	if !bytes.Equal(dstData, oneByteRawData) {
		t.Error("Marshal failed")
	}
}
