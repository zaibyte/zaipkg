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
// Package xdigest provides hash functions on byte sequences by wrapping crc32.
// These hash functions are intended to be used to implement zai object digest
// that need to map byte sequences to a uniform distribution on unsigned 32-bit integers.
//
// For Zai, the needs of object digest:
// 1. Fast.
// candidates: xxh3, xxhash, hashes based on AES, crc32
// 2. Low collisions with 32bits sum.
// candidates: xxh3_low, xxhash_low (hashes based on AES can't pass smhasher sparse test, crc32 can't pass smhasher avalanche test)
// candidates*: crc32
// *What smhasher avalanche does: Flipping a single bit of a key should flip each output bit with 50% probability.
// What we want: we don't need 50%, there is one bit flipping is enough. Anyway, crc32 can't pass all smhasher test,
// but it's still a good one for our needs.
// 3. Stable
// candidates: xxh3, xxhash, crc32
// 4. Could satisfy hash.Hash32 interface.
// candidates: xxhash (xxh3 only has one-shot hash API) & crc32
// 5. *xxhash_low can't work well with hopscotch hashing which being used in zbuf, it cause lots of hashing conflicting,
// crc32 does a good job there.
// final answer: crc32.
// TODO Maybe use wyhash in future.
//
// Warn:
// Don't change the digest algorithm after a zai cluster starting.
package xdigest

import (
	"hash"
	"hash/crc32"

	"g.tesamc.com/IT/zaipkg/xstrconv"
)

type Digest struct {
	h32 hash.Hash32
}

// crcTable uses Castagnoli which has better error detection characteristics than IEEE and faster.
var crcTable = crc32.MakeTable(crc32.Castagnoli)

// New creates a Digest.
func New() *Digest {

	return &Digest{h32: crc32.New(crcTable)}
}

// Write (via the embedded io.Writer interface) adds more data to the running hash.
// It never returns an error.
func (d *Digest) Write(b []byte) (n int, err error) {
	return d.h32.Write(b)
}

// WriteString (via the embedded io.Writer interface) adds more data to the running hash.
// It never returns an error.
func (d *Digest) WriteString(s string) (n int, err error) {
	return d.Write(xstrconv.ToBytes(s))
}

// Sum appends the current hash to b and returns the resulting slice.
// It does not change the underlying hash state.
func (d *Digest) Sum(b []byte) []byte {
	return d.h32.Sum(b)
}

// Sum32 returns the current hash.
func (d *Digest) Sum32() uint32 {
	return d.h32.Sum32()
}

// Reset resets the Hash to its initial state.
func (d *Digest) Reset() {
	d.h32.Reset()
}

// Size returns the number of bytes Sum will return.
func (d *Digest) Size() int {
	return 4
}

// BlockSize returns the hash's underlying block size.
// The Write method must be able to accept any amount
// of data, but it may operate more efficiently if all writes
// are a multiple of the block size.
func (d *Digest) BlockSize() int {
	return d.h32.BlockSize()
}

// Sum32 computes the 32-bit digest of b directly.
func Sum32(b []byte) uint32 {
	return crc32.Checksum(b, crcTable)
}
