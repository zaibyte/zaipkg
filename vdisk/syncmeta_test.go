package vdisk

import (
	"testing"

	"g.tesamc.com/IT/zproto/pkg/metapb"

	"github.com/stretchr/testify/assert"
)

func TestSyncMeta(t *testing.T) {
	d := new(metapb.Disk)
	(*SyncMeta)(d).AddUsed(1)
	assert.Equal(t, uint64(1), d.GetUsed())
	sd := (*SyncMeta)(d)
	sd.AddUsed(1)
	assert.Equal(t, uint64(2), sd.GetUsed())
	assert.Equal(t, uint64(2), d.GetUsed())
}
