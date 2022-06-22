// Package obu implements tools for working with the "Open Bitstream Unit"
package obu

import "errors"

const (
	sevenLsbBitmask = uint(0b01111111)
	msbBitmask      = uint(0b10000000)
)

// ErrFailedToReadLEB128 indicates that a buffer ended before a LEB128 value could be successfully read
var ErrFailedToReadLEB128 = errors.New("payload ended before LEB128 was finished")

// AppendUleb128 appends v to b using unsigned LEB128 encoding.
func AppendUleb128(b []byte, v uint) []byte {
	// If it's less than or equal to 7-bit
	if v < 0x80 {
		return append(b, byte(v))
	}

	for {
		c := uint8(v & 0x7f)
		v >>= 7

		if v != 0 {
			c |= 0x80
		}

		b = append(b, c)

		if c&0x80 == 0 {
			break
		}
	}

	return b
}

func EncodeLEB128(in uint) []byte {
	return AppendUleb128(make([]byte, 0), in)
}

func decodeLEB128(in uint) (out uint) {
	for {
		// Take 7 LSB from in
		out |= (in & sevenLsbBitmask)

		// Discard the MSB
		in >>= 8
		if in == 0 {
			return out
		}

		out <<= 7
	}
}

// ReadLeb128 scans an buffer and decodes a Leb128 value.
// If the end of the buffer is reached and all MSB are set
// an error is returned
func ReadLeb128(in []byte) (uint, uint, error) {
	var encodedLength uint

	for i := range in {
		encodedLength |= uint(in[i])

		if in[i]&byte(msbBitmask) == 0 {
			return decodeLEB128(encodedLength), uint(i + 1), nil
		}

		// Make more room for next read
		encodedLength <<= 8
	}

	return 0, 0, ErrFailedToReadLEB128
}
