package codecs

import (
	"encoding/binary"
	"errors"
)

// H264InterleavedPayloader payloads H264 packets
type H264InterleavedPayloader struct{}

const (
	stapbNALUType = 25
	fuBNALUType   = 29

	fubHeaderSize   = 2
	stapbHeaderSize = 1

	stapbNALULengthSize = 2

	donSize = 2
)

//var donSN uint16 = 0

// convert Nal to STAP-B Payload
func NalToStapBPayload(nalu []byte, donSN uint16) []byte {

	//naluType := nalu[0] & naluTypeBitmask
	naluRefIdc := nalu[0] & naluRefIdcBitmask

	out := make([]byte, stapbHeaderSize+donSize+stapbNALULengthSize+len(nalu))

	// +---------------+
	// |0|1|2|3|4|5|6|7|
	// +-+-+-+-+-+-+-+-+
	// |F|NRI|  Type   |
	// +---------------+
	out[0] = stapbNALUType
	out[0] |= naluRefIdc

	binary.BigEndian.PutUint16(out[stapbHeaderSize:], donSN)
	binary.BigEndian.PutUint16(out[stapbHeaderSize+donSize:], uint16(len(nalu)))

	copy(out[stapbHeaderSize+donSize+stapaNALULengthSize:], nalu)

	// increase donSN only for new frames (nonIdr).
	// not for SPS, PPS
	//if naluType == 1 || naluType == 5{
	//	donSN++
	//}

	return out
}

// fragment NALU based on FU-B scheme
func NalToFuBPayload(mtu uint16, nalu []byte, donSN uint16) [][]byte {

	var payloads [][]byte

	naluType := nalu[0] & naluTypeBitmask
	naluRefIdc := nalu[0] & naluRefIdcBitmask

	// FU-B
	maxFragmentSize := int(mtu) - fubHeaderSize

	// The FU payload consists of fragments of the payload of the fragmented
	// NAL unit so that if the fragmentation unit payloads of consecutive
	// FUs are sequentially concatenated, the payload of the fragmented NAL
	// unit can be reconstructed.  The NAL unit type octet of the fragmented
	// NAL unit is not included as such in the fragmentation unit payload,
	//  but rather the information of the NAL unit type octet of the
	// fragmented NAL unit is conveyed in the F and NRI fields of the FU
	// indicator octet of the fragmentation unit and in the type field of
	// the FU header.  An FU payload MAY have any number of octets and MAY
	// be empty.

	naluData := nalu
	// According to the RFC, the first octet is skipped due to redundant information
	naluDataIndex := 1
	naluDataLength := len(nalu) - naluDataIndex
	naluDataRemaining := naluDataLength

	if min(maxFragmentSize, naluDataRemaining) <= 0 {
		return payloads
	}
	var currentFragmentSize int

	for naluDataRemaining > 0 {
		var out []byte

		// first fragment, create FU-B packet
		if naluDataRemaining == naluDataLength {

			maxFragmentSize = int(mtu) - fubHeaderSize - donSize
			currentFragmentSize = min(maxFragmentSize, naluDataRemaining)
			out = make([]byte, fubHeaderSize+donSize+currentFragmentSize)

			// +---------------+
			// |0|1|2|3|4|5|6|7|
			// +-+-+-+-+-+-+-+-+
			// |F|NRI|  Type   |
			// +---------------+

			out[0] = fuBNALUType
			out[0] |= naluRefIdc

		} else {
			// not first fragment, create FU-A packet
			maxFragmentSize = int(mtu) - fuaHeaderSize
			currentFragmentSize = min(maxFragmentSize, naluDataRemaining)

			out = make([]byte, fuaHeaderSize+currentFragmentSize)

			// +---------------+
			// |0|1|2|3|4|5|6|7|
			// +-+-+-+-+-+-+-+-+
			// |F|NRI|  Type   |
			// +---------------+

			out[0] = fuaNALUType
			out[0] |= naluRefIdc
		}

		// +---------------+
		// |0|1|2|3|4|5|6|7|
		// +-+-+-+-+-+-+-+-+
		// |S|E|R|  Type   |
		// +---------------+

		out[1] = naluType

		// first fragment: FU-B
		if naluDataRemaining == naluDataLength {
			// Set start bit
			out[1] |= 1 << 7

			// copy the don into the array
			binary.BigEndian.PutUint16(out[fuaHeaderSize:], donSN)
			// copy the nal into the array
			copy(out[fubHeaderSize+donSize:], naluData[naluDataIndex:naluDataIndex+currentFragmentSize])

			// last fragment: FU-A
		} else if naluDataRemaining-currentFragmentSize == 0 {
			// Set end bit
			out[1] |= 1 << 6
			copy(out[fuaHeaderSize:], naluData[naluDataIndex:naluDataIndex+currentFragmentSize])

			// not first, not last, FU-A
		} else {
			copy(out[fuaHeaderSize:], naluData[naluDataIndex:naluDataIndex+currentFragmentSize])
		}

		payloads = append(payloads, out)

		naluDataRemaining -= currentFragmentSize
		naluDataIndex += currentFragmentSize
	}

	return payloads
}

// Payload fragments a H264 packet across one or more byte arrays
func (p *H264InterleavedPayloader) Payload(mtu uint16, payload []byte) [][]byte {
	var payloads [][]byte
	if len(payload) < 2 {
		return payloads
	}

	// first 2 bytes hold the DON for the interleaved mode
	// next bytes are the frame itself
	frame := payload[2:]
	don := binary.BigEndian.Uint16(payload)

	emitNalus(frame, func(nalu []byte) {
		if len(nalu) == 0 {
			return
		}

		naluType := nalu[0] & naluTypeBitmask

		if naluType == 9 || naluType == 12 {
			return
		}

		// NALU fits into a sinlge packet, wrap it in an STAP-B packet
		if len(nalu)+stapbHeaderSize+donSize+stapbNALULengthSize <= int(mtu) {
			payloads = append(payloads, NalToStapBPayload(nalu, don))
			return
		}

		// NALU is larger than mtu, fragment according to FU-B scheme
		payloads = append(payloads, NalToFuBPayload(mtu, nalu, don)...)
	})

	return payloads
}

// H264Packet represents the H264 header that is stored in the payload of an RTP Packet
type H264InterlevedPacket struct {
}

// Unmarshal parses the passed byte slice and stores the result in the H264Packet this method is called upon
func (p *H264InterlevedPacket) Unmarshal(payload []byte) ([]byte, error) {
	return nil, errors.New("not implemented")
}
