package rtp

import (
	"bytes"
	"reflect"
	"testing"
)

func TestBasic(t *testing.T) {
	p := &Packet{}

	if err := p.Unmarshal([]byte{}); err == nil {
		t.Fatal("Unmarshal did not error on zero length packet")
	}

	rawPkt := []byte{
		0x90, 0xe0, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0x00, 0x01, 0x00, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0x98, 0x36, 0xbe, 0x88, 0x9e,
	}
	parsedPacket := &Packet{
		Header: Header{
			Marker:           true,
			Extension:        true,
			ExtensionProfile: 1,
			ExtensionPayload: []byte{0xFF, 0xFF, 0xFF, 0xFF},
			Version:          2,
			PayloadOffset:    20,
			PayloadType:      96,
			SequenceNumber:   27023,
			Timestamp:        3653407706,
			SSRC:             476325762,
			CSRC:             []uint32{},
		},
		Payload: rawPkt[20:],
		Raw:     rawPkt,
	}

	if err := p.Unmarshal(rawPkt); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(p, parsedPacket) {
		t.Errorf("TestBasic unmarshal: got %#v, want %#v", p, parsedPacket)
	}

	if parsedPacket.Header.MarshalSize() != 20 {
		t.Errorf("wrong computed header marshal size")
	} else if parsedPacket.MarshalSize() != len(rawPkt) {
		t.Errorf("wrong computed marshal size")
	}

	if p.PayloadOffset != 20 {
		t.Errorf("wrong payload offset: %d != %d", p.PayloadOffset, 20)
	}

	raw, err := p.Marshal()
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(raw, rawPkt) {
		t.Errorf("TestBasic marshal: got %#v, want %#v", raw, rawPkt)
	}

	// TODO This is a BUG but without it, stuff breaks.
	if p.PayloadOffset != 12 {
		t.Errorf("wrong payload offset: %d != %d", p.PayloadOffset, 12)
	}
}

func TestExtension(t *testing.T) {
	p := &Packet{}

	missingExtensionPkt := []byte{
		0x90, 0x60, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82,
	}
	if err := p.Unmarshal(missingExtensionPkt); err == nil {
		t.Fatal("Unmarshal did not error on packet with missing extension data")
	}

	invalidExtensionLengthPkt := []byte{
		0x90, 0x60, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0x99, 0x99, 0x99, 0x99,
	}
	if err := p.Unmarshal(invalidExtensionLengthPkt); err == nil {
		t.Fatal("Unmarshal did not error on packet with invalid extension length")
	}

	p = &Packet{Header: Header{
		Extension:        true,
		ExtensionProfile: 3,
		ExtensionPayload: []byte{0},
	},
		Payload: []byte{},
	}
	if _, err := p.Marshal(); err == nil {
		t.Fatal("Marshal did not error on packet with invalid extension length")
	}
}

func TestRoundtrip(t *testing.T) {
	rawPkt := []byte{
		0x00, 0x10, 0x23, 0x45, 0x12, 0x34, 0x45, 0x67, 0xCC, 0xDD, 0xEE, 0xFF,
		0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77,
	}
	payload := rawPkt[12:]

	p := &Packet{}
	if err := p.Unmarshal(rawPkt); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rawPkt, p.Raw) {
		t.Errorf("p.Raw must be same as rawPkt.\n p.Raw: %+v,\nrawPkt: %+v", p.Raw, rawPkt)
	}
	if !bytes.Equal(payload, p.Payload) {
		t.Errorf("p.Payload must be same as payload.\n  payload: %+v,\np.Payload: %+v",
			payload, p.Payload,
		)
	}

	buf, err := p.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rawPkt, buf) {
		t.Errorf("buf must be same as rawPkt.\n   buf: %+v,\nrawPkt: %+v", buf, rawPkt)
	}
	if !bytes.Equal(rawPkt, p.Raw) {
		t.Errorf("p.Raw must be same as rawPkt.\n p.Raw: %+v,\nrawPkt: %+v", p.Raw, rawPkt)
	}
	if !bytes.Equal(payload, p.Payload) {
		t.Errorf("p.Payload must be same as payload.\n  payload: %+v,\np.Payload: %+v",
			payload, p.Payload,
		)
	}
}

func BenchmarkMarshal(b *testing.B) {
	rawPkt := []byte{
		0x90, 0x60, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0x00, 0x01, 0x00, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0x98, 0x36, 0xbe, 0x88, 0x9e,
	}

	p := &Packet{}
	err := p.Unmarshal(rawPkt)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err = p.Marshal()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshalTo(b *testing.B) {
	rawPkt := []byte{
		0x90, 0x60, 0x69, 0x8f, 0xd9, 0xc2, 0x93, 0xda, 0x1c, 0x64,
		0x27, 0x82, 0x00, 0x01, 0x00, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0x98, 0x36, 0xbe, 0x88, 0x9e,
	}

	p := &Packet{}

	err := p.Unmarshal(rawPkt)
	if err != nil {
		b.Fatal(err)
	}

	buf := [100]byte{}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err = p.MarshalTo(buf[:])
		if err != nil {
			b.Fatal(err)
		}
	}
}
