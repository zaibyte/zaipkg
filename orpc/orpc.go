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
	"io"
	"time"

	"g.tesamc.com/IT/zaipkg/xbytes"
)

// Client is the object RPC client.
type Client interface {
	// Start Client.
	Start() error
	// Stop Client, release resource.
	Close() error
	// Put puts object to the ZBuf node which Client connected.
	PutObj(reqid uint64, oid string, objData []byte, timeout time.Duration) error
	// Get gets object from the ZBuf node which Client connected.
	GetObj(reqid uint64, oid string, timeout time.Duration) (obj io.ReadCloser, err error)
	// Delete deletes object in the ZBuf node which Client connected.
	DeleteObj(reqid uint64, oid string, timeout time.Duration) error
}

// Server is the object RPC server.
type Server interface {
	// RegisterHandler registers handlers for server.
	// Call it before Start().
	RegisterHandler(putFunc PutFunc, getFunc GetFunc, deleteFunc DeleteFunc)
	// Start Server.
	Start() error
	// Stop Server, release resource.
	Close() error
}

// PutFunc is the object put function.
type PutFunc func(reqid, oid uint64, objData xbytes.Buffer) error

// GetFunc is the object get function.
type GetFunc func(reqid, oid uint64) (objData xbytes.Buffer, err error)

// DeleteFunc is the object delete function.
type DeleteFunc func(reqid, oid uint64) error
