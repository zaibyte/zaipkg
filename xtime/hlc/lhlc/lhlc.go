// Package lhlc(local hybrid logical clock) implements HLC interface, that combines the best of logical clocks and physical clocks.
// It's a clock which never goes backwards in one instance.
//
// Warn:
// After instance's HLC rootPath broken, it has chance going backwards because we have no reference time.
// But it's rare to happen because we cannot making a new device that fast(in dozens ms, which is NTP jitter).
package lhlc

type HLC struct {
	// There will be some states persist into local file system.
	rootPath string

	physic int64
	logic  int64
}

const (
	defaultRootPath = "/usr/local/zai"
	lhlcFileName    = "lhlc"
)

var _globalHLC = NewHLC(defaultRootPath)

// NewHLC creates an HLC for application.
// Each instance should have one.
func NewHLC(rootPath string) *HLC {
	h := &HLC{rootPath: rootPath}
	return h
}

// ResetGlobalHLC changes globalHLC's path.
// It's important that we could use a path belongs to a NVMe device,
// it'll improving performance in
func ResetGlobalHLC(rootPath string) {
	_globalHLC = NewHLC(rootPath)
}

// Next returns a timestamp.
func (h *HLC) Next() int64 {
	return 0
}

func HLCNext() {

}
