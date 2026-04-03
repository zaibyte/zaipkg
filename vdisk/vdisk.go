package vdisk

import (
	"github.com/zaibyte/zproto/pkg/metapb"
)

type Disk interface {
	InitUsage(path string, meta *SyncMeta) error
	GetType(path string) metapb.DiskType
	AddUsed(meta *SyncMeta, delta int64)
	GetSN(path string) string
}

func SetState(d *metapb.Disk, state metapb.DiskState) {

	oldState := d.GetState()

	if oldState == state {
		return
	}

	switch oldState {
	case metapb.DiskState_Disk_Broken:
		return
	case metapb.DiskState_Disk_Tombstone:
		return
	case metapb.DiskState_Disk_Offline:
		if state != metapb.DiskState_Disk_Tombstone {
			return
		}

	default:

	}

	d.State = state

	return
}
