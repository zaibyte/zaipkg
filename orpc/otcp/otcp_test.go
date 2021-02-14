// Copyright (c) 2020. Temple3x (temple3x@gmail.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// The MIT License (MIT)
//
// Copyright (c) 2014 Aliaksandr Valialkin
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// This file contains code derived from gorpc.
// The main logic & codes are copied from gorpc.

package otcp

import (
	"bytes"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"g.tesamc.com/IT/zaipkg/config/settings"

	"g.tesamc.com/IT/zaipkg/directio"

	"g.tesamc.com/IT/zaipkg/orpc"
	"g.tesamc.com/IT/zaipkg/uid"
	"g.tesamc.com/IT/zaipkg/xbytes"
	"g.tesamc.com/IT/zaipkg/xdigest"
	_ "g.tesamc.com/IT/zaipkg/xlog/xlogtest"
	"github.com/stretchr/testify/assert"
	"github.com/templexxx/tsc"
)

func init() {
	// Avoiding too many allocations.
	xbytes.ResetLeakyCap(32, 32, 32, 4)
}

type testHandler struct {
	putFn func(reqid uint64, oid uint64, objData []byte) error
	getFn func(reqid uint64, oid uint64) (objData []byte, err error)
	delFn func(reqid uint64, oid uint64) error
}

func (h *testHandler) PutObj(reqid uint64, oid uint64, extID uint32, objData []byte) error {
	return h.putFn(reqid, oid, objData)
}

func (h *testHandler) GetObj(reqid uint64, oid uint64, extID uint32, isClone bool) (objData []byte, err error) {
	return h.getFn(reqid, oid)
}

func (h *testHandler) DeleteObj(reqid uint64, oid uint64, extID uint32) error {
	return h.delFn(reqid, oid)
}

func nopHandler() *testHandler {
	return &testHandler{
		putFn: func(reqid uint64, oid uint64, objData []byte) error {
			return nil
		},
		getFn: func(reqid uint64, oid uint64) (objData []byte, err error) {
			return nil, nil
		},
		delFn: func(reqid uint64, oid uint64) error {
			return nil
		},
	}
}

func init() {
	rand.Seed(tsc.UnixNano())
	rand.Read(randObjData)
}

var randObjData = directio.AlignedBlock(4 * 1024 * 1024)

func getRandomAddr() string {
	rand.Seed(tsc.UnixNano())
	return fmt.Sprintf("127.0.0.1:%d", rand.Intn(20000)+10000)
}

func newTestClient(addr string) *Client {
	c := NewClient(addr)
	c.CloseWait = time.Microsecond
	return c
}

func TestClient_GetObj(t *testing.T) {
	addr := getRandomAddr()

	stor := make(map[uint64][]byte)
	sizes := make(map[uint64]uint32)

	h := nopHandler()
	h.putFn = func(reqid uint64, oid uint64, objData []byte) error {
		o := make([]byte, len(objData))
		copy(o, objData)
		stor[oid] = o
		sizes[oid] = uint32(len(o))
		return nil
	}
	h.getFn = func(reqid uint64, oid uint64) (objData []byte, err error) {
		size := sizes[oid]
		objData = xbytes.GetAlignedBytes(int(size))
		o := stor[oid]
		copy(objData, o)
		return
	}

	s := NewServer(addr, h)
	if err := s.Start(); err != nil {
		t.Fatalf("cannot start server: %s", err)
	}
	defer s.Stop()

	c := newTestClient(addr)
	c.Start()
	defer c.Stop()

	for i := 0; i < 128; i++ {

		size := rand.Intn(1025)
		if size == 0 {
			size = 1
		}
		size *= 4096
		objData := randObjData[:size]
		digest := xdigest.Sum32(objData)
		oid := uid.MakeOID(1, 1, uid.BytesToGrains(uint32(size)), digest, uid.NormalObj)

		err := c.PutObj(0, oid, 1, objData, 0)
		if err != nil {
			t.Fatal(err, size)
		}
	}

	getBuf := make([]byte, settings.MaxObjectSize)
	for oid, objBytes := range stor {
		size := sizes[oid]
		act := getBuf[:size]
		err := c.GetObj(0, oid, 1, act, false, 0)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(act, objBytes) {
			t.Fatal("obj data mismatch")
		}
	}
}

func TestClient_DeleteObj(t *testing.T) {
	addr := getRandomAddr()

	stor := make(map[uint64][]byte)
	sizes := make(map[uint64]uint32)

	h := nopHandler()
	h.putFn = func(reqid uint64, oid uint64, objData []byte) error {
		o := make([]byte, len(objData))
		copy(o, objData)
		stor[oid] = o
		sizes[oid] = uint32(len(o))
		return nil
	}
	h.getFn = func(reqid uint64, oid uint64) (objData []byte, err error) {
		o, ok := stor[oid]
		if !ok {
			return nil, orpc.ErrNotFound
		}
		size := sizes[oid]
		objData = xbytes.GetAlignedBytes(int(size))
		copy(objData, o)
		return
	}
	h.delFn = func(reqid uint64, oid uint64) error {
		_, ok := stor[oid]
		if !ok {
			return orpc.ErrNotFound
		}
		delete(stor, oid)
		return nil
	}

	s := NewServer(addr, h)
	if err := s.Start(); err != nil {
		t.Fatalf("cannot start server: %s", err)
	}
	defer s.Stop()

	c := newTestClient(addr)
	c.Start()
	defer c.Stop()

	for i := 0; i < 128; i++ {

		size := rand.Intn(1025)
		if size == 0 {
			size = 1
		}
		size *= 4096

		objData := randObjData[:size]
		digest := xdigest.Sum32(objData)
		oid := uid.MakeOID(1, 1, uid.BytesToGrains(uint32(size)), digest, uid.NormalObj)

		err := c.PutObj(0, oid, 1, objData, 0)
		if err != nil {
			t.Fatal(err, size)
		}
	}

	deleted := make([]uint64, 3)
	cnt := 0
	for oid := range stor {
		if cnt >= 3 {
			break
		}
		err := c.DeleteObj(0, oid, 1, 0)
		if err != nil {
			t.Fatal(err)
		}
		deleted[cnt] = oid
		cnt++
	}

	getBuf := make([]byte, settings.MaxObjectSize)
	for oid, objBytes := range stor {
		size := sizes[oid]
		act := getBuf[:size]
		err := c.GetObj(0, oid, 1, act, false, 0)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(act, objBytes) {
			t.Fatal("obj data mismatch")
		}
	}

	for _, oid := range deleted {
		err := c.GetObj(0, oid, 1, make([]byte, 0), false, 0)
		assert.Equal(t, orpc.ErrNotFound, err)
	}
}

func TestClient_GetObj_Concurrency(t *testing.T) {
	addr := getRandomAddr()

	stor := new(sync.Map)
	sizes := new(sync.Map)

	h := nopHandler()
	h.putFn = func(reqid uint64, oid uint64, objData []byte) error {

		o := make([]byte, len(objData))
		copy(o, objData)
		stor.Store(oid, o)
		sizes.Store(oid, len(o))
		return nil
	}
	h.getFn = func(reqid uint64, oid uint64) (objData []byte, err error) {

		v, ok := sizes.Load(oid)
		if !ok {
			return nil, orpc.ErrNotFound
		}
		size := v.(int)
		objData = xbytes.GetAlignedBytes(size)

		o, ok := stor.Load(oid)
		if !ok {
			return nil, orpc.ErrNotFound
		}
		copy(objData, o.([]byte))
		return
	}

	s := NewServer(addr, h)
	if err := s.Start(); err != nil {
		t.Fatalf("cannot start server: %s", err)
	}
	defer s.Stop()

	c := newTestClient(addr)
	c.Start()
	defer c.Stop()

	testCnt := 128
	oids := make([]uint64, testCnt)
	for i := 0; i < testCnt; i++ {

		size := rand.Intn(1025)
		if size == 0 {
			size = 1
		}
		size *= 4096
		objData := randObjData[:size]
		digest := xdigest.Sum32(objData)
		oid := uid.MakeOID(1, 1, uid.BytesToGrains(uint32(size)), digest, uid.NormalObj)
		err := c.PutObj(0, oid, 1, objData, 0)
		if err != nil {
			t.Fatal(err, size)
		}
		oids[i] = oid
	}

	var wg sync.WaitGroup
	for _, oid := range oids {
		wg.Add(1)
		go func(oid uint64) {
			defer wg.Done()
			v, ok := sizes.Load(oid)
			if !ok {
				t.Fatal("not found")
			}
			size := v.(int)
			act := make([]byte, size)
			err := c.GetObj(0, oid, 1, act, false, 0)
			if err != nil {
				t.Fatal(err)
			}

			v2, ok := stor.Load(oid)
			if !ok {
				t.Fatal("not found")
			}
			if !bytes.Equal(act, v2.([]byte)) {
				t.Fatal("get obj data mismatch")
			}

		}(oid)
	}

	wg.Wait()
}

func TestClient_GetObj_Error_Concurrency(t *testing.T) {
	addr := getRandomAddr()

	stor := new(sync.Map)
	sizes := new(sync.Map)

	h := nopHandler()
	h.putFn = func(reqid uint64, oid uint64, objData []byte) error {
		o := make([]byte, len(objData))
		copy(o, objData)
		stor.Store(oid, o)
		sizes.Store(oid, len(o))
		return nil
	}
	h.getFn = func(reqid uint64, oid uint64) (objData []byte, err error) {
		err = orpc.ErrNotFound
		return
	}

	s := NewServer(addr, h)
	if err := s.Start(); err != nil {
		t.Fatalf("cannot start server: %s", err)
	}
	defer s.Stop()

	c := newTestClient(addr)
	c.Start()
	defer c.Stop()

	oids := make([]uint64, 1024)

	for i := 0; i < 128; i++ {

		size := rand.Intn(1025)
		if size == 0 {
			size = 1
		}
		size *= 4096
		objData := randObjData[:size]
		digest := xdigest.Sum32(objData)
		oid := uid.MakeOID(1, 1, uid.BytesToGrains(uint32(size)), digest, uid.NormalObj)
		err := c.PutObj(0, oid, 1, objData, 0)
		if err != nil {
			t.Fatal(err, size)
		}
		oids[i] = oid
	}

	var wg sync.WaitGroup
	for _, oid := range oids {
		wg.Add(1)
		go func(oid uint64) {
			defer wg.Done()
			err := c.GetObj(0, oid, 1, make([]byte, 0), false, 0)
			if err != orpc.ErrNotFound {
				t.Fatal("error should be not found")
			}
		}(oid)
	}

	wg.Wait()
}
