package xatomic

// On the latest CPU micro-architectures (Skylake and Zen 2) AVX/AVX2 128b/256b aligned loads and stores are atomic
// even though Intel and AMD officially don’t guarantee this.
// https://rigtorp.se/isatomic/
// We assume that we're using the latest CPU (after Skylake).

// AvxLoad16B atomically loads 16bytes from *addr.
// src & dst must be 16bytes aligned.
//
//go:noescape
func AvxLoad16B(src, dst *byte)

// AvxStore16B atomically stores 16bytes to *addr.
// src & val must be 16bytes aligned.
//
//go:noescape
func AvxStore16B(src, val *byte)
