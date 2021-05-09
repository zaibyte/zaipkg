#include "textflag.h"

// func AvxLoad16B(src, dst *byte)
TEXT ·AvxLoad16B(SB), NOSPLIT, $0
    MOVQ src+0(FP), R8
    MOVQ dst+8(FP), R9
    VMOVDQU (R8), X3
    VMOVDQU X3, (R9)
	RET

// func AvxStore16B(src, val *byte)
TEXT ·AvxStore16B(SB),NOSPLIT,$0
	MOVQ src+0(FP), R8
	MOVQ val+8(FP), R9
	VMOVDQU (R9), X3
	VMOVDQU X3, (R8)
	RET
