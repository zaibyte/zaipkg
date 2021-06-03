package svr

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"g.tesamc.com/IT/zaipkg/uid"

	"g.tesamc.com/IT/zaipkg/vdisk"
	"g.tesamc.com/IT/zaipkg/vfs"
	"g.tesamc.com/IT/zaipkg/xio"
	"g.tesamc.com/IT/zaipkg/xio/sched"
	"g.tesamc.com/IT/zproto/pkg/metapb"
)

// .
// ├── <data_root>
// │    ├── disk_<disk_id0>
//
const (
	diskNamePrefix = "disk_"
)

// ZBufDisks contains all avail disks on single ZBuf server.
type ZBufDisks struct {
	VDisk    vdisk.Disk
	DataRoot string
	// Using sync.Map for online adding/removing disk.
	Disks *sync.Map // k: diskID, v: ZBufDisk

	schedCfg *sched.Config

	ctx context.Context
}

// ZBufDisk
type ZBufDisk struct {
	DiskID       string
	Info         *vdisk.SyncMeta
	Sched        xio.Scheduler
	SchedStarted int64
}

// NewZBufDisks creates a new ZBufDisks instance.
func NewZBufDisks(ctx context.Context, vdisk vdisk.Disk, dataRoot string, schedCfg *sched.Config) *ZBufDisks {
	d := &ZBufDisks{
		VDisk:    vdisk,
		DataRoot: dataRoot,
		Disks:    new(sync.Map),
		schedCfg: schedCfg,
		ctx:      ctx,
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

	v := new(ZBufDisk)

	meta := (*vdisk.SyncMeta)(new(metapb.Disk))
	meta.Id = diskID
	path := MakeDiskDir(diskID, d.DataRoot)
	meta.Type = d.VDisk.GetType(path)
	_ = d.VDisk.InitUsage(path, meta)

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
func (d *ZBufDisks) StartSched(diskIDs ...uint32) {

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
func (d *ZBufDisks) CloseSched(diskIDs ...uint32) {
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
func (d *ZBufDisks) GetInfo(diskID uint32) *vdisk.SyncMeta {
	di, ok := d.Disks.Load(diskID)
	if !ok {
		return nil
	}
	return di.(*ZBufDisk).Info
}

// GetSched gets scheduler by diskID and started or not.
func (d *ZBufDisks) GetSched(diskID uint32) (xio.Scheduler, bool) {
	di, ok := d.Disks.Load(diskID)
	if !ok {
		return nil, false
	}
	zd := di.(*ZBufDisk)
	return zd.Sched, atomic.LoadInt64(&zd.SchedStarted) == 1
}

// GetDisk gets ZBufDisk by diskID.
func (d *ZBufDisks) GetDisk(diskID uint32) *ZBufDisk {
	di, ok := d.Disks.Load(diskID)
	if !ok {
		return nil
	}
	return di.(*ZBufDisk)
}

// ListDiskIDs lists all disk IDs in this ZBuf server.
func (d *ZBufDisks) ListDiskIDs() []uint32 {

	ids := make([]uint32, 0, 32)
	cnt := 0
	d.Disks.Range(func(key, value interface{}) bool {
		id := key.(uint32)
		ids = append(ids, id)
		cnt++
		return true
	})
	return ids[:cnt]
}
