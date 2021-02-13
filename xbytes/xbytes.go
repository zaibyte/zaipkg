package xbytes

// AlignSize calculates size after align n to align.
// The return will be the multiple of align.
func AlignSize(n int64, align int64) int64 {
	return (n + align - 1) &^ (align - 1)
}
