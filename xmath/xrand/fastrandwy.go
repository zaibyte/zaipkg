package xrand

import (
	"math/bits"
	"math/rand"
	"sync/atomic"

	"github.com/templexxx/tsc"
	"github.com/zeebo/xxh3"
)

const falseSharingRange = 128

var (
	_        [falseSharingRange]byte
	fastrand uint64 = 0
	_        [falseSharingRange]byte
)

func init() {
	// TODO need to test how good the fastrand is.
	rand.Seed(int64(tsc.RDTSC()))
	randData := make([]byte, rand.Int31n(33344))
	rand.Read(randData)
	fastrand = xxh3.HashSeed(randData, tsc.RDTSC())
}

func Uint32() uint32 {
	fr := atomic.AddUint64(&fastrand, 0xa0761d6478bd642f)
	hi, lo := bits.Mul64(fr, fr^0xe7037ed1a0b428db)
	return uint32(hi ^ lo)
}

func Uint32n(n uint32) uint32 {
	// This is similar to fastrand() % n, but faster.
	// See https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
	return uint32(uint64(Uint32()) * uint64(n) >> 32)
}

// PickTwo picks up two elements which belong to [0, n).
// I think uint32 is enough for n.
// Useful for implementing Two Randomly Choices algorithm.
//
// Warn:
// a & b maybe equal. It's okay.
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

	return int64(Uint32n(uint32(n))), int64(Uint32n(uint32(n)))
}
