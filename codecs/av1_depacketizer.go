// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package codecs

import (
	"fmt"

	"github.com/pion/rtp/codecs/av1/obu"
)

const (
	av1OBUTemporalDelimiter = 2
	av1OBUTileList          = 8
)

// AV1Depacketizer is a AV1 RTP Packet depacketizer.
type AV1Depacketizer struct {
	// holds the fragmented OBU from the previous packet.
	buffer []byte
}

// Unmarshal parses an AV1 RTP payload into its constituent OBU elements.
// It assumes that the payload is in order (e.g. the caller is responsible for reordering RTP packets).
// If the last OBU in the payload is fragmented, it will be stored in the buffer until the
// it is completed.
//
//nolint:gocognit,cyclop
func (d *AV1Depacketizer) Unmarshal(payload []byte) (buff []byte, err error) {
	if len(payload) <= 1 {
		return nil, errShortPacket
	}

	// |Z|Y| W |N|-|-|-|
	obuZ := (0b10000000 & payload[0]) != 0     // Z
	obuY := (0b01000000 & payload[0]) != 0     // Y
	obuCount := (0b00110000 & payload[0]) >> 4 // W
	buff = make([]byte, 0)

	// Make sure we clear the buffer if Z is not 0.
	if !obuZ && len(d.buffer) > 0 {
		buff = nil
	}

	obuOffset := 0
	for offset := 1; offset < len(payload); obuOffset++ {
		isFirst := obuOffset == 0
		isLast := obuCount != 0 && obuOffset == int(obuCount)-1

		// https://aomediacodec.github.io/av1-rtp-spec/#44-av1-aggregation-header
		// W: two bit field that describes the number of OBU elements in the packet.
		// This field MUST be set equal to 0 or equal to the number of OBU elements contained in the packet.
		// If set to 0, each OBU element MUST be preceded by a length field. If not set to 0
		// (i.e., W = 1, 2 or 3) the last OBU element MUST NOT be preceded by a length field.
		var obuSize, n int
		if obuCount == 0 || !isLast {
			obuSizeVal, nVal, err := obu.ReadLeb128(payload[offset:])
			obuSize = int(obuSizeVal) //nolint:gosec // G115 false positive
			n = int(nVal)             //nolint:gosec // G115 false positive
			if err != nil {
				return nil, err
			}

			offset += n
			if obuCount == 0 && offset+obuSize == len(payload) {
				isLast = true
			}
		} else {
			// https://aomediacodec.github.io/av1-rtp-spec/#44-av1-aggregation-header
			// Length of the last OBU element =
			// length of the RTP payload
			// - length of aggregation header
			// - length of previous OBU elements including length fields
			obuSize = len(payload) - offset
		}

		if offset+obuSize > len(payload) {
			return nil, fmt.Errorf(
				"%w: OBU size %d + %d offset exceeds payload length %d",
				errShortPacket, obuSize, offset, len(payload),
			)
		}

		var obuBuffer []byte
		if isFirst && obuZ {
			// We lost the first fragment of the OBU
			// We drop the buffer and continue
			if len(d.buffer) == 0 {
				if isLast {
					break
				}

				offset += obuSize

				continue
			}

			obuBuffer = make([]byte, len(d.buffer)+obuSize)

			copy(obuBuffer, d.buffer)
			copy(obuBuffer[len(d.buffer):], payload[offset:offset+obuSize])
			d.buffer = nil
		} else {
			obuBuffer = payload[offset : offset+obuSize]
		}
		offset += obuSize

		if isLast && obuY {
			d.buffer = obuBuffer
		} else {
			if len(obuBuffer) == 0 {
				return nil, fmt.Errorf(
					"%w: OBU size %d is 0",
					errShortPacket, obuSize,
				)
			}

			// The temporal delimiter OBU, if present, SHOULD be removed when transmitting,
			// and MUST be ignored by receivers. Tile list OBUs are not supported.
			// They SHOULD be removed when transmitted, and MUST be ignored by receivers.
			// https://aomediacodec.github.io/av1-rtp-spec/#5-packetization-rules

			obuType := (obuBuffer[0] & obuFrameTypeMask) >> obuFrameTypeBitshift

			if obuType != av1OBUTemporalDelimiter && obuType != av1OBUTileList {
				buff = append(buff, obuBuffer...)
			}
		}

		if isLast {
			break
		}
	}

	if obuCount != 0 && obuOffset != int(obuCount-1) {
		return nil, fmt.Errorf(
			"%w: OBU count %d does not match number of OBUs %d",
			errShortPacket, obuCount, obuOffset,
		)
	}

	return buff, nil
}

// IsPartitionTail returns true if RTP packet marker is set.
// Clear the buffer if we are at the end of the partition.
func (d *AV1Depacketizer) IsPartitionTail(marker bool, _ []byte) bool {
	if marker {
		// We make sure we clear the buffer if we are at the end of the partition.
		d.buffer = nil

		return true
	}

	return false
}

// IsPartitionHead returns true if Z in the AV1 Aggregation Header
// is set to 0.
func (d *AV1Depacketizer) IsPartitionHead(payload []byte) bool {
	if len(payload) == 0 {
		return false
	}

	return (payload[0] & 0b11000000) == 0
}
