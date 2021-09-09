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

func TestBytesToGrains(t *testing.T) {
	var n uint32 = GrainSize
	var i uint32

	if BytesToGrains(0) != 0 {
		t.Fatal("mismatch")
	}

	for i = 1; i <= n; i++ {
		g := BytesToGrains(i)
		if g != 1 {
			t.Fatal("mismatch")
		}
	}

	for i = n + 1; i < n*2; i++ {
		g := BytesToGrains(i)
		if g != 2 {
			t.Fatal("mismatch")
		}
	}
}

func TestOIDMinMax(t *testing.T) {

	min := MakeOID(1, 0, 0, 0)
	groupID, grains, digest, otype, err := ParseOID(min)
	if err != nil {
		t.Fatal(err)
	}
	if groupID != 1 || grains != 0 ||
		digest != 0 || otype != NopObj {
		t.Fatal("min mismatch", groupID, grains, digest, otype)
	}

	max := MakeOID(MaxGroupID, MaxGrains, math.MaxUint32, MaxOType)
	groupID, grains, digest, otype, err = ParseOID(max)
	if err != nil {
		t.Fatal(err)
	}
	if groupID != MaxGroupID || grains != MaxGrains ||
		digest != math.MaxUint32 || otype != MaxOType {
		t.Fatal("max mismatch")
	}
}

func TestOIDMakeParse(t *testing.T) {
	rand.Seed(tsc.UnixNano())

	n := 1024
	for i := 0; i < n; i++ {
		groupID := uint32(rand.Intn(MaxGroupID + 1))
		otype := uint8(rand.Intn(MaxOType + 1))
		grain := uint32(rand.Intn(MaxGrains + 1))
		digest := uint32(rand.Intn(math.MaxUint32 + 1))

		if groupID == 0 {
			groupID = 2
		}
		if otype == 0 {
			otype = 3
		}
		oid := MakeOID(groupID, grain, digest, otype)

		groupIDAct, sizeAct, digestAct, otypeAct, err := ParseOID(oid)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, groupID, groupIDAct)
		assert.Equal(t, grain, sizeAct)
		assert.Equal(t, digest, digestAct)
		assert.Equal(t, otype, otypeAct)
	}
}

func BenchmarkMakeOID(b *testing.B) {

	for i := 0; i < b.N; i++ {
		_ = MakeOID(1, 1, 1, 1)
	}
}

func BenchmarkParseOID(b *testing.B) {

	oid := MakeOID(2, 3, 4, 1)

	for i := 0; i < b.N; i++ {
		_, _, _, _, _ = ParseOID(oid)
	}
}
