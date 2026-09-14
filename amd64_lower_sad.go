package plan9asm

import (
	"fmt"
	"strings"
)

// lowerPackedSumAbsoluteDifferences implements Go 1.27's PSADBW yxm table
// and VPSADBW _yvaesdec table. Neither table has masking or suffix forms.
func (c *amd64Ctx) lowerPackedSumAbsoluteDifferences(op Op, ins Instr) (ok bool, terminated bool, err error) {
	rawOp := strings.ToUpper(string(op))
	baseOp := rawOp
	if dot := strings.IndexByte(rawOp, '.'); dot >= 0 {
		baseOp = rawOp[:dot]
	}
	switch baseOp {
	case "PSADBW", "VPSADBW":
		// handled below
	default:
		return false, false, nil
	}
	if rawOp != baseOp {
		return true, false, fmt.Errorf("amd64 %s does not accept instruction suffixes: %q", baseOp, ins.Raw)
	}
	if baseOp == "PSADBW" {
		if len(ins.Args) != 2 || ins.Args[1].Kind != OpReg || !isAMD64XReg(ins.Args[1].Reg) {
			return true, false, fmt.Errorf("amd64 PSADBW expects X/m source and X destination: %q", ins.Raw)
		}
		if ins.Args[0].Kind == OpReg && !isAMD64XReg(ins.Args[0].Reg) {
			return true, false, fmt.Errorf("amd64 PSADBW register source must be X: %q", ins.Raw)
		}
		first, err := c.loadXVecOperand(ins.Args[0])
		if err != nil {
			return true, false, err
		}
		second, err := c.loadX(ins.Args[1].Reg)
		if err != nil {
			return true, false, err
		}
		result := c.emitPackedSumAbsoluteByteDifferences(16, first, second)
		return true, false, c.storeX(ins.Args[1].Reg, result)
	}

	if len(ins.Args) != 3 || ins.Args[2].Kind != OpReg {
		return true, false, fmt.Errorf("amd64 VPSADBW expects src1, src2, destination: %q", ins.Raw)
	}
	byteWidth := amd64VectorByteWidth(ins.Args[2].Reg)
	if byteWidth == 0 {
		return true, false, fmt.Errorf("amd64 VPSADBW destination must be X, Y, or Z: %q", ins.Raw)
	}
	if !amd64VectorRegisterHasWidth(ins.Args[1], byteWidth) {
		return true, false, fmt.Errorf("amd64 VPSADBW second source must match its destination width: %q", ins.Raw)
	}
	if ins.Args[0].Kind == OpReg && !amd64VectorRegisterHasWidth(ins.Args[0], byteWidth) {
		return true, false, fmt.Errorf("amd64 VPSADBW first source must match its destination width: %q", ins.Raw)
	}
	first, err := c.loadPackedCompareBytes(ins.Args[0], byteWidth)
	if err != nil {
		return true, false, err
	}
	second, err := c.loadPackedCompareBytes(ins.Args[1], byteWidth)
	if err != nil {
		return true, false, err
	}
	result := c.emitPackedSumAbsoluteByteDifferences(byteWidth, first, second)
	return true, false, c.storeVectorBytes(ins.Args[2].Reg, byteWidth, result)
}

func (c *amd64Ctx) emitPackedSumAbsoluteByteDifferences(byteWidth int, first, second string) string {
	firstWide := c.newTmp()
	secondWide := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = zext <%d x i8> %s to <%d x i16>\n", firstWide, byteWidth, first, byteWidth)
	fmt.Fprintf(c.b, "  %%%s = zext <%d x i8> %s to <%d x i16>\n", secondWide, byteWidth, second, byteWidth)
	firstMinusSecond := c.newTmp()
	secondMinusFirst := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = sub <%d x i16> %%%s, %%%s\n", firstMinusSecond, byteWidth, firstWide, secondWide)
	fmt.Fprintf(c.b, "  %%%s = sub <%d x i16> %%%s, %%%s\n", secondMinusFirst, byteWidth, secondWide, firstWide)
	firstGreater := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = icmp uge <%d x i16> %%%s, %%%s\n", firstGreater, byteWidth, firstWide, secondWide)
	differences := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = select <%d x i1> %%%s, <%d x i16> %%%s, <%d x i16> %%%s\n", differences, byteWidth, firstGreater, byteWidth, firstMinusSecond, byteWidth, secondMinusFirst)

	groups := byteWidth / 8
	result := "zeroinitializer"
	for group := 0; group < groups; group++ {
		sum := ""
		for lane := group * 8; lane < (group+1)*8; lane++ {
			value := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = extractelement <%d x i16> %%%s, i32 %d\n", value, byteWidth, differences, lane)
			if sum == "" {
				sum = "%" + value
				continue
			}
			added := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = add i16 %s, %%%s\n", added, sum, value)
			sum = "%" + added
		}
		wide := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = zext i16 %s to i64\n", wide, sum)
		inserted := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = insertelement <%d x i64> %s, i64 %%%s, i32 %d\n", inserted, groups, result, wide, group)
		result = "%" + inserted
	}
	bytesValue := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i64> %s to <%d x i8>\n", bytesValue, groups, result, byteWidth)
	return "%" + bytesValue
}
