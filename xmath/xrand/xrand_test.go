package xrand

import (
	"math"
	"math/rand"
	"runtime"
	"testing"
)

func TestPickTwo(t *testing.T) {

	comb := make(map[int64]bool)

	// C(64,2) with order = 2016.
	for i := 0; i < 2016; i++ {
		// It's DefaultMaxWritableGroupsCnt in keeper client.
		a, b := PickTwo(64)

		if comb[a*100+b] || comb[b*100+a] {
			continue
		}
		comb[a*100+b] = true
	}

	if float64(len(comb)) < 2016*0.6 { // After rand.Shuffle in math/rand, it's about 0.6 too.
		t.Fatal("PickTwo is not random enough", float64(len(comb))/2016)
	}
}

func TestPickTwoDistribution(t *testing.T) {

	cnt := make(map[int64]int)

	for i := 0; i < 64*1024; i++ {
		a, b := PickTwo(64)
		cnt[a]++
		cnt[b]++
	}

	avg := float64(64 * 1024 / 64 * 2)

	if len(cnt) != 64 {
		t.Fatal("distribution too bad, even cannot fill all elements")
	}

	for k, v := range cnt {
		if math.Abs(float64(v)-avg) > 0.1*avg {
			t.Fatalf("PickTwo distribution bad for %d", k)
		}
	}
}

func pickTwoMathRand(n int) (a, b int64) {
	s := make([]int64, n)
	for i := range s {
		s[i] = int64(i)
	}
	rand.Shuffle(n, func(i, j int) {
		s[i], s[j] = s[j], s[i]
	})
	return s[0], s[1]
}

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

// Int63n is much faster. 3x faster.
func BenchmarkInt63nParallel(b *testing.B) {

	b.SetParallelism(runtime.NumCPU())
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			Int63n(1000)
		}
	})
}

func BenchmarkInt63nMathRandParallel(b *testing.B) {

	b.SetParallelism(runtime.NumCPU())
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			rand.Int63n(1000)
		}
	})
}

func BenchmarkPickTwo(b *testing.B) {

	for i := 0; i < b.N; i++ {
		_, _ = PickTwo(64)
	}
}

func BenchmarkPickTwoMathRand(b *testing.B) {

	for i := 0; i < b.N; i++ {
		_, _ = pickTwoMathRand(64)
	}
}

func BenchmarkPickTwoParallel(b *testing.B) {

	b.SetParallelism(runtime.NumCPU())
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			_, _ = PickTwo(64)
		}
	})
}

func BenchmarkPickTwoMathRandParallel(b *testing.B) {

	b.SetParallelism(runtime.NumCPU())
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			_, _ = pickTwoMathRand(64)
		}
	})
}
