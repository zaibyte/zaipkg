package extutil

import (
	"sync/atomic"

	"g.tesamc.com/IT/zproto/pkg/metapb"
)

// SyncCloneJob provides thread-safe methods to access metapb.CloneJob.
type SyncCloneJob metapb.CloneJob

// Clone clones SyncCloneJob's metapb.CloneJob for heartbeat or other users.
func (p *SyncCloneJob) Clone() *metapb.CloneJob {
	return &metapb.CloneJob{
		IsSource: p.IsSource,
		State:    p.GetState(),
		Id:       p.Id,
		ParentId: p.ParentId,
		Total:    p.GetTotal(),
		Done:     p.GetDone(),
		OidsOid:  p.GetOidsOid(),
	}
}

// GetState gets clone job state.
func (p *SyncCloneJob) GetState() metapb.CloneJobState {
	return metapb.CloneJobState(atomic.LoadInt32((*int32)(&p.State)))
}

func (p *SyncCloneJob) SetState(state metapb.CloneJobState) bool {
	oldSate := metapb.CloneJobState(atomic.LoadInt32((*int32)(&p.State)))

	if oldSate == state {
		return true
	}

	if oldSate == metapb.CloneJobState_CloneJob_Doing && state == metapb.CloneJobState_CloneJob_Init {
		return false
	}

	switch oldSate {
	case metapb.CloneJobState_CloneJob_Done:
		return false
	default:

	}

	return atomic.CompareAndSwapInt32((*int32)(&p.State), int32(oldSate), int32(state))
}

func (p *SyncCloneJob) SetTotal(cnt uint64) {
	atomic.StoreUint64(&p.Total, cnt)
}

func (p *SyncCloneJob) GetTotal() uint64 {
	return atomic.LoadUint64(&p.Total)
}

func (p *SyncCloneJob) AddDone(d uint64) {
	atomic.AddUint64(&p.Done, d)
}

func (p *SyncCloneJob) GetDone() uint64 {
	return atomic.LoadUint64(&p.Done)
}

func (p *SyncCloneJob) SetOidsOid(oid uint64) {
	atomic.StoreUint64(&p.OidsOid, oid)
}

func (p *SyncCloneJob) GetOidsOid() uint64 {
	return atomic.LoadUint64(&p.OidsOid)
}
