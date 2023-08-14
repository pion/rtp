// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

// Package obu is deprecated.
package obu

import (
	"github.com/pion/rtp/codecs/av1/obu"
)

// ErrFailedToReadLEB128 indicates that a buffer ended before a LEB128 value could be successfully read
//
// Deprecated: moved into codecs/av1/obu.
var ErrFailedToReadLEB128 = obu.ErrFailedToReadLEB128

// EncodeLEB128 encodes a uint as LEB128
//
// Deprecated: moved into codecs/av1/obu.
func EncodeLEB128(in uint) (out uint) {
	return obu.EncodeLEB128(in)
}

// ReadLeb128 scans an buffer and decodes a Leb128 value.
// If the end of the buffer is reached and all MSB are set
// an error is returned
//
// Deprecated: moved into codecs/av1/obu.
func ReadLeb128(in []byte) (uint, uint, error) {
	return obu.ReadLeb128(in)
}
