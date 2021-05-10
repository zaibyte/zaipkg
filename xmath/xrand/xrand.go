// Package xrand provides rand number generating.
// Reference:
// https://lemire.me/blog/2019/03/19/the-fastest-conventional-random-number-generator-that-can-pass-big-crush/
// https://github.com/lemire/testingRNG/blob/master/source/lehmer64.h
package xrand

import (
	"sync"
	"time"

	"g.tesamc.com/IT/zaipkg/xatomic"
	"g.tesamc.com/IT/zaipkg/xbytes"
	"g.tesamc.com/IT/zaipkg/xmath"
)

var (
	State = xbytes.MakeAlignedBlock(16, 16)
)

func init() {

	Seed(time.Now().UnixNano())
}

// Int63n returns, as an int64, a non-negative pseudo-random number in [0,n)
// from the default Source.
// It panics if n <= 0.
func Int63n(n int64) int64 {
	if n <= 0 {
		panic("invalid argument to Int63n")
	}
	if n&(n-1) == 0 { // n is power of two, can mask
		return Int63() & (n - 1)
	}
	m := int64((1 << 63) - 1 - (1<<63)%uint64(n))
	v := Int63()
	for v > m {
		v = Int63()
	}
	return v % n
}

// Seed uses the provided seed value to initialize the default Source to a
// deterministic state. If Seed is not called, the generator behaves as
// if seeded by Seed(time.Now().UnixNano()).
//
// State = (((__uint128_t)splitmix64_stateless(seed)) << 64) +
//                     splitmix64_stateless(seed + 1);
// It's safe for concurrent use.
func Seed(s int64) {
	a := xmath.Uint128{L: splitMix64Stateless(uint64(s))}.ShiftLeft(64)
	newS := a.Add(xmath.Uint128{L: splitMix64Stateless(uint64(s) + 1)})
	xatomic.AvxStore16B(&State[0], &newS.ToArr()[0])
}

var _blockPool = sync.Pool{New: func() interface{} {
	b := xbytes.MakeAlignedBlock(16, 16)
	return &b
}}

const (
	max  = 1 << 63
	mask = max - 1
)

// Int63 returns a non-negative pseudo-random 63-bit integer as an int64
// from the default Source.
func Int63() int64 {
	return int64(Uint64() & mask)
}

// Uint64 returns a non-negative pseudo-random 64-bit integer as an uint64.
//
// g_lehmer64_state *= UINT64_C(0xda942042e4dd58b5);
// return g_lehmer64_state >> 64;
func Uint64() uint64 {
	p := _blockPool.Get().(*[]byte)
	old := *p
	xatomic.AvxLoad16B(&State[0], &old[0])
	s := xmath.FromArrToUint128(old)
	s = s.Mult(xmath.Uint128{L: 0xda942042e4dd58b5})
	s.ToArrDst(old)
	// It's okay to use store directly here, although it may cause dirty.
	// It's rare, enough strong for non-strict cases.
	xatomic.AvxStore16B(&State[0], &old[0])
	_blockPool.Put(p)
	return s.H
}

// PickTwo picks up two elements which belong to [0, n).
// Useful for implementing Two Randomly Choices algorithm.
// Warn:
// a & b maybe equal.
func PickTwo(n int64) (a, b int64) {

	if n < 0 {
		panic("invalid argument to Shuffle")
	}

	if n == 1 {
		return 0, 0
	}

	if n == 2 {
		return 0, 1
	}

	return Int63n(n), Int63n(n)
}

// returns random number,
// compared with D. Lemire against
// http://grepcode.com/file/repository.grepcode.com/java/root/jdk/openjdk/8-b132/java/util/SplittableRandom.java#SplittableRandom.0gamma
// https://github.com/lemire/testingRNG/blob/master/source/splitmix64.h
func splitMix64Stateless(n uint64) uint64 {
	z := n + 0x9E3779B97F4A7C15
	z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
	z = (z ^ (z >> 27)) * 0x94D049BB133111EB
	return z ^ (z >> 31)
}
