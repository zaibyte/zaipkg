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
		Version:  p.Version,
		IsSource: p.IsSource,
		State:    p.GetState(),
		Id:       p.Id,
		ParentId: p.ParentId,
		ObjCnt:   p.GetObjCnt(),
		DoneCnt:  p.GetDoneCnt(),
		OidsOid:  p.GetOidsOid(),
	}
}

// GetState gets clone job state.
func (p *SyncCloneJob) GetState() metapb.CloneJobState {
	return metapb.CloneJobState(atomic.LoadInt32((*int32)(&p.State)))
}

func (p *SyncCloneJob) SetObjCnt(cnt uint64) {
	atomic.StoreUint64(&p.ObjCnt, cnt)
}

func (p *SyncCloneJob) GetObjCnt() uint64 {
	return atomic.LoadUint64(&p.ObjCnt)
}

func (p *SyncCloneJob) AddDoneCnt(d uint64) {
	atomic.AddUint64(&p.DoneCnt, d)
}

func (p *SyncCloneJob) GetDoneCnt() uint64 {
	return atomic.LoadUint64(&p.DoneCnt)
}

func (p *SyncCloneJob) SetOidsOid(oid uint64) {
	atomic.StoreUint64(&p.OidsOid, oid)
}

func (p *SyncCloneJob) GetOidsOid() uint64 {
	return atomic.LoadUint64(&p.OidsOid)
}
