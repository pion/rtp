// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package rtp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransportCCExtensionTooSmall(t *testing.T) {
	t1 := TransportCCExtension{}

	rawData := []byte{}

	err := t1.Unmarshal(rawData)
	assert.ErrorIs(t, err, errTooSmall)
}

func TestTransportCCExtension(t *testing.T) {
	t1 := TransportCCExtension{}

	rawData := []byte{
		0x00, 0x02,
	}

	err := t1.Unmarshal(rawData)
	assert.NoError(t, err)

	t2 := TransportCCExtension{
		TransportSequence: 2,
	}

	assert.Equal(t, t1, t2)

	dstData, _ := t2.Marshal()
	assert.Equal(t, dstData, rawData)
}

func TestTransportCCExtensionExtraBytes(t *testing.T) {
	t1 := TransportCCExtension{}

	rawData := []byte{
		0x00, 0x02, 0x00, 0xff, 0xff,
	}

	err := t1.Unmarshal(rawData)
	assert.NoError(t, err)

	t2 := TransportCCExtension{
		TransportSequence: 2,
	}

	assert.Equal(t, t1, t2)
}
