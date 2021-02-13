package xbytes

import "testing"

func TestAlignSize(t *testing.T) {
	var align int64 = 1 << 12
	var i int64
	for i = 1; i <= align; i++ {
		n := AlignSize(i, align)
		if n != align {
			t.Fatal("align mismatch", n, i)
		}
	}
	for i = align + 1; i <= align*2; i++ {
		n := AlignSize(i, align)
		if n != align*2 {
			t.Fatal("align mismatch")
		}
	}
}
