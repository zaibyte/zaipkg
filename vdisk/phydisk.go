package vdisk

import (
	"sync/atomic"

	"g.tesamc.com/IT/zaipkg/diskutil"
	"g.tesamc.com/IT/zproto/pkg/metapb"
)

// PhyDisk is the physical disk.
type PhyDisk struct{}

func (p *PhyDisk) GetType(path string) metapb.DiskType {
	return diskutil.GetDiskType(path)
}

func (p *PhyDisk) AddUsed(meta *SyncMeta, delta int64) {
	meta.AddUsed(delta)
}

func (p *PhyDisk) InitUsage(path string, meta *SyncMeta) error {
	usage, err := diskutil.GetUsageState(path)
	if err != nil {
		return err
	}
	atomic.StoreUint64(&meta.Size_, usage.Size)
	atomic.StoreUint64(&meta.Used, usage.Used)
	return nil
}
