package limitring

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestLimitRing(t *testing.T) {

	r := New(7)

	a := 1
	err := r.Push(unsafe.Pointer(&a))
	assert.Nil(t, err)

	d, ok := r.Pop()
	assert.True(t, ok)

	assert.Equal(t, a, *(*int)(d))
}
