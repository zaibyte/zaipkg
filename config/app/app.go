package app

import "g.tesamc.com/IT/zaipkg/xlog"

type App struct {
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
}
