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
// Copyright 2017-2019 Lei Ni (nilei81@gmail.com) and other Dragonboat authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// This file contains code derived from Dragonboat.
// The main logic & codes are copied from Dragonboat.

package otcp

import (
	"encoding/binary"

	"g.tesamc.com/IT/zaipkg/xchecksum"

	"g.tesamc.com/IT/zaipkg/orpc"
)

type header interface {
	encode(b []byte) []byte
	decode(b []byte) error
	getBodySize() uint32
}

const reqHeaderSize = 45

// reqHeader is the header for request.
type reqHeader struct {
	method   uint8  // [0, 1)
	msgID    uint64 // [1, 9)
	reqid    uint64 // [9, 17)
	offset   uint32 // [17, 21)
	wantSize uint32 // [21, 25)	// It's wanted response body size.
	bodySize uint32 // [25, 29)	// It's request body size.
	oid      uint64 // [29, 37)
	extID    uint32 // [37, 41)
	crc      uint32 // [41, 45)
}

func (h *reqHeader) encode(buf []byte) []byte {
	if len(buf) < reqHeaderSize {
		panic("input buf too small")
	}
	buf[0] = h.method
	binary.BigEndian.PutUint64(buf[1:9], h.msgID)
	binary.BigEndian.PutUint64(buf[9:17], h.reqid)
	binary.BigEndian.PutUint32(buf[17:21], h.offset)
	binary.BigEndian.PutUint32(buf[21:25], h.wantSize)
	binary.BigEndian.PutUint32(buf[25:29], h.bodySize)
	binary.BigEndian.PutUint64(buf[29:37], h.oid)
	binary.BigEndian.PutUint32(buf[37:41], h.extID)
	binary.BigEndian.PutUint32(buf[41:45], 0)
	crc := xchecksum.Sum32(buf[:reqHeaderSize])
	binary.BigEndian.PutUint32(buf[41:45], crc)
	h.crc = crc
	return buf[:reqHeaderSize]
}

func (h *reqHeader) decode(buf []byte) error {
	if len(buf) < reqHeaderSize {
		panic("input buf too small")
	}

	incoming := binary.BigEndian.Uint32(buf[41:45])
	binary.BigEndian.PutUint32(buf[41:45], 0)
	expected := xchecksum.Sum32(buf[:reqHeaderSize])
	if incoming != expected {
		return orpc.ErrChecksumMismatch
	}
	binary.BigEndian.PutUint32(buf[41:45], incoming)

	h.method = buf[0]
	h.msgID = binary.BigEndian.Uint64(buf[1:9])
	h.reqid = binary.BigEndian.Uint64(buf[9:17])
	h.offset = binary.BigEndian.Uint32(buf[17:21])
	h.wantSize = binary.BigEndian.Uint32(buf[21:25])
	h.bodySize = binary.BigEndian.Uint32(buf[25:29])
	h.oid = binary.BigEndian.Uint64(buf[29:37])
	h.extID = binary.BigEndian.Uint32(buf[37:41])
	h.crc = incoming

	return nil
}

func (h *reqHeader) getBodySize() uint32 {
	return h.bodySize
}

const respHeaderSize = 22

// respHeader is the header for response.
type respHeader struct {
	msgID    uint64 // [0, 8)
	errno    uint16 // [8, 10)
	bodySize uint32 // [10, 14)
	bodyCrc  uint32 // [14, 18)
	crc      uint32 // [18, 22)
}

func (h *respHeader) encode(buf []byte) []byte {
	if len(buf) < respHeaderSize {
		panic("input buf too small")
	}
	binary.BigEndian.PutUint64(buf[0:8], h.msgID)
	binary.BigEndian.PutUint16(buf[8:10], h.errno)
	binary.BigEndian.PutUint32(buf[10:14], h.bodySize)
	binary.BigEndian.PutUint32(buf[14:18], h.bodyCrc)
	binary.BigEndian.PutUint32(buf[18:22], 0)
	crc := xchecksum.Sum32(buf[:respHeaderSize])
	binary.BigEndian.PutUint32(buf[18:22], crc)
	h.crc = crc
	return buf[:respHeaderSize]
}

func (h *respHeader) decode(buf []byte) error {
	if len(buf) < respHeaderSize {
		panic("input buf too small")
	}

	incoming := binary.BigEndian.Uint32(buf[18:22])
	binary.BigEndian.PutUint32(buf[18:22], 0)
	expected := xchecksum.Sum32(buf[:respHeaderSize])
	if incoming != expected {
		return orpc.ErrChecksumMismatch
	}
	binary.BigEndian.PutUint32(buf[18:22], incoming)

	h.msgID = binary.BigEndian.Uint64(buf[0:8])
	h.errno = binary.BigEndian.Uint16(buf[8:10])
	h.bodySize = binary.BigEndian.Uint32(buf[10:14])
	h.bodyCrc = binary.BigEndian.Uint32(buf[14:18])
	h.crc = incoming
	return nil
}

func (h *respHeader) getBodySize() uint32 {
	return h.bodySize
}
