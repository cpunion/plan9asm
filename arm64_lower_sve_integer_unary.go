package plan9asm

import (
	"fmt"
	"strings"
)

type arm64SVEIntegerUnarySpec struct {
	intrinsic string
	unpred    bool
	minBits   int
	maxBits   int
	sve2      bool
}

var arm64SVEIntegerUnarySpecs = map[Op]arm64SVEIntegerUnarySpec{
	"ZABS":     {intrinsic: "abs", minBits: 8},
	"ZCLS":     {intrinsic: "cls", minBits: 8},
	"ZCLZ":     {intrinsic: "clz", minBits: 8},
	"ZCNOT":    {intrinsic: "cnot", minBits: 8, sve2: true},
	"ZCNT":     {intrinsic: "cnt", minBits: 8},
	"ZNEG":     {intrinsic: "neg", minBits: 8},
	"ZNOT":     {intrinsic: "not", minBits: 8},
	"ZRBIT":    {intrinsic: "rbit", minBits: 8, sve2: true},
	"ZREV":     {intrinsic: "vector.reverse", unpred: true, minBits: 8},
	"ZREVB":    {intrinsic: "revb", minBits: 16, sve2: true},
	"ZREVH":    {intrinsic: "revh", minBits: 32, sve2: true},
	"ZREVW":    {intrinsic: "revw", minBits: 64, sve2: true},
	"ZSQABS":   {intrinsic: "sqabs", minBits: 8},
	"ZSQNEG":   {intrinsic: "sqneg", minBits: 8},
	"ZSXTB":    {intrinsic: "sxtb", minBits: 16},
	"ZSXTH":    {intrinsic: "sxth", minBits: 32},
	"ZSXTW":    {intrinsic: "sxtw", minBits: 64},
	"ZUXTB":    {intrinsic: "uxtb", minBits: 16},
	"ZUXTH":    {intrinsic: "uxth", minBits: 32},
	"ZUXTW":    {intrinsic: "uxtw", minBits: 64},
	"ZURECPE":  {intrinsic: "urecpe", minBits: 32, maxBits: 32},
	"ZURSQRTE": {intrinsic: "ursqrte", minBits: 32, maxBits: 32},
}

type arm64SVEIntegerUnaryForm struct {
	elementBits int
	source      int
	predicate   int
	destination int
	zeroing     bool
}

func (c *arm64Ctx) lowerARM64SVEIntegerUnary(op Op, ins Instr) (ok bool, terminated bool, err error) {
	spec, ok := arm64SVEIntegerUnarySpecs[op]
	if !ok {
		return false, false, nil
	}
	if strings.ToUpper(string(ins.Op)) != string(op) {
		return true, false, fmt.Errorf("arm64 %s does not accept an instruction suffix: %q", op, ins.Raw)
	}
	form := arm64SVEIntegerUnaryForm{}
	if spec.unpred {
		if len(ins.Args) != 2 {
			return true, false, fmt.Errorf("arm64 %s expects Zn.T, Zd.T: %q", op, ins.Raw)
		}
		source, sourceBits, sourceOK := arm64ParseSVEZElementReg(ins.Args[0])
		destination, destinationBits, destinationOK := arm64ParseSVEZElementReg(ins.Args[1])
		if !sourceOK || !destinationOK || sourceBits != destinationBits || sourceBits < spec.minBits || (spec.maxBits != 0 && sourceBits > spec.maxBits) {
			return true, false, fmt.Errorf("arm64 %s vector widths do not match its Go 1.27 form: %q", op, ins.Raw)
		}
		form = arm64SVEIntegerUnaryForm{elementBits: sourceBits, source: source, destination: destination}
	} else {
		if len(ins.Args) != 3 {
			return true, false, fmt.Errorf("arm64 %s expects Zn.T, Pg/M|Z, Zd.T: %q", op, ins.Raw)
		}
		source, sourceBits, sourceOK := arm64ParseSVEZElementReg(ins.Args[0])
		destination, destinationBits, destinationOK := arm64ParseSVEZElementReg(ins.Args[2])
		predicate, mergeOK := arm64ParseSVEPredicateMode(ins.Args[1], "M", 7)
		if !mergeOK {
			form.zeroing = true
			var zeroOK bool
			predicate, zeroOK = arm64ParseSVEPredicateMode(ins.Args[1], "Z", 7)
			if !zeroOK {
				return true, false, fmt.Errorf("arm64 %s requires P0..P7/M or P0..P7/Z: %q", op, ins.Raw)
			}
		}
		if !sourceOK || !destinationOK || sourceBits != destinationBits || sourceBits < spec.minBits || (spec.maxBits != 0 && sourceBits > spec.maxBits) {
			return true, false, fmt.Errorf("arm64 %s vector widths do not match its Go 1.27 form: %q", op, ins.Raw)
		}
		form.elementBits = sourceBits
		form.source = source
		form.predicate = predicate
		form.destination = destination
	}
	return true, false, c.lowerARM64SVEIntegerUnaryForm(spec, form)
}

func (c *arm64Ctx) lowerARM64SVEIntegerUnaryForm(spec arm64SVEIntegerUnarySpec, form arm64SVEIntegerUnaryForm) error {
	source, vectorType, err := c.loadZRegElements(form.source, form.elementBits)
	if err != nil {
		return err
	}
	lanes := 128 / form.elementBits
	result := c.newTmp()
	if spec.unpred {
		fmt.Fprintf(c.b, "  %%%s = call %s @llvm.%s.nxv%di%d(%s %s)\n", result, vectorType, spec.intrinsic, lanes, form.elementBits, vectorType, source)
	} else {
		predicate, predicateType, err := c.loadPRegElements(form.predicate, form.elementBits)
		if err != nil {
			return err
		}
		merge := source
		if !form.zeroing {
			merge, _, err = c.loadZRegElements(form.destination, form.elementBits)
			if err != nil {
				return err
			}
		}
		fmt.Fprintf(c.b, "  %%%s = call %s @llvm.aarch64.sve.%s.nxv%di%d(%s %s, %s %s, %s %s)\n", result, vectorType, spec.intrinsic, lanes, form.elementBits, vectorType, merge, predicateType, predicate, vectorType, source)
		if form.zeroing {
			mask := c.newTmp()
			zeroed := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = sext %s %s to %s\n", mask, predicateType, predicate, vectorType)
			fmt.Fprintf(c.b, "  %%%s = and %s %%%s, %%%s\n", zeroed, vectorType, result, mask)
			result = zeroed
		}
	}
	return c.storeZRegElements(form.destination, form.elementBits, "%"+result)
}
