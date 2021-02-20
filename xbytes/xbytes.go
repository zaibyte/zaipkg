package xbytes

// AlignSize calculates size after align.
// The return will be the multiple of align.
func AlignSize(n int64, align int64) int64 {
	if n <= 0 {
		return 0
	}
	return (n + align - 1) &^ (align - 1)
}
