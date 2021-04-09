package lhlc

var (
	_global *HLC
)

// Init Global var.
// warn: It's unsafe for concurrent use.
func InitGlobalHLC(h *HLC) {
	_global = h
}
