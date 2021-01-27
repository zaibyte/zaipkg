package xtime

import "time"

var nopTime = make(chan time.Time)

func init() {
	close(nopTime)
}

// GetTimeEvent gets a single time event.
//
// It's designed for process which need accurate control of time event.
// We want the next event will come after the last event finishing in duration.
// If we use time.Ticker there, the worst case is the cost of event is high, which means
// just after the event finishing, the ticker will tick again, that's not we want.
//
// e.g.
// t := time.NewTimer(duration)
// var tChan <-chan time.Time
// for {
// 	var m *msg
//
// 	select {
// 		case m = <-msgChan:
// 		case <-tChan:
// 			foo()
// 			tChan = nil
// 			continue
// 		}
// 	}
//
// 	if tChan == nil {
// 		tChan = xtime.GetTimeEvent(t, s.FlushDelay)
// 	}
//	...
func GetTimerEvent(t *time.Timer, duration time.Duration) <-chan time.Time {
	if duration <= 0 {
		return nopTime
	}

	if !t.Stop() {
		// Exhaust expired timer's chan.
		select {
		case <-t.C:
		default:
		}
	}
	t.Reset(duration)
	return t.C
}
