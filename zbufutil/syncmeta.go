package zbufutil

import (
	"sync/atomic"

	"g.tesamc.com/IT/zproto/pkg/metapb"
)

// SyncMeta provides thread-safe methods to access metapb.ZBuf.
type SyncMeta metapb.ZBuf

func (p *SyncMeta) GetState() metapb.ZBufState {
	return metapb.ZBufState(atomic.LoadInt32((*int32)(&p.State)))
}

func (p *SyncMeta) SetState(state metapb.ZBufState) (ok bool, oldState metapb.ZBufState) {

	oldState = p.GetState()

	if oldState == state {
		return true, oldState
	}

	switch oldState {
	case metapb.ZBufState_ZBuf_Tombstone:
		return false, oldState
	case metapb.ZBufState_ZBuf_Offline:
		if state == metapb.ZBufState_ZBuf_Tombstone {
			return atomic.CompareAndSwapInt32((*int32)(&p.State), int32(oldState), int32(state)), oldState
		}
		return false, oldState
	default:

	}

	return atomic.CompareAndSwapInt32((*int32)(&p.State), int32(oldState), int32(state)), oldState
}

func (p *SyncMeta) IsTombstone() bool {
	return p.GetState() == metapb.ZBufState_ZBuf_Tombstone
}

func (p *SyncMeta) IsDown() bool {
	return p.GetState() == metapb.ZBufState_ZBuf_Down
}

func (p *SyncMeta) IsDisconnected() bool {
	return p.GetState() == metapb.ZBufState_ZBuf_Disconnected
}

func (p *SyncMeta) IsOffline() bool {
	return p.GetState() == metapb.ZBufState_ZBuf_Offline
}
