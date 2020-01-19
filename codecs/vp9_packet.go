package codecs

import (
	"fmt"
)

// VP9Payloader payloads VP9 packets
type VP9Payloader struct{}

// Payload fragments an VP9 packet across one or more byte arrays
func (p *VP9Payloader) Payload(mtu int, payload []byte) [][]byte {
	if payload == nil {
		return [][]byte{}
	}

	out := make([]byte, len(payload))
	copy(out, payload)
	return [][]byte{out}
}

// VP9Packet represents the VP9 header that is stored in the payload of an RTP Packet
type VP9Packet struct {
	Payload []byte
}

// Unmarshal parses the passed byte slice and stores the result in the VP9Packet this method is called upon
func (p *VP9Packet) Unmarshal(packet []byte) ([]byte, error) {
	if packet == nil {
		return nil, fmt.Errorf("invalid nil packet")
	}

	if len(packet) == 0 {
		return nil, fmt.Errorf("Payload is not large enough")
	}
	p.Payload = packet
	return packet, nil
}

// VP9PartitionHeadChecker checks VP9 partition head
type VP9PartitionHeadChecker struct{}

// IsPartitionHead checks whether if this is a head of the VP9 partition
func (*VP9PartitionHeadChecker) IsPartitionHead(packet []byte) bool {
	p := &VP9Packet{}
	if _, err := p.Unmarshal(packet); err != nil {
		return false
	}
	return true
}
