package codecs

import (
	"fmt"

	"github.com/pions/rtp"
)

// OpusPayloader payloads Opus packets
type OpusPayloader struct{}

// Payload fragments an Opus packet across one or more byte arrays
func (p *OpusPayloader) Payload(mtu int, payload []byte) [][]byte {
	if payload == nil {
		return [][]byte{}
	}

	out := make([]byte, len(payload))
	copy(out, payload)
	return [][]byte{out}
}

// OpusPacket represents the VP8 header that is stored in the payload of an RTP Packet
type OpusPacket struct {
	Payload []byte
}

// Unmarshal parses the passed byte slice and stores the result in the OpusPacket this method is called upon
func (p *OpusPacket) Unmarshal(packet *rtp.Packet) ([]byte, error) {
	if packet == nil {
		return nil, fmt.Errorf("invalid nil packet")
	}
	if packet.Payload == nil {
		return nil, fmt.Errorf("Payload is not large enough")
	}
	p.Payload = packet.Payload
	return p.Payload, nil
}
