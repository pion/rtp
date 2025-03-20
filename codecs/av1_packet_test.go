// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package codecs

import (
	"fmt"
	"testing"

	"github.com/pion/rtp/codecs/av1/obu"
	"github.com/stretchr/testify/assert"
)

type testAV1AggregationHeader struct {
	Z, Y, N bool
	W       byte
}

func (t testAV1AggregationHeader) Marshal() []byte {
	header := byte(0)

	if t.Z {
		header |= 0b10000000
	}
	if t.Y {
		header |= 0b01000000
	}
	if t.N {
		header |= 0b00001000
	}
	header |= (t.W << 4) & 0b00110000

	return []byte{header}
}

type testAV1OBUPayload struct {
	Payload           []byte
	Header            *obu.Header
	HasRTPLengthField bool
}

func (t testAV1OBUPayload) Marshal() []byte {
	payload := make([]byte, 0)

	// obu_size_field() leb128()
	var obuSize []byte
	if t.Header != nil && t.Header.HasSizeField {
		obuSize = obu.WriteToLeb128(uint(len(t.Payload)))
	}

	// RTP length field leb128()
	if t.HasRTPLengthField {
		length := len(t.Payload) + len(obuSize)

		if t.Header != nil {
			length += t.Header.Size()
		}

		payload = append(payload, obu.WriteToLeb128(
			uint(length), //nolint:gosec // G115 false positive
		)...)
	}
	if t.Header != nil {
		payload = append(payload, t.Header.Marshal()...)

		if t.Header.HasSizeField {
			payload = append(payload, obuSize...)
		}
	}
	payload = append(payload, t.Payload...)

	return payload
}

type testAV1MultiOBUsPayload []testAV1OBUPayload

func (t testAV1MultiOBUsPayload) Marshal() []byte {
	payload := make([]byte, 0)

	for _, obu := range t {
		payload = append(payload, obu.Marshal()...)
	}

	return payload
}

type testAV1Tests struct {
	Name           string
	MTU            uint16
	InputPayload   []byte
	OutputPayloads [][]byte
}

func testAV1TestRun(t *testing.T, tests []testAV1Tests) {
	t.Helper()
	payloader := &AV1Payloader{}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result := payloader.Payload(test.MTU, test.InputPayload)

			assert.Equal(t, len(test.OutputPayloads), len(result))

			for i := range result {
				assert.Equal(t, test.OutputPayloads[i], result[i])
			}
		})
	}
}

func TestAV1Payloader_ShortMtU(t *testing.T) {
	p := &AV1Payloader{}

	assert.Len(t, p.Payload(0, []byte{0x00, 0x01, 0x18}), 0, "Expected empty payload")
	assert.Len(t, p.Payload(1, []byte{0x00, 0x01, 0x18}), 0, "Expected empty payload")
	// 2 is the minimum MTU for AV1 (aggregate header + 1 byte)
	assert.Greater(t, len(p.Payload(2, []byte{0x00, 0x01, 0x18})), 0)
}

func TestAV1Payloader_SinglePacket(t *testing.T) {
	tests := []testAV1Tests{
		{
			Name: "Single Sequence Header",
			MTU:  1000,
			InputPayload: (testAV1OBUPayload{
				Payload: []byte{
					0x01, 0x02, 0x03, 0x04, 0x05,
				},
				Header: &obu.Header{
					Type:         obu.OBUSequenceHeader,
					HasSizeField: false,
				},
			}).Marshal(),
			OutputPayloads: [][]byte{
				append(
					(testAV1AggregationHeader{
						N: true,
						W: 1,
					}).Marshal(),
					(testAV1OBUPayload{
						Header: &obu.Header{
							Type: obu.OBUSequenceHeader,
						},
						Payload: []byte{
							0x01, 0x02, 0x03, 0x04, 0x05,
						},
					}).Marshal()...,
				),
			},
		},
		{
			Name: "Single Frame",
			MTU:  1000,
			InputPayload: (testAV1OBUPayload{
				Header: &obu.Header{
					Type: obu.OBUFrameHeader,
				},
				Payload: []byte{
					0x01, 0x02, 0x03, 0x04, 0x05,
				},
			}).Marshal(),
			OutputPayloads: [][]byte{
				append(
					(testAV1AggregationHeader{
						W: 1,
					}).Marshal(),
					(testAV1OBUPayload{
						Header: &obu.Header{
							Type: obu.OBUFrameHeader,
						},
						Payload: []byte{
							0x01, 0x02, 0x03, 0x04, 0x05,
						},
					}).Marshal()...,
				),
			},
		},
		{
			"Should remove size field",
			1000,
			(testAV1OBUPayload{
				Header: &obu.Header{
					Type:         obu.OBUFrameHeader,
					HasSizeField: true,
				},
				Payload: []byte{
					0x01, 0x02, 0x03, 0x04, 0x05,
				},
			}).Marshal(),
			[][]byte{
				append(
					(testAV1AggregationHeader{
						W: 1,
					}).Marshal(),
					(testAV1OBUPayload{
						Header: &obu.Header{
							Type: obu.OBUFrameHeader,
						},
						Payload: []byte{
							0x01, 0x02, 0x03, 0x04, 0x05,
						},
					}).Marshal()...,
				),
			},
		},
		{
			Name: "Should Skip Tile List",
			MTU:  1000,
			InputPayload: (testAV1OBUPayload{
				Header: &obu.Header{
					Type: obu.OBUTileList,
				},
				Payload: []byte{
					0x01, 0x02, 0x03, 0x04, 0x05,
				},
			}).Marshal(),
		},
	}

	testAV1TestRun(t, tests)
}

//nolint:maintidx
func TestAV1Payloader_MultipleOBUsInSinglePacket(t *testing.T) {
	tests := []testAV1Tests{
		{
			Name: "Should pack two OBUs in a single packet with W=2",
			MTU:  1000,
			InputPayload: (testAV1MultiOBUsPayload{
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x06, 0x07, 0x08, 0x09, 0x0A},
				},
			}).Marshal(),
			OutputPayloads: [][]byte{
				append(
					(testAV1AggregationHeader{
						W: 2,
					}).Marshal(),
					(testAV1MultiOBUsPayload{
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x01, 0x02, 0x03, 0x04, 0x05},
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							Payload: []byte{0x06, 0x07, 0x08, 0x09, 0x0A},
						},
					}).Marshal()...,
				),
			},
		},
		{
			Name: "Should pack three OBUs in a single packet with W=3",
			MTU:  1000,
			InputPayload: (testAV1MultiOBUsPayload{
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x06, 0x07, 0x08, 0x09, 0x0A},
				},
				{
					Header: &obu.Header{
						Type: obu.OBUFrameHeader,
					},
					Payload: []byte{0x0B, 0x0C, 0x0D, 0x0E, 0x0F},
				},
			}).Marshal(),
			OutputPayloads: [][]byte{
				append(
					(testAV1AggregationHeader{
						W: 3,
					}).Marshal(),
					(testAV1MultiOBUsPayload{
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x01, 0x02, 0x03, 0x04, 0x05},
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x06, 0x07, 0x08, 0x09, 0x0A},
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							Payload: []byte{0x0B, 0x0C, 0x0D, 0x0E, 0x0F},
						},
					}).Marshal()...,
				),
			},
		},
		{
			Name: "Should pack four OBUs in a single packet with W=0",
			MTU:  1000,
			InputPayload: (testAV1MultiOBUsPayload{
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x06, 0x07, 0x08, 0x09, 0x0A},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x0B, 0x0C, 0x0D, 0x0E, 0x0F},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x10, 0x11, 0x12, 0x13, 0x14},
				},
			}).Marshal(),
			OutputPayloads: [][]byte{
				append(
					(testAV1AggregationHeader{}).Marshal(),
					(testAV1MultiOBUsPayload{
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x01, 0x02, 0x03, 0x04, 0x05},
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x06, 0x07, 0x08, 0x09, 0x0A},
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x0B, 0x0C, 0x0D, 0x0E, 0x0F},
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x10, 0x11, 0x12, 0x13, 0x14},
						},
					}).Marshal()...,
				),
			},
		},
		{
			Name: "Should pack five OBUs in a single packet with W=0",
			MTU:  1000,
			InputPayload: (testAV1MultiOBUsPayload{
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x06, 0x07, 0x08, 0x09, 0x0A},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x0B, 0x0C, 0x0D, 0x0E, 0x0F},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x10, 0x11, 0x12, 0x13, 0x14},
				},
				{
					Header: &obu.Header{
						Type: obu.OBUFrameHeader,
					},
					Payload: []byte{0x15, 0x16, 0x17, 0x18, 0x19},
				},
			}).Marshal(),
			OutputPayloads: [][]byte{
				append(
					(testAV1AggregationHeader{}).Marshal(),
					(testAV1MultiOBUsPayload{
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x01, 0x02, 0x03, 0x04, 0x05},
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x06, 0x07, 0x08, 0x09, 0x0A},
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x0B, 0x0C, 0x0D, 0x0E, 0x0F},
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x10, 0x11, 0x12, 0x13, 0x14},
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x15, 0x16, 0x17, 0x18, 0x19},
						},
					}).Marshal()...,
				),
			},
		},
		{
			Name: "Should read last obu without obu_size_field",
			MTU:  1000,
			InputPayload: (testAV1MultiOBUsPayload{
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
				},
				{
					Header: &obu.Header{
						Type: obu.OBUFrameHeader,
					},
					Payload: []byte{0x06, 0x07, 0x08, 0x09, 0x0A},
				},
			}).Marshal(),
			OutputPayloads: [][]byte{
				append(
					(testAV1AggregationHeader{
						W: 2,
					}).Marshal(),
					(testAV1MultiOBUsPayload{
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x01, 0x02, 0x03, 0x04, 0x05},
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							Payload: []byte{0x06, 0x07, 0x08, 0x09, 0x0A},
						},
					}).Marshal()...,
				),
			},
		},
	}

	testAV1TestRun(t, tests)
}

//nolint:maintidx
func TestAV1Payloader_HandleMTUBasedFragmentation(t *testing.T) {
	tests := []testAV1Tests{
		{
			Name: "Should pack two OBUs in a single packet with W=1 for each",
			MTU:  7,
			InputPayload: (testAV1MultiOBUsPayload{
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x06, 0x07, 0x08, 0x09, 0x0A},
				},
			}).Marshal(),
			OutputPayloads: [][]byte{
				append(
					(testAV1AggregationHeader{
						W: 1,
					}).Marshal(),
					(testAV1OBUPayload{
						Header: &obu.Header{
							Type: obu.OBUFrameHeader,
						},
						Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
					}).Marshal()...,
				),
				append(
					(testAV1AggregationHeader{
						W: 1,
					}).Marshal(),
					(testAV1OBUPayload{
						Header: &obu.Header{
							Type: obu.OBUFrameHeader,
						},
						Payload: []byte{0x06, 0x07, 0x08, 0x09, 0x0A},
					}).Marshal()...,
				),
			},
		},
		{
			Name: "Should split OBU over two packets with each W=1",
			MTU:  7,
			InputPayload: (testAV1MultiOBUsPayload{
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{
						0x01, 0x02, 0x03, 0x04, 0x05,
						0x06, 0x07, 0x08, 0x09, 0x0A,
					},
				},
			}).Marshal(),
			OutputPayloads: [][]byte{
				append(
					(testAV1AggregationHeader{
						W: 1,
						Y: true,
					}).Marshal(),
					(testAV1OBUPayload{
						Header: &obu.Header{
							Type: obu.OBUFrameHeader,
						},
						Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
					}).Marshal()...,
				),
				append(
					(testAV1AggregationHeader{
						W: 1,
						Z: true,
					}).Marshal(),
					(testAV1OBUPayload{
						Payload: []byte{0x06, 0x07, 0x08, 0x09, 0x0A},
					}).Marshal()...,
				),
			},
		},
		{
			Name: "Should split OBU over three packets with each W=1",
			MTU:  7,
			InputPayload: (testAV1MultiOBUsPayload{
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{
						0x01, 0x02, 0x03, 0x04, 0x05,
						0x06, 0x07, 0x08, 0x09, 0x0A,
						0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
					},
				},
			}).Marshal(),
			OutputPayloads: [][]byte{
				append(
					(testAV1AggregationHeader{
						W: 1,
						Y: true,
					}).Marshal(),
					(testAV1OBUPayload{
						Header: &obu.Header{
							Type: obu.OBUFrameHeader,
						},
						Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
					}).Marshal()...,
				),
				append(
					(testAV1AggregationHeader{
						W: 1,
						Z: true,
						Y: true,
					}).Marshal(),
					(testAV1OBUPayload{
						Payload: []byte{0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B},
					}).Marshal()...,
				),
				append(
					(testAV1AggregationHeader{
						W: 1,
						Z: true,
					}).Marshal(),
					(testAV1OBUPayload{
						Payload: []byte{0x0C, 0x0D, 0x0E, 0x0F},
					}).Marshal()...,
				),
			},
		},
		{
			Name: "Should split OBU over three packets and adds extra packet",
			MTU:  7,
			InputPayload: (testAV1MultiOBUsPayload{
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{
						0x01, 0x02, 0x03, 0x04,
						0x05, 0x06, 0x07, 0x08,
						0x09, 0x0A, 0x0B, 0x0C,
					},
				},
				{
					Header: &obu.Header{
						Type: obu.OBUFrame,
					},
					Payload: []byte{
						0x01, 0x02,
					},
				},
			}).Marshal(),
			OutputPayloads: [][]byte{
				append(
					(testAV1AggregationHeader{
						W: 1,
						Y: true,
					}).Marshal(),
					(testAV1OBUPayload{
						Header: &obu.Header{
							Type: obu.OBUFrameHeader,
						},
						Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
					}).Marshal()...,
				),
				append(
					(testAV1AggregationHeader{
						W: 1,
						Z: true,
						Y: true,
					}).Marshal(),
					(testAV1OBUPayload{
						Payload: []byte{0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B},
					}).Marshal()...,
				),
				append(
					(testAV1AggregationHeader{
						W: 2,
						Z: true,
					}).Marshal(),
					(testAV1OBUPayload{
						Payload: append(
							(testAV1OBUPayload{
								Payload:           []byte{0x0C},
								HasRTPLengthField: true,
							}).Marshal(),
							(testAV1OBUPayload{
								Header: &obu.Header{
									Type: obu.OBUFrame,
								},
								Payload: []byte{0x01, 0x02},
							}).Marshal()...,
						),
					}).Marshal()...,
				),
			},
		},
		{
			Name: "Should skip the last byte in the packet if W=0",
			MTU:  14,
			InputPayload: (testAV1MultiOBUsPayload{
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x01},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x02},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x03},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x04},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrame,
						HasSizeField: true,
					},
					Payload: []byte{0x05, 0x06, 0x07, 0x08, 0x09},
				},
			}).Marshal(),
			OutputPayloads: [][]byte{
				append(
					(testAV1AggregationHeader{}).Marshal(),
					(testAV1MultiOBUsPayload{
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x01},
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x02},
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x03},
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x04},
						},
					}).Marshal()...,
				),
				append(
					(testAV1AggregationHeader{
						W: 1,
					}).Marshal(),
					(testAV1OBUPayload{
						Header:  &obu.Header{Type: obu.OBUFrame},
						Payload: []byte{0x05, 0x06, 0x07, 0x08, 0x09},
					}).Marshal()...,
				),
			},
		},
		{
			Name: "Should split OBU after four OBUs in a single packet with W=0",
			MTU:  15,
			InputPayload: (testAV1MultiOBUsPayload{
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x01},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x02},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x03},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x04},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrame,
						HasSizeField: true,
					},
					Payload: []byte{0x05, 0x06, 0x07, 0x08, 0x09},
				},
			}).Marshal(),
			OutputPayloads: [][]byte{
				append(
					(testAV1AggregationHeader{
						Y: true,
					}).Marshal(),
					(testAV1MultiOBUsPayload{
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x01},
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x02},
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x03},
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x04},
						},
						{
							// only the length field and the header.
							HasRTPLengthField: true,

							Header: &obu.Header{
								Type: obu.OBUFrame,
							},
						},
					}).Marshal()...,
				),
				append(
					(testAV1AggregationHeader{
						W: 1,
						Z: true,
					}).Marshal(),
					(testAV1OBUPayload{
						Payload: []byte{0x05, 0x06, 0x07, 0x08, 0x09},
					}).Marshal()...,
				),
			},
		},
		{
			Name: "Should use the correct W size when OBUs ands at the MTU boundary",
			MTU:  9,
			InputPayload: (testAV1MultiOBUsPayload{
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x01},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x02},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x03},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x04},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x05},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x06},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrame,
						HasSizeField: true,
					},
					Payload: []byte{0x07, 0x08, 0x09},
				},
			}).Marshal(),
			OutputPayloads: [][]byte{
				append(
					(testAV1AggregationHeader{
						W: 3,
					}).Marshal(),
					(testAV1MultiOBUsPayload{
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x01},
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x02},
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							Payload: []byte{0x03},
						},
					}).Marshal()...,
				),
				append(
					(testAV1AggregationHeader{
						W: 3,
					}).Marshal(),
					(testAV1MultiOBUsPayload{
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x04},
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x05},
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							Payload: []byte{0x06},
						},
					}).Marshal()...,
				),
				append(
					(testAV1AggregationHeader{
						W: 1,
					}).Marshal(),
					(testAV1OBUPayload{
						Header: &obu.Header{
							Type: obu.OBUFrame,
						},
						Payload: []byte{0x07, 0x08, 0x09},
					}).Marshal()...,
				),
			},
		},
		{
			Name: "Should use the correct W size for the next OBU when OBU fragment ands at the MTU boundary",
			MTU:  9,
			InputPayload: (testAV1MultiOBUsPayload{
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{
						0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
						0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E,
						0x0F,
					},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrame,
						HasSizeField: true,
					},
					Payload: []byte{0x10, 0x11, 0x12},
				},
			}).Marshal(),
			OutputPayloads: [][]byte{
				append(
					(testAV1AggregationHeader{
						W: 1,
						Y: true,
					}).Marshal(),
					(testAV1OBUPayload{
						Header: &obu.Header{
							Type: obu.OBUFrameHeader,
						},
						Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
					}).Marshal()...,
				),
				append(
					(testAV1AggregationHeader{
						Z: true,
						W: 1,
					}).Marshal(),
					(testAV1OBUPayload{
						Payload: []byte{0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F},
					}).Marshal()...,
				),
				append(
					(testAV1AggregationHeader{
						W: 1,
					}).Marshal(),
					(testAV1OBUPayload{
						Header: &obu.Header{
							Type: obu.OBUFrame,
						},
						Payload: []byte{0x10, 0x11, 0x12},
					}).Marshal()...,
				),
			},
		},
		{
			Name: "Should split OBU over three packets and adds extra fragmented packet",
			MTU:  7,
			InputPayload: (testAV1MultiOBUsPayload{
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{
						0x01, 0x02, 0x03, 0x04, 0x05,
						0x06, 0x07, 0x08, 0x09, 0x0A,
						0x0B, 0x0C, 0x0D, 0x0E,
					},
				},
				{
					Header: &obu.Header{
						Type: obu.OBUFrame,
					},
					Payload: []byte{
						0x01, 0x02, 0x03, 0x04, 0x05,
					},
				},
			}).Marshal(),
			OutputPayloads: [][]byte{
				append(
					(testAV1AggregationHeader{
						W: 1,
						Y: true,
					}).Marshal(),
					(testAV1OBUPayload{
						Header: &obu.Header{
							Type: obu.OBUFrameHeader,
						},
						Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
					}).Marshal()...,
				),
				append(
					(testAV1AggregationHeader{
						W: 1,
						Z: true,
						Y: true,
					}).Marshal(),
					(testAV1OBUPayload{
						Payload: []byte{0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B},
					}).Marshal()...,
				),
				append(
					(testAV1AggregationHeader{
						W: 2,
						Z: true,
						Y: true,
					}).Marshal(),
					(testAV1OBUPayload{
						Payload: append(
							(testAV1OBUPayload{
								Payload:           []byte{0x0C, 0x0D, 0x0E},
								HasRTPLengthField: true,
							}).Marshal(),
							(testAV1OBUPayload{
								Header: &obu.Header{
									Type: obu.OBUFrame,
								},
								Payload: []byte{0x01},
							}).Marshal()...,
						),
					}).Marshal()...,
				),
				append(
					(testAV1AggregationHeader{
						W: 1,
						Z: true,
					}).Marshal(),
					(testAV1OBUPayload{
						Payload: []byte{0x02, 0x03, 0x04, 0x05},
					}).Marshal()...,
				),
			},
		},
	}

	testAV1TestRun(t, tests)
}

func TestAV1Payloader_TemporalDelimiter(t *testing.T) {
	tests := []testAV1Tests{
		{
			Name: "Ignore single temporal delimiter",
			MTU:  1000,
			InputPayload: (testAV1OBUPayload{
				Header:  &obu.Header{Type: obu.OBUTemporalDelimiter},
				Payload: []byte{},
			}).Marshal(),
			OutputPayloads: [][]byte{},
		},
		{
			Name: "Ignore mutlitple temporal delimiters",
			MTU:  1000,
			InputPayload: (testAV1MultiOBUsPayload{
				{
					Header: &obu.Header{
						Type:         obu.OBUTemporalDelimiter,
						HasSizeField: true,
					},
				},
				{
					Header: &obu.Header{Type: obu.OBUTemporalDelimiter},
				},
			}).Marshal(),
		},
		{
			Name: "Split payloads at temporal delimiter",
			MTU:  1000,
			InputPayload: (testAV1MultiOBUsPayload{
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUTemporalDelimiter,
						HasSizeField: true,
					},
				},
				{
					Header:  &obu.Header{Type: obu.OBUFrame},
					Payload: []byte{0x06, 0x07, 0x08, 0x09, 0x0A},
				},
			}).Marshal(),
			OutputPayloads: [][]byte{
				append(
					(testAV1AggregationHeader{
						W: 1,
					}).Marshal(),
					(testAV1OBUPayload{
						Header: &obu.Header{
							Type: obu.OBUFrameHeader,
						},
						Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
					}).Marshal()...,
				),
				append(
					(testAV1AggregationHeader{
						W: 1,
					}).Marshal(),
					(testAV1OBUPayload{
						Header:  &obu.Header{Type: obu.OBUFrame},
						Payload: []byte{0x06, 0x07, 0x08, 0x09, 0x0A},
					}).Marshal()...,
				),
			},
		},
	}

	testAV1TestRun(t, tests)
}

func TestAV1Payloader_ExtensionHeaders(t *testing.T) {
	tests := []testAV1Tests{
		{
			Name: "Keeps extension headers",
			MTU:  1000,
			InputPayload: (testAV1OBUPayload{
				Header: &obu.Header{
					Type: obu.OBUFrameHeader,
					ExtensionHeader: &obu.ExtensionHeader{
						TemporalID: 1,
						SpatialID:  2,
					},
				},
				Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
			}).Marshal(),
			OutputPayloads: [][]byte{
				append(
					(testAV1AggregationHeader{
						W: 1,
					}).Marshal(),
					(testAV1OBUPayload{
						Header: &obu.Header{
							Type: obu.OBUFrameHeader,
							ExtensionHeader: &obu.ExtensionHeader{
								TemporalID: 1,
								SpatialID:  2,
							},
						},
						Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
					}).Marshal()...,
				),
			},
		},
		{
			Name: "Keeps OBUs with the same temporal ID and spatial ID in the same packet",
			MTU:  1000,
			InputPayload: (testAV1MultiOBUsPayload{
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
						ExtensionHeader: &obu.ExtensionHeader{
							TemporalID: 1,
							SpatialID:  2,
						},
					},
					Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
				},
				{
					Header: &obu.Header{
						Type: obu.OBUFrame,
						ExtensionHeader: &obu.ExtensionHeader{
							TemporalID: 1,
							SpatialID:  2,
						},
					},
					Payload: []byte{0x06, 0x07, 0x08, 0x09, 0x0A},
				},
			}).Marshal(),
			OutputPayloads: [][]byte{
				append(
					(testAV1AggregationHeader{
						W: 2,
					}).Marshal(),
					(testAV1MultiOBUsPayload{
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
								ExtensionHeader: &obu.ExtensionHeader{
									TemporalID: 1,
									SpatialID:  2,
								},
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x01, 0x02, 0x03, 0x04, 0x05},
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrame,
								ExtensionHeader: &obu.ExtensionHeader{
									TemporalID: 1,
									SpatialID:  2,
								},
							},
							Payload: []byte{0x06, 0x07, 0x08, 0x09, 0x0A},
						},
					}).Marshal()...,
				),
			},
		},
		{
			Name: "Split OBUs with different temporal ID and spatial ID to different packets",
			MTU:  1000,
			InputPayload: (testAV1MultiOBUsPayload{
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
						ExtensionHeader: &obu.ExtensionHeader{
							TemporalID: 1,
							SpatialID:  2,
						},
					},
					Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
				},
				{
					Header: &obu.Header{
						Type: obu.OBUFrame,
						ExtensionHeader: &obu.ExtensionHeader{
							TemporalID: 2,
							SpatialID:  1,
						},
					},
					Payload: []byte{0x06, 0x07, 0x08, 0x09, 0x0A},
				},
			}).Marshal(),
			OutputPayloads: [][]byte{
				append(
					(testAV1AggregationHeader{
						W: 1,
					}).Marshal(),
					(testAV1MultiOBUsPayload{
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
								ExtensionHeader: &obu.ExtensionHeader{
									TemporalID: 1,
									SpatialID:  2,
								},
							},
							Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
						},
					}).Marshal()...,
				),
				append(
					(testAV1AggregationHeader{
						W: 1,
					}).Marshal(),
					(testAV1MultiOBUsPayload{
						{
							Header: &obu.Header{
								Type: obu.OBUFrame,
								ExtensionHeader: &obu.ExtensionHeader{
									TemporalID: 2,
									SpatialID:  1,
								},
							},
							Payload: []byte{0x06, 0x07, 0x08, 0x09, 0x0A},
						},
					}).Marshal()...,
				),
			},
		},
	}

	testAV1TestRun(t, tests)
}

func TestAV1Payloader_SequenceHeader(t *testing.T) {
	tests := []testAV1Tests{
		{
			Name: "Should pack sequence header with frame in a single packet",
			MTU:  1000,
			InputPayload: (testAV1MultiOBUsPayload{
				{
					Header: &obu.Header{
						Type:         obu.OBUSequenceHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
				},
				{
					Header: &obu.Header{
						Type: obu.OBUFrameHeader,
					},
					Payload: []byte{0x06, 0x07, 0x08, 0x09, 0x0A},
				},
			}).Marshal(),
			OutputPayloads: [][]byte{
				append(
					(testAV1AggregationHeader{
						W: 2,
						N: true,
					}).Marshal(),
					(testAV1MultiOBUsPayload{
						{
							Header: &obu.Header{
								Type: obu.OBUSequenceHeader,
							},
							HasRTPLengthField: true,
							Payload:           []byte{0x01, 0x02, 0x03, 0x04, 0x05},
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							Payload: []byte{0x06, 0x07, 0x08, 0x09, 0x0A},
						},
					}).Marshal()...,
				),
			},
		},
		{
			Name: "Sequence header should start a new packet",
			MTU:  1000,
			InputPayload: (testAV1MultiOBUsPayload{
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
				},
				{
					Header: &obu.Header{
						Type: obu.OBUSequenceHeader,
					},
					Payload: []byte{0x06, 0x07, 0x08, 0x09, 0x0A},
				},
			}).Marshal(),
			OutputPayloads: [][]byte{
				append(
					(testAV1AggregationHeader{
						W: 1,
					}).Marshal(),
					(testAV1OBUPayload{
						Header: &obu.Header{
							Type: obu.OBUFrameHeader,
						},
						Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
					}).Marshal()...,
				),
				append(
					(testAV1AggregationHeader{
						W: 1,
						N: true,
					}).Marshal(),
					(testAV1OBUPayload{
						Header: &obu.Header{
							Type: obu.OBUSequenceHeader,
						},
						Payload: []byte{0x06, 0x07, 0x08, 0x09, 0x0A},
					}).Marshal()...,
				),
			},
		},
		{
			Name: "Sequence header should start a new packet and break with temporal delimiter",
			MTU:  1000,
			InputPayload: (testAV1MultiOBUsPayload{
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUSequenceHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x06, 0x07, 0x08, 0x09, 0x0A},
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUTemporalDelimiter,
						HasSizeField: true,
					},
				},
				{
					Header: &obu.Header{
						Type: obu.OBUFrameHeader,
					},
					Payload: []byte{0x0B, 0x0C, 0x0D, 0x0E, 0x0F},
				},
			}).Marshal(),
			OutputPayloads: [][]byte{
				append(
					(testAV1AggregationHeader{
						W: 1,
					}).Marshal(),
					(testAV1OBUPayload{
						Header: &obu.Header{
							Type: obu.OBUFrameHeader,
						},
						Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
					}).Marshal()...,
				),
				append(
					(testAV1AggregationHeader{
						W: 1,
						N: true,
					}).Marshal(),
					(testAV1OBUPayload{
						Header: &obu.Header{
							Type: obu.OBUSequenceHeader,
						},
						Payload: []byte{0x06, 0x07, 0x08, 0x09, 0x0A},
					}).Marshal()...,
				),
				append(
					(testAV1AggregationHeader{
						W: 1,
					}).Marshal(),
					(testAV1OBUPayload{
						Header: &obu.Header{
							Type: obu.OBUFrameHeader,
						},
						Payload: []byte{0x0B, 0x0C, 0x0D, 0x0E, 0x0F},
					}).Marshal()...,
				),
			},
		},
	}

	testAV1TestRun(t, tests)
}

func TestAv1Payloader_FragmentedEdgeLeb128Size(t *testing.T) {
	size := uint16(128)
	payload := make([]byte, 0, size)
	for i := uint16(0); i < size; i++ {
		payload = append(payload, byte(i))
	}

	tests := []testAV1Tests{
		{
			Name: fmt.Sprintf("Should handle leb128 size edge case at %d bytes", size),
			MTU:  (size * 5) + 14,
			InputPayload: (testAV1MultiOBUsPayload{
				{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: payload,
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrame,
						HasSizeField: true,
					},
					Payload: payload,
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrame,
						HasSizeField: true,
					},
					Payload: payload,
				},
				{
					Header: &obu.Header{
						Type:         obu.OBUFrame,
						HasSizeField: true,
					},
					Payload: payload,
				},
				{
					Header: &obu.Header{
						Type: obu.OBUFrame,
					},
					Payload: payload[:size-1],
				},
			}).Marshal(),
			OutputPayloads: [][]byte{
				append(
					(testAV1AggregationHeader{
						Y: true,
					}).Marshal(),
					(testAV1MultiOBUsPayload{
						{
							Header: &obu.Header{
								Type: obu.OBUFrameHeader,
							},
							HasRTPLengthField: true,
							Payload:           payload,
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrame,
							},
							HasRTPLengthField: true,
							Payload:           payload,
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrame,
							},
							HasRTPLengthField: true,
							Payload:           payload,
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrame,
							},
							HasRTPLengthField: true,
							Payload:           payload,
						},
						{
							Header: &obu.Header{
								Type: obu.OBUFrame,
							},
							HasRTPLengthField: true,
							Payload:           payload[:size-2],
						},
					}).Marshal()...,
				),
				append(
					(testAV1AggregationHeader{
						W: 1,
						Z: true,
					}).Marshal(),
					(testAV1OBUPayload{
						Payload: payload[size-2 : size-1],
					}).Marshal()...,
				),
			},
		},
	}

	testAV1TestRun(t, tests)
}

func TestAV1Payloader_ReturnEarlyOnError(t *testing.T) {
	tests := []testAV1Tests{
		{
			Name:         "Should return early on empty payload",
			MTU:          1000,
			InputPayload: []byte{},
		},
		{
			Name:         "Should return early on nil payload",
			MTU:          1000,
			InputPayload: nil,
		},
		{
			Name:         "Should return early on invalid OBU (missing extension header)",
			MTU:          1000,
			InputPayload: []byte{0x04},
		},
		{
			Name:         "Should return early on invalid OBU (invalid obu_size leb128)",
			MTU:          1000,
			InputPayload: []byte{0x4a, 0xff},
		},
		{
			Name: "Should return early on small packets (obu_size is bigger than the payload)",
			MTU:  1000,
			InputPayload: append(
				(testAV1OBUPayload{
					Header: &obu.Header{
						Type:         obu.OBUFrameHeader,
						HasSizeField: true,
					},
					Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
				}).Marshal(),
				append(
					(&obu.Header{
						Type:         obu.OBUFrame,
						HasSizeField: true,
					}).Marshal(),
					0x03,
				)...,
			),
			OutputPayloads: [][]byte{
				append(
					(testAV1AggregationHeader{
						W: 1,
					}).Marshal(),
					(testAV1OBUPayload{
						Header: &obu.Header{
							Type: obu.OBUFrameHeader,
						},
						Payload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
					}).Marshal()...,
				),
			},
		},
	}

	testAV1TestRun(t, tests)
}

func TestAV1Payloader_Leb128Size(t *testing.T) {
	tests := []struct {
		leb128 int
		size   int
		edge   bool
	}{
		{0, 1, false},
		{1, 1, false},
		{127, 1, false},
		{128, 2, true},
		{16383, 2, false},
		{16384, 3, true},
		{2097151, 3, false},
		{2097152, 4, true},
		{268435455, 4, false},
		{268435456, 5, true},
	}
	payloader := &AV1Payloader{}

	for _, test := range tests {
		actual, edge := payloader.leb128Size(test.leb128)

		assert.Equal(t, test.size, actual)
		assert.Equal(t, test.edge, edge)
	}
}

func TestAV1_depacketizer_to_packetizer(t *testing.T) {
	type testOBU struct {
		Type obu.Type
		Size uint64
	}
	obus := []testOBU{
		{Type: obu.OBUSequenceHeader, Size: 10},
		{Type: obu.OBUFrameHeader, Size: 20},
		{Type: obu.OBUFrame, Size: 3000},
		{Type: obu.OBUFrame, Size: 4800},
		{Type: obu.OBUFrame, Size: 3024},
		{Type: obu.OBUFrame, Size: 2841},
		{Type: obu.OBUFrameHeader, Size: 20},
		{Type: obu.OBUFrame, Size: 8000},
		{Type: obu.OBUSequenceHeader, Size: 12},
		{Type: obu.OBUFrameHeader, Size: 20},
		{Type: obu.OBUFrame, Size: 6300},
		{Type: obu.OBUFrame, Size: 53},
		{Type: obu.OBUFrame, Size: 101},
		{Type: obu.OBUFrame, Size: 202},
		{Type: obu.OBUSequenceHeader, Size: 11},
		{Type: obu.OBUFrameHeader, Size: 20},
		{Type: obu.OBUFrame, Size: 9000},
	}
	payload := make([]byte, 0)
	for _, testOBU := range obus {
		header := obu.Header{
			Type:         testOBU.Type,
			HasSizeField: true,
		}
		payload = append(payload, header.Marshal()...)
		payload = append(payload, obu.WriteToLeb128(uint(testOBU.Size))...)
		for j := 0; j < int(testOBU.Size); j++ { //nolint:gosec // G115
			payload = append(payload, byte((j+len(payload))%256))
		}
	}

	mtuSize := []uint16{
		32,
		215,
		1500,
		8192,
		9216,
	}
	for _, mtu := range mtuSize {
		t.Run(fmt.Sprintf("MTU %d", mtu), func(t *testing.T) {
			payloader := &AV1Payloader{}
			depacketizer := &AV1Depacketizer{}
			result := make([]byte, 0)

			packets := payloader.Payload(mtu, payload)
			for _, packet := range packets {
				p, err := depacketizer.Unmarshal(packet)
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, int(mtu), len(packet), "Expected packet size to be smaller or equal to %d", mtu)

				result = append(result, p...)
			}

			assert.Equalf(
				t,
				len(payload),
				len(result),
				"Expected to packetize and depacketize to be the same for MTU=%d",
				mtu,
			)
			assert.Equalf(t, payload, result, "Expected to packetize and depacketize to be the same for MTU=%d", mtu)
		})
	}
}

func TestAV1_Unmarshal_Error(t *testing.T) {
	for _, test := range []struct {
		expectedError error
		input         []byte
	}{
		{errNilPacket, nil},
		{errShortPacket, []byte{0x00}},
		{errIsKeyframeAndFragment, []byte{byte(0b10001000), 0x00}},
		{obu.ErrFailedToReadLEB128, []byte{byte(0b10000000), 0xFF, 0xFF}},
		{errShortPacket, []byte{byte(0b10000000), 0xFF, 0x0F, 0x00, 0x00}},
	} {
		test := test
		av1Pkt := &AV1Packet{}

		_, err := av1Pkt.Unmarshal(test.input)
		assert.ErrorIs(t, err, test.expectedError)
	}
}

func TestAV1_Unmarshal(t *testing.T) {
	// nolint: dupl
	av1Payload := []byte{
		0x68, 0x0c, 0x08, 0x00, 0x00, 0x00, 0x2c,
		0xd6, 0xd3, 0x0c, 0xd5, 0x02, 0x00, 0x80,
		0x30, 0x10, 0xc3, 0xc0, 0x07, 0xff, 0xff,
		0xf8, 0xb7, 0x30, 0xc0, 0x00, 0x00, 0x88,
		0x17, 0xf9, 0x0c, 0xcf, 0xc6, 0x7b, 0x9c,
		0x0d, 0xda, 0x55, 0x82, 0x82, 0x67, 0x2f,
		0xf0, 0x07, 0x26, 0x5d, 0xf6, 0xc6, 0xe3,
		0x12, 0xdd, 0xf9, 0x71, 0x77, 0x43, 0xe6,
		0xba, 0xf2, 0xce, 0x36, 0x08, 0x63, 0x92,
		0xac, 0xbb, 0xbd, 0x26, 0x4c, 0x05, 0x52,
		0x91, 0x09, 0xf5, 0x37, 0xb5, 0x18, 0xbe,
		0x5c, 0x95, 0xb1, 0x2c, 0x13, 0x27, 0x81,
		0xc2, 0x52, 0x8c, 0xaf, 0x27, 0xca, 0xf2,
		0x93, 0xd6, 0x2e, 0x46, 0x32, 0xed, 0x71,
		0x87, 0x90, 0x1d, 0x0b, 0x84, 0x46, 0x7f,
		0xd1, 0x57, 0xc1, 0x0d, 0xc7, 0x5b, 0x41,
		0xbb, 0x8a, 0x7d, 0xe9, 0x2c, 0xae, 0x36,
		0x98, 0x13, 0x39, 0xb9, 0x0c, 0x66, 0x47,
		0x05, 0xa2, 0xdf, 0x55, 0xc4, 0x09, 0xab,
		0xe4, 0xfb, 0x11, 0x52, 0x36, 0x27, 0x88,
		0x86, 0xf3, 0x4a, 0xbb, 0xef, 0x40, 0xa7,
		0x85, 0x2a, 0xfe, 0x92, 0x28, 0xe4, 0xce,
		0xce, 0xdc, 0x4b, 0xd0, 0xaa, 0x3c, 0xd5,
		0x16, 0x76, 0x74, 0xe2, 0xfa, 0x34, 0x91,
		0x4f, 0xdc, 0x2b, 0xea, 0xae, 0x71, 0x36,
		0x74, 0xe1, 0x2a, 0xf3, 0xd3, 0x53, 0xe8,
		0xec, 0xd6, 0x63, 0xf6, 0x6a, 0x75, 0x95,
		0x68, 0xcc, 0x99, 0xbe, 0x17, 0xd8, 0x3b,
		0x87, 0x5b, 0x94, 0xdc, 0xec, 0x32, 0x09,
		0x18, 0x4b, 0x37, 0x58, 0xb5, 0x67, 0xfb,
		0xdf, 0x66, 0x6c, 0x16, 0x9e, 0xba, 0x72,
		0xc6, 0x21, 0xac, 0x02, 0x6d, 0x6b, 0x17,
		0xf9, 0x68, 0x22, 0x2e, 0x10, 0xd7, 0xdf,
		0xfb, 0x24, 0x69, 0x7c, 0xaf, 0x11, 0x64,
		0x80, 0x7a, 0x9d, 0x09, 0xc4, 0x1f, 0xf1,
		0xd7, 0x3c, 0x5a, 0xc2, 0x2c, 0x8e, 0xf5,
		0xff, 0xee, 0xc2, 0x7c, 0xa1, 0xe4, 0xcb,
		0x1c, 0x6d, 0xd8, 0x15, 0x0e, 0x40, 0x36,
		0x85, 0xe7, 0x04, 0xbb, 0x64, 0xca, 0x6a,
		0xd9, 0x21, 0x8e, 0x95, 0xa0, 0x83, 0x95,
		0x10, 0x48, 0xfa, 0x00, 0x54, 0x90, 0xe9,
		0x81, 0x86, 0xa0, 0x4a, 0x6e, 0xbe, 0x9b,
		0xf0, 0x73, 0x0a, 0x17, 0xbb, 0x57, 0x81,
		0x17, 0xaf, 0xd6, 0x70, 0x1f, 0xe8, 0x6d,
		0x32, 0x59, 0x14, 0x39, 0xd8, 0x1d, 0xec,
		0x59, 0xe4, 0x98, 0x4d, 0x44, 0xf3, 0x4f,
		0x7b, 0x47, 0xd9, 0x92, 0x3b, 0xd9, 0x5c,
		0x98, 0xd5, 0xf1, 0xc9, 0x8b, 0x9d, 0xb1,
		0x65, 0xb3, 0xe1, 0x87, 0xa4, 0x6a, 0xcc,
		0x42, 0x96, 0x66, 0xdb, 0x5f, 0xf9, 0xe1,
		0xa1, 0x72, 0xb6, 0x05, 0x02, 0x1f, 0xa3,
		0x14, 0x3e, 0xfe, 0x99, 0x7f, 0xeb, 0x42,
		0xcf, 0x76, 0x09, 0x19, 0xd2, 0xd2, 0x99,
		0x75, 0x1c, 0x67, 0xda, 0x4d, 0xf4, 0x87,
		0xe5, 0x55, 0x8b, 0xed, 0x01, 0x82, 0xf6,
		0xd6, 0x1c, 0x5c, 0x05, 0x96, 0x96, 0x79,
		0xc1, 0x61, 0x87, 0x74, 0xcd, 0x29, 0x83,
		0x27, 0xae, 0x47, 0x87, 0x36, 0x34, 0xab,
		0xc4, 0x73, 0x76, 0x58, 0x1b, 0x4a, 0xec,
		0x0e, 0x4c, 0x2f, 0xb1, 0x76, 0x08, 0x7f,
		0xaf, 0xfa, 0x6d, 0x8c, 0xde, 0xe4, 0xae,
		0x58, 0x87, 0xe7, 0xa0, 0x27, 0x05, 0x0d,
		0xf5, 0xa7, 0xfb, 0x2a, 0x75, 0x33, 0xd9,
		0x3b, 0x65, 0x60, 0xa4, 0x13, 0x27, 0xa5,
		0xe5, 0x1b, 0x83, 0x78, 0x7a, 0xd7, 0xec,
		0x0c, 0xed, 0x8b, 0xe6, 0x4e, 0x8f, 0xfe,
		0x6b, 0x5d, 0xbb, 0xa8, 0xee, 0x38, 0x81,
		0x6f, 0x09, 0x23, 0x08, 0x8f, 0x07, 0x21,
		0x09, 0x39, 0xf0, 0xf8, 0x03, 0x17, 0x24,
		0x2a, 0x22, 0x44, 0x84, 0xe1, 0x5c, 0xf3,
		0x4f, 0x20, 0xdc, 0xc1, 0xe7, 0xeb, 0xbc,
		0x0b, 0xfb, 0x7b, 0x20, 0x66, 0xa4, 0x27,
		0xe2, 0x01, 0xb3, 0x5f, 0xb7, 0x47, 0xa1,
		0x88, 0x4b, 0x8c, 0x47, 0xda, 0x36, 0x98,
		0x60, 0xd7, 0x46, 0x92, 0x0b, 0x7e, 0x5b,
		0x4e, 0x34, 0x50, 0x12, 0x67, 0x50, 0x8d,
		0xe7, 0xc9, 0xe4, 0x96, 0xef, 0xae, 0x2b,
		0xc7, 0xfa, 0x36, 0x29, 0x05, 0xf5, 0x92,
		0xbd, 0x62, 0xb7, 0xbb, 0x90, 0x66, 0xe0,
		0xad, 0x14, 0x3e, 0xe7, 0xb4, 0x24, 0xf3,
		0x04, 0xcf, 0x22, 0x14, 0x86, 0xa4, 0xb8,
		0xfb, 0x83, 0x56, 0xce, 0xaa, 0xb4, 0x87,
		0x5a, 0x9e, 0xf2, 0x0b, 0xaf, 0xad, 0x40,
		0xe1, 0xb5, 0x5c, 0x6b, 0xa7, 0xee, 0x9f,
		0xbb, 0x1a, 0x68, 0x4d, 0xc3, 0xbf, 0x22,
		0x4d, 0xbe, 0x58, 0x52, 0xc9, 0xcc, 0x0d,
		0x88, 0x04, 0xf1, 0xf8, 0xd4, 0xfb, 0xd6,
		0xad, 0xcf, 0x13, 0x84, 0xd6, 0x2f, 0x90,
		0x0c, 0x5f, 0xb4, 0xe2, 0xd8, 0x29, 0x26,
		0x8d, 0x7c, 0x6b, 0xab, 0x91, 0x91, 0x3c,
		0x25, 0x39, 0x9c, 0x86, 0x08, 0x39, 0x54,
		0x59, 0x0d, 0xa4, 0xa8, 0x31, 0x9f, 0xa3,
		0xbc, 0xc2, 0xcb, 0xf9, 0x30, 0x49, 0xc3,
		0x68, 0x0e, 0xfc, 0x2b, 0x9f, 0xce, 0x59,
		0x02, 0xfa, 0xd4, 0x4e, 0x11, 0x49, 0x0d,
		0x93, 0x0c, 0xae, 0x57, 0xd7, 0x74, 0xdd,
		0x13, 0x1a, 0x15, 0x79, 0x10, 0xcc, 0x99,
		0x32, 0x9b, 0x57, 0x6d, 0x53, 0x75, 0x1f,
		0x6d, 0xbb, 0xe4, 0xbc, 0xa9, 0xd4, 0xdb,
		0x06, 0xe7, 0x09, 0xb0, 0x6f, 0xca, 0xb3,
		0xb1, 0xed, 0xc5, 0x0b, 0x8d, 0x8e, 0x70,
		0xb0, 0xbf, 0x8b, 0xad, 0x2f, 0x29, 0x92,
		0xdd, 0x5a, 0x19, 0x3d, 0xca, 0xca, 0xed,
		0x05, 0x26, 0x25, 0xee, 0xee, 0xa9, 0xdd,
		0xa0, 0xe3, 0x78, 0xe0, 0x56, 0x99, 0x2f,
		0xa1, 0x3f, 0x07, 0x5e, 0x91, 0xfb, 0xc4,
		0xb3, 0xac, 0xee, 0x07, 0xa4, 0x6a, 0xcb,
		0x42, 0xae, 0xdf, 0x09, 0xe7, 0xd0, 0xbb,
		0xc6, 0xd4, 0x38, 0x58, 0x7d, 0xb4, 0x45,
		0x98, 0x38, 0x21, 0xc8, 0xc1, 0x3c, 0x81,
		0x12, 0x7e, 0x37, 0x03, 0xa8, 0xcc, 0xf3,
		0xf9, 0xd9, 0x9d, 0x8f, 0xc1, 0xa1, 0xcc,
		0xc1, 0x1b, 0xe3, 0xa8, 0x93, 0x91, 0x2c,
		0x0a, 0xe8, 0x1f, 0x28, 0x13, 0x44, 0x07,
		0x68, 0x5a, 0x8f, 0x27, 0x41, 0x18, 0xc9,
		0x31, 0xc4, 0xc1, 0x71, 0xe2, 0xf0, 0xc4,
		0xf4, 0x1e, 0xac, 0x29, 0x49, 0x2f, 0xd0,
		0xc0, 0x98, 0x13, 0xa6, 0xbc, 0x5e, 0x34,
		0x28, 0xa7, 0x30, 0x13, 0x8d, 0xb4, 0xca,
		0x91, 0x26, 0x6c, 0xda, 0x35, 0xb5, 0xf1,
		0xbf, 0x3f, 0x35, 0x3b, 0x87, 0x37, 0x63,
		0x40, 0x59, 0x73, 0x49, 0x06, 0x59, 0x04,
		0xe0, 0x84, 0x16, 0x3a, 0xe8, 0xc4, 0x28,
		0xd1, 0xf5, 0x11, 0x9c, 0x34, 0xf4, 0x5a,
		0xc0, 0xf8, 0x67, 0x47, 0x1c, 0x90, 0x63,
		0xbc, 0x06, 0x39, 0x2e, 0x8a, 0xa5, 0xa0,
		0xf1, 0x6b, 0x41, 0xb1, 0x16, 0xbd, 0xb9,
		0x50, 0x78, 0x72, 0x91, 0x8e, 0x8c, 0x99,
		0x0f, 0x7d, 0x99, 0x7e, 0x77, 0x36, 0x85,
		0x87, 0x1f, 0x2e, 0x47, 0x13, 0x55, 0xf8,
		0x07, 0xba, 0x7b, 0x1c, 0xaa, 0xbf, 0x20,
		0xd0, 0xfa, 0xc4, 0xe1, 0xd0, 0xb3, 0xe4,
		0xf4, 0xf9, 0x57, 0x8d, 0x56, 0x19, 0x4a,
		0xdc, 0x4c, 0x83, 0xc8, 0xf1, 0x30, 0xc0,
		0xb5, 0xdf, 0x67, 0x25, 0x58, 0xd8, 0x09,
		0x41, 0x37, 0x2e, 0x0b, 0x47, 0x2b, 0x86,
		0x4b, 0x73, 0x38, 0xf0, 0xa0, 0x6b, 0x83,
		0x30, 0x80, 0x3e, 0x46, 0xb5, 0x09, 0xc8,
		0x6d, 0x3e, 0x97, 0xaa, 0x70, 0x4e, 0x8c,
		0x75, 0x29, 0xec, 0x8a, 0x37, 0x4a, 0x81,
		0xfd, 0x92, 0xf1, 0x29, 0xf0, 0xe8, 0x9d,
		0x8c, 0xb4, 0x39, 0x2d, 0x67, 0x06, 0xcd,
		0x5f, 0x25, 0x02, 0x30, 0xbb, 0x6b, 0x41,
		0x93, 0x55, 0x1e, 0x0c, 0xc9, 0x6e, 0xb5,
		0xd5, 0x9f, 0x80, 0xf4, 0x7d, 0x9d, 0x8a,
		0x0d, 0x8d, 0x3b, 0x15, 0x14, 0xc9, 0xdf,
		0x03, 0x9c, 0x78, 0x39, 0x4e, 0xa0, 0xdc,
		0x3a, 0x1b, 0x8c, 0xdf, 0xaa, 0xed, 0x25,
		0xda, 0x60, 0xdd, 0x30, 0x64, 0x09, 0xcc,
		0x94, 0x53, 0xa1, 0xad, 0xfd, 0x9e, 0xe7,
		0x65, 0x15, 0xb8, 0xb1, 0xda, 0x9a, 0x28,
		0x80, 0x51, 0x88, 0x93, 0x92, 0xe3, 0x03,
		0xdf, 0x70, 0xba, 0x1b, 0x59, 0x3b, 0xb4,
		0x8a, 0xb6, 0x0b, 0x0a, 0xa8, 0x48, 0xdf,
		0xcc, 0x74, 0x4c, 0x71, 0x80, 0x08, 0xec,
		0xc8, 0x8a, 0x73, 0xf5, 0x0e, 0x3d, 0xec,
		0x16, 0xf6, 0x32, 0xfd, 0xf3, 0x6b, 0xba,
		0xa9, 0x65, 0xd1, 0x87, 0xe2, 0x56, 0xcd,
		0xde, 0x2c, 0xa4, 0x1b, 0x25, 0x81, 0xb2,
		0xed, 0xea, 0xe9, 0x11, 0x07, 0xf5, 0x17,
		0xd0, 0xca, 0x5d, 0x07, 0xb9, 0xb2, 0xa9,
		0xa9, 0xee, 0x42, 0x33, 0x93, 0x21, 0x30,
		0x5e, 0xd2, 0x58, 0xfd, 0xdd, 0x73, 0x0d,
		0xb2, 0x93, 0x58, 0x77, 0x78, 0x40, 0x69,
		0xba, 0x3c, 0x95, 0x1c, 0x61, 0xc6, 0xc6,
		0x97, 0x1c, 0xef, 0x4d, 0x91, 0x0a, 0x42,
		0x91, 0x1d, 0x14, 0x93, 0xf5, 0x78, 0x41,
		0x32, 0x8a, 0x0a, 0x43, 0xd4, 0x3e, 0x6b,
		0xb0, 0xd8, 0x0e, 0x04,
	}

	av1Pkt := &AV1Packet{}
	_, err := av1Pkt.Unmarshal(av1Payload)
	assert.NoError(t, err)

	expect := &AV1Packet{
		Z: false,
		Y: true,
		W: 2,
		N: true,
		OBUElements: [][]byte{
			av1Payload[2:14],
			av1Payload[14:],
		},
	}
	assert.Equal(t, expect, av1Pkt, "AV1 Unmarshal didn't store the expected results in the packet")
}
