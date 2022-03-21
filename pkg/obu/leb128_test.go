package obu

import (
	"errors"
	"testing"
)

func TestLEB128(t *testing.T) {
	for _, test := range []struct {
		Value   uint
		Encoded uint
	}{
		{0, 0},
		{5, 5},
		{999999, 0xBF843D},
	} {
		test := test

		encoded := EncodeLEB128(test.Value)
		if encoded != test.Encoded {
			t.Fatalf("Actual(%d) did not equal expected(%d)", encoded, test.Encoded)
		}

		decoded := decodeLEB128(encoded)
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
