package vdisk

import (
	"fmt"
	"sync/atomic"

	"g.tesamc.com/IT/zaipkg/config/settings"

	"g.tesamc.com/IT/zproto/pkg/metapb"
)

// SyncMeta provides thread-safe methods to access metapb.Disk.
type SyncMeta metapb.Disk

// Clone clones SyncMeta.
// Heartbeat needs it.
func (d *SyncMeta) Clone() *metapb.Disk {
	return &metapb.Disk{
		State:      d.GetState(),
		Id:         d.Id,
		Size:       d.Size,
		Used:       d.GetUsed(),
		Type:       d.Type,
		InstanceId: d.InstanceId,
		SN:         d.SN,
	}
}

func (d *SyncMeta) Update(newOne *metapb.Disk) {
	d.SetState(newOne.State) // Actually we only need update state by heartbeat response.
}

func (d *SyncMeta) GetState() metapb.DiskState {
	return metapb.DiskState(atomic.LoadInt32((*int32)(&d.State)))
}

func (d *SyncMeta) SetState(state metapb.DiskState) (ok bool, oldState metapb.DiskState) {
	oldState = d.GetState()

	if oldState == state {
		return true, oldState
	}

	switch oldState {
	case metapb.DiskState_Disk_Broken:
		return false, oldState
	case metapb.DiskState_Disk_Tombstone:
		return false, oldState
	case metapb.DiskState_Disk_Offline:
		if state != metapb.DiskState_Disk_Tombstone {
			return false, oldState
		}

	default:

	}

	return atomic.CompareAndSwapInt32((*int32)(&d.State), int32(oldState), int32(state)), oldState
}

// AddUsed adds delta to used. delta could be negative means delta space have been freed.
func (d *SyncMeta) AddUsed(delta int64) {

	if delta < 0 {
		atomic.AddUint64(&d.Used, ^uint64(-delta-1))
		return
	}
	atomic.AddUint64(&d.Used, uint64(delta))
}

func (d *SyncMeta) GetUsed() uint64 {
	return atomic.LoadUint64(&d.Used)
}

// IsAvailForExt checks if the disk has enough space for making a new extent.
func (d *SyncMeta) IsAvailForExt(minSpace uint64) bool {
	avail := d.Size - d.GetUsed()
	if avail < minSpace {
		return false
	}
	return true
}

// IsLowSpace checks if the disk is lack of space.
func (d *SyncMeta) IsLowSpace(lowSpaceRatio float64) bool {
	return d.AvailRatio() < 1-lowSpaceRatio
}

func (d *SyncMeta) AvailRatio() float64 {
	if d.Size == 0 {
		return 0
	}
	avail := d.Size - d.GetUsed()
	return float64(avail) / float64(d.Size)
}

func (d *SyncMeta) IsTombstone() bool {
	return d.GetState() == metapb.DiskState_Disk_Tombstone
}

func (d *SyncMeta) IsBroken() bool {
	return d.GetState() == metapb.DiskState_Disk_Broken
}

func (d *SyncMeta) IsFull() bool {
	return d.GetState() == metapb.DiskState_Disk_Full
}

func (d *SyncMeta) IsOffline() bool {
	return d.GetState() == metapb.DiskState_Disk_Offline
}

// GetIsolationValue gets isolation level value.
func (d *SyncMeta) GetIsolationValue(key string) string {
	switch key {
	case settings.IsolationInstance:
		return d.InstanceId
	case settings.IsolationDisk:
		return d.Id
	case settings.IsolationNone:
		return ""
	default:
		panic(fmt.Sprintf("illegal isolation level: %s", key))
	}
}

// SpaceScore returns the disk's space score.
// The principle is quite simple:
// The more avail_ratio, the higher score.
//
// We'll only pick up the highest one in disk picking process,
// the mid-state is meaningless.
// And with the help of I/O scheduler (see xio for details), there is no need to
// implement a complex algorithm to get a score.
//
// And not like the OLTP database, we care the cost more but not latency, if the disk device has more capacity,
// we should use it more for saving TCO although it may increase the loading for this device.
func (d *SyncMeta) SpaceScore() float64 {

	return d.AvailRatio()
}
