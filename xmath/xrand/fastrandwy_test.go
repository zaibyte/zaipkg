package xrand

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUint32n(t *testing.T) {

	var cnt0, cnt1, cnt2, cnt3 int

	total := 100000000
	for i := 0; i < total; i++ {
		v := Uint32n(3)
		switch v {
		case 0:
			cnt0++
		case 1:
			cnt1++
		case 2:
			cnt2++
		case 3:
			cnt3++
		}
	}

	assert.Equal(t, true, cnt3 == 0)
	assert.Equal(t, total, cnt0+cnt1+cnt2)

	allowDelta := 0.01 // 1% more or less is allowed.
	assert.True(t, math.Abs(float64(cnt0)-float64(total)/3) < float64(total)/3*allowDelta)
	assert.True(t, math.Abs(float64(cnt1)-float64(total)/3) < float64(total)/3*allowDelta)
	assert.True(t, math.Abs(float64(cnt2)-float64(total)/3) < float64(total)/3*allowDelta)
}
