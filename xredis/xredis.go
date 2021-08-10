// Package xredis provides helper functions for communicating with redis.
package xredis

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/go-redis/redis/v8"
)

// NewClient creates a redis.Client by url,
// return fail-over client if it's a set of urls;
// return normal client if it's just one url.
//
// Options are followed the default values in redis lib,
// I think the default values are well enough.
func NewClient(url string) (*redis.Client, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %s", url, err)
	}
	var rdb *redis.Client
	if strings.Contains(opt.Addr, ",") {
		var fopt redis.FailoverOptions
		ps := strings.Split(opt.Addr, ",")
		fopt.MasterName = ps[0]
		fopt.SentinelAddrs = ps[1:]
		_, port, _ := net.SplitHostPort(fopt.SentinelAddrs[len(fopt.SentinelAddrs)-1])
		if port != "" {
			for i := range fopt.SentinelAddrs {
				h, p, _ := net.SplitHostPort(fopt.SentinelAddrs[i])
				if p == "" {
					fopt.SentinelAddrs[i] = net.JoinHostPort(h, port)
				}
			}
		}
		// Assume Redis server and sentinel have the same password.
		fopt.SentinelPassword = opt.Password
		fopt.Username = opt.Username
		fopt.Password = opt.Password
		if fopt.SentinelPassword == "" && os.Getenv("SENTINEL_PASSWORD") != "" {
			fopt.SentinelPassword = os.Getenv("SENTINEL_PASSWORD")
		}
		if fopt.Password == "" && os.Getenv("REDIS_PASSWORD") != "" {
			fopt.Password = os.Getenv("REDIS_PASSWORD")
		}
		fopt.DB = opt.DB
		fopt.TLSConfig = opt.TLSConfig
		rdb = redis.NewFailoverClient(&fopt)
	} else {
		if opt.Password == "" && os.Getenv("REDIS_PASSWORD") != "" {
			opt.Password = os.Getenv("REDIS_PASSWORD")
		}
		rdb = redis.NewClient(opt)
	}
	return rdb, nil
}
