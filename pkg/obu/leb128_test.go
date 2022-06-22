package obu

import (
	"encoding/hex"
	"errors"
	"testing"
)

func TestLEB128(t *testing.T) {
	for _, test := range []struct {
		Value   uint
		Encoded string
	}{
		{0, "00"},
		{5, "05"},
		{999999, "bf843d"},
	} {
		test := test

		encoded := EncodeLEB128(test.Value)
		encodedHex := hex.EncodeToString(encoded)
		if encodedHex != test.Encoded {
			t.Fatalf("Actual(%s) did not equal expected(%s)", encodedHex, test.Encoded)
		}

		decoded, _, _ := ReadLeb128(encoded)
		if decoded != test.Value {
			t.Fatalf("Actual(%d) did not equal expected(%d)", decoded, test.Value)
		}
	}
}

func TestReadLeb128(t *testing.T) {
	if _, _, err := ReadLeb128(nil); !errors.Is(err, ErrFailedToReadLEB128) {
		t.Fatal("ReadLeb128 on a nil buffer should return an error")
	}

	if _, _, err := ReadLeb128([]byte{0xFF}); !errors.Is(err, ErrFailedToReadLEB128) {
		t.Fatal("ReadLeb128 on a buffer with all MSB set should fail")
	}
}
