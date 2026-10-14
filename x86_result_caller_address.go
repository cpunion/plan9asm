package plan9asm

// TEST and CMP's register/immediate read-only rows observe actual incoming
// caller addresses. CMP requires caller provenance for every register operand;
// a numerical register is not a second address participant. The same physical
// typed byte view and original immediate
// encoding are used by lowering. This does not initialize GP bytes, prove a
// pointee, or authorize a numerical/address result, store or native entry.
func x86ResultCallerAddressTest(ins Instr, state *x86ResultState, owned []bool, arch string) (bool, bool) {
	form, known, err := x86ScalarCompareProperties(ins.Op, ins, arch)
	if !known || err != nil {
		return false, false
	}
	address, valid := false, true
	for _, operand := range form.args {
		if operand.Kind == OpReg {
			root, _, physical := x86ResultGP(operand.Reg, int64(form.bits/8), arch)
			// Source-wide shape validation sees all registered numerical
			// registers together. It supplies no reaching caller provenance;
			// the same operands must pass this gate again with shapeOnly=false.
			if physical && (state.callerAddressGP[root] || !form.test && state.shapeOnly) {
				address = true
				continue
			}
		}
		if form.test {
			valid = valid && (operand.Kind == OpReg || operand.Kind == OpImm) &&
				x86ResultRead(operand, int64(form.bits/8), state, owned, arch)
		} else {
			valid = valid && operand.Kind == OpImm && x86ResultRead(operand, int64(form.bits/8), state, owned, arch)
		}
	}
	if !address {
		return false, false
	}
	if !valid || ins.x86Encoded || ins.x86SourceRaw != nil || ins.x86RawNearReturn != nil ||
		ins.x86AddressBits != 0 || ins.x86UnboundAddressOverride || ins.x86SegmentPrefix != 0 {
		return false, true
	}
	state.flags = x86ByteFlagsAll
	if form.test {
		state.flags &^= x86ByteFlagAF
	}
	return true, true
}

// This post-transfer must-domain runs through the shared complete source CFG,
// including dead effects and joins. A partial/full GP overwrite revokes its
// old provenance independently of initialized bytes or opaque pointerGP.
// Unknown effects conservatively revoke all caller addresses, never seed one.
func x86ResultCallerAddressesAfter(ins Instr, before [16]bool, entryFP map[int64]bool, arch string) [16]bool {
	if ins.x86Encoded || ins.x86SourceRaw != nil || ins.x86RawNearReturn != nil ||
		ins.x86AddressBits != 0 || ins.x86UnboundAddressOverride || ins.x86SegmentPrefix != 0 {
		return [16]bool{}
	}
	ins = x86GoMOVQSourceInstruction(ins) // Original source/certificates stay unchanged.
	word := x86ResultPointerWord(arch)
	move, lea, add, sub := Op("MOVQ"), Op("LEAQ"), Op("ADDQ"), Op("SUBQ")
	if arch == "386" {
		move, lea, add, sub = "MOVL", "LEAL", "ADDL", "SUBL"
	}
	if len(ins.Args) == 2 && ins.Args[1].Kind == OpReg {
		destination, _, valid := x86ResultGP(ins.Args[1].Reg, word, arch)
		if valid {
			first := ins.Args[0]
			switch ins.Op {
			case move:
				caller := first.Kind == OpFP && entryFP[first.FPOffset]
				if first.Kind == OpReg {
					source, _, valid := x86ResultGP(first.Reg, word, arch)
					caller = valid && before[source]
				}
				before[destination] = caller
				return before
			case lea:
				base, _, valid := x86ResultGP(first.Mem.Base, word, arch)
				before[destination] = first.Kind == OpMem && first.Mem.Sym == "" && valid && before[base]
				return before // Numerical index/complete address checked by the original transfer.
			case add, sub:
				if ins.Op == sub && first.Kind == OpReg {
					if source, _, valid := x86ResultGP(first.Reg, word, arch); valid && before[source] {
						before[destination] = false // A proved same-origin difference is numerical.
					}
				}
				return before // Only the separately validated full-word address transform preserves it.
			}
		}
	}
	kill := func(operand Operand, width int64) {
		if operand.Kind == OpReg {
			if root, _, valid := x86ResultGP(operand.Reg, width, arch); valid {
				before[root] = false
			}
		}
	}
	if effect, known, err := x86TypedInstructionByteEffect(ins.Op, ins, arch); known && err == nil {
		for _, operand := range effect.operands {
			if operand.write {
				kill(operand.operand, operand.width)
			}
		}
		return before
	}
	if form, known, err := x86MoveMaskProperties(ins.Op, ins, arch); known {
		if err != nil || form.byteWidth == 8 {
			return [16]bool{} // Unbound physical MMX is not a caller effect.
		}
		index, valid := amd64VectorRegisterIndex(ins.Args[0].Reg, form.byteWidth)
		destination, _, physical := x86ResultGP(ins.Args[1].Reg, 4, arch)
		if !valid || arch == "386" && index >= 8 || !physical {
			return [16]bool{}
		}
		before[destination] = false // Complete typed r32 output, no other GP writes.
		return before
	}
	if _, known, err := x86BroadcastProperties(ins.Op, ins, arch); known && err == nil {
		return before // Scalar/block/mask inputs never give this vector output a GP write.
	}
	if _, known, err := x86VectorTestProperties(ins.Op, ins, arch); known && err == nil {
		return before // Only FLAGS change; source/memory definedness is checked independently.
	}
	if form, known, err := x86ScalarBitScanProperties(ins.Op, ins, arch); known && err == nil {
		kill(ins.Args[1], int64(form.bits/8))
		return before // Conditional numerical definedness does not widen the write footprint.
	}
	if form, known, err := x86ConditionalSetProperties(ins.Op, ins, arch); known && err == nil {
		kill(form.destination, 1)
		return before
	}
	if _, known, err := x86ScalarCompareProperties(ins.Op, ins, arch); known && err == nil {
		return before
	}
	if width := x86RawReturnScalarMoveBytes(ins.Op); width != 0 && len(ins.Args) == 2 {
		kill(ins.Args[1], width)
		return before
	}
	if form, known, err := x86GPSIMDTransferProperties(ins, arch); known && err == nil && !form.mmx {
		if !form.destinationX {
			kill(ins.Args[1], int64(form.bits/8))
		}
		return before
	}
	if form, known, err := x86ScalarLaneProperties(ins, arch); known && err == nil {
		if !form.insert {
			kill(Operand{Kind: OpReg, Reg: form.transferReg}, int64(form.gpBits/8))
		}
		return before
	}
	if x86CallerVectorMoveKnown(ins) {
		for _, operand := range ins.Args {
			if operand.Kind == OpReg {
				_, vector := amd64VectorRegisterIndex(operand.Reg, amd64VectorByteWidth(operand.Reg))
				_, mask := amd64ParseKReg(operand.Reg)
				if !vector && !mask {
					return [16]bool{} // GP/alias/native banks need their own exact effect.
				}
			}
		}
		return before // Complete shared typed vector-only form, no GP output.
	}
	if ins.Op == OpJMP || isAMD64ConditionalBranch(ins.Op) || ins.Op == OpTEXT || ins.Op == OpLABEL || ins.Op == "NOP" {
		return before
	}
	if _, known, err := x86DirectiveProperties(ins); known && err == nil {
		return before
	}
	return [16]bool{}
}
