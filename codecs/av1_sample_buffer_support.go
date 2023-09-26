package codecs

import (
	"github.com/pion/rtp/pkg/frame"
	"github.com/pion/rtp/pkg/obu"
)

type AV1PacketSampleBufferSupport struct {
	popFrame bool
	avFrame  *frame.AV1
}

func (d *AV1PacketSampleBufferSupport) IsPartitionTail(marker bool, _ []byte) bool {
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

func sizeLeb128(leb128 uint) uint {
	if (leb128 >> 24) > 0 {
		return 4
	} else if (leb128 >> 16) > 0 {
		return 3
	} else if (leb128 >> 8) > 0 {
		return 2
	}
	return 1
}

func (d *AV1PacketSampleBufferSupport) Unmarshal(payload []byte) ([]byte, error) {

	if d.popFrame {
		d.avFrame = &frame.AV1{}
		d.popFrame = false // start frame assembling
	}

	packet := AV1Packet{}
	_, err := packet.Unmarshal(payload)

	if err != nil {
		return nil, err
	}

	OBUs, _ := d.avFrame.ReadFrames(&packet)

	if len(OBUs) == 0 {
		return nil, nil
	}

	var payloadSize uint = 0

	for i := range OBUs {
		obuLength := uint(len(OBUs[i]))
		if obuLength == 0 {
			continue
		}
		payloadSize += obuLength
		payloadSize += sizeLeb128(obu.EncodeLEB128(obuLength - 1))
	}

	result := make([]byte, payloadSize)

	offset := 0
	for i := range OBUs {
		obuLength := len(OBUs[i])

		if obuLength == 0 {
			continue
		}

		lenMinus := obuLength - 1

		result[offset] = OBUs[i][0] ^ 2 // mark size header exists
		offset++
		payloadSize := obu.EncodeLEB128(uint(lenMinus))

		switch sizeLeb128(payloadSize) {
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

		copy(result[offset:], OBUs[i][1:])
		offset += lenMinus
	}

	return result, nil
}
