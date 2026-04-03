package barrier

import "testing"

func BenchmarkLFence(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LFence()
	}
}

func BenchmarkMFence(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MFence()
	}
}

func BenchmarkSFence(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SFence()
	}
}
