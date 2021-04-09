package lhlc

import (
	"testing"

	"g.tesamc.com/IT/zaipkg/vfs"
)

func BenchmarkLHLC_Next(b *testing.B) {
	l := CreateLHLC("", vfs.GetFS())

	for i := 0; i < b.N; i++ {
		_ = l.Next()
	}
}
