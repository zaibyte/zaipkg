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
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"g.tesamc.com/IT/zaipkg/xtime"

	"g.tesamc.com/IT/zaipkg/uid"

	"g.tesamc.com/IT/zaipkg/orpc"
	"g.tesamc.com/IT/zaipkg/xbytes"
	"g.tesamc.com/IT/zaipkg/xdigest"
	"g.tesamc.com/IT/zaipkg/xerrors"
	"g.tesamc.com/IT/zaipkg/xlog"
)

// Server implements orpc.Server.
//
// Default server settings are optimized for high load, so don't override
// them without valid reason.
type Server struct {
	// Address to listen to for incoming connections.
	// TCP transport is used.
	Addr string

	// The maximum number of concurrent rpc calls the server may perform.
	// Default is DefaultConcurrency.
	Concurrency int

	// The maximum number of pending responses in the queue.
	// Default is DefaultPendingMessages.
	PendingResponses int

	// Size of send buffer per each underlying connection in bytes.
	// Default is DefaultBufferSize.
	SendBufferSize int

	// Size of recv buffer per each underlying connection in bytes.
	// Default is DefaultBufferSize.
	RecvBufferSize int

	// The maximum delay between response flushes to clients.
	//
	// Negative values lead to immediate requests' sending to the client
	// without their buffering. This minimizes rpc latency at the cost
	// of higher CPU and network usage.
	//
	// Default is DefaultFlushDelay.
	FlushDelay time.Duration

	// The server obtains new client connections via Listener.Accept().
	//
	// Override the Listener if you want custom underlying transport
	// and/or client authentication/authorization.
	// Don't forget overriding Client.Dial() callback accordingly.
	//
	// It returns TCP connections accepted from Server.Addr.
	Listener *defaultListener

	Handler orpc.ServerHandler

	serverStopChan chan struct{}
	stopWg         sync.WaitGroup
}

const (
	// DefaultConcurrency is the default number of concurrent rpc calls
	// the server can process.
	DefaultConcurrency = 4 * 1024 // 4096 is enough to hold 256 default clients.
	// DefaultServerSendBufferSize is the default size for Server send buffers.
	DefaultServerSendBufferSize = 64 * 1024
	// DefaultServerRecvBufferSize is the default size for Server receive buffers.
	DefaultServerRecvBufferSize = 64 * 1024
)

// Start starts rpc server.
func (s *Server) Start() error {

	if s.serverStopChan != nil {
		xlog.Panic("server is already running. Stop it before starting it again")
	}
	s.serverStopChan = make(chan struct{})

	if s.Handler == nil {
		xlog.Panic("no handler registered")
	}

	if s.Concurrency <= 0 {
		s.Concurrency = DefaultConcurrency
	}
	if s.PendingResponses <= 0 {
		s.PendingResponses = DefaultPendingMessages
	}
	if s.SendBufferSize <= 0 {
		s.SendBufferSize = DefaultServerSendBufferSize
	}
	if s.RecvBufferSize <= 0 {
		s.RecvBufferSize = DefaultServerRecvBufferSize
	}
	if s.FlushDelay == 0 {
		s.FlushDelay = DefaultFlushDelay
	}

	if s.Listener == nil {
		s.Listener = &defaultListener{}
	}
	if err := s.Listener.Init(s.Addr); err != nil {
		xlog.Errorf("cannot listen to: %s: %s", s.Addr, err.Error())
		return err
	}

	workersCh := make(chan struct{}, s.Concurrency)
	s.stopWg.Add(1)
	go s.serverHandler(workersCh)
	return nil
}

// Stop stops rpc server. Stopped server can be started again.
func (s *Server) Stop() {
	if s.serverStopChan == nil {
		xlog.Panic("server must be started before stopping it")
	}
	close(s.serverStopChan)
	s.stopWg.Wait()
	s.serverStopChan = nil
}

// Serve starts rpc server and blocks until it is stopped.
func (s *Server) Serve() error {
	if err := s.Start(); err != nil {
		return err
	}
	s.stopWg.Wait()
	return nil
}

func (s *Server) serverHandler(workersCh chan struct{}) {
	defer s.stopWg.Done()

	var conn net.Conn
	var err error
	var stopping atomic.Value

	for {
		acceptChan := make(chan struct{})
		go func() {
			if conn, err = s.Listener.Accept(); err != nil {
				xlog.Errorf("failed to accept: %s", err.Error())
				if stopping.Load() == nil {
					xlog.Errorf("cannot accept new connection: %s", err)
				}
			}
			close(acceptChan)
		}()

		select {
		case <-s.serverStopChan:
			stopping.Store(true)
			_ = s.Listener.Close()
			<-acceptChan
			return
		case <-acceptChan:
		}

		if err != nil {
			select {
			case <-s.serverStopChan:
				return
			case <-time.After(time.Second):
			}
			continue
		}

		s.stopWg.Add(1)
		go s.serverHandleConnection(conn, workersCh)
	}
}

func (s *Server) serverHandleConnection(conn net.Conn, workersCh chan struct{}) {
	defer s.stopWg.Done()

	var stopping atomic.Value
	var err error

	okHandshake := make(chan bool, 1)
	go func() {
		var buf [1]byte
		if _, err = conn.Read(buf[:]); err != nil {
			if stopping.Load() == nil {
				xlog.Errorf("failed to reading handshake from client: %s: %s", conn.RemoteAddr().String(), err)
			}
		}
		okHandshake <- buf[0] == 1
	}()

	select {
	case ok := <-okHandshake:
		if !ok || err != nil {
			_ = conn.Close()
			return
		}
	case <-s.serverStopChan:
		stopping.Store(true)
		_ = conn.Close()
		return
	case <-time.After(10 * time.Second):
		xlog.Errorf("cannot obtain handshake from client:%s during 10s", conn.RemoteAddr().String())
		_ = conn.Close()
		return
	}

	responsesChan := make(chan *serverMessage, s.PendingResponses)
	stopChan := make(chan struct{})

	readerDone := make(chan struct{})
	go s.serverReader(conn, responsesChan, stopChan, readerDone, workersCh)

	writerDone := make(chan struct{})
	go s.serverWriter(conn, responsesChan, stopChan, writerDone)

	select {
	case <-readerDone:
		close(stopChan)
		_ = conn.Close()
		<-writerDone
	case <-writerDone:
		close(stopChan)
		_ = conn.Close()
		<-readerDone
	case <-s.serverStopChan:
		close(stopChan)
		_ = conn.Close()
		<-readerDone
		<-writerDone
	}
}

type serverMessage struct {
	method   uint8
	msgID    uint64
	reqid    uint64
	oid      uint64
	extID    uint32
	bodySize uint32
	reqbody  []byte

	resp []byte
	err  error
}

var serverMessagePool = &sync.Pool{
	New: func() interface{} {
		return &serverMessage{}
	},
}

func (s *serverMessage) reset() {
	s.method = 0
	s.msgID = 0
	s.reqid = 0
	s.oid = 0
	s.extID = 0
	s.bodySize = 0
	s.reqbody = nil

	s.resp = nil
	s.err = nil
}

func (s *Server) serverReader(r net.Conn, responsesChan chan<- *serverMessage,
	stopChan <-chan struct{}, done chan<- struct{}, workersCh chan struct{}) {

	defer func() {
		if x := recover(); x != nil {
			stackTrace := make([]byte, 1<<20)
			n := runtime.Stack(stackTrace, false)
			xlog.Errorf("panic when reading data from client: %v\nStack trace: %s", x, stackTrace[:n])
		}
		close(done)
	}()

	hash := xdigest.New()
	dec := newDecoder(r, s.RecvBufferSize, hash)
	rh := new(reqHeader)
	headerBuf := make([]byte, reqHeaderSize)

	for {
		err := dec.decodeHeader(headerBuf, rh)
		if err != nil {
			if err == orpc.ErrTimeout {
				continue // Keeping trying to read request header.
			}
			xlog.Errorf("failed to read request header from %s: %s", r.RemoteAddr().String(), err)
			return
		}

		m := serverMessagePool.Get().(*serverMessage)
		m.method = rh.method
		m.msgID = rh.msgID
		m.reqid = rh.reqid
		m.oid = rh.oid
		m.extID = rh.extID
		m.bodySize = rh.bodySize

		n := int(m.bodySize)
		if n != 0 {
			body := xbytes.GetAlignedBytes(n)
			err = dec.decodeBody(body, n)
			if err != nil {
				xlog.ErrorIDf(m.reqid, "failed to read request body from %s: %s", r.RemoteAddr().String(), err.Error())
				xbytes.PutAlignedBytes(body)
				m.reset()
				serverMessagePool.Put(m)
				return
			}

			digest := uid.GetDigest(m.oid)
			actDigest := hash.Sum32()
			if actDigest != digest {
				xlog.ErrorID(m.reqid, xerrors.WithMessage(orpc.ErrChecksumMismatch, fmt.Sprintf("request exp: %d, but: %d", digest, actDigest)).Error())
				m.err = orpc.ErrChecksumMismatch
				xbytes.PutAlignedBytes(body)
			} else {
				m.reqbody = body
			}
			hash.Reset()
		}

		select {
		case workersCh <- struct{}{}:
		default:
			select {
			case workersCh <- struct{}{}:
			case <-stopChan:
				return
			}
		}

		// Haven read the request, handle request async, free the conn for the next request reading.
		go s.serveRequest(responsesChan, stopChan, m, workersCh)
	}
}

func (s *Server) serveRequest(responsesChan chan<- *serverMessage, stopChan <-chan struct{}, m *serverMessage, workersCh <-chan struct{}) {

	if m.err == nil {
		resp, err := s.callHandlerWithRecover(m.reqid, m.method, m.oid, m.extID, m.reqbody)
		m.resp = resp
		if err != nil {
			m.resp = nil
		}
		m.err = err
	}

	if m.reqbody != nil {
		xbytes.PutAlignedBytes(m.reqbody)
	}
	m.reqbody = nil

	// Select hack for better performance.
	// See https://github.com/valyala/gorpc/pull/1 for details.
	select {
	case responsesChan <- m:
	default:
		select {
		case responsesChan <- m:
		case <-stopChan:
		}
	}

	<-workersCh
}

func (s *Server) callHandlerWithRecover(reqid uint64, method uint8, oid uint64, extID uint32, reqBody []byte) (resp []byte, err error) {
	defer func() {
		if x := recover(); x != nil {
			stackTrace := make([]byte, 1<<20)
			n := runtime.Stack(stackTrace, false)
			err = fmt.Errorf("panic occured: %v\nStack trace: %s", x, stackTrace[:n])
			xlog.ErrorID(reqid, err.Error())
		}
	}()

	switch method {
	case objPutMethod:
		err = s.Handler.PutObj(reqid, oid, extID, reqBody)
	case objGetMethod:
		resp, err = s.Handler.GetObj(reqid, oid, extID, false)
	case objGetCloneMethod:
		resp, err = s.Handler.GetObj(reqid, oid, extID, true)
	case objDelMethod:
		err = s.Handler.DeleteObj(reqid, oid, extID)
	case objDelBatchMethod:
		err = s.Handler.DeleteBatch(reqid, extID, reqBody)
	default:
		err = orpc.ErrNotImplemented
	}

	return
}

func isServerStop(stopChan <-chan struct{}) bool {
	select {
	case <-stopChan:
		return true
	default:
		return false
	}
}

func (s *Server) serverWriter(w net.Conn, responsesChan <-chan *serverMessage, stopChan <-chan struct{}, done chan<- struct{}) {
	defer func() { close(done) }()

	t := time.NewTimer(s.FlushDelay)
	var flushChan <-chan time.Time
	enc := newEncoder(w, s.SendBufferSize)
	msg := new(msgBytes)
	rh := new(respHeader)
	headerBuf := make([]byte, respHeaderSize) // reqHeaderSize is bigger than respHeaderSize.

	for {
		var m *serverMessage

		select {
		case m = <-responsesChan:
		default:
			// Give the last chance for ready goroutines filling responsesChan :)
			runtime.Gosched()

			select {
			case <-stopChan:
				return
			case m = <-responsesChan:
			case <-flushChan:
				if err := enc.flush(); err != nil {
					if !isServerStop(stopChan) {
						xlog.Errorf("server cannot flush requests to: %s: %s", w.RemoteAddr().String(), err)
					}
					return
				}
				flushChan = nil
				continue
			}
		}

		if flushChan == nil {
			flushChan = xtime.GetTimerEvent(t, s.FlushDelay)
		}

		resp := m.resp
		reqid := m.reqid

		rh.msgID = m.msgID
		if resp != nil {
			rh.bodySize = uint32(len(resp))
		} else {
			rh.bodySize = 0
		}
		rh.errno = uint16(orpc.ErrToErrno(m.err))
		msg.header = rh
		msg.body = resp

		m.reset()
		serverMessagePool.Put(m)

		if err := enc.encodeBytesPool(msg, headerBuf); err != nil {

			xlog.ErrorIDf(reqid, "failed to send response to: %s: %s", w.RemoteAddr().String(), err)
			return
		}

		msg.body = nil
	}
}
