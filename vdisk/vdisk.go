package vdisk

import "g.tesamc.com/IT/zproto/pkg/metapb"

type Disk interface {
	InitUsage(path string, meta *SyncMeta) error
	GetType(path string) metapb.DiskType
	AddUsed(meta *SyncMeta, delta int64)
}

// NeedRepair returns if the disk need to be repaired or not.
func NeedRepair(old, newState metapb.DiskState) bool {

	if old != metapb.DiskState_Disk_ReadWrite || old != metapb.DiskState_Disk_Full {
		return false
	}

	if newState == metapb.DiskState_Disk_ReadWrite || newState == metapb.DiskState_Disk_Full {
		return false
	}
	return true
}
