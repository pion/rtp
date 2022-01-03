package rtp

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestHeaderExtension_RFC8285OneByteExtension(t *testing.T) {
	p := &OneByteHeaderExtension{}

	rawPkt := []byte{
		0xBE, 0xDE, 0x00, 0x01, 0x50, 0xAA, 0x00, 0x00,
		0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	if _, err := p.Unmarshal(rawPkt); err != nil {
		t.Fatal("Unmarshal err for valid extension")
	}

	dstData, _ := p.Marshal()
	if !bytes.Equal(dstData, rawPkt) {
		t.Errorf("Marshal failed raw \nMarshaled:\n%s\nrawPkt:\n%s", hex.Dump(dstData), hex.Dump(rawPkt))
	}
}

func TestHeaderExtension_RFC8285OneByteTwoExtensionOfTwoBytes(t *testing.T) {
	p := &OneByteHeaderExtension{}

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
	if _, err := p.Unmarshal(rawPkt); err != nil {
		t.Fatal("Unmarshal err for valid extension")
	}

	ext1 := p.Get(1)
	ext1Expect := []byte{0xAA}
	if !bytes.Equal(ext1, ext1Expect) {
		t.Errorf("Extension has incorrect data. Got: %+v, Expected: %+v", ext1, ext1Expect)
	}

	ext2 := p.Get(2)
	ext2Expect := []byte{0xBB}
	if !bytes.Equal(ext2, ext2Expect) {
		t.Errorf("Extension has incorrect data. Got: %+v, Expected: %+v", ext2, ext2Expect)
	}

	dstData, _ := p.Marshal()
	if !bytes.Equal(dstData, rawPkt) {
		t.Errorf("Marshal failed raw \nMarshaled:\n%s\nrawPkt:\n%s", hex.Dump(dstData), hex.Dump(rawPkt))
	}
}

func TestHeaderExtension_RFC8285OneByteMultipleExtensionsWithPadding(t *testing.T) {
	p := &OneByteHeaderExtension{}

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
	if _, err := p.Unmarshal(rawPkt); err != nil {
		t.Fatal("Unmarshal err for valid extension")
	}

	ext1 := p.Get(1)
	ext1Expect := []byte{0xAA}
	if !bytes.Equal(ext1, ext1Expect) {
		t.Errorf("Extension has incorrect data. Got: %v+, Expected: %v+", ext1, ext1Expect)
	}

	ext2 := p.Get(2)
	ext2Expect := []byte{0xBB, 0xBB}
	if !bytes.Equal(ext2, ext2Expect) {
		t.Errorf("Extension has incorrect data. Got: %v+, Expected: %v+", ext2, ext2Expect)
	}

	ext3 := p.Get(3)
	ext3Expect := []byte{0xCC, 0xCC, 0xCC, 0xCC}
	if !bytes.Equal(ext3, ext3Expect) {
		t.Errorf("Extension has incorrect data. Got: %v+, Expected: %v+", ext3, ext3Expect)
	}

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
			n, err := p.MarshalTo(buf)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(buf[:n], rawPkt) {
				t.Errorf("Marshal failed raw \nMarshaled:\n%s\nrawPkt:\n%s", hex.Dump(buf[:n]), hex.Dump(rawPkt))
			}
		})
	}
}

func TestHeaderExtension_RFC8285TwoByteExtension(t *testing.T) {
	p := &TwoByteHeaderExtension{}

	rawPkt := []byte{
		0x10, 0x00, 0x00, 0x07, 0x05, 0x18, 0xAA, 0xAA,
		0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
		0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,
		0xAA, 0xAA, 0x00, 0x00,
	}
	if _, err := p.Unmarshal(rawPkt); err != nil {
		t.Fatal("Unmarshal err for valid extension")
	}

	dstData, _ := p.Marshal()
	if !bytes.Equal(dstData, rawPkt) {
		t.Errorf("Marshal failed raw \nMarshaled:\n%s\nrawPkt:\n%s", hex.Dump(dstData), hex.Dump(rawPkt))
	}
}

func TestHeaderExtension_RFC8285TwoByteMultipleExtensionsWithPadding(t *testing.T) {
	p := &TwoByteHeaderExtension{}

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

	if _, err := p.Unmarshal(rawPkt); err != nil {
		t.Fatal("Unmarshal err for valid extension")
	}

	ext1 := p.Get(1)
	ext1Expect := []byte{}
	if !bytes.Equal(ext1, ext1Expect) {
		t.Errorf("Extension has incorrect data. Got: %v+, Expected: %v+", ext1, ext1Expect)
	}

	ext2 := p.Get(2)
	ext2Expect := []byte{0xBB}
	if !bytes.Equal(ext2, ext2Expect) {
		t.Errorf("Extension has incorrect data. Got: %v+, Expected: %v+", ext2, ext2Expect)
	}

	ext3 := p.Get(3)
	ext3Expect := []byte{0xCC, 0xCC, 0xCC, 0xCC}
	if !bytes.Equal(ext3, ext3Expect) {
		t.Errorf("Extension has incorrect data. Got: %v+, Expected: %v+", ext3, ext3Expect)
	}
}

func TestHeaderExtension_RFC8285TwoByteMultipleExtensionsWithLargeExtension(t *testing.T) {
	p := &TwoByteHeaderExtension{}

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

	if _, err := p.Unmarshal(rawPkt); err != nil {
		t.Fatal("Unmarshal err for valid extension")
	}

	ext1 := p.Get(1)
	ext1Expect := []byte{}
	if !bytes.Equal(ext1, ext1Expect) {
		t.Errorf("Extension has incorrect data. Got: %v+, Expected: %v+", ext1, ext1Expect)
	}

	ext2 := p.Get(2)
	ext2Expect := []byte{0xBB}
	if !bytes.Equal(ext2, ext2Expect) {
		t.Errorf("Extension has incorrect data. Got: %v+, Expected: %v+", ext2, ext2Expect)
	}

	ext3 := p.Get(3)
	ext3Expect := []byte{
		0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC,
		0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC, 0xCC,
	}
	if !bytes.Equal(ext3, ext3Expect) {
		t.Errorf("Extension has incorrect data. Got: %v+, Expected: %v+", ext3, ext3Expect)
	}

	dstData, _ := p.Marshal()
	if !bytes.Equal(dstData, rawPkt) {
		t.Errorf("Marshal failed raw \nMarshaled: %+v,\nrawPkt:    %+v", dstData, rawPkt)
	}
}

func TestHeaderExtension_RFC8285OneByteDelExtension(t *testing.T) {
	p := &OneByteHeaderExtension{}

	if _, err := p.Unmarshal([]byte{0xBE, 0xDE, 0x00, 0x00}); err != nil {
		t.Fatal("Unmarshal err for valid extension")
	}

	if err := p.Set(1, []byte{0xBB}); err != nil {
		t.Fatal("Set err for valid extension")
	}

	ext := p.Get(1)
	if ext == nil {
		t.Error("Extension should exist")
	}

	err := p.Del(1)
	if err != nil {
		t.Error("Should successfully delete extension")
	}

	ext = p.Get(1)
	if ext != nil {
		t.Error("Extension should not exist")
	}

	err = p.Del(1)
	if err == nil {
		t.Error("Should return error when deleting extension that doesnt exist")
	}
}

func TestHeaderExtension_RFC8285TwoByteDelExtension(t *testing.T) {
	p := &TwoByteHeaderExtension{}

	if _, err := p.Unmarshal([]byte{0x10, 0x00, 0x00, 0x00}); err != nil {
		t.Fatal("Unmarshal err for valid extension")
	}

	if err := p.Set(1, []byte{0xBB}); err != nil {
		t.Fatal("Set err for valid extension")
	}

	ext := p.Get(1)
	if ext == nil {
		t.Error("Extension should exist")
	}

	err := p.Del(1)
	if err != nil {
		t.Error("Should successfully delete extension")
	}

	ext = p.Get(1)
	if ext != nil {
		t.Error("Extension should not exist")
	}

	err = p.Del(1)
	if err == nil {
		t.Error("Should return error when deleting extension that doesnt exist")
	}
}
