package app

import (
	"g.tesamc.com/IT/zaipkg/xlog"
)

type Config struct {
	// Every instance belongs to a certain box.
	// boxID: [1, 255)
	BoxID int64 `toml:"box_id"`
	// Every instance has its own unique instanceID
	InstanceID string `toml:"instance-id"`
	// Every instance belongs to a certain rack.
	// RackID: [1, ?)
	// 0 is reserved, it's default rack, means ignore rack location.
	RackID uint64 `toml:"rack_id"`

	Log xlog.Config `toml:"log"`

	HTTPServerAddr string `toml:"http_server_addr"`

	// Enable Prometheus time histogram may impact performance.
	// Deprecated.
	EnableHandlingTimeHistogram bool `toml:"enable_handling_time_histogram"`
}
