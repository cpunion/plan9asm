package plan9asm

import (
	"fmt"
	"strings"
)

type amd64PackedAddMode uint8

const (
	amd64PackedAddWrap amd64PackedAddMode = iota
	amd64PackedAddSignedSaturating
	amd64PackedAddUnsignedSaturating
)

// lowerPackedIntegerAdd implements the complete Go 1.27 packed-integer add
// family: the legacy ymm/yxm tables and the _yvandnpd VEX/EVEX table.
func (c *amd64Ctx) lowerPackedIntegerAdd(op Op, ins Instr) (ok bool, terminated bool, err error) {
	rawOp := strings.ToUpper(string(op))
	baseOp := rawOp
	suffix := ""
	if dot := strings.IndexByte(rawOp, '.'); dot >= 0 {
		baseOp, suffix = rawOp[:dot], rawOp[dot+1:]
	}
	laneBits, mode, vector, recognized := amd64PackedAddProperties(baseOp)
	if !recognized {
		return false, false, nil
	}
	if !vector {
		if suffix != "" {
			return true, false, fmt.Errorf("amd64 %s does not accept instruction suffixes: %q", baseOp, ins.Raw)
		}
		return c.lowerLegacyPackedIntegerAdd(baseOp, laneBits, mode, ins)
	}
	return c.lowerVectorPackedIntegerAdd(baseOp, suffix, laneBits, mode, ins)
}

func amd64PackedAddProperties(op string) (laneBits int, mode amd64PackedAddMode, vector, ok bool) {
	switch op {
	case "PADDB":
		return 8, amd64PackedAddWrap, false, true
	case "PADDW":
		return 16, amd64PackedAddWrap, false, true
	case "PADDL":
		return 32, amd64PackedAddWrap, false, true
	case "PADDQ":
		return 64, amd64PackedAddWrap, false, true
	case "PADDSB":
		return 8, amd64PackedAddSignedSaturating, false, true
	case "PADDSW":
		return 16, amd64PackedAddSignedSaturating, false, true
	case "PADDUSB":
		return 8, amd64PackedAddUnsignedSaturating, false, true
	case "PADDUSW":
		return 16, amd64PackedAddUnsignedSaturating, false, true
	case "VPADDB":
		return 8, amd64PackedAddWrap, true, true
	case "VPADDW":
		return 16, amd64PackedAddWrap, true, true
	case "VPADDD":
		return 32, amd64PackedAddWrap, true, true
	case "VPADDQ":
		return 64, amd64PackedAddWrap, true, true
	case "VPADDSB":
		return 8, amd64PackedAddSignedSaturating, true, true
	case "VPADDSW":
		return 16, amd64PackedAddSignedSaturating, true, true
	case "VPADDUSB":
		return 8, amd64PackedAddUnsignedSaturating, true, true
	case "VPADDUSW":
		return 16, amd64PackedAddUnsignedSaturating, true, true
	default:
		return 0, 0, false, false
	}
}

func (c *amd64Ctx) lowerLegacyPackedIntegerAdd(baseOp string, laneBits int, mode amd64PackedAddMode, ins Instr) (bool, bool, error) {
	if len(ins.Args) != 2 || ins.Args[1].Kind != OpReg {
		return true, false, fmt.Errorf("amd64 %s expects MMX/m64, MMX or X/m128, X: %q", baseOp, ins.Raw)
	}
	dst := ins.Args[1].Reg
	if _, ok := amd64ParseMReg(dst); ok {
		if baseOp == "PADDQ" {
			return true, false, fmt.Errorf("amd64 PADDQ has no MMX form in Go 1.27's yxm table: %q", ins.Raw)
		}
		firstBits, err := c.loadLegacyMMXPackedSource(ins.Args[0])
		if err != nil {
			return true, false, fmt.Errorf("amd64 %s MMX source: %w", baseOp, err)
		}
		secondBits, err := c.loadReg(dst)
		if err != nil {
			return true, false, err
		}
		lanes := 64 / laneBits
		first := c.bitcastI64ToIntegerLanes(lanes, laneBits, firstBits)
		second := c.bitcastI64ToIntegerLanes(lanes, laneBits, secondBits)
		result := c.emitPackedIntegerAdd(lanes, laneBits, first, second, mode)
		bits := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i%d> %s to i64\n", bits, lanes, laneBits, result)
		return true, false, c.storeReg(dst, "%"+bits)
	}
	if !isAMD64XReg(dst) {
		return true, false, fmt.Errorf("amd64 %s destination must be MMX or X: %q", baseOp, ins.Raw)
	}
	if ins.Args[0].Kind == OpReg && !isAMD64XReg(ins.Args[0].Reg) {
		return true, false, fmt.Errorf("amd64 %s XMM form requires an X source: %q", baseOp, ins.Raw)
	}
	firstBytes, err := c.loadXVecOperand(ins.Args[0])
	if err != nil {
		return true, false, err
	}
	secondBytes, err := c.loadX(dst)
	if err != nil {
		return true, false, err
	}
	lanes := 128 / laneBits
	first := c.bitcastVectorBytesToIntegerLanes(16, lanes, laneBits, firstBytes)
	second := c.bitcastVectorBytesToIntegerLanes(16, lanes, laneBits, secondBytes)
	result := c.emitPackedIntegerAdd(lanes, laneBits, first, second, mode)
	out := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i%d> %s to <16 x i8>\n", out, lanes, laneBits, result)
	return true, false, c.storeX(dst, "%"+out)
}

func (c *amd64Ctx) loadLegacyMMXPackedSource(src Operand) (string, error) {
	if src.Kind == OpReg {
		if _, ok := amd64ParseMReg(src.Reg); !ok {
			return "", fmt.Errorf("expected MMX register or memory, got %s", src.String())
		}
		return c.loadReg(src.Reg)
	}
	if !isAMD64MemoryOperand(src) {
		return "", fmt.Errorf("expected MMX register or memory, got %s", src.String())
	}
	return c.evalIntSized(src, I64)
}

func (c *amd64Ctx) bitcastI64ToIntegerLanes(lanes, laneBits int, value string) string {
	cast := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast i64 %s to <%d x i%d>\n", cast, value, lanes, laneBits)
	return "%" + cast
}

func (c *amd64Ctx) lowerVectorPackedIntegerAdd(baseOp, suffix string, laneBits int, mode amd64PackedAddMode, ins Instr) (bool, bool, error) {
	broadcast, zeroing := false, false
	switch suffix {
	case "":
	case "Z":
		zeroing = true
	case "BCST":
		broadcast = true
	case "BCST.Z":
		broadcast, zeroing = true, true
	default:
		return true, false, fmt.Errorf("amd64 %s has a suffix absent from its Go 1.27 optab: %q", baseOp, ins.Raw)
	}
	if broadcast && baseOp != "VPADDD" && baseOp != "VPADDQ" {
		return true, false, fmt.Errorf("amd64 %s does not enable EVEX broadcast: %q", baseOp, ins.Raw)
	}
	if len(ins.Args) != 3 && len(ins.Args) != 4 {
		return true, false, fmt.Errorf("amd64 %s expects src1, src2, [K mask,] destination: %q", baseOp, ins.Raw)
	}
	dstArg := ins.Args[len(ins.Args)-1]
	if dstArg.Kind != OpReg {
		return true, false, fmt.Errorf("amd64 %s expects an X, Y, or Z destination: %q", baseOp, ins.Raw)
	}
	byteWidth := amd64VectorByteWidth(dstArg.Reg)
	if byteWidth == 0 {
		return true, false, fmt.Errorf("amd64 %s expects an X, Y, or Z destination: %q", baseOp, ins.Raw)
	}
	if !amd64VectorRegisterHasWidth(ins.Args[1], byteWidth) {
		return true, false, fmt.Errorf("amd64 %s second source must match its destination width: %q", baseOp, ins.Raw)
	}
	if broadcast {
		if !isAMD64MemoryOperand(ins.Args[0]) {
			return true, false, fmt.Errorf("amd64 %s.BCST requires a memory first source: %q", baseOp, ins.Raw)
		}
	} else if ins.Args[0].Kind == OpReg && !amd64VectorRegisterHasWidth(ins.Args[0], byteWidth) {
		return true, false, fmt.Errorf("amd64 %s first source must match its destination width: %q", baseOp, ins.Raw)
	}

	masked := len(ins.Args) == 4
	if zeroing && !masked {
		return true, false, fmt.Errorf("amd64 %s zeroing requires a K1-K7 mask: %q", baseOp, ins.Raw)
	}
	var mask string
	if masked {
		maskArg := ins.Args[2]
		if maskArg.Kind != OpReg {
			return true, false, fmt.Errorf("amd64 %s masked form expects K1-K7: %q", baseOp, ins.Raw)
		}
		maskIndex, ok := amd64ParseKReg(maskArg.Reg)
		if !ok || maskIndex == 0 {
			return true, false, fmt.Errorf("amd64 %s masked form expects K1-K7: %q", baseOp, ins.Raw)
		}
		var err error
		mask, err = c.loadK(maskArg.Reg)
		if err != nil {
			return true, false, err
		}
	}

	lanes := byteWidth * 8 / laneBits
	var first string
	var err error
	if broadcast {
		first, err = c.evalIntSized(ins.Args[0], amd64IntegerTypeForBits(laneBits))
		if err == nil {
			first = amd64SplatInteger(c, lanes, laneBits, first)
		}
	} else {
		var firstBytes string
		firstBytes, err = c.loadPackedCompareBytes(ins.Args[0], byteWidth)
		if err == nil {
			first = c.bitcastVectorBytesToIntegerLanes(byteWidth, lanes, laneBits, firstBytes)
		}
	}
	if err != nil {
		return true, false, err
	}
	secondBytes, err := c.loadPackedCompareBytes(ins.Args[1], byteWidth)
	if err != nil {
		return true, false, err
	}
	second := c.bitcastVectorBytesToIntegerLanes(byteWidth, lanes, laneBits, secondBytes)
	result := c.emitPackedIntegerAdd(lanes, laneBits, first, second, mode)
	if masked {
		oldBytes, err := c.loadPackedCompareBytes(dstArg, byteWidth)
		if err != nil {
			return true, false, err
		}
		old := c.bitcastVectorBytesToIntegerLanes(byteWidth, lanes, laneBits, oldBytes)
		result = amd64ApplyIntegerLaneMask(c, lanes, laneBits, result, old, mask, zeroing)
	}
	out := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i%d> %s to <%d x i8>\n", out, lanes, laneBits, result, byteWidth)
	return true, false, c.storeVectorBytes(dstArg.Reg, byteWidth, "%"+out)
}

func (c *amd64Ctx) emitPackedIntegerAdd(lanes, laneBits int, first, second string, mode amd64PackedAddMode) string {
	vecType := fmt.Sprintf("<%d x i%d>", lanes, laneBits)
	if mode == amd64PackedAddWrap {
		result := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = add %s %s, %s\n", result, vecType, second, first)
		return "%" + result
	}

	wideBits := laneBits * 2
	wideType := fmt.Sprintf("<%d x i%d>", lanes, wideBits)
	firstWide := c.newTmp()
	secondWide := c.newTmp()
	extend := "sext"
	if mode == amd64PackedAddUnsignedSaturating {
		extend = "zext"
	}
	fmt.Fprintf(c.b, "  %%%s = %s %s %s to %s\n", firstWide, extend, vecType, first, wideType)
	fmt.Fprintf(c.b, "  %%%s = %s %s %s to %s\n", secondWide, extend, vecType, second, wideType)
	sum := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = add %s %%%s, %%%s\n", sum, wideType, secondWide, firstWide)

	max := int64((uint64(1) << laneBits) - 1)
	predicate := "ugt"
	if mode == amd64PackedAddSignedSaturating {
		max = int64((uint64(1) << (laneBits - 1)) - 1)
		predicate = "sgt"
	}
	above := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = icmp %s %s %%%s, %s\n", above, predicate, wideType, sum, llvmSplatSignedInteger(lanes, wideBits, max))
	capped := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = select <%d x i1> %%%s, %s %s, %s %%%s\n", capped, lanes, above, wideType, llvmSplatSignedInteger(lanes, wideBits, max), wideType, sum)
	clamped := "%" + capped
	if mode == amd64PackedAddSignedSaturating {
		min := -int64(uint64(1) << (laneBits - 1))
		below := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = icmp slt %s %s, %s\n", below, wideType, clamped, llvmSplatSignedInteger(lanes, wideBits, min))
		bounded := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = select <%d x i1> %%%s, %s %s, %s %s\n", bounded, lanes, below, wideType, llvmSplatSignedInteger(lanes, wideBits, min), wideType, clamped)
		clamped = "%" + bounded
	}
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = trunc %s %s to %s\n", result, wideType, clamped, vecType)
	return "%" + result
}

func llvmSplatSignedInteger(lanes, bits int, value int64) string {
	var b strings.Builder
	b.WriteByte('<')
	for lane := 0; lane < lanes; lane++ {
		if lane != 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "i%d %d", bits, value)
	}
	b.WriteByte('>')
	return b.String()
}
