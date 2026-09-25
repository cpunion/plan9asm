package plan9asm

import (
	"fmt"
	"strings"
)

type arm64SVEIntegerReductionSpec struct {
	intrinsic   string
	elementBits int
	quad        bool
}

var arm64SVEIntegerReductionSpecs = map[Op]arm64SVEIntegerReductionSpec{
	"ZADDQV": {intrinsic: "addqv", quad: true},
	"ZANDQV": {intrinsic: "andqv", quad: true},
	"ZANDVB": {intrinsic: "andv", elementBits: 8},
	"ZANDVH": {intrinsic: "andv", elementBits: 16},
	"ZANDVS": {intrinsic: "andv", elementBits: 32},
	"ZANDVD": {intrinsic: "andv", elementBits: 64},
	"ZEORQV": {intrinsic: "eorqv", quad: true},
	"ZEORVB": {intrinsic: "eorv", elementBits: 8},
	"ZEORVH": {intrinsic: "eorv", elementBits: 16},
	"ZEORVS": {intrinsic: "eorv", elementBits: 32},
	"ZEORVD": {intrinsic: "eorv", elementBits: 64},
	"ZORQV":  {intrinsic: "orqv", quad: true},
	"ZORVB":  {intrinsic: "orv", elementBits: 8},
	"ZORVH":  {intrinsic: "orv", elementBits: 16},
	"ZORVS":  {intrinsic: "orv", elementBits: 32},
	"ZORVD":  {intrinsic: "orv", elementBits: 64},
}

type arm64SVEIntegerReductionForm struct {
	elementBits int
	source      int
	predicate   int
	destination int
}

func (c *arm64Ctx) lowerARM64SVEIntegerReduction(op Op, ins Instr) (ok bool, terminated bool, err error) {
	spec, ok := arm64SVEIntegerReductionSpecs[op]
	if !ok {
		return false, false, nil
	}
	if strings.ToUpper(string(ins.Op)) != string(op) {
		return true, false, fmt.Errorf("arm64 %s does not accept an instruction suffix: %q", op, ins.Raw)
	}
	form, err := arm64ParseSVEIntegerReductionForm(op, spec, ins)
	if err != nil {
		return true, false, err
	}
	return true, false, c.lowerARM64SVEIntegerReductionForm(spec, form, ins.Args[2].Reg)
}

func arm64ParseSVEIntegerReductionForm(op Op, spec arm64SVEIntegerReductionSpec, ins Instr) (arm64SVEIntegerReductionForm, error) {
	form := arm64SVEIntegerReductionForm{}
	if len(ins.Args) != 3 {
		return form, fmt.Errorf("arm64 %s expects Zn.T, Pg, Vd%s: %q", op, map[bool]string{true: ".T", false: ""}[spec.quad], ins.Raw)
	}
	source, elementBits, sourceOK := arm64ParseSVEZElementReg(ins.Args[0])
	predicate, predicateOK := arm64ParseSVEPredicate(ins.Args[1])
	destination, destinationOK := arm64ParseVReg(ins.Args[2].Reg)
	if !sourceOK || !predicateOK || predicate > 7 || !destinationOK || ins.Args[2].Kind != OpReg {
		return form, fmt.Errorf("arm64 %s requires a scalable source, bare predicate, and SIMD destination: %q", op, ins.Raw)
	}
	if !spec.quad {
		if strings.Contains(string(ins.Args[2].Reg), ".") {
			return form, fmt.Errorf("arm64 %s destination must be a bare V register: %q", op, ins.Raw)
		}
		// Go's generic Zn.T encoder ORs the operand size bits into the
		// mnemonic's fixed size bits, rather than requiring them to match.
		opcodeSize := map[int]int{8: 0, 16: 1, 32: 2, 64: 3}[spec.elementBits]
		sourceSize := map[int]int{8: 0, 16: 1, 32: 2, 64: 3}[elementBits]
		elementBits = 8 << (opcodeSize | sourceSize)
	} else {
		arrangement, arrangementOK := parseARM64VectorArrangement(ins.Args[2].Reg)
		if !arrangementOK || arrangement.elementBits != elementBits || arrangement.lanes*arrangement.elementBits != 128 {
			return form, fmt.Errorf("arm64 %s destination must be a full-width SIMD vector matching its source elements: %q", op, ins.Raw)
		}
	}
	return arm64SVEIntegerReductionForm{elementBits: elementBits, source: source, predicate: predicate, destination: destination}, nil
}

func (c *arm64Ctx) lowerARM64SVEIntegerReductionForm(spec arm64SVEIntegerReductionSpec, form arm64SVEIntegerReductionForm, destination Reg) error {
	source, vectorType, err := c.loadZRegElements(form.source, form.elementBits)
	if err != nil {
		return err
	}
	predicate, predicateType, err := c.loadPRegElements(form.predicate, form.elementBits)
	if err != nil {
		return err
	}
	lanes := 128 / form.elementBits
	result := c.newTmp()
	if !spec.quad {
		fmt.Fprintf(c.b, "  %%%s = call i%d @llvm.aarch64.sve.%s.nxv%di%d(%s %s, %s %s)\n", result, form.elementBits, spec.intrinsic, lanes, form.elementBits, predicateType, predicate, vectorType, source)
		return c.storeARM64ScalarToVReg(destination, form.elementBits, "%"+result)
	}
	fmt.Fprintf(c.b, "  %%%s = call <%d x i%d> @llvm.aarch64.sve.%s.v%di%d.nxv%di%d(%s %s, %s %s)\n", result, lanes, form.elementBits, spec.intrinsic, lanes, form.elementBits, lanes, form.elementBits, predicateType, predicate, vectorType, source)
	bytes := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i%d> %%%s to <16 x i8>\n", bytes, lanes, form.elementBits, result)
	return c.storeVReg(destination, "%"+bytes)
}
