// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package rtp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaderExtension_RFC8285OneByteExtension(t *testing.T) {
	p := &OneByteHeaderExtension{}

	rawPkt := []byte{
		0xBE, 0xDE, 0x00, 0x01, 0x50, 0xAA, 0x00, 0x00,
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	_, err := p.Unmarshal(rawPkt)
	assert.NoError(t, err, "Unmarshal err for valid extension")

	dstData, _ := p.Marshal()
	assert.Equal(t, rawPkt, dstData)
}

func TestHeaderExtension_RFC8285OneByteTwoExtensionOfTwoBytes(t *testing.T) {
	ext := &OneByteHeaderExtension{}

	//  0                   1                   2                   3
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |       0xBE    |    0xDE       |           length=1            |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |  ID   | L=0   |     data      |  ID   |  L=0  |   data...
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	rawPkt := []byte{
		0xBE, 0xDE, 0x00, 0x01, 0x10, 0xAA, 0x20, 0xBB,
	}
	_, err := ext.Unmarshal(rawPkt)
	assert.NoError(t, err, "Unmarshal err for valid extension")

	ext1 := ext.Get(1)
	ext1Expect := []byte{0xAA}
	assert.Equal(t, ext1Expect, ext1, "Extension has incorrect data")

	ext2 := ext.Get(2)
	ext2Expect := []byte{0xBB}
	assert.Equal(t, ext2Expect, ext2, "Extension has incorrect data")

	dstData, _ := ext.Marshal()
	assert.Equal(t, rawPkt, dstData)
}

func TestHeaderExtension_RFC8285OneByteMultipleExtensionsWithPadding(t *testing.T) {
	ext := &OneByteHeaderExtension{}

	//  0                   1                   2                   3
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |       0xBE    |    0xDE       |           length=3            |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |  ID   | L=0   |     data      |  ID   |  L=1  |   data...
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//       ...data   |    0 (pad)    |    0 (pad)    |  ID   | L=3   |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |                          data                                 |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	rawPkt := []byte{
		0xBE, 0xDE, 0x00, 0x03, 0x10, 0xAA, 0x21, 0xBB,
		0xBB, 0x00, 0x00, 0x33, 0xCC, 0xCC, 0xCC, 0xCC,
	}
	_, err := ext.Unmarshal(rawPkt)
	assert.NoError(t, err, "Unmarshal err for valid extension")

	ext1 := ext.Get(1)
	ext1Expect := []byte{0xAA}
	assert.Equal(t, ext1Expect, ext1, "Extension has incorrect data")

	ext2 := ext.Get(2)
	ext2Expect := []byte{0xBB, 0xBB}
	assert.Equal(t, ext2Expect, ext2, "Extension has incorrect data")

	ext3 := ext.Get(3)
	ext3Expect := []byte{0xCC, 0xCC, 0xCC, 0xCC}
	assert.Equal(t, ext3Expect, ext3, "Extension has incorrect data")

	dstBuf := map[string][]byte{
		"CleanBuffer": make([]byte, 1000),
		"DirtyBuffer": make([]byte, 1000),
	}
	for i := range dstBuf["DirtyBuffer"] {
		dstBuf["DirtyBuffer"][i] = 0xFF
	}
	for name, buf := range dstBuf {
		buf := buf
		t.Run(name, func(t *testing.T) {
			n, err := ext.MarshalTo(buf)
			assert.NoError(t, err)

			assert.Equal(t, rawPkt, buf[:n])
		})
	}
}

func TestHeaderExtension_RFC8285TwoByteExtension(t *testing.T) {
	ext := &TwoByteHeaderExtension{}

	rawPkt := []byte{
		0x10, 0x00, 0x00, 0x07, 0x05, 0x18, 0xAA, 0xAA,
		0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
		0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
		0xAA, 0xAA, 0x00, 0x00,
	}
	_, err := ext.Unmarshal(rawPkt)
	assert.NoError(t, err, "Unmarshal err for valid extension")

	dstData, _ := ext.Marshal()
	assert.Equal(t, rawPkt, dstData)
}

func TestHeaderExtension_RFC8285TwoByteMultipleExtensionsWithPadding(t *testing.T) {
	ext := &TwoByteHeaderExtension{}

	// 0                   1                   2                   3
	// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |       0x10    |    0x00       |           length=3            |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |      ID=1     |     L=0       |     ID=2      |     L=1       |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |       data    |    0 (pad)    |       ID=3    |      L=4      |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |                          data                                 |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	rawPkt := []byte{
		0x10, 0x00, 0x00, 0x03, 0x01, 0x00, 0x02, 0x01,
		0xBB, 0x00, 0x03, 0x04, 0xCC, 0xCC, 0xCC, 0xCC,
	}

	_, err := ext.Unmarshal(rawPkt)
	assert.NoError(t, err, "Unmarshal err for valid extension")

	ext1 := ext.Get(1)
	ext1Expect := []byte{}
	assert.Equal(t, ext1Expect, ext1, "Extension has incorrect data")

	ext2 := ext.Get(2)
	ext2Expect := []byte{0xBB}
	assert.Equal(t, ext2Expect, ext2, "Extension has incorrect data")

	ext3 := ext.Get(3)
	ext3Expect := []byte{0xCC, 0xCC, 0xCC, 0xCC}
	assert.Equal(t, ext3Expect, ext3, "Extension has incorrect data")
}

func TestHeaderExtension_RFC8285TwoByteMultipleExtensionsWithLargeExtension(t *testing.T) {
	ext := &TwoByteHeaderExtension{}

	// 0                   1                   2                   3
	// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |       0x10    |    0x00       |           length=3            |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |      ID=1     |     L=0       |     ID=2      |     L=1       |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |       data    |       ID=3    |      L=17      |    data...
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//                            ...data...
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//                            ...data...
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//                            ...data...
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//                            ...data...                           |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	rawPkt := []byte{
		0x10, 0x00, 0x00, 0x06, 0x01, 0x00, 0x02, 0x01,
		0xBB, 0x03, 0x11, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC,
		0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC,
	}

	_, err := ext.Unmarshal(rawPkt)
	assert.NoError(t, err, "Unmarshal err for valid extension")

	ext1 := ext.Get(1)
	ext1Expect := []byte{}
	assert.Equal(t, ext1Expect, ext1, "Extension has incorrect data")

	ext2 := ext.Get(2)
	ext2Expect := []byte{0xBB}
	assert.Equal(t, ext2Expect, ext2, "Extension has incorrect data")

	ext3 := ext.Get(3)
	ext3Expect := []byte{
		0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC,
		0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC,
	}
	assert.Equal(t, ext3Expect, ext3, "Extension has incorrect data")

	dstData, _ := ext.Marshal()
	assert.Equal(t, rawPkt, dstData)
}

func TestHeaderExtension_RFC8285OneByteDelExtension(t *testing.T) {
	ext := &OneByteHeaderExtension{}

	_, err := ext.Unmarshal([]byte{0xBE, 0xDE, 0x00, 0x00})
	assert.NoError(t, err, "Unmarshal err for valid extension")
	assert.NoError(t, ext.Set(1, []byte{0xBB}), "Set err for valid extension")
	assert.NotNil(t, ext.Get(1), "Extension should exist")
	assert.NoError(t, ext.Del(1), "Should successfully delete extension")
	assert.Nil(t, ext.Get(1), "Extension should not")
	assert.Error(t, ext.Del(1), "Should return error when deleting extension that doesnt exist")
}

func TestHeaderExtension_RFC8285TwoByteDelExtension(t *testing.T) {
	ext := &TwoByteHeaderExtension{}

	_, err := ext.Unmarshal([]byte{0x10, 0x00, 0x00, 0x00})
	assert.NoError(t, err, "Unmarshal err for valid extension")

	assert.NoError(t, ext.Set(1, []byte{0xBB}), "Set err for valid extension")

	extExtension := ext.Get(1)
	assert.NotNil(t, extExtension, "Extension should exist")

	assert.NoError(t, ext.Del(1), "Should successfully delete extension")

	extExtension = ext.Get(1)
	assert.Nil(t, extExtension, "Extension should exist")
	assert.Error(t, ext.Del(1), "Should return error when deleting extension that doesnt exist")
}
