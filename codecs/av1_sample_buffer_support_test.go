package codecs

import (
	"bytes"
	"github.com/pion/rtp"
	"github.com/pion/rtp/pkg/obu"
	"github.com/pion/webrtc/v3/pkg/media/samplebuilder"
	"testing"
	"time"
)

func buildAv1Payload(data byte, padding int) []byte {
	dataSize := 0

	if data > 0 {
		dataSize = 1
	}

	payloadSize := obu.EncodeLEB128(uint(1 + padding + dataSize))
	result := make([]byte, 3+sizeLeb128(payloadSize))

	result[0] = 0 // AV1 RTP header

	offset := 1

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

	result[offset] = 0 // OBU HEADER

	if dataSize > 0 {
		offset++
		result[offset] = data
	}
	return append(result, make([]byte, padding)...)
}
func buildAv1Packages(seqNo *uint16, timestamp *uint32, padding int) []*rtp.Packet {
	s := *seqNo
	t := *timestamp

	*timestamp += 1800
	*seqNo += 5

	return []*rtp.Packet{
		{Header: rtp.Header{SequenceNumber: s, Timestamp: t}, Payload: buildAv1Payload(1, 0)},
		{Header: rtp.Header{SequenceNumber: s + 1, Timestamp: t}, Payload: buildAv1Payload(2, 0)},
		{Header: rtp.Header{SequenceNumber: s + 2, Timestamp: t}, Payload: buildAv1Payload(3, 0)},
		{Header: rtp.Header{SequenceNumber: s + 3, Timestamp: t}, Payload: buildAv1Payload(4, 0)},
		{Header: rtp.Header{SequenceNumber: s + 4, Timestamp: t, Marker: true}, Payload: buildAv1Payload(5, padding)},
	}
}
func TestAV1SampleBufferSupport(t *testing.T) {

	assembledAv1Frame := []byte{2, 1, 1, 2, 1, 2, 2, 1, 3, 2, 1, 4, 2, 1, 5}
	t.Run("AV1 Sample Buffer returning OBU stream", func(t *testing.T) {
		videoStreamBuilder := samplebuilder.New(100, &AV1PacketSampleBufferSupport{}, 90000,
			samplebuilder.WithMaxTimeDelay(time.Millisecond*100))
		var seqNo uint16 = 0
		var timestamp uint32 = 0

		for i := 0; i < 4; i++ {
			for _, pkt := range buildAv1Packages(&seqNo, &timestamp, 0) {
				sample := videoStreamBuilder.Pop()
				if nil != sample {
					if !bytes.Equal(sample.Data, assembledAv1Frame) {
						t.Fatal("issue in unmarshalling")
					}
				}
				videoStreamBuilder.Push(pkt)
			}
		}

		for i := 1; i < 16400; i++ { // check OBU len up to 3 bytes
			for _, pkt := range buildAv1Packages(&seqNo, &timestamp, i) {
				sample := videoStreamBuilder.Pop()
				if nil != sample {

					if !bytes.Equal(sample.Data[0:12], assembledAv1Frame[0:12]) {
						t.Fatal("issue in unmarshalling")
					}

					if sample.Data[len(sample.Data)-i] != 5 {
						t.Fatal("issue in unmarshalling")
					}

					if i > 0 {
						padding := make([]byte, i-1)
						if !bytes.Equal(sample.Data[len(sample.Data)-i+1:], padding) {
							t.Fatal("issue in unmarshalling")
						}
					}
				}
				videoStreamBuilder.Push(pkt)
			}
		}
	})
}

func buildHeaderOnlyAv1Packets(seqNo *uint16, timestamp *uint32, padding int) []*rtp.Packet {
	s := *seqNo
	t := *timestamp

	*timestamp += 1800
	*seqNo += 3

	return []*rtp.Packet{
		// Two header-only OBUs
		{Header: rtp.Header{SequenceNumber: s, Timestamp: t}, Payload: buildAv1Payload(0, 0)},
		{Header: rtp.Header{SequenceNumber: s + 1, Timestamp: t}, Payload: buildAv1Payload(0, 0)},
		{Header: rtp.Header{SequenceNumber: s + 2, Timestamp: t, Marker: true}, Payload: buildAv1Payload(5, padding)},
	}
}

func TestAV1SampleBufferSupport_OBUWIthoutPayload(t *testing.T) {
	assembledAv1Frame := []byte{2, 0, 2, 0, 2, 1, 5}
	t.Run("AV1 Sample Buffer with header only OBU elements", func(t *testing.T) {
		videoStreamBuilder := samplebuilder.New(100, &AV1PacketSampleBufferSupport{}, 90000,
			samplebuilder.WithMaxTimeDelay(time.Millisecond*100))
		var seqNo uint16 = 0
		var timestamp uint32 = 0

		for i := 0; i < 4; i++ {
			for _, pkt := range buildHeaderOnlyAv1Packets(&seqNo, &timestamp, 0) {
				sample := videoStreamBuilder.Pop()
				if nil != sample {
					if !bytes.Equal(sample.Data, assembledAv1Frame) {
						t.Fatal("issue in unmarshalling")
					}
				}
				videoStreamBuilder.Push(pkt)
			}
		}
	})
}
