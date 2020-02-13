package codecs

import (
	"errors"
)

// VP9Payloader payloads VP9 packets
type VP9Payloader struct{}

// Payload fragments an VP9 packet across one or more byte arrays
func (p *VP9Payloader) Payload(mtu int, payload []byte) [][]byte {
	/*
	 * https://www.ietf.org/id/draft-ietf-payload-vp9-09.txt
	 *
	 * Flexible mode (F=1)
	 *        0 1 2 3 4 5 6 7
	 *       +-+-+-+-+-+-+-+-+
	 *       |I|P|L|F|B|E|V|-| (REQUIRED)
	 *       +-+-+-+-+-+-+-+-+
	 *  I:   |M| PICTURE ID  | (REQUIRED)
	 *       +-+-+-+-+-+-+-+-+
	 *  M:   | EXTENDED PID  | (RECOMMENDED)
	 *       +-+-+-+-+-+-+-+-+
	 *  L:   | TID |U| SID |D| (CONDITIONALLY RECOMMENDED)
	 *       +-+-+-+-+-+-+-+-+                             -\
	 *  P,F: | P_DIFF      |N| (CONDITIONALLY REQUIRED)    - up to 3 times
	 *       +-+-+-+-+-+-+-+-+                             -/
	 *  V:   | SS            |
	 *       | ..            |
	 *       +-+-+-+-+-+-+-+-+
	 *
	 * Non-flexible mode (F=1)
	 *        0 1 2 3 4 5 6 7
	 *       +-+-+-+-+-+-+-+-+
	 *       |I|P|L|F|B|E|V|-| (REQUIRED)
	 *       +-+-+-+-+-+-+-+-+
	 *  I:   |M| PICTURE ID  | (RECOMMENDED)
	 *       +-+-+-+-+-+-+-+-+
	 *  M:   | EXTENDED PID  | (RECOMMENDED)
	 *       +-+-+-+-+-+-+-+-+
	 *  L:   | TID |U| SID |D| (CONDITIONALLY RECOMMENDED)
	 *       +-+-+-+-+-+-+-+-+
	 *       |   TL0PICIDX   | (CONDITIONALLY REQUIRED)
	 *       +-+-+-+-+-+-+-+-+
	 *  V:   | SS            |
	 *       | ..            |
	 *       +-+-+-+-+-+-+-+-+
	 */

	if payload == nil {
		return [][]byte{}
	}

	out := make([]byte, len(payload))
	copy(out, payload)
	return [][]byte{out}
}

// VP9Packet represents the VP9 header that is stored in the payload of an RTP Packet
type VP9Packet struct {
	// Required header
	I bool // PictureID is present
	P bool // Inter-picture predicted frame
	L bool // Layer indices is present
	F bool // Flexible mode
	B bool // Start of a frame
	E bool // End of a frame
	V bool // Scalability structure (SS) data present

	// Recommended headers
	PictureID uint16 // 7 or 16 bits, picture ID

	// Conditionally recommended headers
	TID uint8 // Temporal layer ID
	U   bool  // Switching up point
	SID uint8 // Spatial layer ID
	D   bool  // Inter-layer dependency used

	// Conditionally required headers
	PDiff     []uint8 // Reference index (F=1)
	TL0PICIDX uint8   // Temporal layer zero index (F=0)

	Payload []byte
}

// Unmarshal parses the passed byte slice and stores the result in the VP9Packet this method is called upon
func (p *VP9Packet) Unmarshal(packet []byte) ([]byte, error) {
	if packet == nil {
		return nil, errNilPacket
	}
	if len(packet) < 1 {
		return nil, errShortPacket
	}

	p.I = packet[0]&0x80 != 0
	p.P = packet[0]&0x40 != 0
	p.L = packet[0]&0x20 != 0
	p.F = packet[0]&0x10 != 0
	p.B = packet[0]&0x08 != 0
	p.E = packet[0]&0x04 != 0
	p.V = packet[0]&0x02 != 0

	if p.V {
		return nil, errors.New("scalability structure is not yet implemented")
	}

	pos := 1

	// if p.F && !p.I { // It's out of the standard but still possible to unmarshal
	// 	return nil, errors.New("picture ID is required but not present")
	// }

	if p.I {
		if len(packet) <= pos {
			return nil, errShortPacket
		}
		p.PictureID = uint16(packet[pos] & 0x7F)
		if packet[pos]&0x80 != 0 {
			pos++
			p.PictureID = p.PictureID<<8 | uint16(packet[pos])
		}
		pos++
	}

	if p.L {
		if len(packet) <= pos {
			return nil, errShortPacket
		}
		p.TID = packet[pos] >> 5
		p.U = packet[pos]&0x10 != 0
		p.SID = (packet[pos] >> 1) & 0x7
		p.D = packet[pos]&0x01 != 0
		pos++
	}

	if !p.F {
		if len(packet) <= pos {
			return nil, errShortPacket
		}
		p.TL0PICIDX = packet[pos]
		pos++
	}

	if p.F && p.P {
		for {
			if len(packet) <= pos {
				return nil, errShortPacket
			}
			p.PDiff = append(p.PDiff, packet[pos]>>1)
			if packet[pos]&0x01 == 0 {
				break
			}
			if len(p.PDiff) >= 3 {
				return nil, errTooManyPDiff
			}
			pos++
		}
		pos++
	}

	p.Payload = packet[pos:]
	return p.Payload, nil
}

// VP9PartitionHeadChecker checks VP9 partition head
type VP9PartitionHeadChecker struct{}

// IsPartitionHead checks whether if this is a head of the VP9 partition
func (*VP9PartitionHeadChecker) IsPartitionHead(packet []byte) bool {
	p := &VP9Packet{}
	if _, err := p.Unmarshal(packet); err != nil {
		return false
	}
	return p.B
}
