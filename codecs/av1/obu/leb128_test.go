// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package obu

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"
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

func TestWriteToLeb128(t *testing.T) {
	type testVector struct {
		value  uint
		leb128 string
	}
	testVectors := []testVector{
		{150, "9601"},
		{240, "f001"},
		{400, "9003"},
		{720, "d005"},
		{1200, "b009"},
		{999999, "bf843d"},
		{0, "00"},
		{math.MaxUint32, "ffffffff0f"},
	}

	runTest := func(t *testing.T, v testVector) {
		t.Helper()

		b := WriteToLeb128(v.value)
		if v.leb128 != hex.EncodeToString(b) {
			t.Errorf("Expected %s, got %s", v.leb128, hex.EncodeToString(b))
		}
	}

	for _, v := range testVectors {
		t.Run(fmt.Sprintf("encode %d", v.value), func(t *testing.T) {
			runTest(t, v)
		})
	}
}
