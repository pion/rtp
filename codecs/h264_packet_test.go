// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package codecs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestH264Payloader_Payload(t *testing.T) {
	pck := H264Payloader{}
	smallpayload := []byte{0x90, 0x90, 0x90}
	multiplepayload := []byte{0x00, 0x00, 0x01, 0x90, 0x00, 0x00, 0x01, 0x90}
	mixednalupayload := []byte{
		0x00, 0x00, 0x01, 0x90,
		0x00, 0x00, 0x00, 0x01, 0x90,
		0x00, 0x00, 0x01, 0x90,
		0x00, 0x00, 0x00, 0x01, 0x90,
	}

	largepayload := []byte{
		0x00, 0x00, 0x01, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15,
	}
	largePayloadPacketized := [][]byte{
		{0x1c, 0x80, 0x01, 0x02, 0x03},
		{0x1c, 0x00, 0x04, 0x05, 0x06},
		{0x1c, 0x00, 0x07, 0x08, 0x09},
		{0x1c, 0x00, 0x10, 0x11, 0x12},
		{0x1c, 0x40, 0x13, 0x14, 0x15},
	}

	// Positive MTU, nil payload
	res := pck.Payload(1, nil)
	assert.Len(t, res, 0, "Generated payload should be empty")

	// Positive MTU, empty payload
	res = pck.Payload(1, []byte{})
	assert.Len(t, res, 0, "Generated payload should be empty")

	// Positive MTU, empty NAL
	res = pck.Payload(1, []byte{0x00, 0x00, 0x01})
	assert.Len(t, res, 0, "Generated payload should be empty")

	// Negative MTU, small payload
	res = pck.Payload(0, smallpayload)
	assert.Len(t, res, 0, "Generated payload should be empty")

	// 0 MTU, small payload
	res = pck.Payload(0, smallpayload)
	assert.Len(t, res, 0, "Generated payload should be empty")

	// Positive MTU, small payload
	res = pck.Payload(1, smallpayload)
	assert.Len(t, res, 0, "Generated payload should be empty")

	// Positive MTU, small payload
	res = pck.Payload(5, smallpayload)
	assert.Len(t, res, 1, "Generated payload should be the 1")
	assert.Len(t, smallpayload, len(res[0]), "Generated payload should be the same size as original payload size")

	// Multiple NALU in a single payload
	res = pck.Payload(5, multiplepayload)
	assert.Len(t, res, 2, "2 nal units should be broken out")
	for i := 0; i < 2; i++ {
		assert.Lenf(t, res[i], 1, "Payload %d of 2 is packed incorrectly", i+1)
	}

	// Multiple NALU in a single payload with 3-byte and 4-byte start sequences
	res = pck.Payload(5, mixednalupayload)
	assert.Len(t, res, 4, "4 nal units should be broken out")
	for i := 0; i < 4; i++ {
		assert.Lenf(t, res[i], 1, "Payload %d of 4 is packed incorrectly", i+1)
	}

	// Large Payload split across multiple RTP Packets
	res = pck.Payload(5, largepayload)
	assert.Equal(t, largePayloadPacketized, res, "FU-A packetization failed")

	// Nalu type 9 or 12
	res = pck.Payload(5, []byte{0x09, 0x00, 0x00})
	assert.Len(t, res, 0, "Generated payload should be empty")
}

func TestH264Packet_Unmarshal(t *testing.T) {
	singlePayload := []byte{0x90, 0x90, 0x90}
	singlePayloadUnmarshaled := []byte{0x00, 0x00, 0x00, 0x01, 0x90, 0x90, 0x90}
	singlePayloadUnmarshaledAVC := []byte{0x00, 0x00, 0x00, 0x03, 0x90, 0x90, 0x90}

	largepayload := []byte{
		0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15,
	}
	largepayloadAVC := []byte{
		0x00, 0x00, 0x00, 0x10, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15,
	}
	largePayloadPacketized := [][]byte{
		{0x1c, 0x80, 0x01, 0x02, 0x03},
		{0x1c, 0x00, 0x04, 0x05, 0x06},
		{0x1c, 0x00, 0x07, 0x08, 0x09},
		{0x1c, 0x00, 0x10, 0x11, 0x12},
		{0x1c, 0x40, 0x13, 0x14, 0x15},
	}

	singlePayloadMultiNALU := []byte{
		0x78, 0x00, 0x0f, 0x67, 0x42, 0xc0, 0x1f, 0x1a, 0x32, 0x35, 0x01, 0x40, 0x7a, 0x40,
		0x3c, 0x22, 0x11, 0xa8, 0x00, 0x05, 0x68, 0x1a, 0x34, 0xe3, 0xc8,
	}
	singlePayloadMultiNALUUnmarshaled := []byte{
		0x00, 0x00, 0x00, 0x01, 0x67, 0x42, 0xc0, 0x1f, 0x1a, 0x32, 0x35, 0x01, 0x40, 0x7a,
		0x40, 0x3c, 0x22, 0x11, 0xa8, 0x00, 0x00, 0x00, 0x01, 0x68, 0x1a, 0x34, 0xe3, 0xc8,
	}
	singlePayloadMultiNALUUnmarshaledAVC := []byte{
		0x00, 0x00, 0x00, 0x0f, 0x67, 0x42, 0xc0, 0x1f, 0x1a, 0x32, 0x35, 0x01, 0x40, 0x7a,
		0x40, 0x3c, 0x22, 0x11, 0xa8, 0x00, 0x00, 0x00, 0x05, 0x68, 0x1a, 0x34, 0xe3, 0xc8,
	}
	singlePayloadWithBrokenSecondNALU := []byte{
		0x78, 0x00, 0x0f, 0x67, 0x42, 0xc0, 0x1f, 0x1a, 0x32, 0x35, 0x01, 0x40, 0x7a, 0x40,
		0x3c, 0x22, 0x11, 0xa8, 0x00,
	}
	singlePayloadWithBrokenSecondNALUUnmarshaled := []byte{
		0x00, 0x00, 0x00, 0x01, 0x67, 0x42, 0xc0, 0x1f, 0x1a, 0x32, 0x35, 0x01, 0x40, 0x7a,
		0x40, 0x3c, 0x22, 0x11, 0xa8,
	}
	singlePayloadWithBrokenSecondUnmarshaledAVC := []byte{
		0x00, 0x00, 0x00, 0x0f, 0x67, 0x42, 0xc0, 0x1f, 0x1a, 0x32, 0x35, 0x01, 0x40, 0x7a,
		0x40, 0x3c, 0x22, 0x11, 0xa8,
	}

	incompleteSinglePayloadMultiNALU := []byte{
		0x78, 0x00, 0x0f, 0x67, 0x42, 0xc0, 0x1f, 0x1a, 0x32, 0x35, 0x01, 0x40, 0x7a, 0x40,
		0x3c, 0x22, 0x11,
	}

	pkt := H264Packet{}
	avcPkt := H264Packet{IsAVC: true}
	_, err := pkt.Unmarshal(nil)
	assert.Error(t, err, "Unmarshal did not fail on nil payload")

	_, err = pkt.Unmarshal([]byte{})
	assert.Error(t, err, "Unmarshal did not fail on []byte{}")

	_, err = pkt.Unmarshal([]byte{0xFC})
	assert.Error(t, err, "Unmarshal accepted a FU-A packet that is too small for a payload and header")

	_, err = pkt.Unmarshal([]byte{0x0A})
	assert.NoError(t, err, "Unmarshaling end of sequence(NALU Type : 10) should succeed")

	_, err = pkt.Unmarshal([]byte{0xFF, 0x00, 0x00})
	assert.Error(t, err, "Unmarshal accepted a packet with a NALU Type we don't handle")

	_, err = pkt.Unmarshal(incompleteSinglePayloadMultiNALU)
	assert.Error(t, err, "Unmarshal accepted a STAP-A packet with insufficient data")

	res, err := pkt.Unmarshal(singlePayload)
	assert.NoError(t, err)
	assert.Equal(t, singlePayloadUnmarshaled, res)

	res, err = avcPkt.Unmarshal(singlePayload)
	assert.NoError(t, err)
	assert.Equal(t, singlePayloadUnmarshaledAVC, res)

	largePayloadResult := []byte{}
	for i := range largePayloadPacketized {
		res, err = pkt.Unmarshal(largePayloadPacketized[i])
		assert.NoError(t, err)
		largePayloadResult = append(largePayloadResult, res...)
	}
	assert.Equal(t, largepayload, largePayloadResult)

	largePayloadResultAVC := []byte{}
	for i := range largePayloadPacketized {
		res, err = avcPkt.Unmarshal(largePayloadPacketized[i])
		assert.NoError(t, err)
		largePayloadResultAVC = append(largePayloadResultAVC, res...)
	}
	assert.Equal(t, largepayloadAVC, largePayloadResultAVC)

	res, err = pkt.Unmarshal(singlePayloadMultiNALU)
	assert.NoError(t, err)
	assert.Equal(t, singlePayloadMultiNALUUnmarshaled, res)

	res, err = avcPkt.Unmarshal(singlePayloadMultiNALU)
	assert.NoError(t, err)
	assert.Equal(t, singlePayloadMultiNALUUnmarshaledAVC, res)

	res, err = pkt.Unmarshal(singlePayloadWithBrokenSecondNALU)
	assert.NoError(t, err)
	assert.Equal(t, singlePayloadWithBrokenSecondNALUUnmarshaled, res)

	res, err = avcPkt.Unmarshal(singlePayloadWithBrokenSecondNALU)
	assert.NoError(t, err)
	assert.Equal(t, singlePayloadWithBrokenSecondUnmarshaledAVC, res)
}

func TestH264IsPartitionHead(t *testing.T) {
	h264 := H264Packet{}

	assert.False(t, h264.IsPartitionHead(nil), "nil must not be a partition head")
	assert.False(t, h264.IsPartitionHead([]byte{}), "empty nalu must not be a partition head")

	singleNalu := []byte{1, 0}
	assert.True(t, h264.IsPartitionHead(singleNalu), "single nalu must be a partition head")

	stapaNalu := []byte{stapaNALUType, 0}
	assert.True(t, h264.IsPartitionHead(stapaNalu), "stapa nalu must be a partition head")

	fuaStartNalu := []byte{fuaNALUType, fuStartBitmask}
	assert.True(t, h264.IsPartitionHead(fuaStartNalu), "fua start nalu must be a partition head")

	fuaEndNalu := []byte{fuaNALUType, fuEndBitmask}
	assert.False(t, h264.IsPartitionHead(fuaEndNalu), "fua end nalu must not be a partition head")

	fubStartNalu := []byte{fubNALUType, fuStartBitmask}
	assert.True(t, h264.IsPartitionHead(fubStartNalu), "fub start nalu must be a partition head")

	fubEndNalu := []byte{fubNALUType, fuEndBitmask}
	assert.False(t, h264.IsPartitionHead(fubEndNalu), "fub end nalu must not be a partition head")
}

func TestH264Payloader_Payload_SPS_and_PPS_handling(t *testing.T) {
	pck := H264Payloader{}
	expected := [][]byte{
		{0x78, 0x00, 0x03, 0x07, 0x00, 0x01, 0x00, 0x03, 0x08, 0x02, 0x03},
		{0x05, 0x04, 0x05},
	}

	// When packetizing SPS and PPS are emitted with following NALU
	res := pck.Payload(1500, []byte{0x07, 0x00, 0x01})
	assert.Len(t, res, 0, "Generated payload should be empty")

	res = pck.Payload(1500, []byte{0x08, 0x02, 0x03})
	assert.Len(t, res, 0, "Generated payload should be empty")
	assert.Equal(t, expected, pck.Payload(1500, []byte{0x05, 0x04, 0x05}), "SPS and PPS aren't packed together")
}

func TestH264Payloader_Payload_SPS_and_PPS_handling_no_stapA(t *testing.T) {
	pck := H264Payloader{}
	pck.DisableStapA = true

	expectedSps := []byte{0x07, 0x00, 0x01}
	// The SPS is packed as a single NALU
	res := pck.Payload(1500, expectedSps)
	assert.Len(t, res, 1, "Generated payload should not be empty")
	assert.Equal(t, expectedSps, res[0], "SPS has not been packed correctly")
	// The PPS is packed as a single NALU
	expectedPps := []byte{0x08, 0x02, 0x03}
	res = pck.Payload(1500, expectedPps)
	assert.Len(t, res, 1, "Generated payload should not be empty")
	assert.Equal(t, expectedPps, res[0], "PPS has not been packed correctly")
}
