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
	"math"
	"math/rand"
	"testing"

	"github.com/templexxx/tsc"

	"github.com/stretchr/testify/assert"
)

func TestOIDMinMax(t *testing.T) {

	min := MakeOID(1, 1, 0, 0, NormalObj)
	boxID, groupID, size, digest, otype, err := ParseOID(min)
	if err != nil {
		t.Fatal(err)
	}
	if boxID != 1 || groupID != 1 || size != 0 ||
		digest != 0 || otype != NormalObj {
		t.Fatal("min mismatch", boxID, groupID, size, digest, otype)
	}

	max := MakeOID(MaxBoxID, MaxGroupID, MaxSize, math.MaxUint32, MaxOType)
	boxID, groupID, size, digest, otype, err = ParseOID(max)
	if err != nil {
		t.Fatal(err)
	}
	if boxID != MaxBoxID || groupID != MaxGroupID || size != MaxSize ||
		digest != math.MaxUint32 || otype != MaxOType {
		t.Fatal("max mismatch")
	}
}

func TestOIDMakeParse(t *testing.T) {
	rand.Seed(tsc.UnixNano())

	n := 1024
	for i := 0; i < n; i++ {
		boxID := uint32(rand.Intn(MaxBoxID + 1))
		groupID := uint32(rand.Intn(MaxGroupID + 1))
		otype := uint8(rand.Intn(MaxOType + 1))
		size := uint32(rand.Intn(MaxSize + 1))
		digest := uint32(rand.Intn(math.MaxUint32 + 1))

		if boxID == 0 {
			boxID = 1
		}
		if groupID == 0 {
			groupID = 2
		}
		oid := MakeOID(boxID, groupID, size, digest, otype)

		boxIDAct, groupIDAct, sizeAct, digestAct, otypeAct, err := ParseOID(oid)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, boxID, boxIDAct)
		assert.Equal(t, groupID, groupIDAct)
		assert.Equal(t, size, sizeAct)
		assert.Equal(t, digest, digestAct)
		assert.Equal(t, otype, otypeAct)
	}
}

func BenchmarkMakeOID(b *testing.B) {

	for i := 0; i < b.N; i++ {
		_ = MakeOID(1, 1, 1, 1, 1)
	}
}

func BenchmarkParseOID(b *testing.B) {

	oid := MakeOID(1, 2, 3, 4, 1)

	for i := 0; i < b.N; i++ {
		_, _, _, _, _, _ = ParseOID(oid)
	}
}
