package extutil

import (
	"fmt"

	"g.tesamc.com/IT/zaipkg/config/settings"
	"g.tesamc.com/IT/zproto/pkg/metapb"
	"g.tesamc.com/IT/zproto/pkg/stmpb"
)

// ExtPreallocate is the disk size maybe taken by an extent.
// Used in Keeper when it want to create a new extent on a certain disk,
// after picking up disk, we should updating the disk usage by this size,
// then beginning to pick up the next disk.
//
// It's not a precise number, but it's okay because the more accurate usage will be report by disk heartbeat.
//
// params only being invoked inside one machine, if not, it must be under the protection of E2E checksum.
// It unmarshal failed, it must be a serious unrecoverable bug which making data broken.
var ExtPreallocate = map[uint16]func(params []byte) uint64{
	settings.ExtV1: ExtV1Preallocate,
}

func ExtV1Preallocate(params []byte) uint64 {

	p := new(stmpb.ExtV1Params)
	err := p.Unmarshal(params)
	if err != nil {
		panic(fmt.Sprintf("parse ext.v1 params failed: %s", err.Error()))
	}

	return getExtV1Preallocate(p.GetSegmentSize())
}

func getExtV1Preallocate(ss uint64) uint64 {
	if ss == 0 {
		ss = settings.DefaultExtV1SegSize
	}

	return (256 + 1) * ss
}

var DefaultExtV1Params = &stmpb.ExtV1Params{
	SegmentSize: uint64(settings.DefaultExtV1SegSize),
}

func marshalExtV1Params(p *stmpb.ExtV1Params) []byte {
	b, err := p.Marshal()
	if err != nil {
		panic(err)
	}
	return b
}

// DefaultExtParams is the default extent params collection.
var DefaultExtParams = map[uint16]*stmpb.ExtParams{
	settings.ExtV1: &stmpb.ExtParams{
		Size_:  getExtV1Preallocate(settings.DefaultExtV1SegSize),
		Params: marshalExtV1Params(DefaultExtV1Params),
	},
}

// SetState sets extent state, return swap ok or not.
func SetState(ext *metapb.Extent, state metapb.ExtentState) (ok bool, oldState metapb.ExtentState) {

	oldSate := ext.GetState()
	if oldSate == state {
		return true, oldState
	}

	switch oldState {
	case metapb.ExtentState_Extent_Broken:
		return false, oldState
	default:

	}

	ext.State = state

	return true, oldState
}
