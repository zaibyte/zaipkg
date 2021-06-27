package zbufutil

import (
	"g.tesamc.com/IT/zproto/pkg/metapb"
)

func SetState(zBuf *metapb.ZBuf, state metapb.ZBufState) {

	old := zBuf.GetState()

	if old == state {
		return
	}

	switch old {
	case metapb.ZBufState_ZBuf_Tombstone:
		return
	case metapb.ZBufState_ZBuf_Offline:
		if state == metapb.ZBufState_ZBuf_Tombstone {
			zBuf.State = state
			return
		}
		return
	default:

	}

	zBuf.State = state

	return
}
