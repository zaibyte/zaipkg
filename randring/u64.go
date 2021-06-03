package randring

import (
	"sync/atomic"

	"github.com/templexxx/cpu"
)

const falseSharingRange = cpu.X86FalseSharingRange

// U64Ring provides a uint64 ring buckets for multi-producer & one-consumer.
// For buffer deletion.
type U64Ring struct {
	mask       uint64
	_          [falseSharingRange]byte
	writeIndex uint64
	_          [falseSharingRange]byte

	// writeIndex cache for Pop, only get new write index when read catch write.
	// Help to reduce caching missing.
	writeIndexCache uint64
	_               [falseSharingRange]byte
	readIndex       uint64

	buckets []uint64
}

// New creates a ring.
// ring size = 2 ^ n.
func New(n uint64) *U64Ring {

	if n > 16 || n == 0 {
		panic("illegal ring size")
	}

	r := &U64Ring{
		buckets: make([]uint64, 1<<n),
		mask:    (1 << n) - 1,
	}

	// Start write index at the value before 0
	// to allow the first write to use AddUint64
	// and still have a beginning index of 0
	r.writeIndex = ^r.writeIndex
	return r
}

// Push puts the data in ring in the next bucket no matter what in it.
func (r *U64Ring) Push(v uint64) {
	idx := atomic.AddUint64(&r.writeIndex, 1) & r.mask
	atomic.SwapUint64(&r.buckets[idx], v)
}

// TryPop tries to pop data from the next bucket,
// return (nil, false) if no data available.
func (r *U64Ring) TryPop() (uint64, bool) {

	if r.readIndex > r.writeIndexCache {
		r.writeIndexCache = atomic.LoadUint64(&r.writeIndex)
		if r.readIndex > r.writeIndexCache {
			return 0, false
		}
	}

	idx := r.readIndex & r.mask
	data := atomic.SwapUint64(&r.buckets[idx], 0)

	if data == 0 {
		return 0, false
	}

	r.readIndex++
	return data, true
}
