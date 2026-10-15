package plan9asm

import (
	"fmt"
	"strings"
)

type x86X87ValueKind uint8

const (
	x86X87FloatMove x86X87ValueKind = iota
	x86X87Constant
	x86X87Unary
	x86X87Exchange
	x86X87Arithmetic
	x86X87Compare
	x86X87Test
	x86X87Status
)

type x86X87ValueSpec struct {
	kind        x86X87ValueKind
	width       byte // W/L signed integer memory, F/D floating memory.
	stem        string
	pop         uint8
	reverse     bool
	directFlags bool
}

// These are value-availability forms, not a physical x87 entry/control-word,
// native-memory or caller-state contract. The Go rows are shared by both x86
// modes. Environment, BCD/extended/integer moves and transcendental state stay
// outside this independent logical family.
var x86X87ValueSpecs = func() map[Op]x86X87ValueSpec {
	forms := map[Op]x86X87ValueSpec{
		"FMOVD":  {kind: x86X87FloatMove, width: 'D'},
		"FMOVDP": {kind: x86X87FloatMove, width: 'D', pop: 1},
		"FMOVF":  {kind: x86X87FloatMove, width: 'F'},
		"FMOVFP": {kind: x86X87FloatMove, width: 'F', pop: 1},
		"FXCHD":  {kind: x86X87Exchange},
		"FTST":   {kind: x86X87Test},
		"FSTSW":  {kind: x86X87Status},
	}
	for _, op := range []Op{"FLD1", "FLDL2E", "FLDL2T", "FLDLG2", "FLDLN2", "FLDPI", "FLDZ"} {
		forms[op] = x86X87ValueSpec{kind: x86X87Constant}
	}
	for _, op := range []Op{"FABS", "FCHS", "FSQRT"} {
		forms[op] = x86X87ValueSpec{kind: x86X87Unary}
	}
	for _, stem := range []string{"FADD", "FMUL", "FSUB", "FSUBR", "FDIV", "FDIVR"} {
		for _, suffix := range []string{"W", "L", "F", "D", "DP"} {
			spec := x86X87ValueSpec{kind: x86X87Arithmetic, width: suffix[0], stem: stem,
				reverse: stem == "FSUBR" || stem == "FDIVR"}
			if suffix == "DP" {
				spec.pop = 1
			}
			forms[Op(stem+suffix)] = spec
		}
	}
	for _, row := range []struct {
		op    Op
		width byte
		pop   uint8
	}{
		{"FCOMW", 'W', 0}, {"FCOMWP", 'W', 1},
		{"FCOML", 'L', 0}, {"FCOMLP", 'L', 1},
		{"FCOMF", 'F', 0}, {"FCOMFP", 'F', 1},
		{"FCOMD", 'D', 0}, {"FCOMDP", 'D', 1}, {"FCOMDPP", 'D', 2},
		{"FCOMI", 'D', 0}, {"FCOMIP", 'D', 1},
		{"FUCOM", 'D', 0}, {"FUCOMP", 'D', 1}, {"FUCOMPP", 'D', 2},
		{"FUCOMI", 'D', 0}, {"FUCOMIP", 'D', 1},
	} {
		forms[row.op] = x86X87ValueSpec{kind: x86X87Compare, width: row.width, pop: row.pop,
			directFlags: row.op == "FCOMI" || row.op == "FCOMIP" || row.op == "FUCOMI" || row.op == "FUCOMIP"}
	}
	return forms
}()

type x86X87ValueForm struct {
	spec                x86X87ValueSpec
	source, destination Operand
	left, right         Operand // Ordered comparison dependencies, independent of spelling.
	push                bool
	unprovedEncoding    bool // Go's documented botch row is not a NUM/effect witness.
}

func (form x86X87ValueForm) memoryBytes() int64 {
	switch form.spec.width {
	case 'W':
		return 2
	case 'L', 'F':
		return 4
	case 'D':
		return 8
	}
	return 0
}

// One actual instruction head selects the complete Go form. Source FLOAT
// rewrites are validated separately from the underlying memory encoding; a
// resolved integer immediate is not an x87 memory value.
func x86X87ValueProperties(ins Instr, arch string) (x86X87ValueForm, bool, error) {
	name, suffix, dotted := strings.Cut(strings.ToUpper(string(ins.Op)), ".")
	spec, known := x86X87ValueSpecs[Op(name)]
	if !known {
		return x86X87ValueForm{}, false, nil
	}
	form := x86X87ValueForm{spec: spec}
	fail := func(reason string) (x86X87ValueForm, bool, error) {
		return form, true, fmt.Errorf("%s %s %s: %q", arch, ins.Op, reason, ins.Raw)
	}
	if arch != "386" && arch != "amd64" || dotted || suffix != "" {
		return fail("is outside the Go x87 value mode/suffix grammar")
	}
	for _, operand := range ins.Args {
		// Ym's scalar address grammar is independent of MUL's numerical
		// effects or conditional native-read role.
		if err := x86NamedMultiplyMemoryError(operand, arch); err != nil {
			return fail(err.Error())
		}
	}
	memory := func(operand Operand) bool {
		if !isX87MemoryOperand(operand) {
			return false
		}
		if operand.Kind == OpSym {
			symbol := strings.TrimSpace(operand.Sym)
			_, integer := parseInt(symbol)
			return !strings.HasPrefix(symbol, "$") && (strings.HasSuffix(symbol, "(SB)") || integer == nil)
		}
		return true
	}
	constant := func(operand Operand) bool {
		return operand.Kind == OpImm && (spec.width == 'F' || spec.width == 'D') &&
			x86FloatArithmeticLiteral(ins, true, true)
	}
	register := func(operand Operand) bool {
		_, valid := x87OperandReg(operand)
		return valid
	}
	if spec.kind == x86X87Constant || spec.kind == x86X87Unary || spec.kind == x86X87Test {
		if len(ins.Args) != 0 {
			return fail("takes no operands")
		}
		return form, true, nil
	}
	if spec.kind == x86X87Status {
		if len(ins.Args) != 1 || !(memory(ins.Args[0]) || ins.Args[0].Kind == OpReg && ins.Args[0].Reg == AX) {
			return fail("expects memory destination or AX")
		}
		form.destination = ins.Args[0]
		return form, true, nil
	}
	if len(ins.Args) != 2 {
		switch spec.kind {
		case x86X87Compare:
			return fail("expects lhs, rhs")
		case x86X87Exchange:
			return fail("expects Fsrc, Fdst")
		case x86X87FloatMove:
			if spec.pop != 0 {
				return fail("expects F0 source and memory or x87-register destination")
			}
		}
		return fail("expects src, dst")
	}
	first, last := ins.Args[0], ins.Args[1]
	form.source, form.destination = first, last
	switch spec.kind {
	case x86X87FloatMove:
		if spec.pop != 0 {
			if !isX87F0(first) || !(memory(last) || spec.width == 'D' && register(last)) {
				return fail("expects F0 source and memory or x87-register destination")
			}
			return form, true, nil
		}
		if isX87F0(last) && (memory(first) || constant(first) || spec.width == 'D' && register(first)) {
			// yfmvd's Yrf,Yf0 row is FLD ST(i), before its overlapping
			// Yf0,Yrf store row. F0,F0 therefore pushes as well.
			form.push = true
			return form, true, nil
		}
		if isX87F0(first) && (memory(last) || spec.width == 'D' && register(last)) {
			return form, true, nil
		}
		return fail("expects memory,F0, Freg,F0, F0,memory, or F0,Freg")
	case x86X87Exchange:
		if !register(first) || !register(last) || !isX87F0(first) && !isX87F0(last) {
			return fail("expects x87 registers with one F0 operand")
		}
	case x86X87Arithmetic:
		if spec.pop != 0 {
			if !isX87F0(first) || !register(last) {
				return fail("expects F0 source and x87-register destination")
			}
		} else if spec.width == 'D' {
			if !(isX87F0(last) && (memory(first) || constant(first) || register(first)) || isX87F0(first) && register(last)) {
				return fail("expects memory,F0, Freg,F0, or F0,Freg")
			}
		} else if !isX87F0(last) || !(memory(first) || spec.width == 'F' && constant(first)) {
			return fail("expects memory source and F0 destination")
		}
	case x86X87Compare:
		switch {
		case name == "FCOMI" || name == "FCOMIP":
			if !register(first) || !isX87F0(last) {
				return fail("expects x87-register source and F0 destination")
			}
			form.left, form.right = last, first
		case strings.HasPrefix(name, "FUCOM") || name == "FCOMDPP":
			if !isX87F0(first) || !register(last) {
				return fail("expects F0 source and x87-register destination")
			}
			form.left, form.right = first, last
		default:
			if isX87F0(last) && (memory(first) || constant(first) || spec.width == 'D' && register(first)) {
				form.left, form.right = last, first
			} else if spec.width == 'D' && isX87F0(first) && register(last) {
				// Retain Go's complete reverse FCOMD/DP grammar without
				// interpreting its undocumented DC /2,/3 register encoding
				// as a new logical compare proof.
				form.left, form.right, form.unprovedEncoding = last, first, true
			} else {
				return fail("expects memory source or x87-register source and F0 destination")
			}
		}
		if spec.pop == 2 && form.right.Reg != "F1" {
			// ycompp's Go table explicitly notes that its two-pop row is
			// really F0,F1. Recognizing another encoded row is not a proof
			// that the legacy mathematical spelling matches that opcode.
			form.unprovedEncoding = true
		}
	}
	return form, true, nil
}
