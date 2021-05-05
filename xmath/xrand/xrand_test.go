package xrand

import (
	"math/rand"
	"runtime"
	"testing"
)

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

// Int63n is much faster.
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
