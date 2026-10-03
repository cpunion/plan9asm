package plan9asm

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func x86FloatMulDivOpcodes() []string {
	return []string{
		"MULPS", "MULPD", "MULSS", "MULSD",
		"DIVPS", "DIVPD", "DIVSS", "DIVSD",
		"VMULPS", "VMULPD", "VMULSS", "VMULSD",
		"VDIVPS", "VDIVPD", "VDIVSS", "VDIVSD",
	}
}

func x86FloatMulDivFormats(arch string) []string {
	// All eight opcode attributes are independently checked below. Go uses
	// the same yxm/_yvaddpd/_yvaddsd operand grammars as ADD/SUB, not merely
	// the same vector widths. Reuse the independent shape-axis inventory.
	rows := x86FloatAddSubFormats(arch)
	for i, row := range rows {
		rows[i] = strings.NewReplacer("ADD", "MUL", "SUB", "DIV").Replace(row)
	}
	return rows
}

func TestX86FloatMulDivEveryGoOpcodeAttribute(t *testing.T) {
	if !goToolchainAtLeast(runtime.Version(), 1, 27) {
		return // The current-toolchain gate owns encoder-table completeness.
	}
	dir := filepath.Join(runtime.GOROOT(), "src/cmd/internal/obj/x86")
	for _, filename := range []string{"asm6.go", "avx_optabs.go"} {
		data, err := os.ReadFile(filepath.Join(dir, filename))
		if err != nil {
			t.Fatal(err)
		}
		for _, opcode := range x86FloatMulDivOpcodes() {
			legacy := !strings.HasPrefix(opcode, "V")
			if (filename == "asm6.go") != legacy {
				continue
			}
			if legacy {
				if !strings.Contains(string(data), "{A"+opcode+", yxm,") {
					t.Fatalf("Go legacy grammar changed: %s", opcode)
				}
				continue
			}
			scalar := strings.HasSuffix(opcode, "SS") || strings.HasSuffix(opcode, "SD")
			bits := 32
			if strings.HasSuffix(opcode, "PD") || strings.HasSuffix(opcode, "SD") {
				bits = 64
			}
			prefix := "{as: A" + opcode + ", ytab: "
			start := strings.Index(string(data), prefix)
			if start < 0 {
				t.Fatalf("Go omits %s", opcode)
			}
			row := strings.SplitN(string(data)[start:], "\n\t}},", 2)[0]
			table := "_yvaddpd"
			if scalar {
				table = "_yvaddsd"
			}
			if !strings.HasPrefix(row, prefix+table+",") {
				t.Fatalf("Go grammar changed: %s", row)
			}
			seen := 0
			for _, line := range strings.Split(row, "\n") {
				if !strings.Contains(line, "avxEscape | evex") {
					continue
				}
				seen++
				if !strings.Contains(line, "evexZeroingEnabled") || strings.Contains(line, "evexSaeEnabled") ||
					strings.Contains(line, "evexRoundingEnabled") != (scalar || strings.Contains(line, "evex512")) ||
					strings.Contains(line, fmt.Sprintf("evexBcstN%d", bits/8)) != !scalar {
					t.Fatalf("%s opcode attribute changed: %s", opcode, line)
				}
			}
			if scalar && seen != 1 || !scalar && seen != 3 {
				t.Fatalf("%s incomplete EVEX row count %d", opcode, seen)
			}
		}
	}
}

func TestX86FloatMulDivCompleteGoGrammarAndFiveLLVMTargets(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc missing")
	}
	for _, arch := range []string{"386", "amd64"} {
		var source strings.Builder
		sigs := map[string]FuncSig{}
		for index, row := range x86FloatMulDivFormats(arch) {
			name := fmt.Sprintf("muldiv%d", index)
			fmt.Fprintf(&source, "TEXT %s(SB),4,$0-0\n%s\nRET\n", name, row)
			sigs[name] = FuncSig{Name: name, Ret: Void}
		}
		goAssemblerOracleOutput(t, arch, source.String(), "", true)
		file, err := Parse(ArchAMD64, source.String())
		if err != nil {
			t.Fatal(err)
		}
		triples := []string{"x86_64-unknown-linux-gnu", "x86_64-apple-darwin", "x86_64-pc-windows-msvc"}
		if arch == "386" {
			triples = []string{"i386-unknown-linux-gnu", "i686-pc-windows-msvc"}
		}
		for _, triple := range triples {
			ir, err := Translate(file, Options{Goarch: arch, TargetTriple: triple, Sigs: sigs})
			if err != nil {
				t.Fatal(err)
			}
			compileLLVMToObject(t, llc, triple, "muldiv-formats.ll", "muldiv-formats.o", ir)
		}
		t.Logf("%s: %d independent Go MUL/DIV formats", arch, len(sigs))
	}
}

func TestX86FloatMulDivRejectsMalformedBeforeEmission(t *testing.T) {
	for _, row := range []string{
		"MULPS X0,Y1", "MULSD X0,AX", "DIVSS X0,X16", "DIVPD.SAE X0,X1",
		"VMULPS X0,X1", "VMULPD X0,Y1,Y2", "VMULSS Y0,Y1,Y2", "VDIVSD X0,X1,K0,X2",
		"VMULPS.Z X0,X1,X2", "VMULPS.BCST X0,X1,X2", "VMULSS.BCST (AX),X1,X2",
		"VDIVPS.SAE Z0,Z1,Z2", "VMULPS.RD_SAE X0,X1,X2", "VDIVSD.RU_SAE (AX),X1,X2",
		"VMULPS. X0,X1,X2", "VDIVPS.Z.BCST (AX),Z1,K1,Z2",
		"VMULPS $data(SB),X1,X2", "VDIVSS (AX)(SP*2),X1,X2", "VMULPD 2147483648(AX),X1,X2",
	} {
		t.Run(row, func(t *testing.T) {
			goAssemblerOracleOutput(t, "amd64", "TEXT f(SB),4,$0-0\n"+row+"\nRET\n", "", false)
			file, err := Parse(ArchAMD64, "TEXT f(SB),4,$0-0\n"+row+"\nRET\n")
			if err != nil {
				return // Parser rejection also occurs before IR emission.
			}
			ins := file.Funcs[0].Instrs[1]
			ctx, output := newAMD64CtxWithFuncForTest(t, file.Funcs[0], FuncSig{Ret: Void}, nil)
			before := output.Len()
			var known, term bool
			if strings.HasPrefix(row, "V") {
				known, term, err = ctx.lowerBinaryFloating(ins.Op, ins)
			} else {
				known, term, err = ctx.lowerFP(ins.Op, ins)
			}
			if !known || term || err == nil || output.Len() != before {
				t.Fatalf("invalid MUL/DIV emitted IR: %t/%t/%v", known, term, err)
			}
		})
	}
}
