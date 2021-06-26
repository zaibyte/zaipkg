package uid

import (
	"math/rand"
	"testing"

	"github.com/templexxx/tsc"

	"github.com/stretchr/testify/assert"
)

func TestMaxMinExtID(t *testing.T) {
	extIDMax := MakeExtID(MaxGroupID, MaxGroupSeq)
	groupID, seq := ParseExtID(extIDMax)
	assert.Equal(t, uint16(MaxGroupID), groupID)
	assert.Equal(t, uint16(MaxGroupSeq), seq)

	extIDMin := MakeExtID(1, 1)
	groupID, seq = ParseExtID(extIDMin)
	assert.Equal(t, uint16(1), groupID)
	assert.Equal(t, uint16(1), seq)
}

func TestMakeParseExtID(t *testing.T) {
	rand.Seed(tsc.UnixNano())

	n := 1024

	for i := 0; i < n; i++ {
		groupID := uint16(rand.Intn(MaxGroupID + 1))
		groupSeq := uint16(rand.Intn(MaxGroupSeq + 1))

		if groupID == 0 {
			groupID = 1
		}

		if groupSeq == 0 {
			groupSeq = 1
		}

		extID := MakeExtID(groupID, groupSeq)

		actG, actS := ParseExtID(extID)
		assert.Equal(t, groupID, actG)
		assert.Equal(t, actS, groupSeq)
	}
}
