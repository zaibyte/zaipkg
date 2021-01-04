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

	"g.tesamc.com/IT/zaipkg/xbytes"
)

// Client is the object RPC client.
type Client interface {
	// Start Client.
	Start() error
	// Stop Client, release resource.
	Stop() error
	// Put puts object to the ZBuf node which Client connected.
	PutObj(reqid uint64, oid uint64, objData []byte, timeout time.Duration) error
	// Get gets object from the ZBuf node which Client connected.
	GetObj(reqid uint64, oid uint64, objData []byte, timeout time.Duration) error
	// Delete deletes object in the ZBuf node which Client connected.
	DeleteObj(reqid uint64, oid uint64, timeout time.Duration) error
}

// Server is the object RPC server.
type Server interface {
	// Start Server.
	Start() error
	// Stop Server, release resource.
	Stop() error
}

// Handler is the object rpc handler.
type Handler interface {
	PutObj(reqid, oid uint64, objData []byte) error
	GetObj(reqid, oid uint64) (objData xbytes.Buffer, err error)
	DeleteObj(reqid, oid uint64) error
}
