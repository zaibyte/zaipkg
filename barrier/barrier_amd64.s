#include "textflag.h"

// func LFence()
TEXT ·LFence(SB), NOSPLIT,$0-0
	LFENCE
	RET

// func MFence()
TEXT ·MFence(SB), NOSPLIT,$0-0
	MFENCE
	RET

// func SFence()
TEXT ·SFence(SB), NOSPLIT,$0-0
	SFENCE
	RET
