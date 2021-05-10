package app

import (
	"g.tesamc.com/IT/zaipkg/typeutil"
	"g.tesamc.com/IT/zaipkg/xlog"
)

type Config struct {
	// Every instance belongs to a certain box.
	// boxID: [1, 255)
	BoxID uint32 `toml:"box_id"`
	// Every instance has its own unique instanceID
	InstanceID uint32 `toml:"instance-id"`
	// Every instance belongs to a certain rack.
	// RackID: [1, ?)
	// 0 is reserved, it's default rack, means ignore rack location.
	RackID uint64 `toml:"rack_id"`

	Log xlog.Config `toml:"log"`

	// ServerAddr is the server address.
	ServerAddr string `toml:"server_addr"`

	// Enable Prometheus time histogram may impact performance.
	// Deprecated.
	EnableHandlingTimeHistogram bool `toml:"enable_handling_time_histogram"`

	// GOMAXPROCS sets runtime.GOXMAXPROCS manually.
	// Sometimes we want more Go Processes for reducing stall.
	GOMAXPROCS int `toml:"gomaxprocs"`

	TimeCalibrateInterval typeutil.Duration `toml:"time_calibrate_interval"`
}
