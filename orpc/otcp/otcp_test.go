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

	"g.tesamc.com/IT/zaipkg/orpc"
	"g.tesamc.com/IT/zaipkg/uid"
	"g.tesamc.com/IT/zaipkg/xbytes"
	"g.tesamc.com/IT/zaipkg/xdigest"
	"g.tesamc.com/IT/zaipkg/xlog/xlogtest"
	"github.com/stretchr/testify/assert"
	"github.com/templexxx/tsc"
)

func init() {
	xlogtest.New(true)
}

type testHandler struct {
	putFn func(reqid uint64, oid uint64, objData xbytes.Buffer) error
	getFn func(reqid uint64, oid uint64) (objData xbytes.Buffer, err error)
	delFn func(reqid uint64, oid uint64) error
}

func newTestHandler() *testHandler {
	return &testHandler{
		putFn: func(reqid uint64, oid uint64, objData xbytes.Buffer) error {
			return nil
		},
		getFn: func(reqid uint64, oid uint64) (objData xbytes.Buffer, err error) {
			return nil, nil
		},
		delFn: func(reqid uint64, oid uint64) error {
			return nil
		},
	}
}

func init() {
	rand.Read(immutableObjData)
}

var immutableObjData = make([]byte, 4*1024*1024)

// newTestGetHandler creates a testHandler which always returns the same objData with grains in oid when get.
// It could be used in bench testing.
func newTestGetHandler() *testHandler {

	return &testHandler{
		putFn: func(reqid uint64, oid uint64, objData xbytes.Buffer) error {
			return nil
		},
		getFn: func(reqid uint64, oid uint64) (objData xbytes.Buffer, err error) {
			_, _, grains, _, _, _ := uid.ParseOID(oid)
			objData = xbytes.GetNBytes(int(grains * uid.GrainSize))
			objData.Set(immutableObjData[:grains*uid.GrainSize])
			return
		},
		delFn: func(reqid uint64, oid uint64) error {
			return nil
		},
	}
}

func (h *testHandler) PutObj(reqid uint64, oid uint64, objData xbytes.Buffer) error {
	return h.putFn(reqid, oid, objData)
}

func (h *testHandler) GetObj(reqid uint64, oid uint64) (objData xbytes.Buffer, err error) {
	return h.getFn(reqid, oid)
}

func (h *testHandler) DeleteObj(reqid uint64, oid uint64) error {
	return h.delFn(reqid, oid)
}

func getRandomAddr() string {
	rand.Seed(tsc.UnixNano())
	return fmt.Sprintf("127.0.0.1:%d", rand.Intn(20000)+10000)
}

func TestRequestTimeout(t *testing.T) {

	addr := getRandomAddr()

	h := newTestHandler()
	h.putFn = func(reqid uint64, oid uint64, objData xbytes.Buffer) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	}
	s := NewServer(addr, h)
	if err := s.Start(); err != nil {
		t.Fatalf("cannot start server: %s", err)
	}
	defer s.Stop()

	c := NewClient(addr)
	c.Start()
	defer c.Close()

	objData := make([]byte, 16)
	rand.Read(objData)
	digest := xdigest.Sum32(objData)
	oid := uid.MakeOID(1, 1, 1, digest, uid.NormalObj)

	for i := 0; i < 10; i++ {
		err := c.PutObj(0, oid, objData, time.Millisecond)
		if err == nil {
			t.Fatalf("Timeout error must be returned")
		}
		if err != orpc.ErrTimeout {
			t.Fatalf("Unexpected error returned: [%s]", err)
		}
	}
}

func TestClient_GetObj(t *testing.T) {
	addr := getRandomAddr()

	stor := make(map[uint64][]byte)
	sizes := make(map[uint64]uint32)

	h := newTestHandler()
	h.putFn = func(reqid uint64, oid uint64, objData xbytes.Buffer) error {
		o := make([]byte, len(objData.Bytes()))
		copy(o, objData.Bytes())
		stor[oid] = o
		sizes[oid] = uint32(len(o))
		return nil
	}
	h.getFn = func(reqid uint64, oid uint64) (objData xbytes.Buffer, err error) {
		size := sizes[oid]
		objData = xbytes.GetNBytes(int(size))
		o := stor[oid]
		objData.Write(o)
		return
	}

	s := NewServer(addr, h)
	if err := s.Start(); err != nil {
		t.Fatalf("cannot start server: %s", err)
	}
	defer s.Stop()

	c := NewClient(addr)
	c.Start()
	defer c.Close()

	req := make([]byte, xbytes.MaxBytesSizeInPool*2)
	rand.Read(req)

	for i := 0; i < 7; i++ {

		size := (1 << i) * uid.GrainSize
		objData := req[:size]
		digest := xdigest.Sum32(objData)
		oid := uid.MakeOID(1, 1, uid.BytesToGrains(uint32(size)), digest, uid.NormalObj)

		err := c.PutObj(0, oid, objData, 0)
		if err != nil {
			t.Fatal(err, size)
		}
	}

	getBuf := make([]byte, xbytes.MaxBytesSizeInPool*2)
	for oid, objBytes := range stor {
		size := sizes[oid]
		act := getBuf[:size]
		err := c.GetObj(0, oid, act, 0)
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

	h := newTestHandler()
	h.putFn = func(reqid uint64, oid uint64, objData xbytes.Buffer) error {
		o := make([]byte, len(objData.Bytes()))
		copy(o, objData.Bytes())
		stor[oid] = o
		sizes[oid] = uint32(len(o))
		return nil
	}
	h.getFn = func(reqid uint64, oid uint64) (objData xbytes.Buffer, err error) {
		o, ok := stor[oid]
		if !ok {
			return nil, orpc.ErrNotFound
		}
		size := sizes[oid]
		objData = xbytes.GetNBytes(int(size))
		objData.Write(o)
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

	c := NewClient(addr)
	c.Start()
	defer c.Close()

	req := make([]byte, xbytes.MaxBytesSizeInPool*2)
	rand.Read(req)

	for i := 0; i < 7; i++ {

		size := (1 << i) * uid.GrainSize
		objData := req[:size]
		digest := xdigest.Sum32(objData)
		oid := uid.MakeOID(1, 1, uid.BytesToGrains(uint32(size)), digest, uid.NormalObj)

		err := c.PutObj(0, oid, objData, 0)
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
		err := c.DeleteObj(0, oid, 0)
		if err != nil {
			t.Fatal(err)
		}
		deleted[cnt] = oid
		cnt++
	}

	for oid, objBytes := range stor {
		size := sizes[oid]
		act := make([]byte, size)
		err := c.GetObj(0, oid, act, 0)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(act, objBytes) {
			t.Fatal("obj data mismatch")
		}
	}

	for _, oid := range deleted {
		err := c.GetObj(0, oid, make([]byte, 0), 0)
		assert.Equal(t, orpc.ErrNotFound, err)
	}
}

func TestClient_GetObj_Concurrency(t *testing.T) {
	addr := getRandomAddr()

	stor := new(sync.Map)
	sizes := new(sync.Map)

	h := newTestHandler()
	h.putFn = func(reqid uint64, oid uint64, objData xbytes.Buffer) error {

		o := make([]byte, len(objData.Bytes()))
		copy(o, objData.Bytes())
		stor.Store(oid, o)
		sizes.Store(oid, len(o))
		return nil
	}
	h.getFn = func(reqid uint64, oid uint64) (objData xbytes.Buffer, err error) {

		v, ok := sizes.Load(oid)
		if !ok {
			return nil, orpc.ErrNotFound
		}
		size := v.(int)
		objData = xbytes.GetNBytes(size)

		o, ok := stor.Load(oid)
		if !ok {
			return nil, orpc.ErrNotFound
		}
		objData.Write(o.([]byte))
		return
	}

	s := NewServer(addr, h)
	if err := s.Start(); err != nil {
		t.Fatalf("cannot start server: %s", err)
	}
	defer s.Stop()

	c := NewClient(addr)
	c.Start()
	defer c.Close()

	req := make([]byte, 18*uid.GrainSize)
	rand.Read(req)

	oids := make([]uint64, 18)

	for i := 1; i < 18; i++ {

		size := i * uid.GrainSize
		objData := req[:size]
		digest := xdigest.Sum32(objData)
		oid := uid.MakeOID(1, 1, uid.BytesToGrains(uint32(size)), digest, uid.NormalObj)
		err := c.PutObj(0, oid, objData, 0)
		if err != nil {
			t.Fatal(err, size)
		}
		oids[i] = oid
	}

	var wg sync.WaitGroup
	for _, oid := range oids[1:] {
		wg.Add(1)
		go func(oid uint64) {
			defer wg.Done()
			v, ok := sizes.Load(oid)
			if !ok {
				t.Fatal("not found")
			}
			size := v.(int)
			act := make([]byte, size)
			err := c.GetObj(0, oid, act, 0)
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

	h := newTestHandler()
	h.putFn = func(reqid uint64, oid uint64, objData xbytes.Buffer) error {
		o := make([]byte, len(objData.Bytes()))
		copy(o, objData.Bytes())
		stor.Store(oid, o)
		sizes.Store(oid, len(o))
		return nil
	}
	h.getFn = func(reqid uint64, oid uint64) (objData xbytes.Buffer, err error) {
		err = orpc.ErrNotFound
		return
	}

	s := NewServer(addr, h)
	if err := s.Start(); err != nil {
		t.Fatalf("cannot start server: %s", err)
	}
	defer s.Stop()

	c := NewClient(addr)
	c.Start()
	defer c.Close()

	req := make([]byte, 18*uid.GrainSize)
	rand.Read(req)

	oids := make([]uint64, 18)

	for i := 1; i < 18; i++ {

		size := i * 4096
		objData := req[:size]
		digest := xdigest.Sum32(objData)
		oid := uid.MakeOID(1, 1, uid.BytesToGrains(uint32(size)), digest, uid.NormalObj)
		err := c.PutObj(0, oid, objData, 0)
		if err != nil {
			t.Fatal(err, size)
		}
		oids[i] = oid
	}

	var wg sync.WaitGroup
	for _, oid := range oids[1:] {
		wg.Add(1)
		go func(oid uint64) {
			defer wg.Done()
			err := c.GetObj(0, oid, make([]byte, 0), 0)
			if err != orpc.ErrNotFound {
				t.Fatal("error should be not found")
			}
		}(oid)
	}

	wg.Wait()
}
