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
	_defaultPool = NewPool(defaultTinySizeLeaky, defaultSmallSizeLeaky, defaultMidSizeLeaky, defaultMaxSizeLeaky, false)
	_alignPool   = NewPool(defaultTinySizeLeaky, defaultSmallSizeLeaky, defaultMidSizeLeaky, defaultMaxSizeLeaky, true)

	GetBytes = func(n int) []byte {
		if n <= _MaxSyncPoolSize {
			return _defaultPool.spPool.Get().([]byte)[:n]
		}
		if n <= _MaxTinySize {
			return _defaultPool.tinyPool.Get()[:n]
		}
		if n <= _MaxSmallSize {
			return _defaultPool.smallPool.Get()[:n]
		}
		if n <= _MaxMidSize {
			return _defaultPool.MidPool.Get()[:n]
		}
		return _defaultPool.maxPool.Get()[:n]
	}
	PutBytes = func(b []byte) {
		n := len(b)
		if n <= _MaxSyncPoolSize {
			_defaultPool.spPool.Put(b)
			return
		}
		if n <= _MaxTinySize {
			_defaultPool.tinyPool.Put(b)
			return
		}
		if n <= _MaxSmallSize {
			_defaultPool.smallPool.Put(b)
			return
		}
		if n <= _MaxMidSize {
			_defaultPool.MidPool.Put(b)
			return
		}
		_defaultPool.maxPool.Put(b)
	}

	GetAlignedBytes = func(n int) []byte {
		if n <= _MaxSyncPoolSize {
			return _alignPool.spPool.Get().([]byte)[:n]
		}
		if n <= _MaxTinySize {
			return _alignPool.tinyPool.Get()[:n]
		}
		if n <= _MaxSmallSize {
			return _alignPool.smallPool.Get()[:n]
		}
		if n <= _MaxMidSize {
			return _alignPool.MidPool.Get()[:n]
		}
		return _alignPool.maxPool.Get()[:n]
	}
	PutAlignedBytes = func(b []byte) {
		n := len(b)
		if n <= _MaxSyncPoolSize {
			_alignPool.spPool.Put(b)
			return
		}
		if n <= _MaxTinySize {
			_alignPool.tinyPool.Put(b)
			return
		}
		if n <= _MaxSmallSize {
			_alignPool.smallPool.Put(b)
			return
		}
		if n <= _MaxMidSize {
			_alignPool.MidPool.Put(b)
			return
		}
		_alignPool.maxPool.Put(b)
	}
)

const (
	// _MaxSyncPoolSize is copied from runtime/sizeclasses,
	// indicates it's a small object, it'll be malloc from Process's own cache firstly.
	// Beyond this, malloc maybe much slower, it's better to use LeakyPool.
	// We will use sync.Pool to implement bytes pool if size <= _MaxSyncPoolSize
	_MaxSyncPoolSize = 32768
	_MaxTinySize     = _MaxSyncPoolSize * 4   // 128KB.
	_MaxSmallSize    = _MaxTinySize * 4       // 512KB.
	_MaxMidSize      = _MaxSmallSize * 4      // 2MB.
	_MaxSize         = settings.MaxObjectSize // 4MB.

	defaultTinySizeLeaky  = 1024 // 128MB leaky at most for 128KB []byte.
	defaultSmallSizeLeaky = 256  // 128MB leaky at most for 512KB []byte.
	defaultMidSizeLeaky   = 256  // 512MB leaky at most for 2MB []byte.
	// 1GB memory leaky at most for 4MB []byte.
	// In Tesamc, the number of 4MB objects maybe large.
	defaultMaxSizeLeaky = 256
)

// BufferPool is a bytes slice pool helping to reduce GC overhead.
type BufferPool struct {
	spPool    *sync.Pool
	tinyPool  *LeakyPool
	smallPool *LeakyPool
	MidPool   *LeakyPool
	maxPool   *LeakyPool
}

// NewPool creates a bytes slice pool.
func NewPool(tiny, small, mid, max int, isAligned bool) *BufferPool {

	var makeSPFn, makeTinyFn, makeSmallFn, makeMidFn, makeMaxFn func() []byte
	if isAligned {
		makeSPFn = func() []byte {
			return directio.AlignedBlock(_MaxSyncPoolSize)
		}
		makeTinyFn = func() []byte {
			return directio.AlignedBlock(_MaxTinySize)
		}
		makeSmallFn = func() []byte {
			return directio.AlignedBlock(_MaxSmallSize)
		}
		makeMidFn = func() []byte {
			return directio.AlignedBlock(_MaxMidSize)
		}
		makeMaxFn = func() []byte {
			return directio.AlignedBlock(_MaxSize)
		}
	} else {
		makeSPFn = func() []byte {
			return make([]byte, _MaxSyncPoolSize, _MaxSyncPoolSize)
		}
		makeTinyFn = func() []byte {
			return make([]byte, _MaxTinySize, _MaxTinySize)
		}
		makeSmallFn = func() []byte {
			return make([]byte, _MaxSmallSize, _MaxSmallSize)
		}
		makeMidFn = func() []byte {
			return make([]byte, _MaxMidSize, _MaxMidSize)
		}
		makeMaxFn = func() []byte {
			return make([]byte, _MaxSize, _MaxSize)
		}
	}

	return &BufferPool{
		spPool: &sync.Pool{New: func() interface{} {
			return makeSPFn()
		}},
		tinyPool:  NewLeakyBuf(tiny, makeTinyFn),
		smallPool: NewLeakyBuf(small, makeSmallFn),
		MidPool:   NewLeakyBuf(mid, makeMidFn),
		maxPool:   NewLeakyBuf(max, makeMaxFn),
	}
}
