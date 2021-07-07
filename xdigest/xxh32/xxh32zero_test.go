package xxh32

import (
	"hash/crc32"
	"math/rand"
	"testing"
)

func BenchmarkChecksumZero4K(b *testing.B) {

	p := make([]byte, 4096)
	rand.Read(p)

	b.SetBytes(4096)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ChecksumZero(p)
	}
}

// crcTable uses Castagnoli which has better error detection characteristics than IEEE and faster.
var crcTable = crc32.MakeTable(crc32.Castagnoli)

func BenchmarkCRC4K(b *testing.B) {

	p := make([]byte, 4096)
	rand.Read(p)

	b.SetBytes(4096)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = crc32.Checksum(p, crcTable)
	}
}
