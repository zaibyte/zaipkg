package xrand

import (
	"math/rand"
	"runtime"
	"testing"
)

// Int63n is a bit slower than math/rand. about 20ns.
func BenchmarkInt63n(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Int63n(1000)
	}
}

func BenchmarkInt63nMathRand(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rand.Int63n(1000)
	}
}

func BenchmarkInt63(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Int63()
	}
}

func BenchmarkInt63MathRand(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rand.Int63()
	}
}

// Int63 is much faster. 4x faster.
func BenchmarkInt63Parallel(b *testing.B) {

	b.SetParallelism(runtime.NumCPU())
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			Int63()
		}
	})
}

func BenchmarkInt63MathRandParallel(b *testing.B) {

	b.SetParallelism(runtime.NumCPU())
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			rand.Int63()
		}
	})
}
