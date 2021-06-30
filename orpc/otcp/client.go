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
	"encoding/binary"
	"fmt"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"g.tesamc.com/IT/zaipkg/config"

	"g.tesamc.com/IT/zaipkg/xtime"

	"g.tesamc.com/IT/zaipkg/orpc"
	"g.tesamc.com/IT/zaipkg/uid"
	"g.tesamc.com/IT/zaipkg/xdigest"
	"g.tesamc.com/IT/zaipkg/xerrors"
	"g.tesamc.com/IT/zaipkg/xlog"
)

// Client implements orpc.Client.
//
// The client must be started with Client.Start() before use.
//
// It is absolutely safe and encouraged using a single client across arbitrary
// number of concurrently running goroutines.
//
// Default client settings are optimized for high load, so don't override
// them without valid reason.
type Client struct {
	isRunning int64

	// Server address to connect to.
	Addr string

	// The number of concurrent connections the client should establish
	// to the sever.
	// Default is DefaultClientConns.
	Conns int

	// The maximum number of pending requests in the queue.
	//
	// The number of pending requests should exceed the expected number
	// of concurrent goroutines calling client's methods.
	// Otherwise a lot of orpc.ErrRequestQueueOverflow errors may appear.
	//
	// Default is DefaultPendingMessages.
	PendingRequests int

	// Size of send buffer per each underlying connection in bytes.
	// Default value is DefaultClientSendBufferSize.
	SendBufferSize int

	// Size of recv buffer per each underlying connection in bytes.
	// Default value is DefaultClientRecvBufferSize.
	RecvBufferSize int

	// Delay between request flushes.
	//
	// Negative values lead to immediate requests' sending to the server
	// without their buffering. This minimizes rpc latency at the cost
	// of higher CPU and network usage.
	//
	// Default value is DefaultFlushDelay.
	FlushDelay time.Duration

	// The client calls this callback when it needs new connection
	// to the server.
	// The client passes Client.Addr into Dial().
	//
	// By default it returns TCP connections established to the Client.Addr.
	Dial DialFunc

	// CloseWait is the wait duration for Stop.
	CloseWait time.Duration

	requestsChan chan *asyncResult

	stopChan chan struct{}
	stopWg   sync.WaitGroup
}

var _client orpc.Client = new(Client)

// asyncResult is a result returned from Client.callAsync().
type asyncResult struct {
	method  uint8
	reqid   uint64
	oid     uint64
	extID   uint32
	reqData []byte

	respBody []byte

	err chan error
}

const (
	// DefaultClientSendBufferSize is the default size for Client send buffers.
	DefaultClientSendBufferSize = 64 * 1024

	// DefaultClientRecvBufferSize is the default size for Client receive buffers.
	DefaultClientRecvBufferSize = 64 * 1024

	// DefaultClientConns is the default connection numbers for Client.
	DefaultClientConns = 16

	DefaultCloseWait = 3 * time.Second
)

// Start starts rpc client. Establishes connection to the server on Client.Addr.
func (c *Client) Start() error {

	if c.stopChan != nil {
		xlog.Panic("already started")
	}

	config.Adjust(&c.PendingRequests, DefaultPendingMessages)
	config.Adjust(&c.SendBufferSize, DefaultClientSendBufferSize)
	config.Adjust(&c.RecvBufferSize, DefaultClientRecvBufferSize)
	config.Adjust(&c.FlushDelay, DefaultFlushDelay)
	config.Adjust(&c.Conns, DefaultClientConns)
	config.Adjust(&c.CloseWait, DefaultCloseWait)

	c.requestsChan = make(chan *asyncResult, c.PendingRequests)
	c.stopChan = make(chan struct{})

	if c.Dial == nil {
		c.Dial = defaultDial
	}

	for i := 0; i < c.Conns; i++ {
		c.stopWg.Add(1)
		go c.clientHandler()
	}

	atomic.StoreInt64(&c.isRunning, 1)
	return nil
}

// Stop stops rpc client.
func (c *Client) Stop() {

	c.Close(nil)
}

func (c *Client) Close(err error) {
	if !atomic.CompareAndSwapInt64(&c.isRunning, 1, 0) {
		return
	}

	if c.stopChan == nil {
		xlog.Panic("client must be started before stopping it")
	}
	close(c.stopChan)

	c.stopWg.Wait()

	if err == nil {
		err = orpc.ErrServiceClosed
	}

	t := xtime.AcquireTimer(c.CloseWait)

	for {
		select {
		case r := <-c.requestsChan:
			r.err <- err
			continue
		case <-t.C:
			goto reset
		default:
			continue
		}
	}

reset:
	xtime.ReleaseTimer(t)

	c.stopChan = nil
}

// PutObj puts object to the ZBuf node which orpc.Client connected.
func (c *Client) PutObj(reqid, oid uint64, extID uint32, objData []byte, _timeout time.Duration) error {
	return c.call(reqid, objPutMethod, oid, extID, objData)
}

// GetObj gets object from the ZBuf node which orpc.Client connected.
func (c *Client) GetObj(reqid, oid uint64, extID uint32, objData []byte, isClone bool, _timeout time.Duration) error {
	method := objGetMethod
	if isClone {
		method = objGetCloneMethod
	}
	return c.call(reqid, method, oid, extID, objData)
}

// DeleteObj deletes object in the ZBuf node which orpc.Client connected.
func (c *Client) DeleteObj(reqid, oid uint64, extID uint32, _timeout time.Duration) error {
	return c.call(reqid, objDelMethod, oid, extID, nil)
}

func (c *Client) DeleteBatch(reqid uint64, oids []uint64, extID uint32, timeout time.Duration) error {

	body := make([]byte, 8*len(oids))
	for i, oid := range oids {
		binary.LittleEndian.PutUint64(body[i*8:i*8+8], oid)
	}
	digest := xdigest.Sum32(body)
	fakeOID := uint64(digest) << 32 // It's a fake oid, just for passing E2E checksum.
	return c.call(reqid, objDelBatchMethod, fakeOID, extID, body)
}

// call sends the given request to the server and obtains response
// from the server.
//
// Returns non-nil error if the response cannot be obtained.
//
// Don't forget starting the client with Client.Start() before calling Client.call().
func (c *Client) call(reqid uint64, method uint8, oid uint64, extID uint32, body []byte) (err error) {

	if atomic.LoadInt64(&c.isRunning) != 1 {
		return orpc.ErrServiceClosed
	}

	var ar *asyncResult
	if ar, err = c.callAsync(reqid, method, oid, extID, body); err != nil {
		return err
	}

	err = <-ar.err
	releaseAsyncResult(ar)
	return
}

func (c *Client) callAsync(reqid uint64, method uint8, oid uint64, extID uint32, body []byte) (ar *asyncResult, err error) {

	if reqid == 0 {
		reqid = uid.MakeReqID()
	}

	if method == 0 || method > 255 {
		return nil, orpc.ErrNotImplemented
	}

	ar = acquireAsyncResult()

	ar.reqid = reqid
	ar.method = method
	ar.oid = oid
	ar.extID = extID
	ar.err = make(chan error)

	if method == objPutMethod || method == objDelBatchMethod {
		ar.reqData = body
	}
	if method == objGetMethod {
		ar.respBody = body
	}

	select {
	case c.requestsChan <- ar:
		return ar, nil
	default:
		// Try substituting the oldest async request by the new one
		// on requests' queue overflow.
		// This increases the chances for new request to succeed
		// without timeout.
		select {
		case ar2 := <-c.requestsChan:
			ar2.err <- orpc.ErrRequestQueueOverflow
		default:
		}

		// After pop, try to put again.
		select {
		case c.requestsChan <- ar:
			return ar, nil
		default:
			// RequestsChan is filled, release it since m wasn't exposed to the caller yet.
			releaseAsyncResult(ar)
			return nil, orpc.ErrRequestQueueOverflow
		}
	}
}

func (c *Client) clientHandler() {
	defer c.stopWg.Done()

	var conn net.Conn
	var err error
	var stopping atomic.Value

	for {
		dialChan := make(chan struct{})
		go func() {
			if conn, err = c.Dial(c.Addr); err != nil {
				if stopping.Load() == nil {
					xlog.Errorf("cannot establish rpc connection to: %s: %s", c.Addr, err)
				}
			}
			close(dialChan)
		}()

		select {
		case <-c.stopChan:
			stopping.Store(true)
			<-dialChan
			return
		case <-dialChan:
		}

		if err != nil {
			select {
			case <-c.stopChan:
				return
			case <-time.After(300 * time.Millisecond): // After 300ms, try to dial again.
			}
			continue
		}
		c.clientHandleConnection(conn)

		select {
		case <-c.stopChan:
			return
		default:
		}
	}
}

func (c *Client) clientHandleConnection(conn net.Conn) {

	err := sendHandshake(conn)
	if err != nil {
		xlog.Errorf("failed to handshake to: %s: %s", c.Addr, err)
		_ = conn.Close()
		return
	}

	stopChan := make(chan struct{})

	pendingRequests := make(map[uint64]*asyncResult, c.PendingRequests)
	var pendingRequestsLock sync.Mutex // Only two goroutine here, map with mutex is faster than sync.Map.

	writerDone := make(chan error, 1)
	go c.clientWriter(conn, pendingRequests, &pendingRequestsLock, stopChan, writerDone)

	readerDone := make(chan error, 1)
	go c.clientReader(conn, pendingRequests, &pendingRequestsLock, readerDone)

	select {
	case err = <-writerDone:
		close(stopChan)
		_ = conn.Close()
		<-readerDone
	case err = <-readerDone:
		close(stopChan)
		_ = conn.Close()
		<-writerDone
	case <-c.stopChan:
		close(stopChan)
		_ = conn.Close()
		<-readerDone
		<-writerDone
	}

	for _, ar := range pendingRequests {
		select {
		case ar.err <- err:
		default: // Avoiding blocking.
		}
	}
}

func sendHandshake(conn net.Conn) error {
	err := conn.SetWriteDeadline(time.Now().Add(handshakeDuration))
	if err != nil {
		return err
	}
	_, err = conn.Write(handshake[:])
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) clientWriter(w net.Conn, pendingRequests map[uint64]*asyncResult, pendingRequestsLock *sync.Mutex,
	stopChan <-chan struct{}, done chan<- error) {

	var err error
	defer func() { done <- err }()

	enc := newEncoder(w, c.SendBufferSize)
	msg := new(msgBytes)
	rh := new(reqHeader)

	t := time.NewTimer(c.FlushDelay)
	var flushChan <-chan time.Time

	var msgID uint64 = 1
	headerBuf := make([]byte, reqHeaderSize) // reqHeaderSize is bigger than respHeaderSize.

	for {
		msgID++

		var ar *asyncResult

		select {
		case ar = <-c.requestsChan:
		default:
			// Give the last chance for ready goroutines filling c.requestsChan :)
			runtime.Gosched()

			select {
			case <-stopChan:
				return
			case ar = <-c.requestsChan:
			case <-flushChan:
				if err = enc.flush(); err != nil {
					err = fmt.Errorf("client cannot requests to: %s: %s", c.Addr, err)
					return
				}
				flushChan = nil
				continue
			}
		}

		if flushChan == nil {
			flushChan = xtime.GetTimerEvent(t, c.FlushDelay)
		}

		pendingRequestsLock.Lock()
		n := len(pendingRequests)
		pendingRequests[msgID] = ar
		pendingRequestsLock.Unlock()

		if n > 10*c.PendingRequests {
			xlog.ErrorIDf(ar.reqid, "server: %s didn't return %d responses yet: closing connection", c.Addr, n)
			err = orpc.ErrConnection
			return
		}

		rh.method = ar.method
		rh.msgID = msgID
		rh.reqid = ar.reqid
		if ar.reqData != nil {
			rh.bodySize = uint32(len(ar.reqData))
		} else {
			rh.bodySize = 0
		}
		rh.oid = ar.oid
		rh.extID = ar.extID
		msg.header = rh
		msg.body = ar.reqData

		if err = enc.encode(msg, headerBuf); err != nil {
			xlog.ErrorIDf(ar.reqid, "failed to send request to: %s: %s", c.Addr, err)
			return
		}
		msg.header = nil
		msg.body = nil
	}
}

func (c *Client) clientReader(r net.Conn, pendingRequests map[uint64]*asyncResult, pendingRequestsLock *sync.Mutex, done chan<- error) {
	var err error
	defer func() {
		if x := recover(); x != nil {
			if err == nil {
				stackTrace := make([]byte, 1<<20)
				n := runtime.Stack(stackTrace, false)
				xlog.Errorf("panic when reading data from server: %s: %v\nStack trace: %s", r.RemoteAddr().String(), x, stackTrace[:n])
			}
		}

		done <- err
	}()

	hash := xdigest.New()
	dec := newDecoder(r, c.RecvBufferSize, hash)
	rh := new(respHeader)
	headerBuf := make([]byte, respHeaderSize)
	for {

		err = dec.decodeHeader(headerBuf, rh)
		if err != nil {
			if err == orpc.ErrTimeout {
				continue // Keeping trying to read request header.
			}
			xlog.Errorf("failed to read request header from %s: %s", r.RemoteAddr().String(), err)
			return
		}

		msgID := rh.msgID

		pendingRequestsLock.Lock()
		ar, ok := pendingRequests[msgID]
		if ok {
			delete(pendingRequests, msgID)
		}
		pendingRequestsLock.Unlock()

		if !ok {
			xlog.Errorf("unexpected msgID: %d obtained from: %s", msgID, c.Addr)
			err = orpc.ErrInternalServer
			return
		}

		errno := rh.errno
		if errno != 0 { // Ignore response if any error. And the response must be nil.
			ar.err <- orpc.Errno(errno).ToErr()
			continue
		}

		n := rh.bodySize
		if n == 0 {
			ar.err <- nil
			continue
		}

		if n != 0 {
			err = dec.decodeBody(ar.respBody, int(n))
			if err != nil { // If failed to read body, the next read header will be failed too, so just return.
				xlog.ErrorIDf(ar.reqid, "failed to read request body from %s: %s", r.RemoteAddr().String(), err)
				ar.err <- err
				return
			}

			digest := uid.GetDigest(ar.oid)
			actDigest := hash.Sum32()
			if actDigest != digest {
				xlog.ErrorID(ar.reqid, xerrors.WithMessage(orpc.ErrChecksumMismatch, fmt.Sprintf("response exp: %d, but: %d", digest, actDigest)).Error())
				ar.err <- orpc.ErrChecksumMismatch
				hash.Reset()
				continue
			}
			hash.Reset()
		}

		ar.err <- nil
	}
}

var asyncResultPool sync.Pool

func acquireAsyncResult() *asyncResult {
	v := asyncResultPool.Get()
	if v == nil {
		return &asyncResult{}
	}
	return v.(*asyncResult)
}

func releaseAsyncResult(ar *asyncResult) {
	ar.method = 0
	ar.reqid = 0
	ar.oid = 0
	ar.extID = 0
	ar.reqData = nil

	ar.respBody = nil
	ar.err = nil

	asyncResultPool.Put(ar)
}
