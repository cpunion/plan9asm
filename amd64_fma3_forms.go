package plan9asm

import (
	"fmt"
	"strings"
)

type amd64FMA3Form struct {
	spec        amd64FMA3Spec
	suffix      amd64FMA3Suffix
	byteWidth   int
	memoryBytes int
	destination int
	mask        int // -1 when unmasked
	memory      int // -1 for a register first source
}

// The complete _yvaddpd/_yvaddsd axes are independent of numerical input,
// caller-memory, initialized-mask and source-stack continuation proofs. FP16
// shares the architectural grammar but keeps its raw-encoding-only boundary.
func amd64FMA3Properties(ins Instr, goarch string) (amd64FMA3Form, bool, error) {
	base, suffix, dotted := strings.Cut(strings.ToUpper(string(ins.Op)), ".")
	spec, recognized := amd64FMA3Specs[Op(base)]
	if !recognized {
		return amd64FMA3Form{}, false, nil
	}
	form := amd64FMA3Form{
		spec: spec, destination: len(ins.Args) - 1, mask: -1, memory: -1,
	}
	fail := func() (amd64FMA3Form, bool, error) {
		return form, true, fmt.Errorf("%s %s operands/suffix are outside the FMA3 grammar: %q", goarch, base, ins.Raw)
	}
	if goarch != "amd64" && goarch != "386" || spec.rawOnly && !ins.x86Encoded ||
		dotted && suffix == "" || len(ins.Args) != 3 && len(ins.Args) != 4 {
		return fail()
	}
	properties, valid := parseAMD64FMA3Suffix(suffix)
	if !valid || spec.scalar && properties.broadcast {
		return fail()
	}
	form.suffix = properties
	if len(ins.Args) == 4 {
		if goarch == "386" && !ins.x86Encoded || !amd64NonzeroKOperand(ins.Args[2]) {
			return fail()
		}
		form.mask = 2
	} else if properties.zeroing {
		return fail()
	}
	form.byteWidth = amd64VectorByteWidth(ins.Args[form.destination].Reg)
	if spec.scalar && form.byteWidth != 16 ||
		!amd64GoPackedVectorMoveRegister(ins.Args[form.destination], form.byteWidth, goarch) ||
		!amd64GoPackedVectorMoveRegister(ins.Args[1], form.byteWidth, goarch) ||
		properties.rounding != "" && !spec.scalar && form.byteWidth != 64 {
		return fail()
	}
	first := ins.Args[0]
	if properties.broadcast {
		if !isAMD64MemoryOperand(first) {
			return fail()
		}
	} else if first.Kind == OpReg {
		if !amd64GoPackedVectorMoveRegister(first, form.byteWidth, goarch) {
			return fail()
		}
	} else if properties.rounding != "" || !isAMD64MemoryOperand(first) {
		return fail()
	}
	if isAMD64MemoryOperand(first) {
		if !x86PackedShiftMemoryValid(first, goarch) {
			return fail()
		}
		if first.Kind == OpMem {
			if err := x86SymbolIndexedMemoryError(first.Mem, goarch); err != nil {
				return form, true, err
			}
			if first.Mem.OffRaw != "" || first.Mem.IndexExt != "" {
				return fail()
			}
		}
		form.memory = 0
		form.memoryBytes = form.byteWidth
		if spec.scalar || properties.broadcast {
			form.memoryBytes = spec.laneBits / 8
		}
	}
	return form, true, x86Mode32VectorRegisterContext(ins, goarch)
}
