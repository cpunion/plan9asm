package plan9asm

import (
	"fmt"
	"strings"
)

var amd64TwoSourcePermuteLaneBits = map[string]int{
	"VPERMI2B":  8,
	"VPERMI2W":  16,
	"VPERMI2D":  32,
	"VPERMI2PS": 32,
	"VPERMI2Q":  64,
	"VPERMI2PD": 64,
}

type amd64TwoSourcePermuteSuffixes struct {
	broadcast bool
	zeroing   bool
}

func amd64ParseTwoSourcePermuteSuffixes(rawOp, baseOp string) (amd64TwoSourcePermuteSuffixes, error) {
	var suffixes amd64TwoSourcePermuteSuffixes
	dot := strings.IndexByte(rawOp, '.')
	if dot < 0 {
		return suffixes, nil
	}
	parts := strings.Split(rawOp[dot+1:], ".")
	for index, part := range parts {
		switch part {
		case "BCST":
			if suffixes.broadcast || suffixes.zeroing || index != 0 {
				return suffixes, fmt.Errorf("amd64 %s has invalid or duplicate .BCST suffix", baseOp)
			}
			suffixes.broadcast = true
		case "Z":
			if suffixes.zeroing || index != len(parts)-1 {
				return suffixes, fmt.Errorf("amd64 %s has invalid or duplicate .Z suffix", baseOp)
			}
			suffixes.zeroing = true
		default:
			return suffixes, fmt.Errorf("amd64 %s has unsupported suffix .%s", baseOp, part)
		}
	}
	return suffixes, nil
}

// lowerTwoSourcePermute implements Go 1.27's complete _yvblendmpd operand
// table for VPERMI2{B,W,D,Q,PS,PD}. The destination supplies the indices and
// is also the merge-masked fallback value.
func (c *amd64Ctx) lowerTwoSourcePermute(op Op, ins Instr) (ok bool, terminated bool, err error) {
	rawOp := strings.ToUpper(string(op))
	baseOp := rawOp
	if dot := strings.IndexByte(rawOp, '.'); dot >= 0 {
		baseOp = rawOp[:dot]
	}
	laneBits, ok := amd64TwoSourcePermuteLaneBits[baseOp]
	if !ok {
		return false, false, nil
	}
	suffixes, err := amd64ParseTwoSourcePermuteSuffixes(rawOp, baseOp)
	if err != nil {
		return true, false, fmt.Errorf("%w: %q", err, ins.Raw)
	}
	if suffixes.broadcast && laneBits != 32 && laneBits != 64 {
		return true, false, fmt.Errorf("amd64 %s does not enable broadcast in Go 1.27: %q", baseOp, ins.Raw)
	}
	if len(ins.Args) != 3 && len(ins.Args) != 4 {
		return true, false, fmt.Errorf("amd64 %s expects first source, second source, [K mask,] destination: %q", baseOp, ins.Raw)
	}
	masked := len(ins.Args) == 4
	if masked && c.goarch == "386" {
		return true, false, fmt.Errorf("386 %s masked form exceeds the Go assembler frontend's operand limit: %q", baseOp, ins.Raw)
	}
	if suffixes.zeroing && !masked {
		return true, false, fmt.Errorf("amd64 %s .Z requires a K1-K7 mask: %q", baseOp, ins.Raw)
	}

	destination := ins.Args[len(ins.Args)-1]
	byteWidth := 0
	if destination.Kind == OpReg {
		byteWidth = amd64VectorByteWidth(destination.Reg)
	}
	if byteWidth == 0 || !c.isGoPackedVectorMoveRegister(destination, byteWidth) {
		return true, false, fmt.Errorf("amd64 %s expects an in-range X, Y, or Z destination: %q", baseOp, ins.Raw)
	}
	second := ins.Args[1]
	if !c.isGoPackedVectorMoveRegister(second, byteWidth) {
		return true, false, fmt.Errorf("amd64 %s second source must match the destination width: %q", baseOp, ins.Raw)
	}
	first := ins.Args[0]
	if first.Kind == OpReg {
		if !c.isGoPackedVectorMoveRegister(first, byteWidth) {
			return true, false, fmt.Errorf("amd64 %s first source register must match the destination width: %q", baseOp, ins.Raw)
		}
	} else if !isAMD64MemoryOperand(first) {
		return true, false, fmt.Errorf("amd64 %s first source must be a matching vector register or memory: %q", baseOp, ins.Raw)
	}
	if suffixes.broadcast && !isAMD64MemoryOperand(first) {
		return true, false, fmt.Errorf("amd64 %s.BCST requires a memory first source: %q", baseOp, ins.Raw)
	}

	mask := ""
	if masked {
		maskArg := ins.Args[2]
		if maskArg.Kind != OpReg {
			return true, false, fmt.Errorf("amd64 %s masked form expects K1-K7: %q", baseOp, ins.Raw)
		}
		maskIndex, validMask := amd64ParseKReg(maskArg.Reg)
		if !validMask || maskIndex == 0 {
			return true, false, fmt.Errorf("amd64 %s masked form expects K1-K7: %q", baseOp, ins.Raw)
		}
		mask, err = c.loadK(maskArg.Reg)
		if err != nil {
			return true, false, err
		}
	}

	lanes := byteWidth * 8 / laneBits
	firstValue, err := c.loadPackedCompareLanes(first, byteWidth, laneBits, suffixes.broadcast)
	if err != nil {
		return true, false, err
	}
	secondValue, err := c.loadPackedCompareLanes(second, byteWidth, laneBits, false)
	if err != nil {
		return true, false, err
	}
	oldBytes, err := c.loadPackedCompareBytes(destination, byteWidth)
	if err != nil {
		return true, false, err
	}
	indices := c.bitcastVectorBytesToIntegerLanes(byteWidth, lanes, laneBits, oldBytes)
	computed := c.emitTwoSourcePermute(lanes, laneBits, firstValue, secondValue, indices)
	if masked {
		computed = amd64ApplyIntegerLaneMask(c, lanes, laneBits, computed, indices, mask, suffixes.zeroing)
	}
	out := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i%d> %s to <%d x i8>\n", out, lanes, laneBits, computed, byteWidth)
	return true, false, c.storeVectorBytes(destination.Reg, byteWidth, "%"+out)
}

func (c *amd64Ctx) emitTwoSourcePermute(lanes, laneBits int, firstPlan9, secondPlan9, indices string) string {
	// Intel VPERMI2 calls the Plan 9 second source its first table, and the
	// Plan 9 r/m first source its second table.
	result := "zeroinitializer"
	for lane := 0; lane < lanes; lane++ {
		indexValue := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = extractelement <%d x i%d> %s, i32 %d\n", indexValue, lanes, laneBits, indices, lane)
		bounded := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = and i%d %%%s, %d\n", bounded, laneBits, indexValue, 2*lanes-1)
		fromSecondPlan9 := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = icmp uge i%d %%%s, %d\n", fromSecondPlan9, laneBits, bounded, lanes)
		withinTable := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = and i%d %%%s, %d\n", withinTable, laneBits, bounded, lanes-1)
		indexI32 := "%" + withinTable
		if laneBits < 32 {
			converted := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = zext i%d %%%s to i32\n", converted, laneBits, withinTable)
			indexI32 = "%" + converted
		} else if laneBits > 32 {
			converted := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = trunc i%d %%%s to i32\n", converted, laneBits, withinTable)
			indexI32 = "%" + converted
		}
		fromFirst := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = extractelement <%d x i%d> %s, i32 %s\n", fromFirst, lanes, laneBits, firstPlan9, indexI32)
		fromSecond := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = extractelement <%d x i%d> %s, i32 %s\n", fromSecond, lanes, laneBits, secondPlan9, indexI32)
		selected := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = select i1 %%%s, i%d %%%s, i%d %%%s\n", selected, fromSecondPlan9, laneBits, fromFirst, laneBits, fromSecond)
		inserted := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = insertelement <%d x i%d> %s, i%d %%%s, i32 %d\n", inserted, lanes, laneBits, result, laneBits, selected, lane)
		result = "%" + inserted
	}
	return result
}
