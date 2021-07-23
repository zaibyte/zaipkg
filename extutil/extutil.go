package extutil

import (
	"fmt"

	"g.tesamc.com/IT/zaipkg/typeutil"

	"g.tesamc.com/IT/zaipkg/config/settings"
	"g.tesamc.com/IT/zproto/pkg/metapb"
	"g.tesamc.com/IT/zproto/pkg/stmpb"

	"github.com/gogo/protobuf/proto"
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

	return GetExtV1Preallocate(p.GetSegmentSize())
}

func GetExtV1Preallocate(segSize uint64) uint64 {
	if segSize == 0 {
		segSize = uint64(settings.DefaultExtV1SegSize)
	}

	return (settings.ExtV1SegCnt + 1) * segSize
}

var DefaultExtV1Params = &stmpb.ExtV1Params{
	SegmentSize: uint64(settings.DefaultExtV1SegSize),
}

func makeExtV1Params(segSize typeutil.ByteSize) *stmpb.ExtV1Params {
	return &stmpb.ExtV1Params{
		SegmentSize: uint64(segSize),
	}
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
		DiskSize: GetExtV1Preallocate(uint64(settings.DefaultExtV1SegSize)),
		ExtSize:  settings.ExtV1SegCnt * uint64(settings.DefaultExtV1SegSize),
		Params:   marshalExtV1Params(DefaultExtV1Params),
	},
}

// MakeExtParamsV1 make stmpb.ZBufParams with ext.v1 segments_size config.
func MakeExtParamsV1(segSize typeutil.ByteSize) map[uint32]*stmpb.ExtParams {
	return map[uint32]*stmpb.ExtParams{
		uint32(settings.ExtV1): &stmpb.ExtParams{
			DiskSize: GetExtV1Preallocate(uint64(segSize)),
			ExtSize:  settings.ExtV1SegCnt * uint64(segSize),
			Params:   marshalExtV1Params(makeExtV1Params(segSize)),
		},
	}
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
	case metapb.ExtentState_Extent_Clone:
		// Set clone extent to sealed, the data must not be integrity.
		// Regard this extent broken.
		if state == metapb.ExtentState_Extent_Sealed {
			state = metapb.ExtentState_Extent_Broken
		}
		if state != metapb.ExtentState_Extent_Broken {
			return false, oldState
		}
	default:

	}

	ext.State = state

	return true, oldState
}

func SetCloneJobState(cj *metapb.CloneJob, state metapb.CloneJobState) bool {
	oldSate := cj.State

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

	cj.State = state
	return true
}

// Copy copies from src to dst.
// dst is extent existed in state-machine.
// src is created by heartbeat or internal methods which want to update some dst states.
func Copy(dst, src *metapb.Extent, noState bool) {

	// These elements won't be changed after extent put into state-machine.
	// dst.Id
	// dst.Size_
	// dst.DiskId
	// dst.InstanceId

	if !noState {
		SetState(dst, src.GetState())
	}
	dst.Avail = src.GetAvail()
	dst.LastUpdate = src.LastUpdate
	dst.Created = src.Created

	if src.CloneJob == nil {
		return
	}

	if dst.CloneJob == nil {
		dst.CloneJob = proto.Clone(src.CloneJob).(*metapb.CloneJob)
	} else {
		dst.CloneJob.State = src.CloneJob.State
		dst.CloneJob.Done = src.CloneJob.Done
		if dst.CloneJob.IsSource {
			dst.CloneJob.Total = src.CloneJob.Total
			dst.CloneJob.OidsOid = src.CloneJob.OidsOid
		}
	}
}
