package orpc

import (
	"math"
	"sync/atomic"
	"time"
)

// Retryer provides methods to give the sleep duration which indicating when to start the next retrying.
// The major idea is exponential backoff algorithms with jitter (randomized delay):
// 1. Exponential backoff: Use progressively longer waits between retries for consecutive error responses
// 2. Jitter: Prevent successive collisions (we may have concurrent clients to use the Retryer)
// 3. Sublinear function of req/resp size: The size bigger, the delay is higher.
type Retryer struct {
	MinSleep time.Duration

	MaxTried int // After MaxTried, using MaxSleep.
	MaxSleep time.Duration
}

func (r *Retryer) GetSleepDuration(haveTried int, bodySize int64) time.Duration {
	if haveTried == 0 {
		return r.MinSleep
	}
	if haveTried >= r.MaxTried {
		return r.MaxSleep
	}

	s := calcSleepDuration(int64(r.MinSleep), bodySize, int64(haveTried))
	if s > float64(r.MaxSleep) {
		return r.MaxSleep
	}
	return time.Duration(int64(s))
}

func calcSleepDuration(min, n, tried int64) float64 {
	return getJitter() * float64(min) * calcSizeCoeff(n) * calcTriedCoeff(tried)
}

const (
	jitterMin = 0.7
	jitterMax = 1.3
)

var (
	jitterC     = []float64{0.0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9}
	nextJitterC int64
)

func getJitter() float64 {

	c := jitterC[atomic.AddInt64(&nextJitterC, 1)%10]
	return jitterMin + c*(jitterMax-jitterMin)
}

const (
	tryExpCoeff = 0.618 // tryExpCoeff controls the backoff speed.
)

// coeff = e^(tryExpCoeff * tried)
func calcTriedCoeff(tried int64) float64 {
	return math.Pow(math.E, tryExpCoeff*float64(tried))
}

func calcSizeCoeff(n int64) float64 {
	return 1 + (math.Pow(float64(n/128/1024), 0.618) * 0.25) // Using 128KB as size unit.
}
