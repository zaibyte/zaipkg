package app

import (
	"context"
	"time"

	"g.tesamc.com/IT/zaipkg/xlog"

	"github.com/templexxx/tsc"
)

const DefaultTimeCalibrateInterval = 5 * time.Minute // < 1/2 ntpd default sync interval(11min).

func TimeCalibrateLoop(ctx context.Context, interval time.Duration) {

	cancelLoopCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	if !tsc.Enabled() {
		xlog.Info("tsc is not enabled, using system clock")
		return
	}

	if interval == 0 {
		interval = DefaultTimeCalibrateInterval
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
