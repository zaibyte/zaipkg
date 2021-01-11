package app

import (
	"context"
	"sync"
	"time"

	"github.com/templexxx/tsc"
)

const DefaultTimeCalibrateInterval = 15 * time.Minute

func TimeCalibrateLoop(ctx context.Context, wg *sync.WaitGroup, interval time.Duration) {

	defer wg.Done()

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
