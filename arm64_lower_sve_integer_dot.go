package plan9asm

import (
	"fmt"
	"strings"
)

type arm64SVEIntegerDotSpec struct {
	intrinsic   string
	mixed       bool
	indexedOnly bool
}

var arm64SVEIntegerDotSpecs = map[Op]arm64SVEIntegerDotSpec{
	"ZSDOT":  {intrinsic: "sdot"},
	"ZUDOT":  {intrinsic: "udot"},
	"ZSUDOT": {intrinsic: "sudot", mixed: true, indexedOnly: true},
	"ZUSDOT": {intrinsic: "usdot", mixed: true},
}

type arm64SVEIntegerDotForm struct {
	intrinsic       string
	sourceBits      int
	destinationBits int
	first           int
	second          int
	destination     int
	lane            int
	indexed         bool
}

func arm64SVEIntegerDotFeature(ins Instr) string {
	if len(ins.Args) != 3 {
		return ""
	}
	_, sourceBits, ok := arm64ParseSVEZElementReg(ins.Args[0])
	if !ok {
		_, sourceBits, _, ok = arm64ParseSVEZIndexedElementRegUnbounded(ins.Args[0])
	}
	_, destinationBits, destinationOK := arm64ParseSVEZElementReg(ins.Args[2])
	if !ok || !destinationOK {
		return ""
	}
	if sourceBits == 8 && destinationBits == 16 {
		return "+sve2p3"
	}
	if sourceBits == 16 && destinationBits == 32 {
		return "+sve2p1"
	}
	return ""
}

func (c *arm64Ctx) lowerARM64SVEIntegerDot(op Op, ins Instr) (ok bool, terminated bool, err error) {
	spec, ok := arm64SVEIntegerDotSpecs[op]
	if !ok {
		return false, false, nil
	}
	if strings.ToUpper(string(ins.Op)) != string(op) || len(ins.Args) != 3 {
		return true, false, fmt.Errorf("arm64 %s expects one complete Go 1.27 integer-dot form without a suffix: %q", op, ins.Raw)
	}
	form := arm64SVEIntegerDotForm{intrinsic: spec.intrinsic}
	if first, firstBits, firstOK := arm64ParseSVEZElementReg(ins.Args[0]); firstOK {
		form.first = first
		form.sourceBits = firstBits
	} else {
		first, firstBits, lane, firstOK := arm64ParseSVEZIndexedElementRegUnbounded(ins.Args[0])
		if !firstOK {
			return true, false, fmt.Errorf("arm64 %s first source must be a B/H vector or indexed vector: %q", op, ins.Raw)
		}
		form.first = first
		form.sourceBits = firstBits
		form.lane = lane
		form.indexed = true
	}
	second, secondBits, secondOK := arm64ParseSVEZElementReg(ins.Args[1])
	destination, destinationBits, destinationOK := arm64ParseSVEZElementReg(ins.Args[2])
	if !secondOK || secondBits != form.sourceBits || !destinationOK || !arm64SVEIntegerDotWidthOK(spec, form.sourceBits, destinationBits) || spec.indexedOnly && !form.indexed {
		return true, false, fmt.Errorf("arm64 %s has an invalid source/destination width or indexing mode: %q", op, ins.Raw)
	}
	if form.indexed {
		maximumVector, maximumLane := arm64SVEIntegerDotIndexedLimits(form.sourceBits, destinationBits)
		if maximumVector < 0 || form.first > maximumVector || form.lane > maximumLane {
			return true, false, fmt.Errorf("arm64 %s indexed source is outside the Go 1.27 register/lane range: %q", op, ins.Raw)
		}
	}
	form.second = second
	form.destination = destination
	form.destinationBits = destinationBits
	return true, false, c.lowerARM64SVEIntegerDotForm(form)
}

func arm64SVEIntegerDotWidthOK(spec arm64SVEIntegerDotSpec, sourceBits, destinationBits int) bool {
	if spec.mixed {
		return sourceBits == 8 && destinationBits == 32
	}
	return sourceBits == 8 && (destinationBits == 16 || destinationBits == 32) ||
		sourceBits == 16 && (destinationBits == 32 || destinationBits == 64)
}

func arm64SVEIntegerDotIndexedLimits(sourceBits, destinationBits int) (maximumVector, maximumLane int) {
	switch {
	case sourceBits == 8 && destinationBits == 16:
		return 7, 7
	case sourceBits == 8 && destinationBits == 32:
		return 7, 3
	case sourceBits == 16 && destinationBits == 32:
		return 7, 3
	case sourceBits == 16 && destinationBits == 64:
		return 15, 1
	default:
		return -1, -1
	}
}

func (c *arm64Ctx) lowerARM64SVEIntegerDotForm(form arm64SVEIntegerDotForm) error {
	accumulator, destinationType, err := c.loadZRegElements(form.destination, form.destinationBits)
	if err != nil {
		return err
	}
	second, sourceType, err := c.loadZRegElements(form.second, form.sourceBits)
	if err != nil {
		return err
	}
	first, _, err := c.loadZRegElements(form.first, form.sourceBits)
	if err != nil {
		return err
	}
	if form.destinationBits == 16 {
		result := c.newTmp()
		assembly := form.intrinsic + " $0.h, $2.b, $3.b"
		if form.indexed {
			assembly += fmt.Sprintf("[%d]", form.lane)
		}
		fmt.Fprintf(c.b, "  %%%s = call %s asm sideeffect %q, %q(%s %s, %s %s, %s %s)\n", result, destinationType, assembly, "=&w,0,w,w", destinationType, accumulator, sourceType, second, sourceType, first)
		return c.storeZRegElements(form.destination, form.destinationBits, "%"+result)
	}
	lanes := 128 / form.destinationBits
	intrinsic := form.intrinsic
	if form.indexed {
		intrinsic += ".lane"
	}
	if form.destinationBits == form.sourceBits*2 {
		intrinsic += ".x2"
	}
	result := c.newTmp()
	if form.indexed {
		fmt.Fprintf(c.b, "  %%%s = call %s @llvm.aarch64.sve.%s.nxv%di%d(%s %s, %s %s, %s %s, i32 %d)\n", result, destinationType, intrinsic, lanes, form.destinationBits, destinationType, accumulator, sourceType, second, sourceType, first, form.lane)
	} else {
		fmt.Fprintf(c.b, "  %%%s = call %s @llvm.aarch64.sve.%s.nxv%di%d(%s %s, %s %s, %s %s)\n", result, destinationType, intrinsic, lanes, form.destinationBits, destinationType, accumulator, sourceType, second, sourceType, first)
	}
	return c.storeZRegElements(form.destination, form.destinationBits, "%"+result)
}
