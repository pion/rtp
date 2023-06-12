package rtp

import (
	"bytes"
	"errors"
	"testing"
)

func TestTransportCCExtensionTooSmall(t *testing.T) {
	t1 := TransportCCExtension{}

	rawData := []byte{}

	if err := t1.Unmarshal(rawData); !errors.Is(err, errTooSmall) {
		t.Fatal("err != errTooSmall")
	}
}

func TestTransportCCExtension(t *testing.T) {
	t1 := TransportCCExtension{}

	rawData := []byte{
		0x00, 0x02,
	}

	if err := t1.Unmarshal(rawData); err != nil {
		t.Fatal("Unmarshal error on extension data")
	}

	t2 := TransportCCExtension{
		TransportSequence: 2,
	}

	if t1 != t2 {
		t.Error("Unmarshal failed")
	}

	dstData, _ := t2.Marshal()
	if !bytes.Equal(dstData, rawData) {
		t.Error("Marshal failed")
	}
}

func TestTransportCCExtensionExtraBytes(t *testing.T) {
	t1 := TransportCCExtension{}

	rawData := []byte{
		0x00, 0x02, 0x00, 0xff, 0xff,
	}

	if err := t1.Unmarshal(rawData); err != nil {
		t.Fatal("Unmarshal error on extension data")
	}

	t2 := TransportCCExtension{
		TransportSequence: 2,
	}

	if t1 != t2 {
		t.Error("Unmarshal failed")
	}
}
