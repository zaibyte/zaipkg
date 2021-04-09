package hlc

var (
	_global HLC
)

// InitGlobalHLC Inits Global var.
// warn: It's unsafe for concurrent use.
func InitGlobalHLC(h HLC) {
	_global = h
}
