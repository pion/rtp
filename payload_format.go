// SPDX-FileCopyrightText: 2026 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package rtp

// MediaFormat describes the bitstream or sample representation passed to a payload format.
type MediaFormat string

// PacketizeContext contains the RTP state and negotiated parameters available
// to a payload-format packetizer.
type PacketizeContext struct {
	MTU         int
	PayloadType uint8
	SSRC        uint32
	Timestamp   uint32
	Sequencer   Sequencer
	Params      any
}

// MediaSample is the media input passed to a payload-format packetizer.
type MediaSample struct {
	Payload  []byte
	Duration uint32
	Format   MediaFormat
	Metadata any
}

// ExtensionWrite applies a payload-format-owned RTP header extension.
type ExtensionWrite interface {
	Apply(*Header) error
}

// PayloadFragment is an RTP payload fragment plus the RTP header semantics
// owned by the payload format.
type PayloadFragment struct {
	Payload         []byte
	Marker          bool
	TimestampOffset uint32
	Extensions      []ExtensionWrite
	MutateHeader    func(*Header) error
}

// PayloadFormatPacketizer packetizes media samples into RTP payload fragments.
type PayloadFormatPacketizer interface {
	Packetize(ctx PacketizeContext, sample MediaSample, emit func(PayloadFragment) error) error
	Flush(ctx PacketizeContext, emit func(PayloadFragment) error) error
	Reset()
}

// PacketView exposes the RTP header and payload to payload-format depacketizers.
type PacketView struct {
	Header  *Header
	Payload []byte
}

// PacketInfo describes payload-format sample boundaries and parsed packet metadata.
type PacketInfo struct {
	StartsSample bool
	EndsSample   bool
	KeyFrame     bool
	Metadata     any
}

// PayloadFormatDepacketizer inspects RTP packets and appends their media bytes
// to a sample being assembled.
type PayloadFormatDepacketizer interface {
	Inspect(packet PacketView) (PacketInfo, error)
	AppendToSample(dst []byte, packet PacketView) ([]byte, error)
	Reset()
}

// LegacyPayloaderAdapter adapts a legacy Payloader to PayloadFormatPacketizer.
type LegacyPayloaderAdapter struct {
	Payloader Payloader
}

// Packetize adapts the legacy Payloader Payload method to PayloadFormatPacketizer.
func (a LegacyPayloaderAdapter) Packetize(
	ctx PacketizeContext,
	sample MediaSample,
	emit func(PayloadFragment) error,
) error {
	payloads := a.Payloader.Payload(legacyPayloaderMTU(ctx.MTU), sample.Payload)
	for i, payload := range payloads {
		if err := emit(PayloadFragment{
			Payload: payload,
			Marker:  i == len(payloads)-1,
		}); err != nil {
			return err
		}
	}

	return nil
}

// Flush adapts legacy payloaders, which do not buffer pending payload fragments.
func (a LegacyPayloaderAdapter) Flush(_ PacketizeContext, _ func(PayloadFragment) error) error {
	return nil
}

// Reset adapts legacy payloaders, which do not expose reset behavior.
func (a LegacyPayloaderAdapter) Reset() {}

// LegacyDepacketizerAdapter adapts a legacy Depacketizer to PayloadFormatDepacketizer.
type LegacyDepacketizerAdapter struct {
	Depacketizer Depacketizer
}

// Inspect adapts the legacy partition boundary methods to PayloadFormatDepacketizer.
func (a LegacyDepacketizerAdapter) Inspect(packet PacketView) (PacketInfo, error) {
	marker := false
	if packet.Header != nil {
		marker = packet.Header.Marker
	}

	return PacketInfo{
		StartsSample: a.Depacketizer.IsPartitionHead(packet.Payload),
		EndsSample:   a.Depacketizer.IsPartitionTail(marker, packet.Payload),
	}, nil
}

// AppendToSample adapts the legacy Unmarshal method to PayloadFormatDepacketizer.
func (a LegacyDepacketizerAdapter) AppendToSample(dst []byte, packet PacketView) ([]byte, error) {
	media, err := a.Depacketizer.Unmarshal(packet.Payload)
	if err != nil {
		return dst, err
	}

	return append(dst, media...), nil
}

// Reset resets the wrapped depacketizer when it exposes reset behavior.
func (a LegacyDepacketizerAdapter) Reset() {
	if resetter, ok := a.Depacketizer.(interface{ Reset() }); ok {
		resetter.Reset()
	}
}

func legacyPayloaderMTU(mtu int) uint16 {
	const (
		baseRTPHeaderSize = csrcOffset
		maxUint16         = 1<<16 - 1
	)

	payloadMTU := mtu - baseRTPHeaderSize
	if payloadMTU <= 0 {
		return 0
	}
	if payloadMTU > maxUint16 {
		return maxUint16
	}

	return uint16(payloadMTU) //nolint:gosec // payloadMTU is clamped to uint16 above.
}

var (
	_ PayloadFormatPacketizer   = LegacyPayloaderAdapter{}
	_ PayloadFormatDepacketizer = LegacyDepacketizerAdapter{}
)
