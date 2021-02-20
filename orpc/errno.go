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

import "errors"

// An Errno is an unsigned number describing an error condition.
// It implements the error interface. The zero Errno is by convention
// a non-error, so code to convert from Errno to error should use:
//	err = nil
//	if errno != 0 {
//		err = errno
//	}
//
// Errno provides error numbers for indicating errors.
// Saving network I/O and marshal & unmarshal cost in RPC between Zai and ZBuf/ZCold.
//
// We don't need to support all error types, because any error should be
// logged in where it raises. For the client, it just need
// to know there is an error in the request, and what the type it is.
//
// There are two major types of errno:
// 1. Server side
// Not found error & internal server error.
// Not found: means there is no need to retry, so it's important.
// Other errors could be combined as internal server error.
//
// 2. Client side
// Bad request, not implemented*, canceled, timeout, too many request*, connection error
// Bad request: could happen when there is an illegal request.
// Not implemented: request a method which is not found.
// For saving network cost, the method will be checked in client side.
// Connection error means network issues.

type Errno uint16

func (e Errno) Error() string {

	if e == 0 {
		return ""
	}

	if int(e) < len(errnoStr) {
		s := errnoStr[uint16(e)]
		if s != "" {
			return s
		}
	}
	return "unknown error"
}

func (e Errno) ToErr() error {
	if e == 0 {
		return nil
	}

	return e
}

// ErrToErrno returns Errno value by error.
func ErrToErrno(err error) Errno {
	if err == nil {
		return 0
	}

	for {
		err2 := errors.Unwrap(err)
		if err2 == nil {
			break
		}
		err = err2
	}

	u, ok := err.(Errno)
	if ok {
		return u
	}

	return Errno(internalServerError)
}

// Error which supports retrying will start from 10000.
const (
	RetryStart = 10000
	RetryEnd   = 19999
)

const (
	timeout              = 10004
	tooManyRequests      = 10005
	internalServerError  = 10006
	canceled             = 10008
	requestQueueOverflow = 10011

	badRequest     = 1
	notFound       = 2
	notImplemented = 3

	connectionError  = 7
	checksumMismatch = 9
	invalidMethod    = 10
	timeBackwards    = 12
	notBootstrapped  = 13

	instanceDisconnected = 14
	instanceDown         = 15
	instanceOffline      = 16
	instanceTombstone    = 17

	diskFull      = 18
	diskBroken    = 19
	diskOffline   = 20
	diskTombstone = 21

	extentFull   = 22
	extentBroken = 23
	extentGhost  = 26
	extentClone  = 27
	extentSealed = 28

	objDigestExisted = 30

	closed = 31
)

// Error table.
// Please add errno in order.
var errnoStr = map[uint16]string{
	badRequest:           "bad message",
	notFound:             "not found",
	notImplemented:       "not implemented",
	timeout:              "timeout",
	tooManyRequests:      "too many requests",
	internalServerError:  "internal server error",
	connectionError:      "connection error",
	canceled:             "canceled",
	checksumMismatch:     "checksum mismatch",
	invalidMethod:        "invalid method",
	requestQueueOverflow: "request queue overflow",
	timeBackwards:        "time gone backwards",
	notBootstrapped:      "not bootstrapped",

	instanceDisconnected: "instance is disconnected",
	instanceDown:         "instance is down",
	instanceOffline:      "instance is offline",
	instanceTombstone:    "instance is tombstone",

	diskFull:      "disk is full",
	diskBroken:    "disk is broken",
	diskOffline:   "disk is offline",
	diskTombstone: "disk is tombstone",

	extentFull:   "extent is full",
	extentBroken: "extent is broken",
	extentClone:  "extent is in clone",
	extentSealed: "extent is sealed",

	objDigestExisted: "object digest is existed in this group",

	closed: "service is closed",
}

var (
	ErrBadRequest           = Errno(badRequest)
	ErrNotFound             = Errno(notFound) // When server side raises a not found error, using this variable.
	ErrNotImplemented       = Errno(notImplemented)
	ErrTimeout              = Errno(timeout)
	ErrTooManyRequests      = Errno(tooManyRequests)
	ErrInternalServer       = Errno(internalServerError)
	ErrConnection           = Errno(connectionError)
	ErrCanceled             = Errno(canceled)
	ErrChecksumMismatch     = Errno(checksumMismatch)
	ErrInvalidMethod        = Errno(invalidMethod)
	ErrRequestQueueOverflow = Errno(requestQueueOverflow)
	ErrTimeBackwards        = Errno(timeBackwards)
	ErrNotBootstrapped      = Errno(notBootstrapped)

	ErrInstanceDisconnected = Errno(instanceDisconnected)
	ErrInstanceDown         = Errno(instanceDown)
	ErrInstanceOffline      = Errno(instanceOffline)
	ErrInstanceTombstone    = Errno(instanceTombstone)

	ErrDiskFull      = Errno(diskFull)
	ErrDiskBroken    = Errno(diskBroken)
	ErrDiskOffline   = Errno(diskOffline)
	ErrDiskTombstone = Errno(diskTombstone)

	ErrExtentFull   = Errno(extentFull)
	ErrExtentBroken = Errno(extentBroken)
	ErrExtentGhost  = Errno(extentGhost)
	ErrExtentClone  = Errno(extentClone)
	ErrExtentSealed = Errno(extentSealed)

	ErrObjDigestExisted = Errno(objDigestExisted)

	ErrServiceClosed = Errno(closed)
)

// StrError is using for other network transport to convert string message to a certain error type.
func StrError(str string) error {
	var en uint16
	for i := range errnoStr {
		if str == errnoStr[i] {
			en = uint16(i)
		}
	}
	return Errno(en)
}
