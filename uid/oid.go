package uid

import (
	"fmt"

	"g.tesamc.com/IT/zaipkg/xmath"
)

// oid struct(uint64):
// +-------------+------------+----------+------------+
// | groupID(19) | grains(11) | otype(2) | digest(32) |
// +-------------+------------+----------+------------+
// 0                                                 64
//
// Total length: 8B.
//
// groupID: [0, 19), 0 is reserved.
// grains: [19, 30), supports up to 4MB for 4KB grain size.
// otype: [30, 32).
// digest: [32, 64), object digest, it's a kind of hash.Sum32, details of the algorithm is in package xdigest.

const (
	GrainSize = 4096 // 4KiB grain.

	MaxGroupID = (1 << 19) - 1
	MaxGrains  = (1 << 11) - 1
	MaxOType   = 3
)

// IsValidGroupID returns the groupID is valid or not.
func IsValidGroupID(groupID uint32) bool {
	if groupID == 0 {
		return false
	}
	if groupID > MaxGroupID {
		return false
	}
	return true
}

// Object types.
const (
	// NopObj means this object is empty. Useful to indicate there is an object (something is done),
	// but no need to care about the content.
	//
	// e.g., For init clone job source, if the extent is empty,
	// we will set the oidsoid with an oid has NopObj type. Then the clone job destination will find
	// the oidsoid is not zero, but it's NopObj.
	NopObj    uint8 = 0
	NormalObj uint8 = 1 // NormalObj: Normal Object, maximum size is 4MB.
	// LinkObj is Link Object, linking objects together(we could have multi-level links).
	// TODO In present, we only use one-level link, enough in our env.
	LinkObj uint8 = 2
)

// IsValidOID returns the oid is valid or not.
func IsValidOID(oid uint64) bool {
	groupID, grains, _, otype := parseOID(oid)
	return checkOIDElements(groupID, grains, otype)
}

// checkOIDElements checks oid elements, return false if not legal.
func checkOIDElements(groupID, grains uint32, otype uint8) bool {

	if !IsValidGroupID(groupID) {
		return false
	}

	if grains > MaxGrains { // Size could be 0, if the object is deleted.
		return false
	}

	if otype > MaxOType {
		return false
	}

	return true
}

// BytesToGrains counts how many grains should the bytes taken.
func BytesToGrains(bytes uint32) uint32 {
	a := xmath.AlignSize(int64(bytes), GrainSize)
	return uint32(a) / GrainSize
}

// GrainsToBytes returns bytes the grains takes.
func GrainsToBytes(grains uint32) uint32 {
	return GrainSize * grains
}

// MakeNopOID makes an NopObj's oid.
func MakeNopOID() uint64 {
	return MakeOID(1, 0, 0, NopObj)
}

// MakeOID makes a new oid.
func MakeOID(groupID, grains, digest uint32, otype uint8) uint64 {

	if !checkOIDElements(groupID, grains, otype) {
		panic(fmt.Sprintf("illegal OID elements, "+
			"groupID: %d, grains: %d, otype: %d",
			groupID, grains, otype))
	}

	return uint64(digest)<<32 | uint64(otype)<<30 | uint64(grains)<<19 | uint64(groupID)
}

// ParseOID parses oid.
func ParseOID(oid uint64) (groupID, grains, digest uint32, otype uint8, err error) {

	groupID, grains, digest, otype = parseOID(oid)

	if !checkOIDElements(groupID, grains, otype) {
		err = fmt.Errorf("illegal OID elements, "+
			"groupID: %d, grains: %d, otype: %d",
			groupID, grains, otype)
		return
	}
	return
}

func parseOID(oid uint64) (groupID, grains, digest uint32, otype uint8) {

	lowBits := uint32(oid)
	groupID = lowBits & MaxGroupID
	grains = (lowBits >> 19) & MaxGrains
	otype = uint8(lowBits>>30) & MaxOType

	digest = uint32(oid >> 32)

	return
}

// GetDigest gets digest from an oid.
func GetDigest(oid uint64) uint32 {
	return uint32(oid >> 32)
}

// GetGrains gets grains from an oid.
func GetGrains(oid uint64) uint32 {
	_, grains, _, _ := parseOID(oid)
	return grains
}

// GetOType gets otype from an oid.
func GetOType(oid uint64) uint8 {
	_, _, _, otype := parseOID(oid)
	return otype
}

// GetGroupIDFromOID gets group_id from an oid.
func GetGroupIDFromOID(oid uint64) uint32 {
	groupID, _, _, _ := parseOID(oid)
	return groupID
}
