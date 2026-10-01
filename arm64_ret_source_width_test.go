package plan9asm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestARM64RawSourceRETWidthMatchesGoObjects(t *testing.T) {
	type fixture struct {
		name                string
		fn                  Func
		ret                 Instr
		retLine, markerLine int
	}
	var source strings.Builder
	var fixtures []fixture
	line := 1
	for _, header := range []struct {
		flags int
		frame int64
	}{
		{4, -8}, {4, 0}, {516, 0}, {4, 8}, {4, 32}, {4, 240}, {4, 4080},
		{4, 4096}, {4, 65536}, {4, 1 << 20}, {4, 1 << 30},
	} {
		for _, call := range []string{"", "BL anchor(SB)", "CALL anchor(SB)", "WORD $0x94000000", "WORD $0xd63f0000"} {
			for _, ret := range []string{"RET", "RET R30", "RET R0", "RET (R30)", "RET 8(R30)", "RET anchor(SB)"} {
				name := fmt.Sprintf("retWidth%d", len(fixtures))
				body := fmt.Sprintf("TEXT %s(SB),%d,$%d-0\n", name, header.flags, header.frame)
				if call != "" {
					body += call + "\n"
				}
				retLine := line + strings.Count(body, "\n")
				body += ret + "\nWORD $0xd503201f\n"
				file, err := Parse(ArchARM64, body)
				if err != nil {
					t.Fatal(err)
				}
				fn := file.Funcs[0]
				fixtures = append(fixtures, fixture{name, fn, fn.Instrs[len(fn.Instrs)-2], retLine, retLine + 1})
				source.WriteString(body)
				line += strings.Count(body, "\n")
			}
		}
	}
	source.WriteString("TEXT anchor(SB),4,$0-0\nRET\n")
	dir := t.TempDir()
	asm := filepath.Join(dir, "ret-width.s")
	if err := os.WriteFile(asm, []byte(source.String()), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "tool", "asm", "-S", "-o", filepath.Join(dir, "ret-width.o"), asm)
	cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=arm64")
	listing, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Go assembler RET object oracle: %v\n%s", err, listing)
	}
	positions := make(map[int]int64)
	rows := regexp.MustCompile(`(?m)^\s*0x([0-9a-f]+)\s+\d+\s+\([^\n]+:(\d+)\)\s+[A-Z]`)
	for _, row := range rows.FindAllStringSubmatch(string(listing), -1) {
		pc, err := strconv.ParseInt(row[1], 16, 64)
		if err != nil {
			t.Fatal(err)
		}
		line, err := strconv.Atoi(row[2])
		if err != nil {
			t.Fatal(err)
		}
		if old, exists := positions[line]; !exists || pc < old {
			positions[line] = pc
		}
	}
	for _, fixture := range fixtures {
		start, startOK := positions[fixture.retLine]
		end, endOK := positions[fixture.markerLine]
		if !startOK || !endOK || end <= start {
			t.Fatalf("missing RET object span for %s: %d..%d", fixture.name, start, end)
		}
		got, known := arm64RawKnownSourceWidth(fixture.fn, fixture.ret)
		if !known || got != end-start {
			t.Errorf("%s %s: source RET width=(%d,%v), Go object=%d bytes", fixture.fn.Instrs[0].Raw, fixture.ret.Raw, got, known, end-start)
		}
	}
	t.Logf("%d source RET spans independently measured from Go objects: frame/NOFRAME, Go BL/CALL vs raw BL/BLR, default/register/symbol returns", len(fixtures))
}

func TestARM64RawSourceRETWidthFailsClosed(t *testing.T) {
	for _, source := range []string{
		"TEXT badWidth(SB),4,$3-0\nRET\n",
		"TEXT badWidth(SB),4,$-16-0\nRET\n",
		"TEXT badWidth(SB),516,$8-0\nRET\n",
		"TEXT badWidth(SB),UNKNOWN_FLAGS,$0-0\nCALL anchor(SB)\nRET\n",
		"TEXT badWidth(SB),4,$0-0\nRET F0\n",
		"TEXT badWidth(SB),4,$0-0\nRET.P R30\n",
	} {
		requireARM64GoAssemblerResult(t, source, false)
		file, err := Parse(ArchARM64, source)
		if err != nil {
			continue
		}
		fn := file.Funcs[0]
		if width, known := arm64RawKnownSourceWidth(fn, fn.Instrs[len(fn.Instrs)-1]); known {
			t.Fatalf("invalid/unresolved RET source established a %d-byte boundary: %s", width, source)
		}
	}
	for _, frame := range []int64{1<<31 - 8, 1 << 32, 1 << 62} {
		file, err := Parse(ArchARM64, fmt.Sprintf("TEXT unknownWidth(SB),4,$%d-0\nRET\n", frame))
		if err != nil {
			t.Fatal(err)
		}
		fn := file.Funcs[0]
		if width, known := arm64RawKnownSourceWidth(fn, fn.Instrs[1]); known {
			t.Fatalf("overflowing Go int32 autosize established a %d-byte boundary", width)
		}
	}
}

func TestCrossLinuxRuntimeMatrixARM64RawBranchCrossesSourceRET(t *testing.T) {
	llc, triple, compiler, runner := arm64FPPairRuntimeTools(t)
	var source, goDecl, cDecl, goChecks, cChecks strings.Builder
	sigs := make(map[string]FuncSig)
	index := 0
	for _, frame := range []int{0, 8, 32, 240} {
		for _, call := range []string{"", "BL ·widthVoid(SB)", "CALL ·widthVoid(SB)", "rawBL", "rawBLR"} {
			name := fmt.Sprintf("retWidthRuntime%d", index)
			index++
			retWidth := 12
			if frame == 0 && (call == "" || strings.HasPrefix(call, "raw")) {
				retWidth = 4
			}
			body := call
			if strings.HasPrefix(call, "raw") {
				link := "WORD $0x97ffffff"
				if call == "rawBLR" {
					link = "ADR helper,R9\nWORD $0xd63f0120"
				}
				body = "MOVD R30,R22\nB callsite\nhelper:\nB (R30)\ncallsite:\n" + link + "\nMOVD R22,R30"
			}
			fmt.Fprintf(&source, `TEXT ·%s(SB),4,$%d-16
%s
MOVD a+0(FP),R0
WORD $%#08x
RET
finish:
WORD $0x91002c00
MOVD R0,ret+8(FP)
RET
`, name, frame, body, 0x14000000+uint32((retWidth+4)/4))
			sigs[name] = arm64LocalRegisterBranchSig(name)
			fmt.Fprintf(&goDecl, "func %s(uint64) uint64\n", name)
			fmt.Fprintf(&cDecl, "extern uint64_t %s(uint64_t);\n", name)
			fmt.Fprintf(&goChecks, `
for _, a := range []uint64{0,1,123,0x8000000000000000,^uint64(0)} {
	if got := %s(a); got != a+11 {
		println(got)
		panic("raw branch across source RET mismatch")
	}
}
`, name)
			fmt.Fprintf(&cChecks, `
for (unsigned i=0;i<5;i++) {
	if (%s(inputs[i])!=inputs[i]+11) {
		fprintf(stderr,"raw branch across source RET mismatch\n");
		return 1;
	}
}
`, name)
		}
	}
	source.WriteString("TEXT ·widthVoid(SB),4,$0-0\nRET\n")
	goDecl.WriteString("func widthVoid()\n")
	sigs["widthVoid"] = FuncSig{Name: "widthVoid", Ret: Void}
	requireARM64GoAssemblerResult(t, source.String(), true)
	runARM64LocalRegisterGoOracle(t, source.String(), "package main\n"+goDecl.String()+"func main(){\n"+goChecks.String()+"}\n", len(runner) != 0)
	file, err := Parse(ArchARM64, source.String())
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{
		"aarch64-unknown-linux-gnu", "aarch64-unknown-linux-musl",
		"aarch64-apple-darwin", "aarch64-pc-windows-msvc",
	} {
		ir, err := Translate(file, Options{Goarch: "arm64", TargetTriple: target, Sigs: sigs,
			ResolveSym: func(sym string) string { return strings.TrimPrefix(sym, "·") },
		})
		if err != nil {
			t.Fatal(err)
		}
		compileLLVMToObject(t, llc, target, "ret-width-runtime.ll", "ret-width-runtime.o", ir)
	}
	ir, err := Translate(file, Options{Goarch: "arm64", TargetTriple: triple, Sigs: sigs,
		ResolveSym: func(sym string) string { return strings.TrimPrefix(sym, "·") },
	})
	if err != nil {
		t.Fatal(err)
	}
	main := "#include <stdint.h>\n#include <stdio.h>\n" + cDecl.String() + "int main(void){ uint64_t inputs[]={0,1,123,UINT64_C(0x8000000000000000),UINT64_MAX};\n" + cChecks.String() + "return 0;}\n"
	compileAndRunRuntimeTestWithCompiler(t, llc, compiler, "ret_width_runtime", triple, ir, main, runner)
	t.Logf("%d raw branch across source RET fixtures x 5 inputs passed native Go and LLVM22, plus four-target objects", index)
}
