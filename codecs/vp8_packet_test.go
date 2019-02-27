package codecs

import (
	"fmt"
	"testing"

	"github.com/pions/rtp"
	"github.com/stretchr/testify/assert"
)

func TestVP8Packet_Unmarshal(t *testing.T) {
	assert := assert.New(t)
	pck := VP8Packet{}

	errSmallerThanHeaderLen := fmt.Errorf("Payload is not large enough to container header")
	errPayloadTooSmall := fmt.Errorf("Payload is not large enough")

	// Nil payload
	raw, err := pck.Unmarshal(&rtp.Packet{
		Payload: nil,
	})
	assert.Nil(raw, "Result should be nil in case of error")
	assert.Equal(err, errSmallerThanHeaderLen, "Error shouldn't nil in case of error")

	// Empty payload
	raw, err = pck.Unmarshal(&rtp.Packet{
		Payload: nil,
	})
	assert.Nil(raw, "Result should be nil in case of error")
	assert.Equal(err, errSmallerThanHeaderLen, "Error shouldn't nil in case of error")

	// Payload smaller than header size
	raw, err = pck.Unmarshal(&rtp.Packet{
		Payload: []byte{0x00, 0x11, 0x22},
	})
	assert.Nil(raw, "Result should be nil in case of error")
	assert.Equal(err, errSmallerThanHeaderLen, "Error shouldn't nil in case of error")

	// Normal payload
	raw, err = pck.Unmarshal(&rtp.Packet{
		Payload: []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x90},
	})
	assert.NotNil(raw, "Result shouldn't be nil in case of succeess")
	assert.Nil(err, "Error should be nil in case of success")

	// Header size, only X
	raw, err = pck.Unmarshal(&rtp.Packet{
		Payload: []byte{0x80, 0x00, 0x00, 0x00},
	})
	assert.NotNil(raw, "Result shouldn't be nil in case of succeess")
	assert.Nil(err, "Error should be nil in case of success")

	// Header size, X and I
	raw, err = pck.Unmarshal(&rtp.Packet{
		Payload: []byte{0x80, 0x80, 0x00, 0x00},
	})
	assert.NotNil(raw, "Result shouldn't be nil in case of succeess")
	assert.Nil(err, "Error should be nil in case of success")

	// Header size, X and I, PID 16bits
	raw, err = pck.Unmarshal(&rtp.Packet{
		Payload: []byte{0x80, 0x80, 0x81, 0x00},
	})
	assert.Nil(raw, "Result should be nil in case of error")
	assert.Equal(err, errPayloadTooSmall, "Error shouldn't nil in case of error")

	// Header size, X and L
	raw, err = pck.Unmarshal(&rtp.Packet{
		Payload: []byte{0x80, 0x40, 0x00, 0x00},
	})
	assert.NotNil(raw, "Result shouldn't be nil in case of succeess")
	assert.Nil(err, "Error should be nil in case of success")

	// Header size, X and T
	raw, err = pck.Unmarshal(&rtp.Packet{
		Payload: []byte{0x80, 0x20, 0x00, 0x00},
	})
	assert.NotNil(raw, "Result shouldn't be nil in case of succeess")
	assert.Nil(err, "Error should be nil in case of success")

	// Header size, X and K
	raw, err = pck.Unmarshal(&rtp.Packet{
		Payload: []byte{0x80, 0x10, 0x00, 0x00},
	})
	assert.NotNil(raw, "Result shouldn't be nil in case of succeess")
	assert.Nil(err, "Error should be nil in case of success")

	// Header size, all flags
	raw, err = pck.Unmarshal(&rtp.Packet{
		Payload: []byte{0xff, 0xff, 0x00, 0x00},
	})
	assert.Nil(raw, "Result should be nil in case of error")
	assert.Equal(err, errPayloadTooSmall, "Error shouldn't nil in case of error")
}
