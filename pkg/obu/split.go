package obu

import (
	"errors"
	"io"
)

type OBUType uint8

type OBUHeader struct {
	ObuType         OBUType // 1-4
	ExtensionFlag   bool    // 5
	HasSizeField    bool    // 6
	ExtensionHeader byte    // 8-16, if ExtensionFlag=1
}

type OBUReader struct {
	buffer []byte
	idx    uint
	size   uint
}

var (
	errInvalidObuData = errors.New("invalid obu data")
	errForbidenBit    = errors.New("forbidenBit=1 in OBU Header")
)

const (
	forbiddenBitMask  = uint8(0b10000000)
	typeMask          = uint8(0b01111000)
	typeShift         = 3
	extensionFlagMask = uint8(0b00000100)
	hasSizeFlagMask   = uint8(0b00000010)
	reserved1BitMask  = uint8(0b00000001)
)

const (
	OBU_SEQUENCE_HEADER        OBUType = 1
	OBU_TEMPORAL_DELIMITER     OBUType = 2
	OBU_FRAME_HEADER           OBUType = 3
	OBU_TILE_GROUP             OBUType = 4
	OBU_METADATA               OBUType = 5
	OBU_FRAME                  OBUType = 6
	OBU_REDUNDANT_FRAME_HEADER OBUType = 7
	OBU_TILE_LIST              OBUType = 8
	OBU_PADDING                OBUType = 15
	// Others are Reserved
)

type OBU struct {
	Header *OBUHeader
	Data   []byte
}

func (h *OBUHeader) Marshal() []byte {
	// header size
	size := 1
	if h.ExtensionFlag {
		size = 2
	}
	data := make([]byte, size)
	// Type
	data[0] |= byte(h.ObuType << typeShift)
	if h.HasSizeField {
		data[0] |= hasSizeFlagMask
	}
	if h.ExtensionFlag {
		data[0] |= extensionFlagMask
		data[1] = h.ExtensionHeader
	}
	return data
}

func (or *OBUReader) ReadLeb128() (uint, error) {
	val, nread, err := ReadLeb128(or.buffer[or.idx:])
	or.idx += nread
	return val, err
}

func (or *OBUReader) ReadHeader() (header OBUHeader, err error) {
	num := or.buffer[or.idx]
	or.idx += 1
	// Check ForbidenBit
	if num&0x80 != 0 {
		err = errForbidenBit
		return
	}
	header.ObuType = OBUType((num & typeMask) >> typeShift)
	header.ExtensionFlag = (num & extensionFlagMask) != 0
	header.HasSizeField = (num & hasSizeFlagMask) != 0

	if header.ExtensionFlag {
		num = or.buffer[or.idx]
		or.idx += 1
		header.ExtensionHeader = num
	}
	return
}

// read next obu
func (or *OBUReader) ParseNext() (*OBU, error) {
	if or.idx == or.size {
		return nil, io.EOF
	} else if or.idx > or.size {
		return nil, errInvalidObuData
	}
	var obuData OBU
	header, err := or.ReadHeader()
	if err != nil {
		return nil, err
	}
	obuData.Header = &header
	if header.HasSizeField {
		size, err := or.ReadLeb128()
		if err != nil {
			return nil, err
		}
		obuData.Data = or.buffer[or.idx : or.idx+size]
		or.idx += size
	} else {
		obuData.Data = or.buffer[or.idx:]
		or.idx = or.size
	}
	return &obuData, nil
}

func (obu *OBU) Marshal() []byte {
	// https://aomediacodec.github.io/av1-rtp-spec/#45-payload-structure
	// To minimize overhead, the obu_has_size_field flag SHOULD be set to zero in all OBUs.
	data := obu.Header.Marshal()
	if obu.Header.HasSizeField {
		AppendUleb128(data, uint(len(obu.Data)))
	}
	data = append(data, obu.Data...)
	return data
}

// Extract obus from frame data
func SplitOBU(payload []byte) (obus []OBU, err error) {
	reader := OBUReader{buffer: payload, size: uint(len(payload))}
	for {
		obu, err := reader.ParseNext()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		obus = append(obus, *obu)
	}
	return obus, nil
}
