package uid

import (
	"strings"
	"testing"

	"github.com/templexxx/tsc"
	"github.com/zaibyte/zaipkg/xmath/xrand"

	"github.com/stretchr/testify/assert"
)

func TestGetIDCFromInstanceID(t *testing.T) {
	assert.Equal(t, "cn-sz-001", GetIDCFromInstanceID("cn-sz-001-0001"))
}

func TestGetMachineFromInstanceID(t *testing.T) {
	assert.Equal(t, "0001", string(GetMachineFromInstanceID("cn-sz-001-0001")))
}

func TestGetMachineNumFromInstanceID(t *testing.T) {

	ids := GenSeqInstanceID(9999)
	for i, id := range ids {
		assert.Equal(t, uint64(i+1), GetMachineNumFromInstanceID(id))
	}
}

func TestIsValidInstanceID(t *testing.T) {

	xrand.Seed(tsc.UnixNano())
	assert.True(t, IsValidInstanceID(GenRandInstanceID()))

	assert.False(t, IsValidInstanceID(strings.ToUpper(GenRandInstanceID())))

	assert.False(t, IsValidInstanceID(GenRandInstanceID()+"0"))
}
