package xbytes

// AlignSize calculates size after align n to align.
// The return will be the multiple of align.
func AlignSize(n int64, align int64) int64 {
	if n <= 0 {
		return align
	}
	return (n + align - 1) &^ (align - 1)
}
