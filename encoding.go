package rtp

// These exist because binary.BigEndian writes to a slice instead of appending.

// Appends a uint16 to a slice in big endian order.
func appendUint16(buf []byte, v uint16) []byte {
	return append(buf, byte(v>>8), byte(v))
}

// Appends a uint32 to a slice in big endian order.
func appendUint32(buf []byte, v uint32) []byte {
	return append(buf, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}
