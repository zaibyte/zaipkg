// Copyright 2014, David Minor. All rights reserved.
// Use of this source code is governed by the MIT
// license which can be found in the LICENSE file.

package xmath

import (
	"encoding/binary"

	"github.com/zaibyte/zaipkg/xbytes"
)

type Uint128 struct {
	H, L uint64
}

func (u Uint128) ToArr() []byte {
	ret := xbytes.MakeAlignedBlock(16, 16)
	binary.LittleEndian.PutUint64(ret[:8], u.L)
	binary.LittleEndian.PutUint64(ret[8:16], u.H)
	return ret
}

func (u Uint128) ToArrDst(dst []byte) {
	binary.LittleEndian.PutUint64(dst[:8], u.L)
	binary.LittleEndian.PutUint64(dst[8:16], u.H)
}

func FromArrToUint128(a []byte) Uint128 {
	var u Uint128
	u.L = binary.LittleEndian.Uint64(a[:8])
	u.H = binary.LittleEndian.Uint64(a[8:16])
	return u
}

func (u Uint128) ShiftLeft(bits uint) Uint128 {
	if bits >= 128 {
		u.H = 0
		u.L = 0
	} else if bits >= 64 {
		u.H = u.L << (bits - 64)
		u.L = 0
	} else {
		u.H <<= bits
		u.H |= u.L >> (64 - bits)
		u.L <<= bits
	}
	return u
}

func (u Uint128) ShiftRight(bits uint) Uint128 {
	if bits >= 128 {
		u.H = 0
		u.L = 0
	} else if bits >= 64 {
		u.L = u.H >> (bits - 64)
		u.H = 0
	} else {
		u.L >>= bits
		u.L |= u.H << (64 - bits)
		u.H >>= bits
	}
	return u
}

func (u Uint128) And(y Uint128) Uint128 {
	u.H &= y.H
	u.L &= y.L
	return u
}

func (u Uint128) Xor(y Uint128) Uint128 {
	u.H ^= y.H
	u.L ^= y.L
	return u
}

func (u Uint128) Or(y Uint128) Uint128 {
	u.H |= y.H
	u.L |= y.L
	return u
}

func (u Uint128) Add(addend Uint128) Uint128 {
	origlow := u.L
	u.L += addend.L
	u.H += addend.H
	if u.L < origlow { // wrapping occurred, so carry the 1
		u.H += 1
	}
	return u
}

// (Adapted from go's math/big)
// z1<<64 + z0 = x*y
// Adapted from Warren, Hacker's Delight, p. 132.
func mult(x, y uint64) (z1, z0 uint64) {
	z0 = x * y // lower 64 bits are easy
	// break the multiplication into (x1 << 32 + x0)(y1 << 32 + y0)
	// which is x1*y1 << 64 + (x0*y1 + x1*y0) << 32 + x0*y0
	// so now we can do 64 bit multiplication and addition and
	// shift the results into the right place
	x0, x1 := x&0x00000000ffffffff, x>>32
	y0, y1 := y&0x00000000ffffffff, y>>32
	w0 := x0 * y0
	t := x1*y0 + w0>>32
	w1 := t & 0x00000000ffffffff
	w2 := t >> 32
	w1 += x0 * y1
	z1 = x1*y1 + w2 + w1>>32
	return
}

func (u Uint128) Mult(multiplier Uint128) Uint128 {
	hl := u.H*multiplier.L + u.L*multiplier.H
	u.H, u.L = mult(u.L, multiplier.L)
	u.H += hl
	return u
}
