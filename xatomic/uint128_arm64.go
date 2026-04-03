package xatomic

import (
	"sync"
	"unsafe"
)

// arm64 fallback: serialize all 16-byte operations to preserve atomic semantics.
// This is slower than amd64 CMPXCHG16B but correct and portable on Apple Silicon.
var uint128FallbackMu sync.Mutex

// AtomicLoad16B atomically loads 16bytes from *addr.
// addr should be 16bytes aligned for API compatibility with amd64.
func AtomicLoad16B(addr *byte) [16]byte {
	uint128FallbackMu.Lock()
	defer uint128FallbackMu.Unlock()

	var v [16]byte
	copy(v[:], unsafe.Slice(addr, 16))
	return v
}

// AtomicStore16B atomically stores 16bytes to *addr.
// addr should be 16bytes aligned for API compatibility with amd64.
func AtomicStore16B(addr *byte, val [16]byte) {
	uint128FallbackMu.Lock()
	defer uint128FallbackMu.Unlock()

	copy(unsafe.Slice(addr, 16), val[:])
}

// AtomicCAS16B executes compare-and-swap for a 16-byte value.
// addr should be 16bytes aligned for API compatibility with amd64.
func AtomicCAS16B(addr, old, new *byte) (swapped bool) {
	uint128FallbackMu.Lock()
	defer uint128FallbackMu.Unlock()

	cur := unsafe.Slice(addr, 16)
	oldV := unsafe.Slice(old, 16)
	for i := 0; i < 16; i++ {
		if cur[i] != oldV[i] {
			return false
		}
	}
	copy(cur, unsafe.Slice(new, 16))
	return true
}
