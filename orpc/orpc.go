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

package orpc

import (
	"time"
)

// Client is the object RPC client.
type Client interface {
	// Start starts Client.
	Start() error
	// Close closes Client with an error which will be passed to the pending requests.
	Close(err error)

	// The reason why we need extID for these methods:
	// 1. For PutObj, we must know which extent want to put. Same as DeleteObj / DeleteBatch.
	// 2. In test env, we may deploy extents in the same group in the same device(or don't care the geo disaster).
	// If we don't pass extID in this situation, we may cannot find the right extent.
	//
	//
	// PutObj puts object to the ZBuf node which Client connected.
	PutObj(reqid uint64, oid uint64, extID uint32, objData []byte, timeout time.Duration) error
	// GetObj gets object from the ZBuf node which Client connected.
	GetObj(reqid uint64, oid uint64, extID uint32, objData []byte, isClone bool, timeout time.Duration) error
	// DeleteObj deletes object in the ZBuf node which Client connected.
	DeleteObj(reqid uint64, oid uint64, extID uint32, timeout time.Duration) error
	// DeleteBatch deletes multi objects in a single RPC call.
	DeleteBatch(reqid uint64, oids []uint64, extID uint32, timeout time.Duration) error
}

// Server is the object RPC server.
type Server interface {
	// Start Server.
	Start() error
	// Stop Server, release resource.
	Stop() error
}

// ServerHandler is the object rpc handler.
type ServerHandler interface {
	PutObj(reqid, oid uint64, extID uint32, objData []byte) error
	GetObj(reqid, oid uint64, extID uint32, isClone bool) (objData []byte, err error)
	DeleteObj(reqid, oid uint64, extID uint32) error
	DeleteBatch(reqid uint64, extID uint32, oids []byte) error
}
