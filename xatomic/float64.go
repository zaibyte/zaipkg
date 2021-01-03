package xatomic

import (
	"math"
	"sync/atomic"
)

// Float64 is an atomic type-safe wrapper for float64 values.
type Float64 struct {
	v uint64
}

var _zeroFloat64 float64

// NewFloat64 creates a new Float64.
func NewFloat64(v float64) *Float64 {
	x := &Float64{}
	if v != _zeroFloat64 {
		x.Store(v)
	}
	return x
}

// Load atomically loads the wrapped float64.
func (x *Float64) Load() float64 {
	return math.Float64frombits(atomic.LoadUint64(&x.v))
}

// Store atomically stores the passed float64.
func (x *Float64) Store(v float64) {
	atomic.StoreUint64(&x.v, math.Float64bits(v))
}

// CAS is an atomic compare-and-swap for float64 values.
func (x *Float64) CAS(o, n float64) bool {
	return atomic.CompareAndSwapUint64(&x.v, math.Float64bits(o), math.Float64bits(n))
}