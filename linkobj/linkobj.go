package linkobj

import (
	"encoding/binary"
	"fmt"
	"sort"

	"g.tesamc.com/IT/zaipkg/uid"
	"g.tesamc.com/IT/zaipkg/xmath"
)

// link object links multi normal object together.
// It's a sequence of normal objects.
// It's users responsibility to maintain the real size of each normal object.
// The maximum length of a link object is about 1.33TB (up to MaxObjsInLink normal objects), large enough.
//
// link object struct(bytes) (little endian):
// +--------+----------------+------------------------+---------------------+---------+
// | cnt(4) | total_size(8B) | grains_cums(4B/object) | oid_list(8B/object) | padding |
// +--------+----------------+------------------------+---------------------+---------+
// 0                                                                               4MB
//
// cnt: total oids number
// total_size: link_object size in bytes (sum of normal_object_size in this link)
// grains_cums: cums of objects grains in order. e.g., link 3 objects, each grains is 1,2,3. grains_cums will be 1, 3, 6
// oid_list: object_oid in order
// padding: aligned to uid.GrainSize may need padding

const (
	// MaxObjsInLink is the max_result (4MiB - 12Bytes) / 12Bytes after aligned to uid.GrainSize that won't > settings.MaxObjectSize.
	MaxObjsInLink = 349523
)

// CalcLen calculates link_obj length by oids cnt (n).
func CalcLen(n int64) int64 {

	if n > MaxObjsInLink {
		panic(fmt.Sprintf("too many objects for link, max: %d, but: %d", n, MaxObjsInLink))
	}

	nn := 4 + 8 + 12*n
	return xmath.AlignSize(nn, uid.GrainSize)
}

// Make makes link_obj with oids list and buf.
// The len(buf) must be >= CalcLen(len(oids)).
func Make(oids []uint64, buf []byte) {

	if len(oids) > MaxObjsInLink {
		panic(fmt.Sprintf("too many objects for link, max: %d, but: %d", len(oids), MaxObjsInLink))
	}

	binary.LittleEndian.PutUint32(buf[:4], uint32(len(oids)))

	grainsCumsOffset := 4 + 8
	oidListOffset := grainsCumsOffset + 4*len(oids)

	var grainCum uint32
	for i, oid := range oids {
		grain := uid.GetGrains(oid)
		grainCum += grain
		binary.LittleEndian.PutUint32(buf[grainsCumsOffset+4*i:grainsCumsOffset+4*i+4], grainCum)
		binary.LittleEndian.PutUint64(buf[oidListOffset+8*i:oidListOffset+8*i+8], oid)
	}
	totalSize := uid.GrainSize * uint64(grainCum)

	binary.LittleEndian.PutUint64(buf[4:12], totalSize)
}

// GetTotalSize gets link_obj's total size.
func GetTotalSize(p []byte) uint64 {
	return binary.LittleEndian.Uint64(p[4:12])
}

// ObjOffset is offset & size of an object.
type ObjOffset struct {
	Oid    uint64
	Offset uint32 // Offset in this object.
	Size   uint32 // Size in this object.
}

// GetOffsets gets needed objects & their offset by offset & needed bytes in link_object.
// All offsets & n are aligned to uid.GrainSize.
func GetOffsets(linkO []byte, offset, n uint64) []ObjOffset {

	if n == 0 {
		return nil
	}

	totalSize := binary.LittleEndian.Uint64(linkO[4:12])
	if offset+n > totalSize {
		panic(fmt.Sprintf("link objects out of range [%d] with length %d", offset+n, totalSize))
	}

	offGrain := uint32(offset / uid.GrainSize)
	sizeGrain := uint32(n / uid.GrainSize)

	cnt := binary.LittleEndian.Uint32(linkO[:4])

	firstIdx := sort.Search(int(cnt), func(i int) bool { // At least has one.
		cums := binary.LittleEndian.Uint32(linkO[12+4*i : 12+4*i+4])
		return cums >= offGrain
	})

	grainsCumsOffset := 4 + 8
	oidListOffset := grainsCumsOffset + int(cnt)*4

	var sizeSum uint64

	ret := make([]ObjOffset, 0, 2) // 2 is enough for most cases.
	firstCums := binary.LittleEndian.Uint32(linkO[12+4*firstIdx : 12+4*firstIdx+4])
	if firstCums > offGrain { // Part of wanted bytes is in this object.

		sizeGrainInObj := firstCums - offGrain
		couldRet := false
		if sizeGrainInObj > sizeGrain {
			sizeGrainInObj = sizeGrain
			couldRet = true
		}

		sizeSum += uint64(sizeGrainInObj) * uid.GrainSize

		ret = append(ret, ObjOffset{
			Oid:    binary.LittleEndian.Uint64(linkO[oidListOffset+firstIdx*8 : oidListOffset+firstIdx*8+8]),
			Offset: (firstCums - offGrain) * uid.GrainSize,
			Size:   sizeGrainInObj * uid.GrainSize,
		})
		if couldRet {
			return ret
		}
	}

	// We could use binary search for end object, but the cost maybe higher in most cases.
	// Because in present, the n won't be >= 4MB, and except last object in link, other object's size will be 4MB,
	// which means we just need one more step, it should be much faster than binary search.
	nextIdx := firstIdx + 1
	for {
		if nextIdx == int(cnt) {
			break
		}
		if sizeSum == n {
			break
		}

		cums := binary.LittleEndian.Uint32(linkO[12+4*nextIdx : 12+4*nextIdx+4])
		delta := int64(cums) - int64(offGrain+sizeGrain)
		if delta <= 0 { // We need this object and maybe more.
			oid := binary.LittleEndian.Uint64(linkO[oidListOffset+nextIdx*8 : oidListOffset+nextIdx*8+8])
			size := uid.GetGrains(oid) * uid.GrainSize
			sizeSum += uint64(size)
			ret = append(ret, ObjOffset{
				Oid:    oid,
				Offset: 0,
				Size:   size,
			})
			if delta == 0 { // We just need this object.
				return ret
			}
		} else { // > 0, last part is in this object or
			oid := binary.LittleEndian.Uint64(linkO[oidListOffset+nextIdx*8 : oidListOffset+nextIdx*8+8])

			ret = append(ret, ObjOffset{
				Oid:    oid,
				Offset: 0,
				Size:   (uid.GetGrains(oid) - uint32(delta)) * uid.GrainSize,
			})
			return ret
		}

		nextIdx++
	}
	return ret
}
