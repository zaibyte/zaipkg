/*
 * Copyright (c) 2020. Temple3x (temple3x@gmail.com)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package uid

import (
	"fmt"

	"g.tesamc.com/IT/zaipkg/xmath"
)

// oid struct(uint64):
// +----------+-------------+----------+----------+------------+
// | boxID(3) | groupID(17) | grains(11) | otype(1) | digest(32) |
// +----------+-------------+----------+----------+------------+
// 0                                                          64
//
// Total length: 8B.
//
// boxID: [0, 3), 0 is reserved.
// groupID: [3, 20), 0 is reserved.
// grains: [20, 31), supports 4MB for 4KB grain.
// otype: [31, 32)
// digest: [32, 64), object digest.

const (
	GrainSize = 4096 // 4KiB grain.

	MaxBoxID   = (1 << 3) - 1
	MaxGroupID = (1 << 17) - 1
	MaxGrains  = (1 << 11) - 1
	MaxOType   = 1
)

// Object types.
const (
	NormalObj uint8 = 0 // NormalObj: Normal Object, maximum size is 4MB.
	LinkObj   uint8 = 1 // LinkObj: Link Object, it links 262144 objects together (at most 1TB).
)

func isOkOID(boxID, groupID, grains uint32, otype uint8) bool {
	if boxID == 0 || boxID > MaxBoxID {
		return false
	}

	if groupID == 0 || groupID > MaxGroupID {
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

// MakeOID makes a new oid.
func MakeOID(boxID, groupID, grains, digest uint32, otype uint8) uint64 {

	if !isOkOID(boxID, groupID, grains, otype) {
		panic(fmt.Sprintf("illegal OID elements, "+
			"boxID: %d, groupID: %d, grains: %d, otype: %d",
			boxID, groupID, grains, otype))
	}

	return uint64(digest)<<32 | uint64(otype)<<31 | uint64(grains)<<20 | uint64(groupID)<<3 | uint64(boxID)
}

// ParseOID parses oid.
func ParseOID(oid uint64) (boxID, groupID, grains, digest uint32, otype uint8, err error) {

	lowBits := uint32(oid)
	boxID = lowBits & MaxBoxID
	groupID = (lowBits >> 3) & MaxGroupID
	grains = (lowBits >> 20) & MaxGrains
	otype = uint8(lowBits>>31) & MaxOType

	digest = uint32(oid >> 32)

	if !isOkOID(boxID, groupID, grains, otype) {
		err = fmt.Errorf("illegal OID elements, "+
			"boxID: %d, groupID: %d, grains: %d, otype: %d",
			boxID, groupID, grains, otype)
		return
	}
	return
}

// GetDigest gets digest from an oid.
func GetDigest(oid uint64) uint32 {
	return uint32(oid >> 32)
}
