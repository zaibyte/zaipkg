// fastrand is copied from https://github.com/valyala/fastrand/blob/master/fastrand.go
// it's fast enough for our usage, what I've modified is just using tsc clock replacing
// system wall clock.
//
// I think uint32 is enough for the most cases.
// We don't need real big rand number.

package xrand

import (
	"sync"
)

var rngPool sync.Pool

// Uint32Fr returns pseudorandom uint32.
//
// It is safe calling this function from concurrent goroutines.
func Uint32Fr() uint32 {
	v := rngPool.Get()
	if v == nil {
		v = &RNG{}
	}
	r := v.(*RNG)
	x := r.Uint32()
	rngPool.Put(r)
	return x
}

// Uint32nFr returns pseudorandom uint32 in the range [0..maxN).
//
// It is safe calling this function from concurrent goroutines.
func Uint32nFr(maxN uint32) uint32 {
	x := Uint32Fr()
	// See http://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
	return uint32((uint64(x) * uint64(maxN)) >> 32)
}

// RNG is a pseudorandom number generator.
//
// It is unsafe to call RNG methods from concurrent goroutines.
type RNG struct {
	x uint32
}

// Uint32 returns pseudorandom uint32.
//
// It is unsafe to call this method from concurrent goroutines.
func (r *RNG) Uint32() uint32 {
	for r.x == 0 {
		r.x = getRandomUint32()
	}

	// See https://en.wikipedia.org/wiki/Xorshift
	x := r.x
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	r.x = x
	return x
}

// Uint32n returns pseudorandom uint32 in the range [0..maxN).
//
// It is unsafe to call this method from concurrent goroutines.
func (r *RNG) Uint32n(maxN uint32) uint32 {
	x := r.Uint32()
	// See http://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
	return uint32((uint64(x) * uint64(maxN)) >> 32)
}

func getRandomUint32() uint32 {
	x := Uint64()
	return uint32((x >> 32) ^ x)
}

// PickTwoFr picks up two elements which belong to [0, n).
// I think uint32 is enough for n.
// Useful for implementing Two Randomly Choices algorithm.
//
// Warn:
// a & b maybe equal. It's okay.
func PickTwoFr(n int64) (a, b int64) {

	if n < 0 {
		panic("invalid argument to Shuffle")
	}

	if n == 1 {
		return 0, 0
	}

	if n == 2 {
		return 0, 1
	}

	return int64(Uint32nFr(uint32(n))), int64(Uint32nFr(uint32(n)))
}
