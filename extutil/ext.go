package extutil

import (
	"sync/atomic"

	"g.tesamc.com/IT/zproto/pkg/metapb"
)

// SyncExt provides thread-safe methods to access metapb.Extent.
type SyncExt metapb.Extent

// Clone clones SyncExt's metapb.Extent for heartbeat or other users.
func (p *SyncExt) Clone() *metapb.Extent {
	return &metapb.Extent{
		State:      p.GetState(),
		Id:         p.Id,
		Size_:      p.Size_,
		Avail:      p.GetAvail(),
		Version:    p.Version,
		DiskId:     p.DiskId,
		InstanceId: p.InstanceId,
		PreferDisk: p.PreferDisk,
		CloneJob:   (*SyncCloneJob)(p.CloneJob).Clone(),
	}
}

func (p *SyncExt) GetState() metapb.ExtentState {
	return metapb.ExtentState(atomic.LoadInt32((*int32)(&p.State)))
}

// SetState sets extent state, return swap ok or not.
// groupSealed in return values indicates should the group which extent belongs to being set sealed.
func (p *SyncExt) SetState(state metapb.ExtentState) (ok, groupSealed bool, oldState metapb.ExtentState) {

	defer func() {
		if p.GetState() != metapb.ExtentState_Extent_ReadWrite {
			groupSealed = false
		}
	}()

	oldSate := p.GetState()
	if oldSate == state {
		return true, groupSealed, oldState
	}

	if state == metapb.ExtentState_Extent_Tombstone {
		return atomic.CompareAndSwapInt32((*int32)(&p.State), int32(oldSate), int32(state)), groupSealed, oldState
	}

	switch oldState {
	case metapb.ExtentState_Extent_Broken:
		return false, groupSealed, oldState
	case metapb.ExtentState_Extent_Ghost:
		return false, groupSealed, oldState
	default:

	}

	return atomic.CompareAndSwapInt32((*int32)(&p.State), int32(oldSate), int32(state)), groupSealed, oldState
}

// AddAvail adds delta to avail. delta could be negative means delta space have been used.
func (p *SyncExt) AddAvail(delta int64) {
	if delta < 0 {
		atomic.AddUint64(&p.Avail, ^uint64(-delta-1))
		return
	}
	atomic.AddUint64(&p.Avail, uint64(delta))
}

// GetAvail gets avail.
func (p *SyncExt) GetAvail() uint64 {
	return atomic.LoadUint64(&p.Avail)
}

// CouldClose returns whether we could close this Extenter or not.
func (p *SyncExt) CouldClose() bool {
	state := p.GetState()
	if state == metapb.ExtentState_Extent_Broken ||
		state == metapb.ExtentState_Extent_Ghost {
		return true
	}
	if state == metapb.ExtentState_Extent_Tombstone {
		return true
	}
	return false
}

// CouldRemove returns whether we could remove all files about this Extenter or not.
func (p *SyncExt) CouldRemove() bool {
	state := p.GetState()
	if state == metapb.ExtentState_Extent_Broken {
		return true
	}
	if state == metapb.ExtentState_Extent_Tombstone {
		return true
	}
	return false
}
