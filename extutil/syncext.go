package extutil

import (
	"sync/atomic"
	"unsafe"

	"g.tesamc.com/IT/zproto/pkg/metapb"
)

// SyncExt provides thread-safe methods to access metapb.Extent.
type SyncExt metapb.Extent

// Clone clones SyncExt's metapb.Extent for heartbeat or other users.
func (p *SyncExt) Clone() *metapb.Extent {

	cj := p.GetCloneJob()
	if cj != nil {
		cj = (*SyncCloneJob)(cj).Clone()
	}
	return &metapb.Extent{
		State:      p.GetState(),
		Id:         p.Id,
		Size_:      p.Size_,
		Avail:      p.GetAvail(),
		DiskId:     p.DiskId,
		InstanceId: p.InstanceId,
		CloneJob:   cj,
		LastUpdate: p.LastUpdate,
		Created:    p.GetCreated(),
	}
}

func (p *SyncExt) GetCloneJobState() metapb.CloneJobState {
	if p.GetCloneJob() == nil {
		return metapb.CloneJobState_CloneJob_Init
	}
	return (*SyncCloneJob)(p.GetCloneJob()).GetState()
}

func (p *SyncExt) SetCloneJobState(state metapb.CloneJobState) {
	if p.GetCloneJob() == nil {
		return
	}
	(*SyncCloneJob)(p.GetCloneJob()).SetState(state)
}

func (p *SyncExt) UpdateBy(v *metapb.Extent) {
	p.SetState(v.State)
	atomic.StoreInt64(&p.LastUpdate, v.LastUpdate)
	atomic.StoreInt32(&p.Created, v.Created)
	p.SetCloneJob(v.CloneJob)
}

func (p *SyncExt) GetCreated() int32 {
	return atomic.LoadInt32(&p.Created)
}

func (p *SyncExt) SetCreated(v int32) {
	atomic.StoreInt32(&p.Created, v)
}

func (p *SyncExt) GetState() metapb.ExtentState {
	return metapb.ExtentState(atomic.LoadInt32((*int32)(&p.State)))
}

// SetState sets extent state, return swap ok or not.
func (p *SyncExt) SetState(state metapb.ExtentState) (ok bool, oldState metapb.ExtentState) {

	oldSate := p.GetState()
	if oldSate == state {
		return true, oldState
	}

	switch oldState {
	case metapb.ExtentState_Extent_Broken:
		return false, oldState
	default:

	}

	return atomic.CompareAndSwapInt32((*int32)(&p.State), int32(oldSate), int32(state)), oldState
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
	if state == metapb.ExtentState_Extent_Broken {
		return true
	}

	return false
}

// GetCloneJob gets clone_job.
func (p *SyncExt) GetCloneJob() *metapb.CloneJob {

	c := unsafe.Pointer(p.CloneJob)
	ret := atomic.LoadPointer(&c)
	return (*metapb.CloneJob)(ret)
}

func (p *SyncExt) SetCloneJob(v *metapb.CloneJob) {

	c := unsafe.Pointer(p.CloneJob)
	atomic.StorePointer(&c, unsafe.Pointer(v))
}
