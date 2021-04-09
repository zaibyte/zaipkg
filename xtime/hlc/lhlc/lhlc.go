// Package lhlc(local hybrid logical clock) implements LHLC interface, that combines the best of logical clocks and physical clocks.
// It's a clock which never goes backwards in one instance.
//
// Warn:
// After instance's LHLC rootPath broken, it has chance going backwards because we have no reference time.
// But it's rare to happen because we cannot making a new device that fast(in dozens ms, which is NTP jitter).
package lhlc

import (
	"encoding/binary"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/templexxx/tsc"

	"g.tesamc.com/IT/zaipkg/xtime/hlc"

	"g.tesamc.com/IT/zaipkg/xdigest"

	"g.tesamc.com/IT/zaipkg/vfs"
)

type LHLC struct {
	rwm *sync.RWMutex

	// There will be some states persist into local file system.
	rootPath string
	fs       vfs.FS

	lastTS uint64

	physical uint64
	logical  uint64
}

const (
	defaultRootPath = "/usr/local/zai"
	lhlcFileName    = "lhlc"
)

// CreateLHLC creates an LHLC for application and set global hlc.
// Each instance should have one.
func CreateLHLC(rootPath string, fs vfs.FS) *LHLC {

	if rootPath == "" {
		rootPath = defaultRootPath
	}

	h := &LHLC{
		rootPath: rootPath,
		fs:       fs,
		rwm:      new(sync.RWMutex),
	}
	return h
}

func (c *LHLC) openOrCreateLocalBase() error {
	return nil
}

func makeBaseFile(path string) string {
	return filepath.Join(path, lhlcFileName)
}

const baseBlockSize = 4 * 1024

// makeBaseBlock makes timestamp data block for next persisting.
func makeBaseBlock(ts uint64, buf []byte) {
	binary.LittleEndian.PutUint64(buf[:8], ts)
	binary.LittleEndian.PutUint32(buf[baseBlockSize-4:], xdigest.Sum32(buf[:baseBlockSize-4]))
}

// Sync syncs base block to local file system.
func (c *LHLC) Sync() {

}

// Load loads ts from local file system.
func (c *LHLC) Load() {

}

// Next returns a timestamp.
func (c *LHLC) Next() (ts uint64) {
	for {
		last, p, l, ok := c.next()
		if !ok {
			time.Sleep(200 * time.Microsecond)
			continue
		}

		ts = hlc.MakeTS(p, l)
		if atomic.CompareAndSwapUint64(&c.lastTS, last, ts) {
			return
		}
	}
}

func (c *LHLC) next() (last, phy, logic uint64, ok bool) {
	last = atomic.LoadUint64(&c.lastTS)
	lp, ll := hlc.ParseTS(last)

	phy = lp

	logic = (ll + 1) & hlc.LogicalMask
	if logic == 0 { // Logical overflow, need new physical.
		now := uint64(tsc.UnixNano() / int64(time.Millisecond))
		if lp >= now {
			return 0, 0, 0, false // Time go backwards.
		}
		phy = now
	}
	return last, phy, logic, true
}
