package xdigest

import (
	"hash"

	"g.tesamc.com/IT/zaipkg/xstrconv"

	"github.com/cespare/xxhash/v2"
)

type Digest64 struct {
	h64 hash.Hash64
}

// New64 creates a Digest64.
func New64() *Digest64 {

	return &Digest64{h64: xxhash.New()}
}

// Write (via the embedded io.Writer interface) adds more data to the running hash.
// It never returns an error.
func (d *Digest64) Write(b []byte) (n int, err error) {
	return d.h64.Write(b)
}

// WriteString (via the embedded io.Writer interface) adds more data to the running hash.
// It never returns an error.
func (d *Digest64) WriteString(s string) (n int, err error) {
	return d.Write(xstrconv.ToBytes(s))
}

// Sum appends the current hash to b and returns the resulting slice.
// It does not change the underlying hash state.
func (d *Digest64) Sum(b []byte) []byte {
	return d.h64.Sum(b)
}

// Sum64 returns the current hash.
func (d *Digest64) Sum64() uint64 {
	return d.h64.Sum64()
}

// Reset resets the Hash to its initial state.
func (d *Digest64) Reset() {
	d.h64.Reset()
}

// Size returns the number of bytes Sum will return.
func (d *Digest64) Size() int {
	return 8
}

// BlockSize returns the hash's underlying block size.
// The Write method must be able to accept any amount
// of data, but it may operate more efficiently if all writes
// are a multiple of the block size.
func (d *Digest64) BlockSize() int {
	return d.h64.BlockSize()
}

// Sum64 computes the 64-bit digest of b directly.
func Sum64(b []byte) uint64 {
	return xxhash.Sum64(b)
}
