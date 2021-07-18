package extutil

import (
	"testing"

	"g.tesamc.com/IT/zproto/pkg/metapb"
	"github.com/stretchr/testify/assert"
)

func TestSyncExt_GetCloneJob(t *testing.T) {

	m := new(metapb.Extent)
	assert.Nil(t, (*SyncExt)(m).GetCloneJob())
}
