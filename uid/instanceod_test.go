package uid

import (
	"strings"
	"testing"

	"g.tesamc.com/IT/zaipkg/xmath/xrand"
	"github.com/templexxx/tsc"

	"github.com/stretchr/testify/assert"
)

func TestGetIDCFromInstanceID(t *testing.T) {
	assert.Equal(t, "cn-sz-001", GetIDCFromInstanceID("cn-sz-001-000001"))
}

func TestGetMachineFromInstanceID(t *testing.T) {
	assert.Equal(t, "000001", GetMachineFromInstanceID("cn-sz-001-000001"))
}

func TestIsValidInstanceID(t *testing.T) {

	xrand.Seed(tsc.UnixNano())
	assert.True(t, IsValidInstanceID(GenRandInstanceID()))

	assert.False(t, IsValidInstanceID(strings.ToUpper(GenRandInstanceID())))

	assert.False(t, IsValidInstanceID(GenRandInstanceID()+"0"))
}
