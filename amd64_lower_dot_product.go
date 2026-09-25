package plan9asm

import (
	"fmt"
	"strings"
)

type amd64PackedDotProductSpec struct {
	inputBits int
	saturate  bool
}

// lowerPackedDotProduct implements the four VNNI dot-product opcodes sharing
// Go 1.27's _yvblendmpd table. Plan 9 lists the signed r/m source first and
// the encoded V source second; VPDPBUSD treats that second source as unsigned.
func (c *amd64Ctx) lowerPackedDotProduct(op Op, ins Instr) (ok bool, terminated bool, err error) {
	rawOp := strings.ToUpper(string(op))
	baseOp, suffix := rawOp, ""
	if dot := strings.IndexByte(rawOp, '.'); dot >= 0 {
		baseOp, suffix = rawOp[:dot], rawOp[dot+1:]
	}
	var spec amd64PackedDotProductSpec
	switch baseOp {
	case "VPDPBUSD":
		spec = amd64PackedDotProductSpec{inputBits: 8}
	case "VPDPBUSDS":
		spec = amd64PackedDotProductSpec{inputBits: 8, saturate: true}
	case "VPDPWSSD":
		spec = amd64PackedDotProductSpec{inputBits: 16}
	case "VPDPWSSDS":
		spec = amd64PackedDotProductSpec{inputBits: 16, saturate: true}
	default:
		return false, false, nil
	}

	properties, validSuffix := parseAMD64BinaryFloatingSuffix(suffix)
	if !validSuffix || properties.sae || properties.rounding != "" {
		return true, false, fmt.Errorf("%s %s suffix is absent from Go 1.27's _yvblendmpd encodings: %q", c.goarch, baseOp, ins.Raw)
	}
	if len(ins.Args) != 3 && len(ins.Args) != 4 {
		return true, false, fmt.Errorf("%s %s expects source2, source1, [K mask,] accumulator: %q", c.goarch, baseOp, ins.Raw)
	}
	masked := len(ins.Args) == 4
	if c.goarch == "386" && masked {
		return true, false, fmt.Errorf("386 %s mask forms exceed the Go assembler frontend's three-operand limit: %q", baseOp, ins.Raw)
	}
	if properties.zeroing && !masked {
		return true, false, fmt.Errorf("%s %s zeroing requires a K1-K7 mask: %q", c.goarch, baseOp, ins.Raw)
	}
	if masked && !amd64NonzeroKOperand(ins.Args[2]) {
		return true, false, fmt.Errorf("amd64 %s masked form expects K1-K7 as its third operand: %q", baseOp, ins.Raw)
	}

	destination := ins.Args[len(ins.Args)-1]
	if destination.Kind != OpReg {
		return true, false, fmt.Errorf("%s %s accumulator must be X, Y, or Z: %q", c.goarch, baseOp, ins.Raw)
	}
	byteWidth := amd64VectorByteWidth(destination.Reg)
	if byteWidth != 16 && byteWidth != 32 && byteWidth != 64 ||
		!c.isGoPackedVectorMoveRegister(destination, byteWidth) ||
		!c.isGoPackedVectorMoveRegister(ins.Args[1], byteWidth) {
		return true, false, fmt.Errorf("%s %s source1 and accumulator must be matching Go vector registers: %q", c.goarch, baseOp, ins.Raw)
	}
	rmSource := ins.Args[0]
	if properties.broadcast {
		if !isAMD64MemoryOperand(rmSource) {
			return true, false, fmt.Errorf("%s %s.BCST requires a scalar memory source2: %q", c.goarch, baseOp, ins.Raw)
		}
	} else if rmSource.Kind == OpReg {
		if !c.isGoPackedVectorMoveRegister(rmSource, byteWidth) {
			return true, false, fmt.Errorf("%s %s source2 register must match the accumulator width: %q", c.goarch, baseOp, ins.Raw)
		}
	} else if !isAMD64MemoryOperand(rmSource) {
		return true, false, fmt.Errorf("%s %s source2 must be a matching vector register or memory: %q", c.goarch, baseOp, ins.Raw)
	}

	dwordLanes := byteWidth / 4
	inputLanes := byteWidth * 8 / spec.inputBits
	var rmBytes string
	if properties.broadcast {
		scalar, err := c.evalIntSized(rmSource, I32)
		if err != nil {
			return true, false, err
		}
		dwords := amd64SplatInteger(c, dwordLanes, 32, scalar)
		cast := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i32> %s to <%d x i8>\n", cast, dwordLanes, dwords, byteWidth)
		rmBytes = "%" + cast
	} else {
		rmBytes, err = c.loadPackedCompareBytes(rmSource, byteWidth)
		if err != nil {
			return true, false, err
		}
	}
	vBytes, err := c.loadPackedCompareBytes(ins.Args[1], byteWidth)
	if err != nil {
		return true, false, err
	}
	oldBytes, err := c.loadPackedCompareBytes(destination, byteWidth)
	if err != nil {
		return true, false, err
	}
	rm := c.bitcastVectorBytesToIntegerLanes(byteWidth, inputLanes, spec.inputBits, rmBytes)
	v := c.bitcastVectorBytesToIntegerLanes(byteWidth, inputLanes, spec.inputBits, vBytes)
	old := c.bitcastVectorBytesToIntegerLanes(byteWidth, dwordLanes, 32, oldBytes)
	computed := c.emitPackedDotProduct(spec, dwordLanes, rm, v, old)
	if masked {
		mask, err := c.loadK(ins.Args[2].Reg)
		if err != nil {
			return true, false, err
		}
		computed = amd64ApplyI32LaneMask(c, dwordLanes, computed, old, mask, properties.zeroing)
	}
	out := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i32> %s to <%d x i8>\n", out, dwordLanes, computed, byteWidth)
	return true, false, c.storeVectorBytes(destination.Reg, byteWidth, "%"+out)
}

func (c *amd64Ctx) emitPackedDotProduct(spec amd64PackedDotProductSpec, dwordLanes int, rm, v, accumulator string) string {
	inputsPerDword := 32 / spec.inputBits
	result := "poison"
	for lane := 0; lane < dwordLanes; lane++ {
		accumulatorLane := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = extractelement <%d x i32> %s, i32 %d\n", accumulatorLane, dwordLanes, accumulator, lane)
		wideAccumulator := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = sext i32 %%%s to i64\n", wideAccumulator, accumulatorLane)
		total := "%" + wideAccumulator
		for item := 0; item < inputsPerDword; item++ {
			inputLane := lane*inputsPerDword + item
			rmLane := c.newTmp()
			vLane := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = extractelement <%d x i%d> %s, i32 %d\n", rmLane, dwordLanes*inputsPerDword, spec.inputBits, rm, inputLane)
			fmt.Fprintf(c.b, "  %%%s = extractelement <%d x i%d> %s, i32 %d\n", vLane, dwordLanes*inputsPerDword, spec.inputBits, v, inputLane)
			wideRM := c.newTmp()
			wideV := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = sext i%d %%%s to i64\n", wideRM, spec.inputBits, rmLane)
			if spec.inputBits == 8 {
				fmt.Fprintf(c.b, "  %%%s = zext i8 %%%s to i64\n", wideV, vLane)
			} else {
				fmt.Fprintf(c.b, "  %%%s = sext i16 %%%s to i64\n", wideV, vLane)
			}
			product := c.newTmp()
			next := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = mul i64 %%%s, %%%s\n", product, wideRM, wideV)
			fmt.Fprintf(c.b, "  %%%s = add i64 %s, %%%s\n", next, total, product)
			total = "%" + next
		}
		if spec.saturate {
			above := c.newTmp()
			below := c.newTmp()
			capped := c.newTmp()
			bounded := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = icmp sgt i64 %s, 2147483647\n", above, total)
			fmt.Fprintf(c.b, "  %%%s = select i1 %%%s, i64 2147483647, i64 %s\n", capped, above, total)
			fmt.Fprintf(c.b, "  %%%s = icmp slt i64 %%%s, -2147483648\n", below, capped)
			fmt.Fprintf(c.b, "  %%%s = select i1 %%%s, i64 -2147483648, i64 %%%s\n", bounded, below, capped)
			total = "%" + bounded
		}
		narrow := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = trunc i64 %s to i32\n", narrow, total)
		inserted := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = insertelement <%d x i32> %s, i32 %%%s, i32 %d\n", inserted, dwordLanes, result, narrow, lane)
		result = "%" + inserted
	}
	return result
}
