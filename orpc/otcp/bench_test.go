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
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"g.tesamc.com/IT/zaipkg/uid"
	"g.tesamc.com/IT/zaipkg/xdigest"
	"g.tesamc.com/IT/zaipkg/xtest"

	"github.com/elastic/go-hdrhistogram"
	"github.com/templexxx/tsc"
)

func makeChans(n int) (bCh chan struct{}, nbCh chan struct{}) {
	nbCh = make(chan struct{}, n)
	bCh = make(chan struct{})
	return
}

func BenchmarkBlockingSelect(b *testing.B) {
	bCh, nbCh := makeChans(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		select {
		case nbCh <- struct{}{}: // query queue emulation
		case <-bCh: // timer emulation
		}
	}
}

func BenchmarkNonBlockingSelect(b *testing.B) {
	bCh, nbCh := makeChans(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		select {
		case nbCh <- struct{}{}:
		default:
			b.Fatalf("Unexpected code path")
			select {
			case nbCh <- struct{}{}:
			case <-bCh:
			}
		}
	}
}

// Single thread request.
func TestClient_Put_Latency_Single(t *testing.T) {

	if !xtest.IsPropEnabled() {
		t.Skip("skip prop testing")
	}

	addr := getRandomAddr()

	s := NewServer(addr, nopHandler())

	if err := s.Start(); err != nil {
		t.Fatalf("cannot start server: %s", err)
	}
	defer s.Stop()

	n := 10000

	c := newTestClient(addr)

	value := make([]byte, 128)
	rand.Read(value)
	key := make([]byte, 8)
	rand.Read(key)

	lat := hdrhistogram.New(100, time.Second.Nanoseconds(), 3)

	err := c.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close(nil)

	objData := make([]byte, 4096)
	rand.Read(objData)
	digest := xdigest.Sum32(objData)
	oid := uid.MakeOID(1, 1, digest, uid.NormalObj)

	jobStart := tsc.UnixNano()
	for i := 0; i < n; i++ {
		start := tsc.UnixNano()
		err := c.PutObj(uid.MakeReqID(), oid, 1, objData, 0)
		if err != nil {
			t.Fatal(err)
		}
		_ = lat.RecordValue(tsc.UnixNano() - start)
	}
	cost := tsc.UnixNano() - jobStart

	printLat("set", lat, cost)
}

// Multi threads request.
func TestClient_Put_Latency_Concurrency(t *testing.T) {

	if !xtest.IsPropEnabled() {
		t.Skip("skip prop testing")
	}

	addr := getRandomAddr()

	s := NewServer(addr, nopHandler())

	if err := s.Start(); err != nil {
		t.Fatalf("cannot start server: %s", err)
	}
	defer s.Stop()

	threads := 64

	c := newTestClient(addr)

	value := make([]byte, 128)
	rand.Read(value)
	key := make([]byte, 8)
	rand.Read(key)

	lat := hdrhistogram.New(100, time.Second.Nanoseconds(), 3)

	n := 10000
	wg := new(sync.WaitGroup)
	wg.Add(threads)

	err := c.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close(nil)

	objData := make([]byte, 4096)
	rand.Read(objData)
	digest := xdigest.Sum32(objData)
	oid := uid.MakeOID(1, 1, digest, uid.NormalObj)

	jobStart := tsc.UnixNano()
	for j := 0; j < threads; j++ {
		go func() {
			defer wg.Done()
			for i := 0; i < n; i++ {
				start := tsc.UnixNano()
				err := c.PutObj(uid.MakeReqID(), oid, 1, objData, 0)
				if err != nil {
					t.Error(err)
					return
				}
				_ = lat.RecordValueAtomic(tsc.UnixNano() - start)
			}
		}()
	}
	wg.Wait()
	cost := tsc.UnixNano() - jobStart

	printLat("set", lat, cost)
}

func printLat(name string, lats *hdrhistogram.Histogram, cost int64) {
	fmt.Println(fmt.Sprintf("%s min: %d, avg: %.2f, max: %d, iops: %.2f",
		name, lats.Min(), lats.Mean(), lats.Max(), float64(lats.TotalCount())/(float64(cost)/float64(time.Second))))
	fmt.Println("percentiles (nsec):")
	fmt.Print(fmt.Sprintf(
		"|  1.00th=[%d],  5.00th=[%d], 10.00th=[%d], 20.00th=[%d],\n"+
			"| 30.00th=[%d], 40.00th=[%d], 50.00th=[%d], 60.00th=[%d],\n"+
			"| 70.00th=[%d], 80.00th=[%d], 90.00th=[%d], 95.00th=[%d],\n"+
			"| 99.00th=[%d], 99.50th=[%d], 99.90th=[%d], 99.95th=[%d],\n"+
			"| 99.99th=[%d]\n",
		lats.ValueAtQuantile(1), lats.ValueAtQuantile(5), lats.ValueAtQuantile(10), lats.ValueAtQuantile(20),
		lats.ValueAtQuantile(30), lats.ValueAtQuantile(40), lats.ValueAtQuantile(50), lats.ValueAtQuantile(60),
		lats.ValueAtQuantile(70), lats.ValueAtQuantile(80), lats.ValueAtQuantile(90), lats.ValueAtQuantile(95),
		lats.ValueAtQuantile(99), lats.ValueAtQuantile(99.5), lats.ValueAtQuantile(99.9), lats.ValueAtQuantile(99.95),
		lats.ValueAtQuantile(99.99)))
}
