package plan9asm

import (
	"fmt"
	"strings"
)

type arm64SVEAddSubSpec struct {
	intrinsic    string
	reverse      bool
	unpredicated bool
	predicated   bool
	immediate    bool
}

var arm64SVEAddSubSpecs = map[Op]arm64SVEAddSubSpec{
	"ZSUB":    {unpredicated: true, predicated: true, immediate: true},
	"ZSUBR":   {reverse: true, predicated: true, immediate: true},
	"ZSQADD":  {intrinsic: "aarch64.sve.sqadd.x", unpredicated: true, predicated: true, immediate: true},
	"ZSQSUB":  {intrinsic: "aarch64.sve.sqsub.x", unpredicated: true, predicated: true, immediate: true},
	"ZSQSUBR": {intrinsic: "aarch64.sve.sqsub.x", reverse: true, predicated: true},
	"ZUQADD":  {intrinsic: "aarch64.sve.uqadd.x", unpredicated: true, predicated: true, immediate: true},
	"ZUQSUB":  {intrinsic: "aarch64.sve.uqsub.x", unpredicated: true, predicated: true, immediate: true},
	"ZUQSUBR": {intrinsic: "aarch64.sve.uqsub.x", reverse: true, predicated: true},
}

// decodeARM64RawSVESubtract reuses the ADD field grammar after removing the
// fixed SUB/SUBR opcode bits. This covers every Go 1.27 SUB/SUBR form and
// inherits ADD's reserved-size and immediate checks.
func decodeARM64RawSVESubtract(word uint32) (Op, arm64RawSVEAdd, bool) {
	var op Op
	var normalized uint32
	switch {
	case word&0xff20fc00 == 0x04200400:
		op, normalized = "ZSUB", word^(1<<10)
	case word&0xff3fe000 == 0x04010000,
		word&0xff3fc000 == 0x2521c000:
		op, normalized = "ZSUB", word^(1<<16)
	case word&0xff3fe000 == 0x04030000,
		word&0xff3fc000 == 0x2523c000:
		op, normalized = "ZSUBR", word^(3<<16)
	default:
		return "", arm64RawSVEAdd{}, false
	}
	form, ok := decodeARM64RawSVEAdd(normalized)
	return op, form, ok
}

func (c *arm64Ctx) lowerARM64SVEAddSub(op Op, ins Instr) (ok bool, terminated bool, err error) {
	spec, ok := arm64SVEAddSubSpecs[op]
	if !ok {
		return false, false, nil
	}
	if strings.ToUpper(string(ins.Op)) != string(op) {
		return true, false, fmt.Errorf("arm64 %s does not accept an instruction suffix: %q", op, ins.Raw)
	}

	form := arm64RawSVEAdd{}
	switch {
	case len(ins.Args) == 3 && ins.Args[0].Kind == OpImm:
		first, firstBits, firstOK := arm64ParseSVEZElementReg(ins.Args[1])
		destination, destinationBits, destinationOK := arm64ParseSVEZElementReg(ins.Args[2])
		if !spec.immediate || ins.Args[0].ImmRaw != "" || !firstOK || !destinationOK || first != destination || firstBits != destinationBits || !arm64SVEAddImmediateRepresentable(firstBits, ins.Args[0].Imm) {
			return true, false, fmt.Errorf("arm64 %s does not accept this immediate form: %q", op, ins.Raw)
		}
		form = arm64RawSVEAdd{mode: arm64SVEAddImmediate, elementBits: firstBits, first: first, destination: destination, immediate: int(ins.Args[0].Imm)}
	case len(ins.Args) == 3:
		second, secondBits, secondOK := arm64ParseSVEZElementReg(ins.Args[0])
		first, firstBits, firstOK := arm64ParseSVEZElementReg(ins.Args[1])
		destination, destinationBits, destinationOK := arm64ParseSVEZElementReg(ins.Args[2])
		if !spec.unpredicated || !secondOK || !firstOK || !destinationOK || secondBits != firstBits || firstBits != destinationBits {
			return true, false, fmt.Errorf("arm64 %s does not accept this unpredicated vector form: %q", op, ins.Raw)
		}
		form = arm64RawSVEAdd{mode: arm64SVEAddUnpredicated, elementBits: firstBits, first: first, second: second, destination: destination}
	case len(ins.Args) == 4:
		second, secondBits, secondOK := arm64ParseSVEZElementReg(ins.Args[0])
		first, firstBits, firstOK := arm64ParseSVEZElementReg(ins.Args[1])
		predicate, predicateOK := arm64ParseSVEPredicateMerge(ins.Args[2])
		destination, destinationBits, destinationOK := arm64ParseSVEZElementReg(ins.Args[3])
		if !spec.predicated || !secondOK || !firstOK || !predicateOK || !destinationOK || secondBits != firstBits || firstBits != destinationBits || first != destination {
			return true, false, fmt.Errorf("arm64 %s does not accept this predicated destructive form: %q", op, ins.Raw)
		}
		form = arm64RawSVEAdd{mode: arm64SVEAddPredicated, elementBits: firstBits, first: first, second: second, predicate: predicate, destination: destination}
	default:
		return true, false, fmt.Errorf("arm64 %s expects one of its Go 1.27 SVE forms: %q", op, ins.Raw)
	}
	return true, false, c.lowerRawSVEAddSub(spec, form)
}

func (c *arm64Ctx) lowerRawSVEAddSub(spec arm64SVEAddSubSpec, form arm64RawSVEAdd) error {
	first, vectorType, err := c.loadZRegElements(form.first, form.elementBits)
	if err != nil {
		return err
	}
	second := fmt.Sprintf("splat (i%d %d)", form.elementBits, form.immediate)
	if form.mode != arm64SVEAddImmediate {
		second, _, err = c.loadZRegElements(form.second, form.elementBits)
		if err != nil {
			return err
		}
	}
	left, right := first, second
	if spec.reverse {
		left, right = right, left
	}
	calculated := c.newTmp()
	if spec.intrinsic == "" {
		fmt.Fprintf(c.b, "  %%%s = sub %s %s, %s\n", calculated, vectorType, left, right)
	} else {
		lanes := 128 / form.elementBits
		fmt.Fprintf(c.b, "  %%%s = call %s @llvm.%s.nxv%di%d(%s %s, %s %s)\n", calculated, vectorType, spec.intrinsic, lanes, form.elementBits, vectorType, left, vectorType, right)
	}
	result := "%" + calculated
	if form.mode == arm64SVEAddPredicated {
		predicate, predicateType, err := c.loadPRegElements(form.predicate, form.elementBits)
		if err != nil {
			return err
		}
		selected := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = select %s %s, %s %s, %s %s\n", selected, predicateType, predicate, vectorType, result, vectorType, first)
		result = "%" + selected
	}
	return c.storeZRegElements(form.destination, form.elementBits, result)
}
