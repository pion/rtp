package codecs

import (
	"testing"
)

func TestCommon_Min(t *testing.T) {
	res := min(1, -1)
	if res != -1 {
		t.Fatal("Error: -1 < 1")
	}

	res = min(1, 2)
	if res != 1 {
		t.Fatal("Error: 1 < 2")
	}

	res = min(3, 3)
	if res != 3 {
		t.Fatal("Error: 3 == 3")
	}
}
