package plan9asm

import "strings"

// Both result observers consume the same typed numerical transfer. The public
// observer enters only for register/register operands; the private observer
// retains its existing logical-memory roles. Neither acquires native backing,
// caller permission or numerical pointer bytes.
func x86ResultGPSIMDTransfer(ins Instr, form x86GPSIMDTransferForm, state *x86ResultState,
	owned []bool, resultOffset int64, arch string) bool {
	return x86ResultGPSIMDTransferWithRead(ins, form, state, owned, resultOffset, arch, false)
}

func x86ResultGPSIMDTransferWithRead(ins Instr, form x86GPSIMDTransferForm, state *x86ResultState,
	owned []bool, resultOffset int64, arch string, conditionalNativeRead bool) bool {
	if form.mmx || form.spec.rawOnly {
		return false
	}
	width := int64(form.bits / 8)
	lowMask := uint16(1)<<uint(width) - 1
	if form.sourceX {
		index, valid := amd64ParseXReg(ins.Args[0].Reg)
		if !valid || state.xmm[index]&lowMask != lowMask {
			return false
		}
	} else if !x86ResultReadWithNativeContinuation(ins, ins.Args[0], width, state, owned, arch, conditionalNativeRead) {
		return false
	}
	if form.destinationX {
		index, valid := amd64ParseXReg(ins.Args[1].Reg)
		if !valid {
			return false
		}
		// Each valid GP/X move writes its low 32/64 bits and clears the
		// remainder of X. The encoding chooses preservation or clearing of
		// the wider physical aliases, independently of their previous state.
		state.writeVector(index, 16, 0xffff, form.spec.legacy)
		return true
	}
	if form.memoryWrite && ins.Args[1].Kind == OpSym {
		// Only the existing private memory consumer enters this branch.
		// An ordinary SB store defines no FP result bytes and supplies no
		// address or extent certificate.
		symbol := strings.TrimSpace(ins.Args[1].Sym)
		base, _, valid := parseSBRef(symbol)
		return valid && strings.HasSuffix(symbol, "(SB)") &&
			strings.IndexAny(base, "$+(),[] \t") < 0
	}
	return x86ResultStore(ins.Args[1], width, state, owned, resultOffset, arch)
}
