// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package codecs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVP8Packet_Unmarshal(t *testing.T) {
	pck := VP8Packet{}

	// Nil packet
	raw, err := pck.Unmarshal(nil)
	assert.ErrorIs(t, err, errNilPacket)
	assert.Nil(t, raw, "Result should be nil in case of error")

	// Nil payload
	raw, err = pck.Unmarshal([]byte{})
	assert.ErrorIs(t, err, errShortPacket)
	assert.Nil(t, raw, "Result should be nil in case of error")

	// Normal payload
	raw, err = pck.Unmarshal([]byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x90})
	assert.NoError(t, err)
	assert.NotNil(t, raw, "Result shouldn't be nil in case of success")

	// Header size, only X
	raw, err = pck.Unmarshal([]byte{0x80, 0x00, 0x00, 0x00})
	assert.NoError(t, err)
	assert.NotNil(t, raw, "Result shouldn't be nil in case of success")

	// Header size, X and I
	raw, err = pck.Unmarshal([]byte{0x80, 0x80, 0x00, 0x00})
	assert.NoError(t, err)
	assert.NotNil(t, raw, "Result shouldn't be nil in case of success")

	// Header size, X and I, PID 16bits
	raw, err = pck.Unmarshal([]byte{0x80, 0x80, 0x81, 0x00})
	assert.NoError(t, err)
	assert.NotNil(t, raw, "Result shouldn't be nil in case of success")

	// Header size, X and L
	raw, err = pck.Unmarshal([]byte{0x80, 0x40, 0x00, 0x00})
	assert.NoError(t, err)
	assert.NotNil(t, raw, "Result shouldn't be nil in case of success")

	// Header size, X and T
	raw, err = pck.Unmarshal([]byte{0x80, 0x20, 0x00, 0x00})
	assert.NoError(t, err)
	assert.NotNil(t, raw, "Result shouldn't be nil in case of success")

	// Header size, X and K
	raw, err = pck.Unmarshal([]byte{0x80, 0x10, 0x00, 0x00})
	assert.NoError(t, err)
	assert.NotNil(t, raw, "Result shouldn't be nil in case of success")

	// Header size, all flags
	raw, err = pck.Unmarshal([]byte{0xff, 0xff, 0x00, 0x00})
	assert.ErrorIs(t, err, errShortPacket)
	assert.Nil(t, raw, "Result should be nil in case of error")

	// According to RFC 7741 Section 4.4, the packetizer need not pay
	// attention to partition boundaries.  In that case, it may
	// produce packets with minimal headers.

	// The next three have been witnessed in nature.
	_, err = pck.Unmarshal([]byte{0x00})
	assert.NoError(t, err, "Empty packet with trivial header")
	_, err = pck.Unmarshal([]byte{0x00, 0x2a, 0x94})
	assert.NoError(t, err, "Non-empty packet with trivial header")

	raw, err = pck.Unmarshal([]byte{0x81, 0x81, 0x94})
	assert.ErrorIs(t, err, errShortPacket)
	assert.Nil(t, raw, "Result should be nil in case of error")

	// The following two were invented.
	_, err = pck.Unmarshal([]byte{0x80, 0x00})
	assert.NoError(t, err, "Empty packet with trivial extension")

	_, err = pck.Unmarshal([]byte{0x80, 0x80, 42})
	assert.NoError(t, err, "Header with PictureID")
}

func TestVP8Payloader_Payload(t *testing.T) {
	testCases := map[string]struct {
		payloader VP8Payloader
		mtu       uint16
		payload   [][]byte
		expected  [][][]byte
	}{
		"WithoutPictureID": {
			payloader: VP8Payloader{},
			mtu:       2,
			payload: [][]byte{
				{0x90, 0x90, 0x90},
				{0x91, 0x91},
			},
			expected: [][][]byte{
				{{0x10, 0x90}, {0x00, 0x90}, {0x00, 0x90}},
				{{0x10, 0x91}, {0x00, 0x91}},
			},
		},
		"WithPictureID_1byte": {
			payloader: VP8Payloader{
				EnablePictureID: true,
				pictureID:       0x20,
			},
			mtu: 5,
			payload: [][]byte{
				{0x90, 0x90, 0x90},
				{0x91, 0x91},
			},
			expected: [][][]byte{
				{
					{0x90, 0x80, 0x20, 0x90, 0x90},
					{0x80, 0x80, 0x20, 0x90},
				},
				{
					{0x90, 0x80, 0x21, 0x91, 0x91},
				},
			},
		},
		"WithPictureID_2bytes": {
			payloader: VP8Payloader{
				EnablePictureID: true,
				pictureID:       0x120,
			},
			mtu: 6,
			payload: [][]byte{
				{0x90, 0x90, 0x90},
				{0x91, 0x91},
			},
			expected: [][][]byte{
				{
					{0x90, 0x80, 0x81, 0x20, 0x90, 0x90},
					{0x80, 0x80, 0x81, 0x20, 0x90},
				},
				{
					{0x90, 0x80, 0x81, 0x21, 0x91, 0x91},
				},
			},
		},
	}
	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			pck := testCase.payloader

			for i := range testCase.payload {
				res := pck.Payload(testCase.mtu, testCase.payload[i])
				assert.Equal(t, testCase.expected[i], res, "Generated packet differs")
			}
		})
	}

	t.Run("Error", func(t *testing.T) {
		pck := VP8Payloader{}
		payload := []byte{0x90, 0x90, 0x90}

		// Positive MTU, nil payload
		res := pck.Payload(1, nil)
		assert.Len(t, res, 0, "Generated payload should be empty")

		// Positive MTU, small payload
		// MTU of 1 results in fragment size of 0
		res = pck.Payload(1, payload)
		assert.Len(t, res, 0, "Generated payload should be empty")
	})
}

func TestVP8IsPartitionHead(t *testing.T) {
	vp8 := &VP8Packet{}
	t.Run("SmallPacket", func(t *testing.T) {
		assert.False(t, vp8.IsPartitionHead([]byte{0x00}), "Small packet should not be the head of a new partition")
	})
	t.Run("SFlagON", func(t *testing.T) {
		assert.True(
			t, vp8.IsPartitionHead([]byte{0x10, 0x00, 0x00, 0x00}),
			"Packet with S flag should be the head of a new partition",
		)
	})
	t.Run("SFlagOFF", func(t *testing.T) {
		assert.False(
			t, vp8.IsPartitionHead([]byte{0x00, 0x00, 0x00, 0x00}),
			"Packet without S flag should not be the head of a new partition",
		)
	})
}
