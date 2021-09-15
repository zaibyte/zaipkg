// Package limitring is inspired by https://github.com/cloudfoundry/go-diodes with these changes:
// 1. Inject Push when ring is full
// 2. Overlap/losing data is not acceptable

package limitring

import (
	"sync/atomic"
	"unsafe"

	"g.tesamc.com/IT/zaipkg/orpc"

	"github.com/templexxx/cpu"
)

const falseSharingRange = cpu.X86FalseSharingRange

// Ring is optimal for many writers (go-routines B-n) and a single
// reader (go-routine A). It is not thread safe for multiple readers.
type Ring struct {
	mask uint64

	_          [falseSharingRange]byte
	writeIndex uint64
	_          [falseSharingRange]byte
	// writeIndex cache for Pop, only get new write index when read catch write.
	// Help to reduce caching missing.
	writeIndexCache uint64
	_               [falseSharingRange]byte
	cnt             uint64
	_               [falseSharingRange]byte
	readIndex       uint64
	buffer          []unsafe.Pointer
}

// New creates a new diode (ring buffer). The Ring diode
// is optimzed for many writers (on go-routines B-n) and a single reader
// (on go-routine A).
//
// ring size = 2 ^ n.
// The min ring size is 128.
// The max ring size is 2^20.
func New(n uint64) *Ring {

	if n <= 7 {
		n = 7
	}
	if n >= 20 {
		n = 20
	}

	d := &Ring{
		buffer: make([]unsafe.Pointer, 1<<n),
		mask:   (1 << n) - 1,
	}

	// Start write index at the value before 0
	// to allow the first write to use AddUint64
	// and still have a beginning index of 0
	d.writeIndex = ^d.writeIndex
	return d
}

// Push pushes the data in the next slot of the ring buffer.
// If Ring is almost full or there is a collision, return orpc.ErrRequestQueueOverflow directly.
func (r *Ring) Push(data unsafe.Pointer) error {

	cnt := atomic.LoadUint64(&r.cnt)
	if cnt+16 > r.mask { // 16 for reducing collision rate.
		return orpc.ErrRequestQueueOverflow
	}

	writeIndex := atomic.AddUint64(&r.writeIndex, 1)
	idx := writeIndex & r.mask
	old := atomic.LoadPointer(&r.buffer[idx])

	if old != nil {
		return orpc.ErrRequestQueueOverflow
	}

	if !atomic.CompareAndSwapPointer(&r.buffer[idx], old, data) {
		return orpc.ErrRequestQueueOverflow // Actually it's another thread's side effect.
	}

	atomic.AddUint64(&r.cnt, 1)
	return nil
}

// Pop will attempt to read from the next slot of the ring buffer.
// If there is no data available, it will return (nil, false).
func (r *Ring) Pop() (data unsafe.Pointer, ok bool) {

	if r.readIndex > r.writeIndexCache {
		r.writeIndexCache = atomic.LoadUint64(&r.writeIndex)
		if r.readIndex > r.writeIndexCache {
			return nil, false // Read catch up write.
		}
	}

	// Read a value from the ring buffer based on the readIndex.
	idx := r.readIndex & r.mask
	result := atomic.SwapPointer(&r.buffer[idx], nil)

	// When the result is nil that means the writer has not had the
	// opportunity to write a value into the diode. This value must be ignored
	// and the read head must not increment.
	if result == nil {
		return nil, false
	}

	r.readIndex++
	atomic.AddUint64(&r.cnt, ^uint64(0))
	return result, true
}
