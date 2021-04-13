package app

import (
	"context"
	"time"

	"github.com/templexxx/tsc"
)

const DefaultTimeCalibrateInterval = 15 * time.Minute

func TimeCalibrateLoop(ctx context.Context, interval time.Duration) {

	cancelLoopCtx, cancel := context.WithCancel(ctx)
	defer cancel()

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
