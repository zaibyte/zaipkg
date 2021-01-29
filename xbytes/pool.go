// xbytes provdies bytes slice pool.
//
// Warning:
// 1. Do not use it when you only need <= 32KB byte slice, and this slice will not escape to the heap.
// In this situation, using sync.Pool will let the slice escapes to the heap, bringing the extra GC overhead.
// Discussion in: https://g.tesamc.com/IT/zaipkg/issues/11
// 2. Do not use it when you want bytes more than 4MB. Otherwise, it'll panic.
package xbytes

import (
	"sync"

	"g.tesamc.com/IT/zaipkg/config/settings"
	"g.tesamc.com/IT/zaipkg/directio"
)

var (
	_defaultPool = NewPool(defaultMaxLeaky, false)
	_alignPool   = NewPool(defaultMaxLeaky, true)

	GetBytes = func(n int) []byte {
		if n <= _MaxSmallSize {
			return _defaultPool.smallPool.Get().([]byte)[:n]
		}
		return _defaultPool.largePool.Get()[:n]
	}
	PutBytes = func(b []byte) {
		n := len(b)
		if n <= _MaxSmallSize {
			_defaultPool.smallPool.Put(b)
			return
		}
		_defaultPool.largePool.Put(b)
	}

	GetAlignedBytes = func(n int) []byte {
		if n <= _MaxSmallSize {
			return _alignPool.smallPool.Get().([]byte)[:n]
		}
		return _alignPool.largePool.Get()[:n]
	}
	PutAlignedBytes = func(b []byte) {
		n := len(b)
		if n <= _MaxSmallSize {
			_alignPool.smallPool.Put(b)
			return
		}
		_alignPool.largePool.Put(b)
	}
)

const (
	// _MaxSmallSize is copied from runtime/sizeclasses,
	// indicates it's a small object, it'll be malloc from Process's own cache firstly.
	// We will use sync.Pool to implement bytes pool if size <= _MaxSmallSize
	_MaxSmallSize   = 32768
	_MaxSize        = settings.MaxObjectSize
	defaultMaxLeaky = 256 // Which means it'll reach 1GB memory never being freed.
)

// BufferPool is a bytes slice pool helping to reduce GC overhead.
type BufferPool struct {
	smallPool *sync.Pool
	largePool *LeakyBuf
}

// NewPool creates a bytes slice pool.
func NewPool(maxLarge int, isAligned bool) *BufferPool {

	var makeSmallFn, makeLargeFn func() []byte
	if isAligned {
		makeSmallFn = func() []byte {
			return directio.AlignedBlock(_MaxSmallSize)
		}
		makeLargeFn = func() []byte {
			return directio.AlignedBlock(_MaxSize)
		}
	} else {
		makeSmallFn = func() []byte {
			return make([]byte, 0, _MaxSmallSize)
		}
		makeLargeFn = func() []byte {
			return make([]byte, 0, _MaxSize)
		}
	}

	return &BufferPool{
		smallPool: &sync.Pool{New: func() interface{} {
			return makeSmallFn()
		}},
		largePool: NewLeakyBuf(maxLarge, makeLargeFn),
	}
}
