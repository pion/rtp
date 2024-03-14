// SPDX-FileCopyrightText: 2024 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package rtp

// https://www.iana.org/assignments/rtp-parameters/rtp-parameters.xhtml
// https://en.wikipedia.org/wiki/RTP_payload_formats

// Known static audio payload types.
const (
	// PayloadPCMU is a payload type for ITU-T G.711 PCM μ-Law audio 64 kbit/s (RFC 3551).
	PayloadPCMU = 0
	// PayloadGSM is a payload type for European GSM Full Rate audio 13 kbit/s (GSM 06.10).
	PayloadGSM = 3
	// PayloadG723 is a payload type for ITU-T G.723.1 audio (RFC 3551).
	PayloadG723 = 4
	// PayloadDVI4_8000 is a payload type for IMA ADPCM audio 32 kbit/s (RFC 3551).
	PayloadDVI4_8000 = 5
	// PayloadDVI4_16000 is a payload type for IMA ADPCM audio 64 kbit/s (RFC 3551).
	PayloadDVI4_16000 = 6
	// PayloadLPC is a payload type for Experimental Linear Predictive Coding audio 5.6 kbit/s (RFC 3551).
	PayloadLPC = 7
	// PayloadPCMA is a payload type for ITU-T G.711 PCM A-Law audio 64 kbit/s (RFC 3551).
	PayloadPCMA = 8
	// PayloadG722 is a payload type for ITU-T G.722 audio 64 kbit/s (RFC 3551).
	PayloadG722 = 9
	// PayloadL16Stereo is a payload type for Linear PCM 16-bit Stereo audio 1411.2 kbit/s, uncompressed (RFC 3551).
	PayloadL16Stereo = 10
	// PayloadL16Mono is a payload type for Linear PCM 16-bit audio 705.6 kbit/s, uncompressed (RFC 3551).
	PayloadL16Mono = 11
	// PayloadQCELP is a payload type for Qualcomm Code Excited Linear Prediction (RFC 2658, RFC 3551).
	PayloadQCELP = 12
	// PayloadCN is a payload type for Comfort noise (RFC 3389).
	PayloadCN = 13
	// PayloadMPA is a payload type for MPEG-1 or MPEG-2 audio only (RFC 3551, RFC 2250).
	PayloadMPA = 14
	// PayloadG728 is a payload type for ITU-T G.728 audio 16 kbit/s (RFC 3551).
	PayloadG728 = 15
	// PayloadDVI4_11025 is a payload type for IMA ADPCM audio 44.1 kbit/s (RFC 3551).
	PayloadDVI4_11025 = 16
	// PayloadDVI4_22050 is a payload type for IMA ADPCM audio 88.2 kbit/s (RFC 3551).
	PayloadDVI4_22050 = 17
	// PayloadG729 is a payload type for ITU-T G.729 and G.729a audio 8 kbit/s (RFC 3551, RFC 3555).
	PayloadG729 = 18
)

// Known static video payload types.
const (
	// PayloadCELLB is a payload type for Sun CellB video (RFC 2029).
	PayloadCELLB = 25
	// PayloadJPEG is a payload type for JPEG video (RFC 2435).
	PayloadJPEG = 26
	// PayloadNV is a payload type for Xerox PARC's Network Video (nv, RFC 3551).
	PayloadNV = 28
	// PayloadH261 is a payload type for ITU-T H.261 video (RFC 4587).
	PayloadH261 = 31
	// PayloadMPV is a payload type for MPEG-1 and MPEG-2 video (RFC 2250).
	PayloadMPV = 32
	// PayloadMP2T is a payload type for MPEG-2 transport stream (RFC 2250).
	PayloadMP2T = 33
	// PayloadH263 is a payload type for H.263 video, first version (1996, RFC 3551, RFC 2190).
	PayloadH263 = 34
)

const (
	// PayloadTypeFirstDynamic is a first non-static payload type.
	PayloadTypeFirstDynamic = 35
	// PayloadTypeDefaultDynamic is a default dynamic payload type used in the wild.
	PayloadTypeDefaultDynamic = 101
)
