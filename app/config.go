package app

import (
	"fmt"

	"g.tesamc.com/IT/zaipkg/config"

	"g.tesamc.com/IT/zaipkg/xruntime"

	"g.tesamc.com/IT/zaipkg/uid"

	"g.tesamc.com/IT/zaipkg/typeutil"
	"g.tesamc.com/IT/zaipkg/xlog"
)

type Config struct {
	// KeeperClusterID is the cluster ID of keeper,
	// used for ensuring connecting the correct cluster.
	KeeperClusterID uint64 `toml:"keeper_cluster_id"`
	// Every instance has its own unique instanceID
	InstanceID string `toml:"instance_id"`

	Log xlog.Config `toml:"log"`

	// ServerAddr is the server address.
	// For std. daemon in Zai, it'll have an HTTP/1.1 server at least,
	// and other servers will use mux to share the same port with HTTP/1.1.
	ServerAddr string `toml:"server_addr"`

	// Sometimes we want more Go Processes for reducing stall.
	// GOMAXPROCS sets runtime.GOXMAXPROCS manually.
	GOMAXPROCS int `toml:"gomaxprocs"`

	TimeCalibrateInterval typeutil.Duration `toml:"time_calibrate_interval"`
}

// Adjust adjusts Config:
// checking the values first, then filling part of the empty with default values.
func (c *Config) Adjust() {

	if !uid.IsValidInstanceID(c.InstanceID) {
		panic(fmt.Sprintf("illegal instance_id: %s", c.InstanceID))
	}

	if c.GOMAXPROCS <= 0 {
		xruntime.AutoGOMAXPROCS()
	}

	config.Adjust(&c.TimeCalibrateInterval, DefaultTimeCalibrateInterval)
}
