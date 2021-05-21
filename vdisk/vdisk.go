package vdisk

import "g.tesamc.com/IT/zproto/pkg/metapb"

type Disk interface {
	InitUsage(path string, meta *SyncMeta) error
	GetType(path string) metapb.DiskType
	AddUsed(meta *SyncMeta, delta int64)
}
