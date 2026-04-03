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
	"testing"

	"github.com/zaibyte/zaipkg/orpc"

	"github.com/stretchr/testify/assert"

	"github.com/zaibyte/zaipkg/uid"
)

func TestRequestHeaderCanBeEncodedAndDecoded(t *testing.T) {
	r := &reqHeader{
		method:   1,
		msgID:    2,
		reqid:    uid.MakeReqID(),
		offset:   3,
		wantSize: 7,
		bodySize: 4,
		oid:      5,
		extID:    6,
	}
	buf := make([]byte, reqHeaderSize)
	result := r.encode(buf)
	assert.Equal(t, reqHeaderSize, len(result))

	rr := &reqHeader{}
	assert.Nil(t, rr.decode(result))

	assert.Equal(t, r, rr)
}

func TestRequestHeaderCRCIsChecked(t *testing.T) {
	r := &reqHeader{
		method:   1,
		msgID:    2,
		reqid:    uid.MakeReqID(),
		offset:   3,
		wantSize: 7,
		bodySize: 4,
		oid:      5,
		extID:    6,
	}
	buf := make([]byte, reqHeaderSize)
	result := r.encode(buf)
	assert.Equal(t, reqHeaderSize, len(result))

	rr := &reqHeader{}
	assert.Nil(t, rr.decode(result))

	crc := binary.BigEndian.Uint32(result[41:])
	binary.BigEndian.PutUint32(result[41:], crc+1)
	assert.Equal(t, orpc.ErrChecksumMismatch, rr.decode(result))

	binary.BigEndian.PutUint32(result[41:], crc)
	assert.Nil(t, rr.decode(result))

	result[0] = 0
	assert.Equal(t, orpc.ErrChecksumMismatch, rr.decode(result))
}

func TestRespHeaderCanBeEncodedAndDecoded(t *testing.T) {
	r := &respHeader{
		msgID:    2048,
		errno:    22,
		bodySize: 1024,
		bodyCrc:  1025,
	}
	buf := make([]byte, respHeaderSize)
	result := r.encode(buf)
	assert.Equal(t, respHeaderSize, len(result))

	rr := &respHeader{}
	assert.Nil(t, rr.decode(result))

	assert.Equal(t, r, rr)
}

func TestRespHeaderCRCIsChecked(t *testing.T) {
	r := &respHeader{
		msgID:    2048,
		errno:    22,
		bodySize: 1024,
		bodyCrc:  1024,
	}
	buf := make([]byte, respHeaderSize)
	result := r.encode(buf)
	assert.Equal(t, respHeaderSize, len(result))

	rr := respHeader{}
	assert.Nil(t, rr.decode(result))

	crc := binary.BigEndian.Uint32(result[18:])
	binary.BigEndian.PutUint32(result[18:], crc+1)
	assert.Equal(t, orpc.ErrChecksumMismatch, rr.decode(result))

	binary.BigEndian.PutUint32(result[18:], crc)
	assert.Nil(t, rr.decode(result))

	binary.BigEndian.PutUint64(result[0:], 0)
	assert.Equal(t, orpc.ErrChecksumMismatch, rr.decode(result))
}
