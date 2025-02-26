/*
H264 payload reducer.
MTU here is related to the RTP and its content (not taking acount the IP/UDP/etc layers)
Assumes that RTP packets will enter already sequenced (but may contains sequence holes )
Assumes that if a given FU-A packet is bigger than the requested MTU, the previous related FU-A will be too big as well.
*/

package codecs

import (
	"encoding/binary"
	"fmt"

	"github.com/pion/rtp/v2"
)

type H264spreader struct {
	Mtu          int
	Spreading    bool
	RtpOffset    uint16
	fuInProgress *fuInProgress
	trailingBuf  []byte
}

type fuInProgress struct {
	LastSeq      uint16
	RtpHeader    []byte
	Trailing     []byte
	FuStartBytes [2]byte
}

const (
	minRtpHeaderSize = 12
	rtpVPECsrcOffset = 0
	rtpMPtOffset     = 1
	rtpSeqNumOffset  = 2
	rtpSeqNumLength  = 2

	nalUnitTypeOffset  = 0
	nalUnitTypeSize    = 1
	fuaOverhead        = 2
	fuaIndicatorOffset = 0
	fuaHeaderOffest    = 1

	stapbNALUType  = 25
	mtap16NALUType = 26
	mtap24NALUType = 27

	rtpPaddingBitMask = byte(0x20)
	rtpMarkerBitMask  = byte(0x80)
)

// From rfc3550
// ===================================
//       RTP header (minimal part)
// ===================================
//
// 0                   1                   2                   3
// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |V=2|P|X|  CC   |M|     PT      |       sequence number         |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                           timestamp                           |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |           synchronization source (SSRC) identifier            |
// +=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
//
//
//
// From rfc6184
// ===================================
//       Single NAL Unit Packet
// ===================================
//
// 0                   1                   2                   3
// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |F|NRI|  Type   |                                               |
// +-+-+-+-+-+-+-+-+                                               |
// |                                                               |
// |               Bytes 2..n of a single NAL unit                 |
// |                                                               |
// |                               +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                               :...OPTIONAL RTP padding        |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//
// ===================================
//       FU-A
// ===================================
//
// RTP payload format for FU-A :
// 0                   1                   2                   3
// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// | FU indicator  |   FU header   |                               |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+                               |
// |                         FU payload                            |
// |                               +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                               :...OPTIONAL RTP padding        |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//
// The FU indicator octet has the following format:
// +---------------+
// |0|1|2|3|4|5|6|7|
// +-+-+-+-+-+-+-+-+
// |F|NRI|  Type   |
// +---------------+
//
// The FU header has the following format:
// +---------------+
// |0|1|2|3|4|5|6|7|
// +-+-+-+-+-+-+-+-+
// |S|E|R|  Type   |
// +---------------+
//
// ===================================
//        STAP-A
// ===================================
//
// An example of an RTP packet including an STAP-A containing two single-time aggregation units
// 0                   1                   2                   3
// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                          RTP Header                           |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |STAP-A NAL HDR |         NALU 1 Size           | NALU 1 HDR    |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                         NALU 1 Data                           |
// :                                                               :
// +               +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |               | NALU 2 Size                   | NALU 2 HDR    |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                         NALU 2 Data                           |
// :                                                               :
// |                               +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                               :...OPTIONAL RTP padding        |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

func NewH264spreader(mtu uint16) H264spreader {
	return H264spreader{
		Mtu:          int(mtu),
		Spreading:    false,
		RtpOffset:    0,
		fuInProgress: nil,
		trailingBuf:  make([]byte, mtu),
	}
}

func (s *H264spreader) Process(payload []byte) (outPayloads [][]byte, err error) {
	outPayloads = make([][]byte, 0, 4)
	payLen := len(payload)
	if payLen == 0 {
		return outPayloads, nil
	} else if payLen < minRtpHeaderSize {
		return nil, fmt.Errorf("payload is too small: %d", payLen)
	} else if !s.Spreading && (payLen <= s.Mtu) {
		//best case scenario : all RTP pkts were small enough up to now, nothing to do! Pkt goes straight!
		outPayloads = append(outPayloads, payload)
		return outPayloads, nil
	}

	s.Spreading = true

	//rtp seq offset to compensate for the previous extra pkts we inserted
	seqNum := binary.BigEndian.Uint16(payload[rtpSeqNumOffset : rtpSeqNumOffset+rtpSeqNumLength])
	seqNum += s.RtpOffset
	binary.BigEndian.PutUint16(payload[rtpSeqNumOffset:rtpSeqNumOffset+rtpSeqNumLength], seqNum)

	if s.fuInProgress == nil && (payLen <= s.Mtu) {
		//whenever possible, forward RTP pkts without any Unmarshal()
		outPayloads = append(outPayloads, payload)
		return outPayloads, nil
	}

	rtpPkt := &rtp.Packet{}
	err = rtpPkt.Unmarshal(payload)
	if err != nil {
		return nil, err
	} else if len(rtpPkt.Payload) < 2 {
		return nil, fmt.Errorf("nal content is too small: %d", len(rtpPkt.Payload))
	}

	//avoiding repetitive RTP Marshal() by passing around the RTP header slice (as a data template)
	nalData := rtpPkt.Payload
	rtpHeaderSize := payLen - len(rtpPkt.Payload) - int(rtpPkt.PaddingSize)
	rtpHeaderData := payload[:rtpHeaderSize]
	rtpHeaderData[rtpVPECsrcOffset] &= ^rtpPaddingBitMask

	naluType := nalData[nalUnitTypeOffset] & naluTypeBitmask
	if naluType != fuaNALUType && s.fuInProgress != nil {
		outPayloads, seqNum = s.flushFuPending(outPayloads, seqNum)

		if payLen <= s.Mtu {
			outPayloads = append(outPayloads, payload)
			s.RtpOffset += uint16(len(outPayloads) - 1)
			return outPayloads, nil
		}
	}

	outPayloads, _, err = s.handleNalTooBigOrFua(outPayloads, seqNum, naluType, rtpHeaderData, nalData)
	if err != nil {
		return nil, err
	}
	s.RtpOffset += uint16(len(outPayloads) - 1)
	return outPayloads, nil
}

func (s *H264spreader) handleNalTooBigOrFua(cumulRtp [][]byte, seqNum uint16, naluType byte, rtpHeader []byte, nalData []byte) ([][]byte, uint16, error) {
	switch {
	case naluType == stapaNALUType:
		return s.explodeStapA(cumulRtp, seqNum, rtpHeader, nalData)
	case naluType == fuaNALUType:
		return s.spreadFua(cumulRtp, seqNum, rtpHeader, nalData)
	case naluType == stapbNALUType || naluType == mtap16NALUType || naluType == mtap24NALUType || naluType == fubNALUType:
		return nil, seqNum, fmt.Errorf("DON or MTAP are not supported")
	default:
		return s.spreadSingleNalToFua(cumulRtp, seqNum, rtpHeader, nalData)
	}
}

//relying on continuous seq number & start/end FU bits to sync ourselve, so not looking at RtpTimestamp
func (s *H264spreader) spreadFua(cumulRtp [][]byte, firtSeqNum uint16, rtpHeader []byte, fua []byte) ([][]byte, uint16, error) {
	seqNum := firtSeqNum
	if s.fuInProgress != nil {
		expectedSeq := s.fuInProgress.LastSeq + 1
		if firtSeqNum != expectedSeq {
			cumulRtp, seqNum = s.flushFuPending(cumulRtp, seqNum)
			//restart over clean (recurse)
			return s.spreadFua(cumulRtp, seqNum, rtpHeader, fua)
		}
	}

	entryMarker := rtpHeader[rtpMPtOffset] & rtpMarkerBitMask
	rtpHeader[rtpMPtOffset] &= ^rtpMarkerBitMask

	lenRtpHeader := len(rtpHeader)
	if s.fuInProgress == nil {
		rtpHeaderCpy := make([]byte, lenRtpHeader)
		copy(rtpHeaderCpy, rtpHeader)
		s.fuInProgress = &fuInProgress{
			LastSeq:   seqNum,
			RtpHeader: rtpHeaderCpy,
			Trailing:  nil,
		}
		s.fuInProgress.FuStartBytes[fuaIndicatorOffset] = fua[fuaIndicatorOffset]
		s.fuInProgress.FuStartBytes[fuaHeaderOffest] = fua[fuaHeaderOffest] & (^fuEndBitmask)
	}

	var lastFuHeader *byte
	mustFinish := (fua[fuaHeaderOffest] & fuEndBitmask) != 0
	reqSubSize := s.Mtu - lenRtpHeader - fuaOverhead
	newData := fua[fuaOverhead:]
	currentDataSize := len(s.fuInProgress.Trailing) + len(newData)
	for currentDataSize > reqSubSize || (mustFinish && currentDataSize > 0) {
		bufSize := min(s.Mtu, lenRtpHeader+fuaOverhead+currentDataSize)
		rtp := make([]byte, bufSize)
		binary.BigEndian.PutUint16(rtpHeader[rtpSeqNumOffset:rtpSeqNumOffset+rtpSeqNumLength], seqNum)
		copy(rtp, rtpHeader)
		copy(rtp[lenRtpHeader:], s.fuInProgress.FuStartBytes[:])
		lastFuHeader = &rtp[lenRtpHeader+1]

		lenTrailing := len(s.fuInProgress.Trailing)
		if lenTrailing > 0 {
			copy(rtp[lenRtpHeader+fuaOverhead:], s.fuInProgress.Trailing)
			s.fuInProgress.Trailing = nil
		}
		toCopyFromNew := min(reqSubSize-lenTrailing, len(newData))
		if toCopyFromNew > 0 {
			copy(rtp[lenRtpHeader+fuaOverhead+lenTrailing:], newData[:toCopyFromNew])
			newData = newData[toCopyFromNew:]
		}

		cumulRtp = append(cumulRtp, rtp)

		s.fuInProgress.FuStartBytes[fuaHeaderOffest] &= ^fuStartBitmask
		s.fuInProgress.LastSeq = seqNum
		seqNum += 1
		currentDataSize = len(newData)
	}

	if mustFinish {
		*lastFuHeader |= fuEndBitmask
		s.fuInProgress = nil
	} else {
		copy(s.trailingBuf, newData)
		s.fuInProgress.Trailing = s.trailingBuf[:len(newData)]
	}

	cumulRtp[len(cumulRtp)-1][rtpMPtOffset] |= entryMarker
	return cumulRtp, seqNum, nil
}

func (s *H264spreader) flushFuPending(cumulRtp [][]byte, entrySeq uint16) ([][]byte, uint16) {
	seqNum := entrySeq
	fuInProgress := s.fuInProgress
	s.fuInProgress = nil
	if fuInProgress != nil && len(fuInProgress.Trailing) > 0 {
		lenPrevRtpHeader := len(fuInProgress.RtpHeader)
		rtp := make([]byte, lenPrevRtpHeader+fuaOverhead+len(fuInProgress.Trailing))
		newSeq := fuInProgress.LastSeq + 1
		binary.BigEndian.PutUint16(fuInProgress.RtpHeader[rtpSeqNumOffset:rtpSeqNumOffset+rtpSeqNumLength], newSeq)
		fuInProgress.FuStartBytes[fuaHeaderOffest] &= ^(fuStartBitmask | fuEndBitmask) //can't have trailing if was 'ending' before
		copy(rtp, fuInProgress.RtpHeader)
		copy(rtp[lenPrevRtpHeader:], fuInProgress.FuStartBytes[:])
		copy(rtp[lenPrevRtpHeader+fuaOverhead:], fuInProgress.Trailing)

		seqNum += 1
		return append(cumulRtp, rtp), seqNum
	}

	return cumulRtp, seqNum
}
func (s *H264spreader) spreadSingleNalToFua(cumulRtp [][]byte, firtSeqNum uint16, rtpHeader []byte, nal []byte) ([][]byte, uint16, error) {
	entryMarker := rtpHeader[rtpMPtOffset] & rtpMarkerBitMask
	rtpHeader[rtpMPtOffset] &= ^rtpMarkerBitMask
	naluType := nal[nalUnitTypeOffset] & naluTypeBitmask
	fuHeader := naluType | fuStartBitmask
	fuIndicator := (nal[nalUnitTypeOffset] ^ naluTypeBitmask) | fuaNALUType
	lenRtpHeader := len(rtpHeader)
	reqSubSize := s.Mtu - lenRtpHeader - fuaOverhead

	// rfc6184:
	// The NAL unit type octet of the fragmented NAL unit is not included as such in the fragmentation unit payload,
	// but rather the information of the NAL unit type octet of the fragmented NAL unit is conveyed in the F and NRI
	// fields of the FU indicator octet of the fragmentation unit and in the type field of the FU header.
	nalWithoutHeader := nal[nalUnitTypeSize:]
	chunks := sliceTo(reqSubSize, nalWithoutHeader)
	nbChunks := len(chunks)
	buf := make([]byte, len(nalWithoutHeader)+((fuaOverhead+lenRtpHeader)*nbChunks))
	offset := 0
	seqNum := firtSeqNum
	var lastFuHeader *byte
	for _, chunk := range chunks {
		cumulRtp = append(cumulRtp, buf[offset:offset+lenRtpHeader+fuaOverhead+len(chunk)])
		binary.BigEndian.PutUint16(rtpHeader[rtpSeqNumOffset:rtpSeqNumOffset+rtpSeqNumLength], seqNum)
		copy(buf[offset:], rtpHeader)
		offset += lenRtpHeader
		buf[offset] = fuIndicator
		offset += 1
		buf[offset] = fuHeader
		lastFuHeader = &buf[offset]
		offset += 1
		copy(buf[offset:], chunk)
		offset += len(chunk)

		seqNum += 1
		fuHeader &= ^fuStartBitmask
	}
	*lastFuHeader |= fuEndBitmask
	cumulRtp[len(cumulRtp)-1][rtpMPtOffset] |= entryMarker
	return cumulRtp, seqNum, nil
}

func (s *H264spreader) explodeStapA(cumulRtp [][]byte, firtSeqNum uint16, rtpHeader []byte, stapa []byte) ([][]byte, uint16, error) {
	entryMarker := rtpHeader[rtpMPtOffset] & rtpMarkerBitMask
	rtpHeader[rtpMPtOffset] &= ^rtpMarkerBitMask
	lenRtpHeader := len(rtpHeader)
	maxSize := s.Mtu - lenRtpHeader
	currOffset := int(stapaHeaderSize)
	lenStapA := len(stapa)
	seqNum := firtSeqNum
	var err error
	for currOffset < lenStapA {
		naluSize := int(binary.BigEndian.Uint16(stapa[currOffset:]))
		currOffset += stapaNALULengthSize

		if lenStapA < currOffset+naluSize {
			return nil, seqNum, fmt.Errorf("STAP-A declared size(%d) is larger than buffer(%d)", naluSize, lenStapA-currOffset)
		}

		subNal := stapa[currOffset : currOffset+naluSize]
		currOffset += naluSize
		if naluSize <= maxSize {
			rtp := make([]byte, lenRtpHeader+naluSize)
			binary.BigEndian.PutUint16(rtpHeader[rtpSeqNumOffset:rtpSeqNumOffset+rtpSeqNumLength], seqNum)
			copy(rtp, rtpHeader)
			copy(rtp[lenRtpHeader:], subNal)
			cumulRtp = append(cumulRtp, rtp)
			seqNum += 1
		} else {
			cumulRtp, seqNum, err = s.spreadSingleNalToFua(cumulRtp, seqNum, rtpHeader, subNal)
			if err != nil {
				return nil, seqNum, err
			}
		}
	}

	cumulRtp[len(cumulRtp)-1][rtpMPtOffset] |= entryMarker
	return cumulRtp, seqNum, nil
}

func sliceTo(reqSize int, data []byte) [][]byte {
	chunkNb := (len(data) + reqSize - 1) / reqSize
	chunks := make([][]byte, chunkNb)
	for i := 0; i < (chunkNb - 1); i++ {
		rangeStart := i * reqSize
		chunks[i] = data[rangeStart : rangeStart+reqSize]
	}
	chunks[chunkNb-1] = data[(chunkNb-1)*reqSize:]
	return chunks
}
