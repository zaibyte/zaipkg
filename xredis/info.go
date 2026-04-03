package xredis

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/zaibyte/zaipkg/xlog"
)

type version struct {
	major, minor, patch int
}

var oldestSupportedVer = version{2, 2, 0}

func parseVersion(v string) (ver version, err error) {
	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		err = fmt.Errorf("invalid version: %v", v)
		return
	}
	ver.major, err = strconv.Atoi(parts[0])
	if err != nil {
		return
	}
	ver.minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return
	}
	ver.patch, err = strconv.Atoi(parts[2])
	return
}

func (ver version) olderThan(v2 version) bool {
	if ver.major < v2.major {
		return true
	}
	if ver.major > v2.major {
		return false
	}
	if ver.minor < v2.minor {
		return true
	}
	if ver.minor > v2.minor {
		return false
	}
	return ver.patch < v2.patch
}

func (ver version) String() string {
	return fmt.Sprintf("%d.%d.%d", ver.major, ver.minor, ver.patch)
}

type Info struct {
	aofEnabled      bool
	clusterEnabled  bool
	maxMemoryPolicy string
	version         string
}

// CheckInfo checks whether redis could support all features.
// It's really hard to find such a old version redis :p in production env，
// log warn if really happen.
func CheckInfo(rawInfo string) (info Info, err error) {
	lines := strings.Split(strings.TrimSpace(rawInfo), "\n")
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		kvPair := strings.SplitN(l, ":", 2)
		key, val := kvPair[0], kvPair[1]
		switch key {
		case "aof_enabled":
			info.aofEnabled = val == "1"
			if val == "0" {
				xlog.Warnf("AOF is not enabled, you may lose data if Redis is not shutdown properly.")
			}
		case "cluster_enabled":
			info.clusterEnabled = val == "1"
			if val != "0" {
				xlog.Warnf("Redis cluster is not supported, some operation may fail unexpected.")
			}
		case "maxmemory_policy":
			info.maxMemoryPolicy = val
			if val != "noeviction" {
				xlog.Warnf("maxmemory_policy is %q, please set it to 'noeviction'.", val)
			}
		case "redis_version":
			info.version = val
			ver, err2 := parseVersion(val)
			if err2 != nil {
				xlog.Warnf("Failed to parse Redis server version: %q", ver)
			} else {
				if ver.olderThan(oldestSupportedVer) {
					xlog.Warnf("Redis version should not be older than %s", oldestSupportedVer)
				}
			}
		}
	}
	return
}
