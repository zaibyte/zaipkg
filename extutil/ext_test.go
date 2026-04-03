package extutil

import (
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/zaibyte/zaipkg/uid"
	"github.com/zaibyte/zaipkg/xmath/xrand"
	"github.com/zaibyte/zaipkg/xtest"
	"github.com/zaibyte/zproto/pkg/metapb"
)

func TestExtentSize(t *testing.T) {

	if !xtest.IsPropEnabled() {
		t.Skip("prop testing is not enabled")
	}

	ext := &metapb.Extent{
		Id:         uint32(xrand.Int63()),
		State:      metapb.ExtentState_Extent_Broken,
		Size:       xrand.Uint64(),
		Avail:      xrand.Uint64(),
		DiskId:     uid.GenRandDiskID(),
		InstanceId: uid.GenRandInstanceID(),
		LastUpdate: xrand.Int63(),
		CloneJob: &metapb.CloneJob{
			IsSource: false,
			State:    metapb.CloneJobState_CloneJob_Done,
			Id:       xrand.Uint64(),
			ParentId: xrand.Uint64(),
			Total:    xrand.Uint64(),
			Done:     xrand.Uint64(),
			OidsOid:  xrand.Uint64(),
		},
	}

	d, err := proto.Marshal(ext)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("one rand ext will take: %d bytes", len(d))

	ext.CloneJob = nil

	d, err = proto.Marshal(ext)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("one rand ext without clone_job will take: %d bytes", len(d))
}
