package app

import (
	"g.tesamc.com/IT/zaipkg/typeutil"
	"g.tesamc.com/IT/zaipkg/xlog"
)

type Config struct {
	// Every instance belongs to a certain box.
	// boxID: [1, 255)
	BoxID uint32 `toml:"box_id"`
	IDC   string `toml:"idc"`
	// Every instance has its own unique instanceID
	InstanceID string `toml:"instance_id"`
	// Every instance belongs to a certain rack.
	RackID string `toml:"rack_id"`

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
