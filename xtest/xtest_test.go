package xtest

import "testing"

func BenchmarkDoNothing(b *testing.B) {

	for i := 0; i < b.N; i++ {
		DoNothing(1000)
	}
}
