package app

import (
	"context"
	"time"

	"g.tesamc.com/IT/zaipkg/xlog"

	"github.com/templexxx/tsc"
)

const DefaultTimeCalibrateInterval = 15 * time.Minute

func TimeCalibrateLoop(ctx context.Context, interval time.Duration) {

	cancelLoopCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	if !tsc.Enabled {
		xlog.Warn("tsc is not enabled, using system clock")
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			tsc.Calibrate()
		case <-cancelLoopCtx.Done():
			return
		}
	}
}
