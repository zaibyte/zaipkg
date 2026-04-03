package uid

import (
	"math/rand"

	"github.com/templexxx/tsc"
	"github.com/zaibyte/zaipkg/xdigest"
)

// GenRandOIDs generates cnt unique random oids.
func GenRandOIDs(cnt int) []uint64 {

	rand.Seed(tsc.UnixNano())

	digests := GenRandDigests(cnt)

	oids := make([]uint64, cnt)

	for i := range oids {
		gid := rand.Intn(MaxGroupID)
		if gid == 0 {
			gid = 1
		}
		grains := rand.Intn(MaxGrains)
		if grains == 0 {
			grains = 1
		}
		ot := uint8(rand.Intn(MaxOType))
		if ot == NopObj { // There is no need to test NopObj.
			ot = NormalObj
		}
		oids[i] = MakeOID(uint32(gid), uint32(grains), digests[i], ot)
	}

	return oids
}

// GenRandDigests generates cnt unique random digests.
func GenRandDigests(cnt int) []uint32 {

	digests := make([]uint32, cnt)

	buf := make([]byte, 8)
	rand.Seed(tsc.UnixNano())

	has := make(map[uint32]bool)

	for i := 0; i < cnt; i++ {

		for {
			rand.Read(buf)
			digest := xdigest.Sum32(buf)
			if has[digest] {
				continue
			}
			has[digest] = true
			digests[i] = digest
			break
		}

	}
	return digests
}
