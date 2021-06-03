package uid

import (
	"testing"

	"g.tesamc.com/IT/zaipkg/xmath/xrand"
	"github.com/templexxx/tsc"

	"github.com/stretchr/testify/assert"
)

func TestGetIDCFromInstanceID(t *testing.T) {
	assert.Equal(t, "cn-sz-001", GetIDCFromInstanceID("cn-sz-001-0000001"))

	xrand.Seed(tsc.UnixNano())
	for i := 0; i < 1024; i++ {
		assert.Equal(t, "cn-sz-001", GetIDCFromInstanceID(GenRandInstanceID()))
	}
}
