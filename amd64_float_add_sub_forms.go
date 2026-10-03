package plan9asm

import (
	"fmt"
	"strings"
)

type x86FloatAddSubSpec struct {
	bits                     int
	legacy, scalar, subtract bool
}

// Go's eight yxm rows and eight _yvaddpd/_yvaddsd opcodes. Horizontal,
// alternating and raw-only FP16 arithmetic have separate contracts.
var x86FloatAddSubSpecs = map[Op]x86FloatAddSubSpec{
	"ADDPS": {32, true, false, false}, "ADDPD": {64, true, false, false},
	"ADDSS": {32, true, true, false}, "ADDSD": {64, true, true, false},
	"SUBPS": {32, true, false, true}, "SUBPD": {64, true, false, true},
	"SUBSS": {32, true, true, true}, "SUBSD": {64, true, true, true},
	"VADDPS": {32, false, false, false}, "VADDPD": {64, false, false, false},
	"VADDSS": {32, false, true, false}, "VADDSD": {64, false, true, false},
	"VSUBPS": {32, false, false, true}, "VSUBPD": {64, false, false, true},
	"VSUBSS": {32, false, true, true}, "VSUBSD": {64, false, true, true},
}

type x86FloatAddSubForm struct {
	spec                                     x86FloatAddSubSpec
	first, second, destination, mask, memory int
	byteWidth, memoryBytes                   int
	broadcast, zeroing                       bool
	rounding                                 string
}

func x86FloatAddSubProperties(ins Instr, goarch string) (x86FloatAddSubForm, bool, error) {
	base, suffix, dotted := strings.Cut(strings.ToUpper(string(ins.Op)), ".")
	spec, recognized := x86FloatAddSubSpecs[Op(base)]
	if !recognized {
		return x86FloatAddSubForm{}, false, nil
	}
	form := x86FloatAddSubForm{spec: spec, first: 0, second: 1, destination: len(ins.Args) - 1, mask: -1, memory: -1}
	fail := func() (x86FloatAddSubForm, bool, error) {
		return form, true, fmt.Errorf("%s %s operands/suffix outside Go floating add/sub grammar: %q", goarch, ins.Op, ins.Raw)
	}
	if goarch != "386" && goarch != "amd64" {
		return fail()
	}
	if spec.legacy {
		if dotted || len(ins.Args) != 2 {
			return fail()
		}
		form.byteWidth = 16
	} else {
		properties, valid := parseAMD64FMA3Suffix(suffix)
		if !valid || dotted && suffix == "" || spec.scalar && properties.broadcast || len(ins.Args) != 3 && len(ins.Args) != 4 {
			return fail()
		}
		form.broadcast, form.zeroing, form.rounding = properties.broadcast, properties.zeroing, properties.rounding
		if len(ins.Args) == 4 {
			if goarch == "386" && !ins.x86Encoded || !amd64NonzeroKOperand(ins.Args[2]) {
				return fail()
			}
			form.mask = 2
		} else if form.zeroing {
			return fail()
		}
		form.byteWidth = amd64VectorByteWidth(ins.Args[form.destination].Reg)
		if spec.scalar && form.byteWidth != 16 || form.rounding != "" && !spec.scalar && form.byteWidth != 64 {
			return fail()
		}
	}
	register := func(operand Operand) bool {
		if spec.legacy {
			index, valid := amd64ParseXReg(operand.Reg)
			return operand.Kind == OpReg && valid && (goarch == "amd64" && index < 16 || goarch == "386" && index < 8)
		}
		return amd64GoPackedVectorMoveRegister(operand, form.byteWidth, goarch)
	}
	if !register(ins.Args[form.second]) || !register(ins.Args[form.destination]) {
		return fail()
	}
	first := ins.Args[form.first]
	if isAMD64MemoryOperand(first) {
		if form.rounding != "" {
			return fail()
		}
		if first.Kind == OpSym {
			symbol := strings.TrimSpace(first.Sym)
			_, absolute := parseInt(symbol)
			if strings.HasPrefix(symbol, "$") || !strings.HasSuffix(symbol, "(SB)") && absolute != nil {
				return fail()
			}
		}
		if first.Kind == OpMem {
			displacement := first.Mem.Off
			if displacement != int64(int32(displacement)) && (goarch != "386" || displacement != int64(uint32(displacement))) ||
				!x86MemoryRegistersValidForArch(first.Mem, goarch) || first.Mem.Index != "" &&
				(first.Mem.Index == SP || first.Mem.Scale != 1 && first.Mem.Scale != 2 && first.Mem.Scale != 4 && first.Mem.Scale != 8) {
				return fail()
			}
		}
		form.memory, form.memoryBytes = 0, form.byteWidth
		if spec.scalar || form.broadcast {
			form.memoryBytes = spec.bits / 8
		}
	} else if form.broadcast || !register(first) {
		return fail()
	}
	return form, true, nil
}
