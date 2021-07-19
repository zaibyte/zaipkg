package svr

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"g.tesamc.com/IT/zaipkg/uid"
	"g.tesamc.com/IT/zaipkg/vdisk"
	"g.tesamc.com/IT/zaipkg/vfs"
	"g.tesamc.com/IT/zaipkg/xio"
	"g.tesamc.com/IT/zaipkg/xio/sched"
	"g.tesamc.com/IT/zaipkg/xlog"
	"g.tesamc.com/IT/zaipkg/xtime"
	"g.tesamc.com/IT/zproto/pkg/metapb"
)

// .
// ├── <data_root>
// │    ├── disk_<disk_id0>
//
const (
	diskNamePrefix = "disk_"
)

// ZBufDisks contains all avail disks on single zBuf server.
type ZBufDisks struct {
	InstanceID string
	FS         vfs.FS
	VDisk      vdisk.Disk
	DataRoot   string
	// Using sync.Map for online adding/removing disk.
	Disks *sync.Map // k: diskID, v: *ZBufDisk

	schedCfg *sched.Config

	ctx context.Context
	wg  *sync.WaitGroup
}

type ZBufDisk struct {
	DiskID       string
	Info         *vdisk.SyncMeta
	Sched        xio.Scheduler
	SchedStarted int64
}

// NewZBufDisks creates a new ZBufDisks instance.
func NewZBufDisks(ctx context.Context, wg *sync.WaitGroup, fs vfs.FS, vdisk vdisk.Disk, instanceID, dataRoot string, schedCfg *sched.Config) *ZBufDisks {
	d := &ZBufDisks{
		InstanceID: instanceID,
		FS:         fs,
		VDisk:      vdisk,
		DataRoot:   dataRoot,
		Disks:      new(sync.Map),
		schedCfg:   schedCfg,
		ctx:        ctx,
		wg:         wg,
	}
	return d
}

// Init inits ZBufDisks at starting.
func (d *ZBufDisks) Init(fs vfs.FS) {
	if d.Disks == nil {
		d.Disks = new(sync.Map)
	}

	diskIDs, _ := ListDiskIDs(fs, d.DataRoot)
	d.AddDisks(diskIDs)
}

var ErrNoDisk = errors.New("no disk for ZBuf in this instance")

// ListDiskIDs lists all disk ids according to the disk path.
func ListDiskIDs(fs vfs.FS, root string) (diskIDs []string, err error) {
	diskFns, err := fs.List(root)
	if err != nil {
		return
	}

	diskIDs = make([]string, 0, len(diskFns))
	cnt := 0
	for _, fn := range diskFns {
		if strings.HasPrefix(fn, diskNamePrefix) {
			cnt++
			id := strings.TrimPrefix(fn, diskNamePrefix)
			if uid.IsValidDiskID(id) {
				diskIDs = append(diskIDs, id)
			}
		}
	}
	if cnt == 0 {
		return nil, ErrNoDisk
	}
	return diskIDs[:cnt], nil
}

// AddDisks adds zbuf disk one by one.
func (d *ZBufDisks) AddDisks(diskIDs []string) {

	for _, diskID := range diskIDs {
		d.AddDisk(diskID)
	}
}

// AddDisk adds single disk.
func (d *ZBufDisks) AddDisk(diskID string) {

	_, ok := d.Disks.Load(diskID)
	if ok {
		return // Already has.
	}

	v := new(ZBufDisk)

	meta := (*vdisk.SyncMeta)(new(metapb.Disk))
	meta.State = metapb.DiskState_Disk_ReadWrite
	meta.Id = diskID
	path := MakeDiskDir(diskID, d.DataRoot)
	meta.Type = d.VDisk.GetType(path)
	meta.SN = d.VDisk.GetSN(path)
	_ = d.VDisk.InitUsage(path, meta)
	meta.InstanceId = d.InstanceID

	if meta.Used >= meta.Size_ {
		meta.State = metapb.DiskState_Disk_Full
	}

	v.Info = meta
	v.DiskID = diskID
	if d.schedCfg != nil {
		v.Sched = sched.New(d.ctx, d.schedCfg, v.Info)
	} else {
		v.Sched = new(xio.NopScheduler)
	}
	d.Disks.Store(diskID, v)
}

// StartSched starts disk I/O scheduler.
// If diskIDs is not empty, using diskIDs, if diskID is not found, ignore.
// If it's empty, starting all schedulers which haven't started.
func (d *ZBufDisks) StartSched(diskIDs ...string) {

	if len(diskIDs) != 0 {
		for _, diskID := range diskIDs {
			zd := d.GetDisk(diskID)
			if zd == nil {
				continue // Just ignore not found disk.
			}
			if atomic.LoadInt64(&zd.SchedStarted) == 1 {
				continue
			}
			zd.Sched.Start()
			atomic.CompareAndSwapInt64(&zd.SchedStarted, 0, 1)
		}
	} else {
		d.Disks.Range(func(key, value interface{}) bool {
			disk := value.(*ZBufDisk)
			if atomic.LoadInt64(&disk.SchedStarted) == 1 {
				return true
			}
			disk.Sched.Start()
			atomic.CompareAndSwapInt64(&disk.SchedStarted, 0, 1)
			return true
		})
	}
}

// CloseSched closes disk I/O scheduler.
// If diskIDs is not empty, using diskIDs, if diskID is not found, ignore.
// If it's empty, closing all schedulers which have started.
func (d *ZBufDisks) CloseSched(diskIDs ...string) {
	if len(diskIDs) != 0 {
		for _, diskID := range diskIDs {
			zd := d.GetDisk(diskID)
			if zd == nil {
				continue // Just ignore not found disk.
			}
			if atomic.LoadInt64(&zd.SchedStarted) == 0 {
				continue
			}
			zd.Sched.Close()
			atomic.CompareAndSwapInt64(&zd.SchedStarted, 1, 0)
		}
	} else {
		d.Disks.Range(func(key, value interface{}) bool {
			disk := value.(*ZBufDisk)
			if atomic.LoadInt64(&disk.SchedStarted) == 0 {
				return true
			}
			disk.Sched.Close()
			atomic.CompareAndSwapInt64(&disk.SchedStarted, 1, 0)
			return true
		})
	}
}

// MakeDiskDir makes disk path according diskID
func MakeDiskDir(diskID string, root string) string {
	return filepath.Join(root, diskNamePrefix+diskID)
}

// GetInfo gets disk info by diskID.
func (d *ZBufDisks) GetInfo(diskID string) *vdisk.SyncMeta {
	di, ok := d.Disks.Load(diskID)
	if !ok {
		return nil
	}
	return di.(*ZBufDisk).Info
}

// GetSched gets scheduler by diskID and started or not.
func (d *ZBufDisks) GetSched(diskID string) (xio.Scheduler, bool) {
	di, ok := d.Disks.Load(diskID)
	if !ok {
		return nil, false
	}
	zd := di.(*ZBufDisk)
	return zd.Sched, atomic.LoadInt64(&zd.SchedStarted) == 1
}

// GetDisk gets ZBufDisk by diskID.
func (d *ZBufDisks) GetDisk(diskID string) *ZBufDisk {
	di, ok := d.Disks.Load(diskID)
	if !ok {
		return nil
	}
	return di.(*ZBufDisk)
}

func (d *ZBufDisks) GetDiskMeta(diskID string) *vdisk.SyncMeta {
	zd := d.GetDisk(diskID)
	if zd == nil {
		return nil
	}
	return zd.Info
}

// CloneAllDiskMeta clones all *metapb.Disk.
// Usually for heartbeat.
func (d *ZBufDisks) CloneAllDiskMeta() map[string]*metapb.Disk {

	ret := make(map[string]*metapb.Disk)

	d.Disks.Range(func(key, value interface{}) bool {
		sm := value.(*ZBufDisk).Info
		ret[key.(string)] = sm.Clone()
		return true
	})

	return ret
}

// UpdateDiskStates updates all disk states.
// For heartbeat only.
func (d *ZBufDisks) UpdateDiskStates(dss map[string]metapb.DiskState) {

	for id, s := range dss {
		v, ok := d.Disks.Load(id)
		if ok {
			mo := v.(*ZBufDisk).Info
			mo.SetState(s)
		} else {
			zd := new(ZBufDisk)

			meta := (*vdisk.SyncMeta)(new(metapb.Disk))
			meta.State = metapb.DiskState_Disk_Broken
			meta.Id = id
			path := MakeDiskDir(id, d.DataRoot)
			meta.Type = d.VDisk.GetType(path)
			meta.SN = d.VDisk.GetSN(path)
			_ = d.VDisk.InitUsage(path, meta)
			meta.InstanceId = d.InstanceID

			if meta.Used >= meta.Size_ {
				meta.State = metapb.DiskState_Disk_Full
			}

			zd.Info = meta
			zd.DiskID = id

			zd.Sched = new(xio.NopScheduler)
			d.Disks.Store(id, zd)
		}
	}
}

// ListDiskIDs lists all disk IDs in this ZBuf server.
func (d *ZBufDisks) ListDiskIDs() []string {

	ids := make([]string, 0, 32)
	cnt := 0
	d.Disks.Range(func(key, value interface{}) bool {
		id := key.(string)
		ids = append(ids, id)
		cnt++
		return true
	})
	return ids[:cnt]
}

// DetectLoop detects local disks round and round.
// Helping us to add new disk automatically.
func (d *ZBufDisks) DetectLoop() {
	defer d.wg.Done()

	ctx, cancel := context.WithCancel(d.ctx)
	defer cancel()

	t := xtime.AcquireTimer(10 * time.Minute)
	defer xtime.ReleaseTimer(t)

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			diskIDs, err := ListDiskIDs(d.FS, d.DataRoot)
			if err != nil {
				xlog.Warn(fmt.Sprintf("failed to list disk ids in detect loop: %s", err.Error()))
				continue
			}
			d.AddDisks(diskIDs)
			d.StartSched()
		}
	}
}
