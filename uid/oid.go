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
)

// oid struct(uint64):
// +----------+-------------+----------+------------+
// | boxID(8) | groupID(20) | otype(4) | digest(32) |
// +----------+-------------+----------+------------+
// 0                                               64
//
// Total length: 8B.
//
// boxID: [0, 8)
// groupID: [8, 28)
// otype: [28, 32)
// digest: [32, 64)

const (
	MaxBoxID   = (1 << 8) - 1
	MaxGroupID = (1 << 20) - 1
	MaxOType   = (1 << 4) - 1
)

// Object types.
// 0 is reserved.
const (
	NormalObj uint8 = 1 // NormalObj: Normal Object, maximum size is 4MB.
	LinkObj   uint8 = 2 // LinkObj: Link Object, it links 262144 objects together (at most 1TB).
)

func isOkOID(boxID, groupID uint32, otype uint8) bool {
	if boxID == 0 || boxID > MaxBoxID {
		return false
	}

	if groupID == 0 || groupID > MaxGroupID {
		return false
	}

	if otype == 0 || otype > MaxOType {
		return false
	}

	return true
}

// MakeOID makes a new oid.
func MakeOID(boxID, groupID, digest uint32, otype uint8) uint64 {

	if !isOkOID(boxID, groupID, otype) {
		panic(fmt.Sprintf("illegal OID elements, "+
			"boxID: %d, groupID: %d, otype: %d",
			boxID, groupID, otype))
	}

	return uint64(digest)<<32 | uint64(otype)<<28 | uint64(groupID)<<8 | uint64(boxID)
}

// ParseReqID parses reqID.
func ParseOID(oid uint64) (boxID, groupID, digest uint32, otype uint8, err error) {

	bgo := uint32(oid)
	boxID = bgo & MaxBoxID
	groupID = (bgo >> 8) & MaxGroupID
	otype = uint8(bgo >> 28)

	digest = uint32(oid >> 32)

	if !isOkOID(boxID, groupID, otype) {
		err = fmt.Errorf("illegal OID elements, "+
			"boxID: %d, groupID: %d, otype: %d",
			boxID, groupID, otype)
		return
	}
	return
}
