package plan9asm

// inferX86ABI0Forwarder proves the complete straight-line adapter shape before
// assigning a typed ABI0 signature to an indirect CALL. The final incoming
// word is the code pointer, not an argument of the callback. This is structural
// inference: no library, function or parameter names participate.
func inferX86ABI0Forwarder(fn Func, sig FuncSig, goarch string) *FuncSig {
	if goarch != "amd64" || len(sig.ArgRegs) != 0 || len(sig.Args) == 0 {
		return nil
	}
	codeIndex := len(sig.Args) - 1
	if sig.Args[codeIndex] != I64 && sig.Args[codeIndex] != Ptr {
		return nil
	}
	var code FrameSlot
	var params []FrameSlot
	codeSlots := 0
	for _, slot := range sig.Frame.Params {
		if slot.Index == codeIndex {
			code = slot
			codeSlots++
		} else {
			params = append(params, slot)
		}
	}
	if codeSlots != 1 || len(frameSlotFields(code)) != 0 {
		return nil
	}

	var instructions []Instr
	for _, ins := range fn.Instrs {
		switch ins.Op {
		case "TEXT", "PCDATA", "FUNCDATA", "GO_ARGS", "NO_LOCAL_POINTERS":
			continue
		}
		instructions = append(instructions, ins)
	}
	// Two instructions per copied slot, code-pointer load, CALL, then RET.
	if len(instructions) != 2*len(params)+2*len(sig.Frame.Results)+3 {
		return nil
	}
	for i, slot := range params {
		load, store := instructions[2*i], instructions[2*i+1]
		if !x86ABI0ForwardCopy(load, store, slot, true) ||
			store.Args[1].Mem.Off != slot.Offset || slot.Offset >= code.Offset {
			return nil
		}
	}
	callIndex := 2*len(params) + 1
	load, call := instructions[callIndex-1], instructions[callIndex]
	if load.Op != "MOVQ" || len(load.Args) != 2 || load.Args[0].Kind != OpFP ||
		load.Args[0].FPOffset != code.Offset || load.Args[1].Kind != OpReg ||
		call.Op != "CALL" || len(call.Args) != 1 || call.Args[0].Kind != OpReg ||
		call.Args[0].Reg != load.Args[1].Reg {
		return nil
	}
	results := append([]FrameSlot(nil), sig.Frame.Results...)
	for i, slot := range results {
		load, store := instructions[callIndex+1+2*i], instructions[callIndex+2+2*i]
		if !x86ABI0ForwardCopy(load, store, slot, false) {
			return nil
		}
		// Removing the trailing pointer shifts the callback's results one
		// pointer word earlier. Do not infer a signature from arbitrary loads.
		if load.Args[0].Mem.Off != slot.Offset-8 || slot.Offset-8 < code.Offset {
			return nil
		}
		results[i].Offset -= 8
	}
	last := instructions[len(instructions)-1]
	if last.Op != "RET" || len(last.Args) != 0 {
		return nil
	}
	return &FuncSig{
		Args: append([]LLVMType(nil), sig.Args[:codeIndex]...), Ret: sig.Ret,
		Frame: FrameLayout{Params: params, Results: results},
	}
}

func x86ABI0ForwardCopy(load, store Instr, slot FrameSlot, incoming bool) bool {
	if len(load.Args) != 2 || len(store.Args) != 2 || load.Op != store.Op ||
		load.Args[1].Kind != OpReg || store.Args[0].Kind != OpReg ||
		load.Args[1].Reg != store.Args[0].Reg {
		return false
	}
	var width int64
	switch load.Op {
	case "MOVB":
		width = 1
	case "MOVW":
		width = 2
	case "MOVL", "MOVSS":
		width = 4
	case "MOVQ", "MOVSD":
		width = 8
	default:
		return false
	}
	if width != frameTypeSize(slot.Type, 8) {
		return false
	}
	fp, sp := load.Args[0], store.Args[1]
	if !incoming {
		fp, sp = store.Args[1], load.Args[0]
	}
	return fp.Kind == OpFP && fp.FPOffset == slot.Offset &&
		sp.Kind == OpMem && sp.Mem.Base == SP && sp.Mem.Index == "" && sp.Mem.Sym == ""
}
