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
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOIDMinMax(t *testing.T) {

	min := MakeOID(1, 1, 0, NormalObj)
	boxID, groupID, digest, otype, err := ParseOID(min)
	if err != nil {
		t.Fatal(err)
	}
	if boxID != 1 || groupID != 1 ||
		digest != 0 || otype != NormalObj {
		t.Fatal("min mismatch", boxID, groupID, digest, otype)
	}

	max := MakeOID(MaxBoxID, MaxGroupID, 1<<32-1, MaxOType)
	boxID, groupID, digest, otype, err = ParseOID(max)
	if err != nil {
		t.Fatal(err)
	}
	if boxID != MaxBoxID || groupID != MaxGroupID ||
		digest != 1<<32-1 || otype != MaxOType {
		t.Fatal("max mismatch")
	}
}

func TestOID(t *testing.T) {

	oids := new(sync.Map)

	wg := new(sync.WaitGroup)
	n := runtime.NumCPU()
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(seed int) {
			defer wg.Done()

			boxID := uint32(seed + 1)
			extID := uint32(seed + 2)
			digest := uint32(seed + 3)
			otype := uint8((seed + 1) & 7)

			oid := MakeOID(boxID, extID, digest, otype)
			oids.Store(seed, oid)

		}(i)
	}
	wg.Wait()

	wg2 := new(sync.WaitGroup)
	wg2.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg2.Done()

			v, ok := oids.Load(i)
			assert.True(t, ok)

			oid := v.(uint64)
			boxID, extID, digest, otype, err := ParseOID(oid)
			assert.Nil(t, err)

			expboxID := uint32(i + 1)
			expextID := uint32(i + 2)
			expdigest := uint32(i + 3)
			expotype := uint8((i + 1) & 7)

			assert.Equal(t, expboxID, boxID)
			assert.Equal(t, expextID, extID)
			assert.Equal(t, expdigest, digest)
			assert.Equal(t, expotype, otype)
		}(i)
	}
	wg2.Wait()
}

func BenchmarkMakeOID(b *testing.B) {

	for i := 0; i < b.N; i++ {
		_ = MakeOID(1, 1, 1, 1)
	}
}

func BenchmarkMakeOID_Parallel(b *testing.B) {

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = MakeOID(1, 1, 1, 1)
		}
	})
}

func BenchmarkParseOID(b *testing.B) {

	oid := MakeOID(1, 2, 3, 4)

	for i := 0; i < b.N; i++ {
		_, _, _, _, _ = ParseOID(oid)
	}
}

func BenchmarkParseOID_Parallel(b *testing.B) {

	oid := MakeOID(1, 2, 3, 4)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _, _, _, _ = ParseOID(oid)
		}
	})
}
