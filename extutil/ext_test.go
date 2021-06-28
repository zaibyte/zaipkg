package extutil

import (
	"testing"

	"g.tesamc.com/IT/zaipkg/uid"
	"g.tesamc.com/IT/zaipkg/xmath/xrand"
	"g.tesamc.com/IT/zaipkg/xtest"
	"g.tesamc.com/IT/zproto/pkg/metapb"
)

func TestExtentSize(t *testing.T) {

	if !xtest.IsPropEnabled() {
		t.Skip("prop testing is not enabled")
	}

	ext := &metapb.Extent{
		Id:         uint32(xrand.Int63()),
		State:      metapb.ExtentState_Extent_Broken,
		Size_:      xrand.Uint64(),
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

	d, err := ext.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("one rand ext will take: %d bytes", len(d))

	ext.CloneJob = nil

	d, err = ext.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("one rand ext without clone_job will take: %d bytes", len(d))
}
