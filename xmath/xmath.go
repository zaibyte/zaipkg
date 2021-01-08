package xmath

import "math"

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
