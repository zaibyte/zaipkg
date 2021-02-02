package xdigest

import "sync"

var pool sync.Pool

// Acquire acquires a Digest from pool.
func Acquire() *Digest {
	v := pool.Get()
	if v == nil {
		return New()
	}
	return v.(*Digest)
}

// Release puts Digest back into pool.
func Release(d *Digest) {
	d.Reset()

	pool.Put(d)
}
