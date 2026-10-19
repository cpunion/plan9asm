package plan9asm

import (
	"fmt"
	"strings"
)

func (c *amd64Ctx) lowerFloatMinMax(op Op, ins Instr) (bool, bool, error) {
	base, _, _ := strings.Cut(strings.ToUpper(string(op)), ".")
	if _, known := x86FloatMinMaxSpecs[Op(base)]; !known {
		return false, false, nil
	}
	if ins.Op != "" && !strings.EqualFold(string(ins.Op), string(op)) {
		return true, false, fmt.Errorf("floating MIN/MAX dispatch/source opcode mismatch: %s/%s", op, ins.Op)
	}
	ins.Op = op
	f, known, err := x86FloatMinMaxProperties(ins, c.goarch)
	if err != nil || !known {
		return known, false, err
	}
	if (ins.x86Encoded || ins.x86SourceRaw != nil) && !x86RawFloatMinMaxEncodingSource(ins, c.goarch) {
		return true, false, fmt.Errorf("%w: MIN/MAX original physical source identity changed", ErrProbeNeedsContext)
	}
	if x86ScalarFloatIntegerNativeTarget(c.goarch, c.targetTriple) {
		if f.legacy && !f.scalar && f.memory >= 0 {
			return true, false, c.lowerLegacyPackedFloatingMemory(ins, f.x86FloatBinaryForm)
		}
		lanes := f.byteWidth * 8 / f.bits
		logicalFPBits := 0
		if f.scalar {
			lanes = 1
			logicalFPBits = f.bits
		}
		evex := x86FloatMinMaxNativeEVEX(ins, f)
		return true, false, c.emitNativeFloatingOperands(ins, x86NativeFloatingOperands{
			byteWidth: f.byteWidth, lanes: lanes, memoryBytes: f.memoryBytes,
			second: f.second, destination: f.destination, mask: f.mask,
			controls: amd64FMA3Suffix{broadcast: f.broadcast, zeroing: f.zeroing},
			sae:      f.sae, legacy: f.legacy, logicalFPBits: logicalFPBits,
			lowBank: !evex, forceEVEX: evex && f.byteWidth < 64 && f.mask < 0 && !f.sae && !f.broadcast,
		})
	}
	// Portable numerical values do not certify x86 MXCSR or native exceptions.
	return c.lowerFloatBinaryOperands(ins, f.x86FloatBinaryForm,
		func(lanes int, left, right, _ string) string {
			typ := amd64FMA3LLVMType(lanes, f.bits)
			predicate := "ogt"
			if f.spec.minimum {
				predicate = "olt"
			}
			condition, value := c.newTmp(), c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = fcmp %s %s %s, %s\n", condition, predicate, typ, left, right)
			predicateType := "i1"
			if lanes > 1 {
				predicateType = fmt.Sprintf("<%d x i1>", lanes)
			}
			// Native MIN/MAX chooses the first explicit Go source for equal
			// values and unordered pairs. Keep its original NaN/zero bits.
			fmt.Fprintf(c.b, "  %%%s = select %s %%%s, %s %s, %s %s\n", value, predicateType, condition, typ, left, typ, right)
			return "%" + value
		})
}
