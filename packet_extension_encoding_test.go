// SPDX-FileCopyrightText: 2026 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package rtp

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFirstExtensionWireEncoding(t *testing.T) {
	for _, scenario := range []struct {
		id      uint8
		size    int
		profile uint16
		invalid bool
	}{
		{0, 2, 0, true},
		{1, 0, ExtensionProfileTwoByte, false},
		{1, 1, ExtensionProfileOneByte, false},
		{14, 16, ExtensionProfileOneByte, false},
		{15, 2, ExtensionProfileTwoByte, false},
		{20, 2, ExtensionProfileTwoByte, false},
		{255, 2, ExtensionProfileTwoByte, false},
		{1, 17, ExtensionProfileTwoByte, false},
		{1, 255, ExtensionProfileTwoByte, false},
		{1, 256, 0, true},
	} {
		t.Run(fmt.Sprintf("id=%d/size=%d", scenario.id, scenario.size), func(t *testing.T) {
			header := Header{Version: 2, SSRC: 123, CSRC: []uint32{42}}
			before := header.Clone()
			payload := bytes.Repeat([]byte{0x42}, scenario.size)
			err := header.SetExtension(scenario.id, payload)
			if scenario.invalid {
				require.Error(t, err)
				require.Equal(t, before, header, "invalid input must leave the entire header unchanged")

				return
			}
			require.NoError(t, err)
			encoded, err := header.Marshal()
			require.NoError(t, err)
			var decoded Header
			_, err = decoded.Unmarshal(encoded)
			require.NoError(t, err)
			assert.Equal(t, scenario.profile, decoded.ExtensionProfile)
			assert.Equal(t, []uint8{scenario.id}, decoded.GetExtensionIDs())
			assert.Equal(t, payload, decoded.GetExtension(scenario.id))
			assert.Equal(t, header.SSRC, decoded.SSRC)
			assert.Equal(t, header.CSRC, decoded.CSRC)
		})
	}
}

func TestExplicitExtensionProfileWireEncoding(t *testing.T) {
	for _, scenario := range []struct {
		name           string
		id             uint8
		size           int
		profile        uint16
		initialProfile uint16
		initialID      uint8
		invalid        bool
	}{
		{"first two-byte short", 1, 2, ExtensionProfileTwoByte, 0, 0, false},
		{"first one-byte high ID", 20, 2, ExtensionProfileOneByte, 0, 0, true},
		{"first one-byte empty", 1, 0, ExtensionProfileOneByte, 0, 0, true},
		{"first raw", 0, 4, 0x1234, 0, 0, false},
		{"first raw invalid ID", 1, 4, 0x1234, 0, 0, true},
		{"upgrade existing", 20, 2, ExtensionProfileTwoByte, ExtensionProfileOneByte, 1, false},
		{"invalid downgrade", 2, 2, ExtensionProfileOneByte, ExtensionProfileTwoByte, 20, true},
		{"existing one-byte empty", 2, 0, ExtensionProfileOneByte, ExtensionProfileOneByte, 1, true},
		{"invalid raw to two-byte", 2, 2, ExtensionProfileTwoByte, 0x1234, 0, true},
		{"invalid two-byte to raw", 0, 4, 0x1234, ExtensionProfileTwoByte, 20, true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			header := Header{Version: 2, SSRC: 123}
			if scenario.initialProfile != 0 {
				header.Extension = true
				header.ExtensionProfile = scenario.initialProfile
				require.NoError(t, header.SetExtension(scenario.initialID, []byte{0xAA, 0xBB}))
			}
			before := header.Clone()
			payload := bytes.Repeat([]byte{0x42}, scenario.size)
			err := header.SetExtensionWithProfile(scenario.id, payload, scenario.profile)
			if scenario.invalid {
				require.Error(t, err)
				require.Equal(t, before, header)

				return
			}
			require.NoError(t, err)
			encoded, err := header.Marshal()
			require.NoError(t, err)
			var decoded Header
			_, err = decoded.Unmarshal(encoded)
			require.NoError(t, err)
			assert.Equal(t, scenario.profile, decoded.ExtensionProfile)
			assert.Equal(t, payload, decoded.GetExtension(scenario.id))
			if scenario.initialProfile != 0 {
				assert.Equal(t, before.GetExtension(scenario.initialID), decoded.GetExtension(scenario.initialID))
			}
		})
	}
}

func TestOneByteExtensionRejectsEmptyPayload(t *testing.T) {
	for _, existing := range []bool{false, true} {
		extension := &OneByteHeaderExtension{}
		if existing {
			require.NoError(t, extension.Set(1, []byte{0xAA}))
		}
		before := bytes.Clone(extension.payload)
		require.ErrorIs(t, extension.Set(1, nil), errRFC8285OneByteHeaderSize)
		require.Equal(t, before, extension.payload)
	}
}
