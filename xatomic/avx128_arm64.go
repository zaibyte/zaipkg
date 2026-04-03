package xatomic

import "unsafe"

// AvxLoad16B is a compatibility fallback on arm64.
// arm64 has no AVX; this keeps behavior by using the atomic 16-byte load path.
func AvxLoad16B(src, dst *byte) {
	v := AtomicLoad16B(src)
	copy(unsafe.Slice(dst, 16), v[:])
}

// AvxStore16B is a compatibility fallback on arm64.
// arm64 has no AVX; this keeps behavior by using the atomic 16-byte store path.
func AvxStore16B(src, val *byte) {
	var v [16]byte
	copy(v[:], unsafe.Slice(val, 16))
	AtomicStore16B(src, v)
}
