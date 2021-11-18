package xmath

import (
	"math"
	"math/bits"
)

// Round rounds a float64 and cuts it by n.
// n: decimal places.
// e.g.
// f = 1.001, n = 2, return 1.00
func Round(f float64, n int) float64 {
	pow10n := math.Pow10(n)
	return math.Trunc(f*pow10n+0.5) / pow10n
}

// AlignSize returns size after n aligns to align.
func AlignSize(n int64, align int64) int64 {
	return (n + align - 1) &^ (align - 1)
}

// AlignToLast aligns n to the last align.
func AlignToLast(n int64, align int64) int64 {
	return AlignSize(n-(align-1), align)
}

// NextPower2 gets next number which is pow(2,x).
func NextPower2(n uint64) uint64 {
	if n <= 1 {
		return 1
	}

	return 1 << (64 - bits.LeadingZeros64(n-1)) // TODO may use BSR instruction.
}
