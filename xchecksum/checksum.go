// Package xchecksum provides Application layer checksum,
// avoiding silent data corruption in header,
// checksum should be ignored only when TLS is enabled.
//
// These hash functions are intended to be used to implement zai object digest
// that need to map byte sequences to a uniform distribution on unsigned 32-bit integers.
//
// CRC32 is enough fast and reliable.
package xchecksum

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

// Sum32 computes the 32-bit checksum of b directly.
func Sum32(b []byte) uint32 {
	return crc32.Checksum(b, crcTable)
}
