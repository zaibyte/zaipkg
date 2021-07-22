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

	// TODO testing first one only
	// TODO testing last one only
	// TODO testing first one +1
	// TODO testing +1 last one

	for i := 0; i < 64; i++ { // 64 loops will cost more than 1 second.

		offset := uint64(xmath.AlignSize(rand.Int63n(int64(totalSize)), uid.GrainSize))
		if offset == totalSize {
			offset -= uid.GrainSize
		}
		n := uint64(xmath.AlignSize(rand.Int63n(int64(totalSize)), uid.GrainSize))
		if offset+n > totalSize {
			n = uid.GrainSize
		}

		offs := GetOffsets(buf, offset, n)

		firstIdx := offset / uint64(grains*uid.GrainSize)
		expFirstOID := oids[firstIdx]
		assert.Equal(t, expFirstOID, offs[0].Oid)

		actN := uint64(offs[0].Size)
		for j := 1; j < len(offs); j++ {
			off := offs[j]
			actN += uint64(off.Size)
			assert.Equal(t, off.Oid, oids[int(firstIdx)+j])
		}

		assert.Equal(t, n, actN)
	}
}

func BenchmarkGetOffsets(b *testing.B) {

	grains := 256
	oids := genFixedSizeOIDs(MaxObjsInLink, grains)

	bl := CalcLen(int64(len(oids)))
	buf := make([]byte, bl)

	Make(oids, buf)

	totalSize := GetTotalSize(buf)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		offset := uint64(xmath.AlignSize(rand.Int63n(int64(totalSize)), uid.GrainSize))
		if offset >= totalSize-uint64(grains)*uid.GrainSize {
			offset -= uint64(grains) * uid.GrainSize
		}
		_ = GetOffsets(buf, offset, uint64(grains)*uid.GrainSize)
	}
}

// Link_Obj total size about 100MB.
func BenchmarkGetOffsets100MB(b *testing.B) {

	grains := 256
	oids := genFixedSizeOIDs(128, grains)

	bl := CalcLen(int64(len(oids)))
	buf := make([]byte, bl)

	Make(oids, buf)

	totalSize := GetTotalSize(buf)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		offset := uint64(xmath.AlignSize(rand.Int63n(int64(totalSize)), uid.GrainSize))
		if offset >= totalSize-uint64(grains)*uid.GrainSize {
			offset -= uint64(grains) * uid.GrainSize
		}
		_ = GetOffsets(buf, offset, uint64(grains)*uid.GrainSize)
	}
}
