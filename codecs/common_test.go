package codecs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommon_Min(t *testing.T) {
	assert := assert.New(t)

	res := min(1, -1)
	assert.Equal(res, -1, "-1 < 1")

	res = min(1, 2)
	assert.Equal(res, 1, "1 < 2")

	res = min(3, 3)
	assert.Equal(res, 3, "3 == 3")
}
