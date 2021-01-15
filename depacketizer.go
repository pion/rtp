package rtp

// Depacketizer depacketizes a RTP payload, removing any RTP specific data from the payload
type Depacketizer interface {
	IsDetectedFinalPacketInSequence(*Packet) bool
	Unmarshal(packet []byte) ([]byte, error)
}
