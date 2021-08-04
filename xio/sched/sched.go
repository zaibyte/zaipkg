package sched

import (
	"context"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"g.tesamc.com/IT/zaipkg/config"
	"g.tesamc.com/IT/zaipkg/orpc"
	"g.tesamc.com/IT/zaipkg/vdisk"
	"g.tesamc.com/IT/zaipkg/xio"
	"g.tesamc.com/IT/zaipkg/xlog"
	"g.tesamc.com/IT/zproto/pkg/metapb"

	"github.com/panjf2000/ants/v2"
	"github.com/templexxx/tsc"
)

const (
	// DefaultThreads is the max concurrent readers/writers in single disk.
	// Beyond 128, we may get higher IOPS, but much higher latency.
	//
	// In an enterprise-class TLC/QLC NVMe driver, 16-64 would be a good choice for daily using.
	// For large I/O, 16 is a better choice.
	// I set a big number here for hitting the largest IOPS with randomly read small data blocks,
	// after peak, the extra goroutines will be recycled, won't cause any wasting.
	//
	// This value is the result of combination of Intel manual & my experience & testing results.
	DefaultThreads = 128

	// DefaultThreadsSATA is the max concurrent readers/writers in single SATA disk.
	// The concurrency ability of SATA disk really poor.
	DefaultThreadsSATA = 8
)

// Config is Scheduler's config.
type Config struct {
	Threads     int
	NVMeThread  int          `toml:"nv_me_thread"`
	SATAThread  int          `toml:"sata_thread"`
	QueueConfig *QueueConfig `toml:"queue_config"`

	noReqSleep    time.Duration
	balanceWindow int64
}

// Scheduler is disk I/O scheduler provides fair scheduling with priority classes.
type Scheduler struct {
	isRunning int64

	cfg *Config

	diskMeta *vdisk.SyncMeta

	queue *Queue

	ctx    context.Context
	cancel func()
	stopWg *sync.WaitGroup
}

func (s *Scheduler) DoAsync(reqType uint64, f xio.File, offset int64, d []byte) (ar *xio.AsyncRequest, err error) {

	return s.queue.Add(reqType, f, offset, d)
}

func (s *Scheduler) DoSync(reqType uint64, f xio.File, offset int64, d []byte) (err error) {

	var ar *xio.AsyncRequest
	if ar, err = s.DoAsync(reqType, f, offset, d); err != nil {
		return err
	}
	err = <-ar.Err
	xio.ReleaseAsyncRequest(ar)
	return err
}

// New creates a scheduler instance.
func New(ctx context.Context, cfg *Config, dm *vdisk.SyncMeta) *Scheduler {

	cfg.adjust(dm.Type)

	ctx2, cancel := context.WithCancel(ctx)

	return &Scheduler{
		cfg: cfg,

		diskMeta: dm,
		queue:    NewQueue(cfg.QueueConfig),

		ctx:    ctx2,
		cancel: cancel,
		stopWg: new(sync.WaitGroup),
	}
}

func (s *Scheduler) Start() {
	if !atomic.CompareAndSwapInt64(&s.isRunning, 0, 1) {
		return // Already started.
	}

	s.stopWg.Add(1)

	go s.FindRunnableLoop()
	xlog.Info(fmt.Sprintf("disk: %s scheduler is running", s.diskMeta.Id))
}

func (s *Scheduler) Close() {
	if !atomic.CompareAndSwapInt64(&s.isRunning, 1, 0) {
		return // Already closed.
	}

	s.cancel()
	s.stopWg.Wait()

	xlog.Info(fmt.Sprintf("disk: %s scheduler is closed", s.diskMeta.Id))
}

func (c *Config) adjust(dt metapb.DiskType) {

	if dt == metapb.DiskType_Disk_SATA {
		config.Adjust(&c.SATAThread, DefaultThreadsSATA)
		config.Adjust(&c.Threads, &c.SATAThread)
		config.Adjust(&c.balanceWindow, balanceWindowsSATA)
		config.Adjust(&c.noReqSleep, noReqSleepSATA)
	} else {
		config.Adjust(&c.NVMeThread, DefaultThreads)
		config.Adjust(&c.Threads, &c.NVMeThread)
		config.Adjust(&c.balanceWindow, defaultBalanceWindows)
		config.Adjust(&c.noReqSleep, defaultNoReqSleep)
	}
}

// That balancing is expected to happen over a specific time window,
// default is 10ms. Good enough for NVMe device.
const (
	defaultBalanceWindows = int64(10 * time.Millisecond)
	balanceWindowsSATA    = int64(100 * time.Millisecond)
)

const (
	defaultNoReqSleep = 10 * time.Microsecond
	noReqSleepSATA    = 2 * time.Millisecond
)

// FindRunnableLoop finds runnable request by scheduler rules round and round.
func (s *Scheduler) FindRunnableLoop() {
	defer s.stopWg.Done()

	// ioWorkers is a Goroutine pool, for saving goroutine creating/destroy/scheduling cost.
	// error here could be ignore because we won't pass illegal params.
	ioWorkers, _ := ants.NewPoolWithFunc(s.cfg.Threads, func(i interface{}) {

		r := i.(*xio.AsyncRequest)
		var err error
		if xio.IsReqRead(r.Type) {
			_, err = r.File.ReadAt(r.Data, r.Offset)
		} else {
			_, err = r.File.WriteAt(r.Data, r.Offset)
			if err == nil {
				// I don't want update ctime, utime etc. at the same time.
				// The file size is pre-allocated, data sync is enough.
				// (data sync will allocate space too, even we've already used pre-allocate,
				// some file system are using lazy allocation)
				err = r.File.Fdatasync()
			}
		}
		r.Err <- err
	}, ants.WithLogger(xlog.GetLogger()), ants.WithExpiryDuration(3*time.Second), ants.WithPreAlloc(true))
	defer ioWorkers.Release()

	start := tsc.UnixNano()
	for {

		if atomic.LoadInt64(&s.isRunning) != 1 {
			return
		}

		var min = math.MaxFloat64
		var minQ = -1
		for i, pq := range s.queue.pqs {
			if atomic.LoadInt64(&pq.pending) > 0 {
				if pq.totalCost < min {
					min = pq.totalCost
					minQ = i
				}
			}
		}

		var ar *xio.AsyncRequest
		if minQ != -1 {
			ar = <-s.queue.pqs[minQ].reqQueue.queue
			atomic.AddInt64(&s.queue.pqs[minQ].pending, -1)
		} else {
			time.Sleep(s.cfg.noReqSleep)
			continue
		}

		if err := s.preproc(ar.Type); err != nil {
			ar.Err <- err
			continue
		}

		now := tsc.UnixNano()

		_ = ioWorkers.Invoke(ar)

		if now-start >= s.cfg.balanceWindow {
			s.setCostsZero()
			start = now
			continue
		}

		c := calcCost(int64(len(ar.Data)), ar.PTS, now, s.queue.pqs[minQ].shares)
		s.queue.pqs[minQ].totalCost += c
	}
}

// preproc preprocess the request.
// Returns error if this request cannot be executed in present.
func (s *Scheduler) preproc(reqType uint64) error {
	state := s.diskMeta.GetState()
	isRead := xio.IsReqRead(reqType)
	if state == metapb.DiskState_Disk_Broken {
		return orpc.ErrDiskBroken
	}
	if state == metapb.DiskState_Disk_Tombstone {
		return orpc.ErrDiskTombstone
	}
	if state == metapb.DiskState_Disk_Offline && !isRead {
		return orpc.ErrDiskOffline
	}
	if state == metapb.DiskState_Disk_Full && !isRead {
		return orpc.ErrDiskFull
	}
	return nil
}

// calcCost calculates the cost of a request.
// n is request length,
// pts is the put in queue timestamp,
// now is the executing timestamp,
// shares is the queue shares.
func calcCost(n, pts, now, shares int64) float64 {
	c0 := calcWeight(n) / float64(shares)
	return c0 * calcWaitCoeff(pts, now)
}

const (
	// waitExpCoeff controls the decay speed.
	waitExpCoeff   = -0.003
	waitDeltaCoeff = float64(time.Microsecond)
)

// calcWaitCoeff calculates coefficient according request waiting time in queue,
// it's an exponential decay.
// It helps to let request which wait longer be executed faster.
//
// coeff = e^(waitExpCoeff * waiting_time)
func calcWaitCoeff(pts, now int64) float64 {
	delta := float64(now-pts) / waitDeltaCoeff // Using microsecond as unit.
	return math.Pow(math.E, waitExpCoeff*delta)
}

const pageSize = 4 * 1024

// calcWeight calculates I/O request weight in scheduler.
// It's sublinear function: w = 200 + 0.25*n^0.6.
// 200 is the init weight,
// n is the request length/4KB,
// 0.6 is an experience value,
// 0.25 makes the result in a reasonable range
// (each request won't be out of 4MB, so in 0.6, the shares still matters.)
func calcWeight(n int64) float64 {
	n = n / pageSize
	return 200 + (math.Pow(float64(n), 0.6) * 0.25)
}

// set all totalCost zero after meet the balance window.
func (s *Scheduler) setCostsZero() {
	for _, q := range s.queue.pqs {
		q.totalCost = 0
	}
}
