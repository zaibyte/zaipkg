package extutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zaibyte/zproto/pkg/metapb"
)

func TestSyncExt_GetCloneJob(t *testing.T) {

	m := new(metapb.Extent)
	assert.Nil(t, (*SyncExt)(m).GetCloneJob())
}
