package vdisk

import (
	"math"
	"sync/atomic"

	"g.tesamc.com/IT/zaipkg/config/settings"

	"g.tesamc.com/IT/zproto/pkg/metapb"
)

// SyncMeta provides thread-safe methods to access *metapb.Disk.
type SyncMeta struct {
	*metapb.Disk
}

// Clone clones SyncMeta.
// Heartbeat needs it.
func (d *SyncMeta) Clone() *metapb.Disk {
	return &metapb.Disk{
		State:      d.GetState(),
		Id:         d.Id,
		Size_:      d.Size_,
		Used:       d.GetUsed(),
		Weight:     d.Weight,
		Type:       d.Type,
		InstanceId: d.InstanceId,
	}
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
	avail := d.GetSize_() - d.GetUsed()
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
	if d.GetSize_() == 0 {
		return 0
	}
	avail := d.GetSize_() - d.GetUsed()
	return float64(avail) / float64(d.GetSize_())
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
// Return instanceID or diskID.
func (d *SyncMeta) GetIsolationValue(key string) uint32 {
	switch key {
	case settings.IsolationDisk:
		return d.GetId()
	default:
		return d.GetInstanceId()
	}
}

const (
	mb = 1 << 20 // megabyte
	// Because in Tesamc, most of the objects' sizes are large (1/n*100), the on-disk index snapshot won't that big,
	// and the write operations won't be frequent, so the storage overhead is low, compare to the size taken
	// by extent, that could be ignored.
	defaultAmplification = 1
	maxScore             = 1024 * 1024 * 1024
	minWeight            = 1e-6
)

// SpaceScore returns the disk's space score.
func (d *SyncMeta) SpaceScore(highSpaceRatio, lowSpaceRatio float64, delta int64) float64 {
	var score float64
	var amplification float64 = defaultAmplification
	available := float64(d.GetSize_()-d.GetUsed()) / mb
	used := float64(d.GetUsed()) / mb
	capacity := float64(d.GetSize_()) / mb

	// highSpaceBound is the lower bound of the high space stage.
	highSpaceBound := (1 - highSpaceRatio) * capacity
	// lowSpaceBound is the upper bound of the low space stage.
	lowSpaceBound := (1 - lowSpaceRatio) * capacity
	if available-float64(delta)/amplification >= highSpaceBound {
		score = float64(delta) + used
	} else if available-float64(delta)/amplification <= lowSpaceBound {
		score = maxScore - (available - float64(delta)/amplification)
	} else {
		// to make the score function continuous, we use linear function y = k * x + b as transition period
		// from above we know that there are two points must on the function image
		// note that it is possible that other irrelative files occupy a lot of storage, so capacity == available + used + irrelative
		// and we regarded irrelative as a fixed value.
		// Then amp = size / used = size / (capacity - irrelative - available)
		//
		// When available == highSpaceBound,
		// we can conclude that size = (capacity - irrelative - highSpaceBound) * amp = (used + available - highSpaceBound) * amp
		// Similarly, when available == lowSpaceBound,
		// we can conclude that size = (capacity - irrelative - lowSpaceBound) * amp = (used + available - lowSpaceBound) * amp
		// These are the two fixed points' x-coordinates, and y-coordinates which can be easily obtained from the above two functions.
		x1, y1 := (used+available-highSpaceBound)*amplification, (used+available-highSpaceBound)*amplification
		x2, y2 := (used+available-lowSpaceBound)*amplification, maxScore-lowSpaceBound

		k := (y2 - y1) / (x2 - x1)
		b := y1 - k*x1
		score = k*(float64(delta)+used) + b
	}

	return score / math.Max(d.GetWeight(), minWeight)
}
