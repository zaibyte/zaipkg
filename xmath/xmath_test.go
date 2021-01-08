package xmath

import "testing"

func TestRound(t *testing.T) {
	f := 1.1
	var i float64
	for i = 0; i < 0.05; i += 0.01 {
		if Round(f+i, 1) != 1.1 {
			t.Fatal("mismatch")
		}
	}
	for i = 0.05; i < 0.1; i += 0.01 {
		if Round(f+i, 1) != 1.2 {
			t.Fatal("mismatch")
		}
	}
}

func TestAlignTo(t *testing.T) {
	var align int64 = 1 << 12
	var i int64

	for i = 1; i <= align; i++ {
		n := AlignTo(i, align)
		if n != align {
			t.Fatal("align mismatch")
		}
	}

	for i = align + 1; i < align*2; i++ {
		n := AlignTo(i, align)
		if n != align*2 {
			t.Fatal("align mismatch")
		}
	}
}
