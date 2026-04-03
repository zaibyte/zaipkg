// Package xbytes provides bytes slice pool.
//
// Warning:
// 1. Do not use it when you only need <= 32KB byte slice, and this slice will not escape to the heap.
// In this situation, using sync.Pool will let the slice escapes to the heap, bringing the extra GC overhead.
// Discussion in: https://github.com/zaibyte/zaipkg/issues/11
// 2. Do not use it when you want bytes more than 4MB. Otherwise, it'll panic.
package xbytes

import (
	"sync"

	"github.com/zaibyte/zaipkg/config/settings"
	"github.com/zaibyte/zaipkg/directio"
)

const (
	NeedAligned = iota
	NeedUnAligned
	NeedBoth
)

const (
	MaxSizeInPool = _MaxSize
)

func init() {
	_pool = NewPool(defaultTinySizeLeaky, defaultSmallSizeLeaky, defaultMidSizeLeaky, defaultMaxSizeLeaky, false)
	_alignPool = NewPool(defaultTinySizeLeaky, defaultSmallSizeLeaky, defaultMidSizeLeaky, defaultMaxSizeLeaky, true)
}

// Need indicates what kinds of bytes do we need.
// Set it before using.
var Need = NeedBoth

// ResetLeakyCap resets leaky pools capacities.
// Warn:
// Not thread safe.
func ResetLeakyCap(tiny, small, mid, max int) {

	switch Need {
	case NeedAligned:
		_alignPool = NewPool(tiny, small, mid, max, true)
	case NeedUnAligned:
		_pool = NewPool(tiny, small, mid, max, false)
	default:
		_alignPool = NewPool(tiny, small, mid, max, true)
		_pool = NewPool(tiny, small, mid, max, false)
	}
}

// EnableDefault enables default memory pool.
// Use it in testing env. (saving memory)
func EnableDefault() {
	ResetLeakyCap(defaultTinySizeLeaky, defaultSmallSizeLeaky, defaultMidSizeLeaky, defaultMaxSizeLeaky)
}

// EnableMax enables max memory pool.
// Use it in production env when start an application.
func EnableMax() {
	ResetLeakyCap(maxTinySizeLeaky, maxSmallSizeLeaky, maxMidSizeLeaky, maxMaxSizeLeaky)
}

var (
	_alignPool *BufferPool
	_pool      *BufferPool

	GetBytes = func(n int) []byte {
		if n <= _MaxSyncPoolSize {
			return _pool.spPool.Get().([]byte)[:n]
		}
		if n <= _MaxTinySize {
			return _pool.tinyPool.Get()[:n]
		}
		if n <= _MaxSmallSize {
			return _pool.smallPool.Get()[:n]
		}
		if n <= _MaxMidSize {
			return _pool.MidPool.Get()[:n]
		}
		return _pool.maxPool.Get()[:n]
	}
	PutBytes = func(b []byte) {
		n := cap(b)
		b = b[:0]
		if n >= _MaxMidSize {
			_pool.MidPool.Put(b)
			return
		}
		if n >= _MaxSmallSize {
			_pool.smallPool.Put(b)
			return
		}
		if n >= _MaxTinySize {
			_pool.tinyPool.Put(b)
			return
		}

		if n >= _MaxSyncPoolSize {
			_pool.spPool.Put(b)
			return
		}

		_pool.maxPool.Put(b)
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
		n := cap(b)
		b = b[:0]
		if n >= _MaxMidSize {
			_alignPool.MidPool.Put(b)
			return
		}
		if n >= _MaxSmallSize {
			_alignPool.smallPool.Put(b)
			return
		}
		if n >= _MaxTinySize {
			_alignPool.tinyPool.Put(b)
			return
		}
		if n >= _MaxSyncPoolSize {
			_alignPool.spPool.Put(b)
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

	defaultTinySizeLeaky  = 16 // 2MB leaky at most for 128KB []byte.
	defaultSmallSizeLeaky = 4  // 2MB leaky at most for 512KB []byte.
	defaultMidSizeLeaky   = 2  // 4MB leaky at most for 2MB []byte.
	defaultMaxSizeLeaky   = 2  // 8MB memory leaky at most for 4MB []byte.

	// Don't set them too big, Go has no generational GC.
	maxTinySizeLeaky  = 1024 // 128MB leaky at most for 128KB []byte.
	maxSmallSizeLeaky = 256  // 128MB leaky at most for 512KB []byte.
	maxMidSizeLeaky   = 64   // 128MB leaky at most for 2MB []byte.
	maxMaxSizeLeaky   = 128  // 512MB memory leaky at most for 4MB []byte.
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

	if tiny <= 0 {
		tiny = defaultTinySizeLeaky
	}
	if small <= 0 {
		small = defaultSmallSizeLeaky
	}
	if mid <= 0 {
		mid = defaultMidSizeLeaky
	}
	if max <= 0 {
		max = defaultMaxSizeLeaky
	}

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

type PoolBytesCloser struct {
	P []byte
}

func (r PoolBytesCloser) Close() error {
	PutBytes(r.P)
	return nil
}

type PoolAlignedBytesCloser struct {
	P []byte
}

func (r PoolAlignedBytesCloser) Close() error {
	PutAlignedBytes(r.P)
	return nil
}
