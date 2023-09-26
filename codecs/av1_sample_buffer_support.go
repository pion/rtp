package codecs

import (
	"github.com/pion/rtp/pkg/frame"
	"github.com/pion/rtp/pkg/obu"
)

type AV1PacketSampleBufferSupport struct {
	popFrame bool
	avframe  *frame.AV1
}

func (d *AV1PacketSampleBufferSupport) IsPartitionTail(marker bool, payload []byte) bool {
	d.popFrame = true
	return marker
}

// IsPartitionHead checks whether if this is a head of the AV1 partition
func (d *AV1PacketSampleBufferSupport) IsPartitionHead(payload []byte) bool {
	d.popFrame = true
	if len(payload) == 0 {
		return false
	}
	return (payload[0] & byte(0b10000000)) == 0
}

func SizeLeb128(leb128 uint) uint {
	if (leb128 >> 24) > 0 {
		return 4
	} else if (leb128 >> 16) > 0 {
		return 3
	} else if (leb128 >> 8) > 0 {
		return 2
	}
	return 1
}

func (p *AV1PacketSampleBufferSupport) Unmarshal(payload []byte) ([]byte, error) {

	if p.popFrame {
		p.avframe = &frame.AV1{}
		p.popFrame = false // start frame assembling
	}

	packet := AV1Packet{}
	_, err := packet.Unmarshal(payload)

	if err != nil {
		return nil, err
	}

	OBUS, _ := p.avframe.ReadFrames(&packet)

	if len(OBUS) == 0 {
		return nil, nil
	}

	var payloadSize uint = 0

	for i := range OBUS {
		obulength := uint(len(OBUS[i]))
		payloadSize += obulength
		payloadSize += SizeLeb128(obu.EncodeLEB128(obulength))
	}

	result := make([]byte, payloadSize)

	offset := 0
	for i := range OBUS {
		result[offset] = OBUS[i][0] ^ 2 // mark size header exists
		offset++
		len_minus := len(OBUS[i]) - 1
		payloadSize := obu.EncodeLEB128(uint(len_minus))

		switch SizeLeb128(payloadSize) {
		case 4:
			result[offset] = byte(payloadSize >> 24)
			offset++
			fallthrough
		case 3:
			result[offset] = byte(payloadSize >> 16)
			offset++
			fallthrough
		case 2:
			result[offset] = byte(payloadSize >> 8)
			offset++
			fallthrough
		case 1:
			result[offset] = byte(payloadSize)
			offset++
		}

		copy(result[offset:], OBUS[i][1:])
		offset += len_minus
	}

	return result, nil
}
