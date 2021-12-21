package randring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestU64Ring_TryPop(t *testing.T) {

	r := New(4)
	for i := 0; i <= 1<<4; i++ {
		r.Push(uint64(i))
	}

	assert.Equal(t, uint64(16), r.writeIndex)

	cnt := 0
	for {
		_, ok := r.TryPop()
		if !ok {
			break
		}
		cnt++
	}
	assert.Equal(t, 16, cnt)
}
