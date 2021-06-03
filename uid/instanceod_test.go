package uid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetIDCFromInstanceID(t *testing.T) {
	assert.Equal(t, "cn-sz-001", GetIDCFromInstanceID("cn-sz-001-0000001"))
}
