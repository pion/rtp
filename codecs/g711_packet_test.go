// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package codecs //nolint:dupl

import (
	"bytes"
	"crypto/rand"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestG711Payloader(t *testing.T) {
	payloader := G711Payloader{}

	const (
		testlen = 10000
		testmtu = 1500
	)

	// generate random 8-bit g722 samples
	samples := make([]byte, testlen)
	_, err := rand.Read(samples)
	assert.NoError(t, err)

	// make a copy, for payloader input
	samplesIn := make([]byte, testlen)
	copy(samplesIn, samples)

	// split our samples into payloads
	payloads := payloader.Payload(testmtu, samplesIn)

	outcnt := int(math.Ceil(float64(testlen) / testmtu))
	assert.Len(t, payloads, outcnt)
	assert.Equal(t, samplesIn, samples, "Modified input samples")

	samplesOut := bytes.Join(payloads, []byte{})
	assert.Equal(t, samplesIn, samplesOut)

	payload := []byte{0x90, 0x90, 0x90}

	// 0 MTU, small payload
	res := payloader.Payload(0, payload)
	assert.Len(t, res, 0, "Generated payload should be empty")

	// Positive MTU, small payload
	res = payloader.Payload(1, payload)
	assert.Len(t, res, len(payload), "Generated payload should be the same size as original payload size")

	// Positive MTU, small payload
	res = payloader.Payload(uint16(len(payload)-1), payload) // nolint: gosec // G115
	assert.Len(t, res, len(payload)-1, "Generated payload should be the same smaller than original payload size")

	// Positive MTU, small payload
	res = payloader.Payload(10, payload)
	assert.Len(t, res, 1, "Generated payload should be the 1")
}
