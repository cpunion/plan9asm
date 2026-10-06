package plan9asm

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var x86AverageNames = []string{"PAVGB", "PAVGW", "VPAVGB", "VPAVGW"}

// Independent of the production grammar. Named mode32 MMX is excluded by
// Go's Py1 selector, while low X/Y/Z operands have physical mode32 identities.
func x86AveragePhysicalRows(arch string) []string {
	var rows []string
	for _, name := range x86AverageNames[:2] {
		for _, source := range []string{"X0", "X7", "(AX)", "16(AX)(CX*4)", "·data(SB)", "(FS)", "(GS)"} {
			rows = append(rows, name+" "+source+",X1")
		}
		if arch == "amd64" {
			rows = append(rows, name+" X15,X0", name+" M0,M7", name+" (AX),M7")
		}
	}
	for _, name := range x86AverageNames[2:] {
		for _, bank := range []string{"X", "Y", "Z"} {
			for _, source := range []string{bank + "0", bank + "7", "(AX)", "16(AX)(CX*4)", "·data(SB)", "(FS)", "(GS)"} {
				rows = append(rows, fmt.Sprintf("%s %s,%s1,%s2", name, source, bank, bank))
				if arch == "amd64" {
					rows = append(rows, fmt.Sprintf("%s %s,%s31,K1,%s30", name, source, bank, bank),
						fmt.Sprintf("%s.Z %s,%s31,K7,%s30", name, source, bank, bank))
				}
			}
		}
	}
	return rows
}

func TestX86PackedAverageActualGoCompleteTables(t *testing.T) {
	root := filepath.Join(runtime.GOROOT(), "src/cmd/internal/obj/x86")
	legacy, err := os.ReadFile(filepath.Join(root, "asm6.go"))
	if err != nil {
		t.Fatal(err)
	}
	avx, err := os.ReadFile(filepath.Join(root, "avx_optabs.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range []string{
		"{APAVGB, ymm, Py1, opBytes{0xe0, Pe, 0xe0}}",
		"{APAVGW, ymm, Py1, opBytes{0xe3, Pe, 0xe3}}",
	} {
		if strings.Count(string(legacy), row) != 1 {
			t.Fatalf("actual Go legacy registration changed: %s", row)
		}
	}
	start := strings.Index(string(legacy), "var ymm =")
	if start < 0 {
		t.Fatal("missing actual ymm table")
	}
	ymm := strings.SplitN(string(legacy)[start:], "\n}", 2)[0]
	for _, row := range []string{"{Zm_r_xm, 1, argList{Ymm, Ymr}}", "{Zm_r_xm, 2, argList{Yxm, Yxr}}"} {
		if strings.Count(ymm, "argList{") != 2 || strings.Count(ymm, row) != 1 {
			t.Fatal("actual ymm class/selector table changed", ymm)
		}
	}
	for index, name := range x86AverageNames[2:] {
		opcode := []string{"E0", "E3"}[index]
		expected := fmt.Sprintf(`{as: A%s, ytab: _yvandnpd, prefix: Pavx, op: opBytes{
		avxEscape | vex128 | vex66 | vex0F | vexW0, 0x%s,
		avxEscape | vex256 | vex66 | vex0F | vexW0, 0x%s,
		avxEscape | evex128 | evex66 | evex0F | evexW0, evexN16 | evexZeroingEnabled, 0x%s,
		avxEscape | evex256 | evex66 | evex0F | evexW0, evexN32 | evexZeroingEnabled, 0x%s,
		avxEscape | evex512 | evex66 | evex0F | evexW0, evexN64 | evexZeroingEnabled, 0x%s,
	}},`, name, opcode, opcode, opcode, opcode, opcode)
		if err := x86SignedPackedFloatValidateGoOpcodeRows(string(avx), name, expected); err != nil {
			t.Fatal(err)
		}
	}
	start = strings.Index(string(avx), "var _yvandnpd =")
	if start < 0 {
		t.Fatal("missing actual ytab")
	}
	table := strings.SplitN(string(avx)[start:], "\n}", 2)[0]
	for _, classes := range []string{"Yxm, Yxr, Yxr", "Yym, Yyr, Yyr",
		"YxmEvex, YxrEvex, YxrEvex", "YxmEvex, YxrEvex, Yknot0, YxrEvex",
		"YymEvex, YyrEvex, YyrEvex", "YymEvex, YyrEvex, Yknot0, YyrEvex",
		"Yzm, Yzr, Yzr", "Yzm, Yzr, Yknot0, Yzr"} {
		if strings.Count(table, "argList{") != 8 || !strings.Contains(table, "argList{"+classes+"}") {
			t.Fatal("complete actual Go ytab changed", classes)
		}
	}
}

func TestX86PackedAverageActualGoModeAndPhysicalBoundary(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		var source strings.Builder
		source.WriteString("TEXT f(SB),4,$0-0\n")
		for _, row := range x86AveragePhysicalRows(arch) {
			source.WriteString(row + "\n")
		}
		source.WriteString("RET\n")
		goAssemblerOracleOutput(t, arch, source.String(), "", true)
	}
	for _, name := range x86AverageNames[:2] {
		for _, source := range []string{"M0", "(AX)", "AX"} {
			goAssemblerOracleOutput(t, "386", "TEXT f(SB),4,$0-0\n"+name+" "+source+",M1\nRET\n", "", false)
		}
	}
	for _, name := range x86AverageNames[2:] {
		for _, bank := range []string{"X", "Y"} {
			for position := 0; position < 3; position++ {
				args := []string{bank + "0", bank + "1", bank + "2"}
				args[position] = bank + "20"
				assembly := "TEXT f(SB),4,$0-0\n" + name + " " + strings.Join(args, ",") + "\nRET\n"
				goAssemblerOracleOutput(t, "386", assembly, "", true)
				file, err := Parse(ArchAMD64, assembly)
				if err != nil {
					t.Fatal(err)
				}
				ctx, out := newAMD64CtxWithFuncForTest(t, file.Funcs[0], FuncSig{Ret: Void}, nil)
				ctx.goarch = "386"
				before := out.String()
				known, _, err := ctx.lowerPackedUnsignedAverage(file.Funcs[0].Instrs[1].Op, file.Funcs[0].Instrs[1])
				if !known || !errors.Is(err, ErrProbeNeedsContext) || out.String() != before {
					t.Errorf("unproved physical high bank emitted IR: %s: %v", assembly, err)
				}
			}
		}
	}
}

func TestX86PackedAverageNativeStorageAndIdentityBoundaries(t *testing.T) {
	for _, name := range x86AverageNames[:2] {
		for _, row := range []string{name + " (SP),X0", name + " (BP),X0", name + " (AX)(BP*2),X0", name + " value+0(FP),X0"} {
			file, err := Parse(ArchAMD64, "TEXT f(SB),4,$0-64\n"+row+"\nRET\n")
			if err != nil {
				t.Fatal(err)
			}
			ctx, out := newAMD64CtxWithFuncForTest(t, file.Funcs[0], x86WordMaddFPType(t), nil)
			before := out.String()
			_, _, err = ctx.lowerPackedUnsignedAverage(file.Funcs[0].Instrs[1].Op, file.Funcs[0].Instrs[1])
			if !errors.Is(err, ErrProbeNeedsContext) || out.String() != before {
				t.Errorf("virtual storage gained native m128 alignment: %s: %v", row, err)
			}
		}
	}
	for _, name := range x86AverageNames {
		row := name + " X0,X1"
		if strings.HasPrefix(name, "V") {
			row += ",X2"
		}
		file, _ := Parse(ArchAMD64, "TEXT f(SB),4,$0-0\n"+row+"\nRET\n")
		ins := file.Funcs[0].Instrs[1]
		ins.Op = "PADDW"
		ctx, out := newAMD64CtxWithFuncForTest(t, file.Funcs[0], FuncSig{Ret: Void}, nil)
		before := out.String()
		if _, _, err := ctx.lowerPackedUnsignedAverage(Op(name), ins); err == nil || out.String() != before {
			t.Errorf("contradictory opcode emitted PAVG IR: %s: %v", name, err)
		}
	}
}
