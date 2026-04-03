package xruntime

import (
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/klauspost/cpuid/v2"
	"github.com/panjf2000/ants/v2"
	"github.com/templexxx/tsc"
	"github.com/zaibyte/zaipkg/xtest"
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
	ProcYield(10) // About 400ns.
}

func TestProcYield(t *testing.T) {

	if !xtest.IsPropEnabled() {
		t.Skip("properties testing isn't enabled")
	}

	cs := []int{10, 20, 30, 60, 120, 240}

	for _, c := range cs {
		// TODO using TSC register to report cost (as property testing)
		tsc.ForbidOutOfOrder()
		start := tsc.UnixNano()
		for i := 0; i < 1000; i++ {
			ProcYield(uint32(c))
		}
		cost := (tsc.UnixNano() - start) / 1000
		t.Logf("%d cycles cost: %dns on %s", c, cost, getCPUBrand())
	}
}

func getCPUBrand() string {

	brand := cpuid.CPU.BrandName
	if brand != "" {
		return cpuid.CPU.BrandName
	}

	if runtime.GOOS == `darwin` {
		brand = getCPUBrandOnDarwin()
	}
	if brand == "" {
		return "unknown"
	}
	return brand
}

func getCPUBrandOnDarwin() string {
	grep := exec.Command("grep", "machdep.cpu.brand_string")
	sysctl := exec.Command("sysctl", "-a")

	pipe, err := sysctl.StdoutPipe()
	if err != nil {
		return ""
	}
	defer pipe.Close()

	grep.Stdin = pipe

	err = sysctl.Start()
	if err != nil {
		return ""
	}
	res, err := grep.Output()
	if err != nil {
		return ""
	}
	return strings.Split(string(res), ": ")[1]
}
