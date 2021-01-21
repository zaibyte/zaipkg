#include "textflag.h"

// func LFence()
TEXT ·LFence+0(SB), NOSPLIT,$0-0
	LFENCE
	RET

// func MFence()
TEXT ·MFence+0(SB), NOSPLIT,$0-0
	MFENCE
	RET

// func SFence()
TEXT ·SFence+0(SB), NOSPLIT,$0-0
	SFENCE
	RET
