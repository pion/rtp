package frame

import (
	"github.com/pion/rtp/codecs"
	"github.com/pion/rtp/codecs/av1/obu"
)

type AV1PacketSampleBufferSupport struct {
	popFrame bool
	avFrame  *AV1
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

func (d *AV1PacketSampleBufferSupport) Unmarshal(payload []byte) ([]byte, error) {

	if d.popFrame {
		d.avFrame = &AV1{}
		d.popFrame = false // start frame assembling
	}

	packet := codecs.AV1Packet{}
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
		payloadSize += obu.SizeLeb128(obu.EncodeLEB128(obuLength - 1))
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

		switch obu.SizeLeb128(payloadSize) {
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
