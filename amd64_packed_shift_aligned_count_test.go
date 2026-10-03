package plan9asm

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

var x86LegacyShiftMemoryOpcodes = []struct {
	name string
	code byte
}{
	{"PSLLW", 0xf1}, {"PSLLL", 0xf2}, {"PSLLQ", 0xf3},
	{"PSRLW", 0xd1}, {"PSRLL", 0xd2}, {"PSRLQ", 0xd3},
}

func TestX86PackedShiftLegacyMemoryActualGoEncoding(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		for _, operation := range x86LegacyShiftMemoryOpcodes {
			source := "TEXT shifts(SB),4,$0-0\n" + operation.name + " (CX),X0\nRET\n"
			dir := t.TempDir()
			input, object := filepath.Join(dir, "original.s"), filepath.Join(dir, "original.o")
			if err := os.WriteFile(input, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := runtimeOracleStageContext(t)
			command := exec.CommandContext(ctx, "go", "tool", "asm", "-S", "-o", object, input)
			command.Env = append(os.Environ(), "GOOS=linux", "GOARCH="+arch, "GOTOOLCHAIN=local")
			output, err := runtimeOracleCombinedOutput(ctx, command)
			cancel()
			if err != nil {
				t.Fatalf("actual Go %s %s: %v\n%s", arch, operation.name, err, output)
			}
			var code []byte
			for _, line := range regexp.MustCompile(`(?m)^\s*0x[[:xdigit:]]+\s+((?:[[:xdigit:]]{2} )+)`).FindAllSubmatch(output, -1) {
				part, err := hex.DecodeString(strings.ReplaceAll(string(line[1]), " ", ""))
				if err != nil {
					t.Fatal(err)
				}
				code = append(code, part...)
			}
			want := []byte{0x66, 0x0f, operation.code, 0x01, 0xc3}
			if !bytes.Equal(code, want) {
				t.Fatalf("actual Go %s %s memory encoding=%x, want=%x\n%s", arch, operation.name, code, want, output)
			}
		}
	}
}

func TestX86PackedShiftAlignedReadThreeAPIsFiveO2Objects(t *testing.T) {
	llc, disassembler := findLLVM22Tool("llc"), findLLVM22Tool("llvm-objdump")
	if llc == "" || disassembler == "" {
		t.Fatal("LLVM22 llc and llvm-objdump required")
	}
	var source strings.Builder
	source.WriteString("TEXT shifts(SB),4,$0-0\n")
	for _, operation := range x86LegacyShiftMemoryOpcodes {
		fmt.Fprintf(&source, "%s (AX),X0\n", operation.name)
	}
	source.WriteString("RET\n")
	file, err := Parse(ArchAMD64, source.String())
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []struct{ arch, triple string }{
		{"amd64", "x86_64-apple-darwin"}, {"amd64", "x86_64-unknown-linux-gnu"}, {"amd64", "x86_64-pc-windows-msvc"},
		{"386", "i386-unknown-linux-gnu"}, {"386", "i686-pc-windows-msvc"},
	} {
		options := Options{Goarch: target.arch, TargetTriple: target.triple, Sigs: map[string]FuncSig{"shifts": {Name: "shifts", Ret: Void}}}
		ir, err := Translate(file, options)
		if err != nil {
			t.Fatal(err)
		}
		context := llvm.NewContext()
		func() {
			defer context.Dispose()
			module, err := TranslateModuleInContext(context, file, options)
			if !module.IsNil() {
				defer module.Dispose()
			}
			if err != nil {
				t.Fatal(err)
			}
			result, err := translateGoModuleInContext(context, x86PackedShiftGoPackage(t, target.arch),
				[]byte(strings.Replace(source.String(), "shifts(SB)", "·shifts(SB)", 1)), GoModuleOptions{GOARCH: target.arch, TargetTriple: target.triple})
			if result != nil && !result.Module.IsNil() {
				defer result.Module.Dispose()
			}
			if err != nil {
				t.Fatal(err)
			}
			for api, body := range []string{ir, module.String(), result.Module.String()} {
				if strings.Count(body, `asm sideeffect "movdqa ($1), $0", "=x,r,~{memory}"`) != 6 {
					t.Fatalf("%s API%d lost complete aligned native count reads", target.triple, api)
				}
				dir := t.TempDir()
				input, object := filepath.Join(dir, "aligned.ll"), filepath.Join(dir, "aligned.o")
				if err := os.WriteFile(input, []byte(body), 0600); err != nil {
					t.Fatal(err)
				}
				stage, cancel := runtimeOracleStageContext(t)
				command := exec.CommandContext(stage, llc, "-O2", "-mtriple="+target.triple, "-filetype=obj", input, "-o", object)
				output, err := runtimeOracleCombinedOutput(stage, command)
				cancel()
				if err != nil {
					t.Fatalf("O2 %s API%d: %v\n%s", target.triple, api, err, output)
				}
				stage, cancel = runtimeOracleStageContext(t)
				command = exec.CommandContext(stage, disassembler, "-d", object)
				output, err = runtimeOracleCombinedOutput(stage, command)
				cancel()
				if err != nil || strings.Count(string(output), "movdqa") < 6 {
					t.Fatalf("O2 actual aligned machine reads %s API%d: %v\n%s", target.triple, api, err, output)
				}
			}
		}()
	}
}

func TestX86PackedShiftLegacyFPAlignmentAndCallerPermissionsRemainContext(t *testing.T) {
	for _, operation := range x86LegacyShiftMemoryOpcodes {
		file, err := Parse(ArchAMD64, "TEXT f(SB),4,$0-0\n"+operation.name+" count+0(FP),X0\nRET\n")
		if err != nil {
			t.Fatal(err)
		}
		ins := file.Funcs[0].Instrs[1]
		ctx, output := newAMD64CtxWithFuncForTest(t, file.Funcs[0], FuncSig{Ret: Void}, nil)
		before := output.Len()
		if _, _, err := ctx.lowerX86PackedShift(ins.Op, ins); !errors.Is(err, ErrProbeNeedsContext) || output.Len() != before {
			t.Fatalf("unproved FP alignment emitted IR: %s %v", operation.name, err)
		}
	}
	// Initialized vectors do not supply alignment or memory permission.
	for _, operation := range x86LegacyShiftMemoryOpcodes {
		file, err := Parse(ArchAMD64, "TEXT f(SB),4,$0-0\n"+operation.name+" (CX),X0\nRET\n")
		if err != nil {
			t.Fatal(err)
		}
		ins := file.Funcs[0].Instrs[1]
		form, recognized, err := x86PackedShiftProperties(ins.Op, ins, "amd64")
		if !recognized || err != nil {
			t.Fatal(err)
		}
		state := x86SourceStackState{vectors: map[Reg]uint64{"Z0": ^uint64(0)}}
		if x86CallerPackedShiftStep(&ins, form, &state, nil) {
			t.Fatal("initialized vector granted legacy memory alignment")
		}
	}
}
