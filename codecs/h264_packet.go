package codecs

// H264Payloader payloads H264 packets
type H264Payloader struct{}

const (
	fuaHeaderSize = 2
)

// I am sorry world, for the OPTIMIZATIONS I have bestowed upon thee.
// This parses an Annex.B bitstream and returns the next NALU and any remaining data.
// Annex.B uses a start-code of either 001 or 0001.
// Each NALU is prefixed with a start-code, which runs until the next start-code (or EOF)
func nextNALU(data []byte) (nalu []byte, remaining []byte) {
	i := 0

	// Loop over the entire data, except for the last 3 bytes.
	// We know that we can't fit a start code in that size plus it would cause a range error.
	for i+2 < len(data) {
		// We want to check if the current index is the start of the start code, otherwise advance.

		// Check the 3rd byte first, because it's the most useful.
		switch data[i+2] {
		case 0:
			// state: ? ? 0

			switch {
			case data[i+1] != 0:
				// state: ? n 0

				// The 3rd byte could be the start of a code.
				i += 2
			case data[i] != 0:
				// state: n 0 0

				// The 2nd byte could be the start of a code.
				i++
			default:
				// state: 0 0 0 ?

				// I don't think it's possible to have three zeroes except for a start code.
				// But to be safe, let's handle edge cases.

				if i+3 < len(data) {
					switch data[i+3] {
					case 0:
						// state: 0 0 0 0

						// This shouldn't be possible but I guess it could be part of a start code.
						i++
					case 1:
						// state: 0 0 0 1
						return data[:i], data[i+4:]
					default:
						// state: 0 0 0 n

						// No start code possible so we can advance 4 bytes.
						i += 4
					}
				}

				// state: 0 0 0 EOF

				return data, nil
			}
		case 1:
			// state: ? ? 1

			if data[i] == 0 && data[i+1] == 0 {
				// state: 0 0 1
				return data[:i], data[i+3:]
			}

			// state: n ? 1 or ? n 1

			// Neither of these could be part of a start code.
			i += 3
		default:
			// state: ? ? n

			// If it's not 0 or 1, then there's no start code nor can it be part of a start code.
			// So most of the time, we advance 3 bytes at once instead of just 1 byte at a time.
			i += 3
		}
	}

	// At the end of the file, return what we've got.
	return data, nil
}

// Payload fragments a H264 packet across one or more byte arrays
func (p *H264Payloader) Payload(mtu int, payload []byte) [][]byte {

	var payloads [][]byte
	if payload == nil {
		return payloads
	}

	for len(payload) > 0 {
		nalu, remaining := nextNALU(payload)
		payload = remaining

		if len(nalu) == 0 {
			// This will only be true for the first start code.
			continue
		}

		naluType := nalu[0] & 0x1F
		naluRefIdc := nalu[0] & 0x60

		if naluType == 9 || naluType == 12 {
			continue
		}

		// Single NALU
		if len(nalu) <= mtu {
			out := make([]byte, len(nalu))
			copy(out, nalu)
			payloads = append(payloads, out)
			continue
		}

		// FU-A
		maxFragmentSize := mtu - fuaHeaderSize

		// The FU payload consists of fragments of the payload of the fragmented
		// NAL unit so that if the fragmentation unit payloads of consecutive
		// FUs are sequentially concatenated, the payload of the fragmented NAL
		// unit can be reconstructed.  The NAL unit type octet of the fragmented
		// NAL unit is not included as such in the fragmentation unit payload,
		// 	but rather the information of the NAL unit type octet of the
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
			continue
		}

		for naluDataRemaining > 0 {
			currentFragmentSize := min(maxFragmentSize, naluDataRemaining)
			out := make([]byte, fuaHeaderSize+currentFragmentSize)

			// +---------------+
			// |0|1|2|3|4|5|6|7|
			// +-+-+-+-+-+-+-+-+
			// |F|NRI|  Type   |
			// +---------------+
			out[0] = 28
			out[0] |= naluRefIdc

			// +---------------+
			//|0|1|2|3|4|5|6|7|
			//+-+-+-+-+-+-+-+-+
			//|S|E|R|  Type   |
			//+---------------+

			out[1] = naluType
			if naluDataRemaining == naluDataLength {
				// Set start bit
				out[1] |= 1 << 7
			} else if naluDataRemaining-currentFragmentSize == 0 {
				// Set end bit
				out[1] |= 1 << 6
			}

			copy(out[fuaHeaderSize:], naluData[naluDataIndex:naluDataIndex+currentFragmentSize])
			payloads = append(payloads, out)

			naluDataRemaining -= currentFragmentSize
			naluDataIndex += currentFragmentSize
		}

	}

	return payloads
}
