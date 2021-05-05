// Package xrand provides rand number generating.
// Reference:
// https://lemire.me/blog/2019/03/19/the-fastest-conventional-random-number-generator-that-can-pass-big-crush/
// https://github.com/lemire/testingRNG/blob/master/source/lehmer64.h
package xrand

import (
	"sync/atomic"
	"unsafe"

	"g.tesamc.com/IT/zaipkg/xmath"
)

var State unsafe.Pointer

func init() {
	s := unsafe.Pointer(new(xmath.Uint128))
	atomic.StorePointer(&State, s)
	Seed(1)
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
	max := int64((1 << 63) - 1 - (1<<63)%uint64(n))
	v := Int63()
	for v > max {
		v = Int63()
	}
	return v % n
}

// Seed uses the provided seed value to initialize the default Source to a
// deterministic state. If Seed is not called, the generator behaves as
// if seeded by Seed(1).
//
// State = (((__uint128_t)splitmix64_stateless(seed)) << 64) +
//                     splitmix64_stateless(seed + 1);
//
// Warn:
// It's not safe for concurrent use.
func Seed(s int64) {
	a := xmath.Uint128{L: splitMix64Stateless(uint64(s))}.ShiftLeft(64)
	newS := a.Add(xmath.Uint128{L: splitMix64Stateless(uint64(s) + 1)})
	atomic.StorePointer(&State, unsafe.Pointer(&newS))
}

// Int63 returns a non-negative pseudo-random 63-bit integer as an int64
// from the default Source.
// g_lehmer64_state *= UINT64_C(0xda942042e4dd58b5);
// return g_lehmer64_state >> 64;
func Int63() int64 {
	sp := (*xmath.Uint128)(atomic.LoadPointer(&State))
	s := *sp
	s = s.Mult(xmath.Uint128{L: 0xda942042e4dd58b5})
	atomic.StorePointer(&State, unsafe.Pointer(&s))
	return int64(s.ShiftRight(64).L)
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
