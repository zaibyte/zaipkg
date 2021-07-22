package linkobj

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/templexxx/tsc"

	"g.tesamc.com/IT/zaipkg/uid"
	"g.tesamc.com/IT/zaipkg/xmath"
)

func genFixedSizeOIDs(cnt int, grains int) []uint64 {

	rand.Seed(tsc.UnixNano())

	digests := uid.GenRandDigests(cnt)

	oids := make([]uint64, cnt)

	for i := range oids {
		gid := rand.Intn(uid.MaxGroupID)
		if gid == 0 {
			gid = 1
		}

		ot := uint8(rand.Intn(uid.MaxOType))
		ot = uid.NormalObj
		oids[i] = uid.MakeOID(uint32(uid.TestBoxID), uint32(gid), uint32(grains), digests[i], ot)
	}

	return oids
}

func TestGetOffsets(t *testing.T) {

	grains := 256
	oids := genFixedSizeOIDs(MaxObjsInLink, grains)

	bl := CalcLen(int64(len(oids)))
	buf := make([]byte, bl)

	Make(oids, buf)

	totalSize := GetTotalSize(buf)

	for i := 0; i < MaxObjsInLink/2; i++ {

		offset := uint64(xmath.AlignSize(rand.Int63n(int64(totalSize)), uid.GrainSize))
		if offset == totalSize {
			offset -= uid.GrainSize
		}
		n := uint64(xmath.AlignSize(rand.Int63n(int64(totalSize)), uid.GrainSize))
		if offset+n > totalSize {
			n = uid.GrainSize
		}

		offs := GetOffsets(buf, offset, n)

		expFirstOID := oids[offset/uint64(grains*uid.GrainSize)]
		assert.Equal(t, expFirstOID, offs[0].Oid)

		// actOff := offs[0].Offset
		// actN := offs[0].Size
		// // TODO how to verify it? oids with fixed size?

	}
}
