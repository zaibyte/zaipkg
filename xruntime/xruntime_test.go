package xruntime

import (
	"runtime"
	"testing"

	"g.tesamc.com/IT/zaipkg/xtest"
	"github.com/panjf2000/ants/v2"
)

// Compare goroutine pool & using chan to limit goroutine numbers.
// TODO should I use goroutine pool?
func BenchmarkGoroutinePool(b *testing.B) {
	p, _ := ants.NewPool(128, ants.WithMaxBlockingTasks(1<<30))
	defer p.Release()

	b.SetParallelism(128)

	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {

			_ = p.Submit(func() {
				demoFunc()
			})
		}
	})

	b.StopTimer()
}

// When I raise parallelism, the chan version gets better performance.
func BenchmarkChanLimitGoroutine(b *testing.B) {

	c := make(chan struct{}, 128)

	b.SetParallelism(128)

	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {

			c <- struct{}{}
			go func() {
				demoFunc()
				<-c
			}()
		}
	})

	b.StopTimer()
}

// Bench channel.
func BenchmarkStructChan(b *testing.B) {
	ch := make(chan struct{}, 1024)
	go func() {
		for {
			<-ch
		}
	}()

	for i := 0; i < b.N; i++ {
		ch <- struct{}{}
	}
}

func BenchmarkChanContended(b *testing.B) {
	const C = 100
	myc := make(chan int, C*runtime.GOMAXPROCS(0))
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < C; i++ {
				myc <- 0
			}
			for i := 0; i < C; i++ {
				<-myc
			}
		}
	})
}

func BenchmarkChanMultiProds(b *testing.B) {

	const C = 2048
	myc := make(chan int, C*runtime.GOMAXPROCS(0))
	go func() {
		for {
			<-myc
		}
	}()
	b.SetParallelism(1024)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			myc <- 1
		}
	})

}

func demoFunc() {
	xtest.DoNothing(10) // About 400ns.
}
