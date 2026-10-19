package plan9asm

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

type x86AbsSignAlignmentRow struct {
	opcode       string
	bits         int
	sign, legacy bool
}

// Type4 has different legacy/VEX alignment rules. Do not derive either the
// opcode directory or that distinction from the implementation's form table.
func x86AbsSignAlignmentRows() []x86AbsSignAlignmentRow {
	var rows []x86AbsSignAlignmentRow
	for _, vector := range []bool{false, true} {
		for _, sign := range []bool{false, true} {
			prefix := "PABS"
			if sign {
				prefix = "PSIGN"
			}
			if vector {
				prefix = "V" + prefix
			}
			for index, width := range []string{"B", "W", "D"} {
				rows = append(rows, x86AbsSignAlignmentRow{prefix + width, 8 << index, sign, !vector})
			}
		}
	}
	return rows
}

func (row x86AbsSignAlignmentRow) instruction() string {
	if row.sign && !row.legacy {
		return row.opcode + " (BX),X0,X0"
	}
	return row.opcode + " (BX),X0"
}

func x86AbsSignAlignmentOriginalBytes(t *testing.T, arch string) [][]byte {
	t.Helper()
	rows := x86AbsSignAlignmentRows()
	var source strings.Builder
	for _, row := range rows {
		source.WriteString(row.instruction() + "\n")
	}
	code, pcs := x86NumericalBoundaryGoBytes(t, arch, source.String()+"RET")
	if len(pcs) != 13 || len(code) == 0 || code[len(code)-1] != 0xc3 {
		t.Fatal("incomplete original Go legacy/VEX alignment opcode inventory")
	}
	result := make([][]byte, len(rows))
	for index, row := range rows {
		opcode := byte(0x1c + index%3)
		if row.sign {
			opcode = byte(0x08 + index%3)
		}
		want := []byte{0x66, 0x0f, 0x38, opcode, 0x03}
		if !row.legacy {
			want = []byte{0xc4, 0xe2, 0x79, opcode, 0x03}
		}
		result[index] = append([]byte(nil), code[pcs[index]:pcs[index+1]]...)
		if !bytes.Equal(result[index], want) {
			t.Fatal("actual Go opcode differs from independent Intel ModRM/prefix row", arch, row.opcode, result[index], want)
		}
	}
	return result
}

func x86AbsSignAlignmentProgram(arch string, vectorControls bool) (string, string) {
	word, move := goWordSize(arch), "MOVQ"
	if arch == "386" {
		move = "MOVL"
	}
	rows := x86AbsSignAlignmentRows()
	if !vectorControls {
		rows = rows[:6]
	}
	var source, declarations strings.Builder
	for index, row := range rows {
		for _, role := range []string{"live", "dead"} {
			name := fmt.Sprintf("AbsAlignment%d_%s", index, role)
			fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-%d\n%s out+0(FP),DI\n%s input+%d(FP),BX\n%s old+%d(FP),CX\nMOVOU (CX),X0\nSTC\n%s\n",
				name, 3*word, move, move, word, move, 2*word, row.instruction())
			if role == "live" {
				source.WriteString("MOVOU X0,(DI)\n")
			}
			source.WriteString("SETCS 16(DI)\nRET\n")
			fmt.Fprintf(&declarations, "func %s(out *[32]byte,input,old *[16]byte)\n", name)
		}
	}
	return source.String(), declarations.String()
}

func TestX86LegacyAbsSignAlignmentIndependentOriginalBytes(t *testing.T) {
	x86PackedAbsSignOriginalProgramUnchanged(t)
	for _, arch := range []string{"386", "amd64"} {
		t.Run(arch, func(t *testing.T) { x86AbsSignAlignmentOriginalBytes(t, arch) })
	}
}

func TestX86LegacyAbsSignAlignmentNeedsActualNativeBackingAndTarget(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		for _, row := range x86AbsSignAlignmentRows()[:6] {
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(target.triple+"/"+row.opcode+"/"+api, func(t *testing.T) {
					source := "TEXT ·Value(SB),4,$0-16\nPXOR X0,X0\n" + row.opcode + " a+0(FP),X0\nRET\n"
					ir, err := x86ScalarFPBitViewAPI(t, source, "func Value(a,b uint64)", target.arch, target.triple, api)
					if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
						t.Fatal("logical FP bytes acquired original native m128 alignment and backing", err)
					}
				})
			}
		}
	}
	for _, triple := range []string{"aarch64-unknown-linux-gnu", "armv7-unknown-linux-gnueabihf", "wasm32-unknown-unknown"} {
		t.Run(triple, func(t *testing.T) {
			source, declarations := x86AbsSignAlignmentProgram("amd64", false)
			ir, err := x86ScalarFPBitViewAPI(t, source, declarations, "amd64", triple, "Go")
			if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatal("a nonmatching target silently erased original x86 alignment/fault behavior", err)
			}
		})
	}
}

func TestX86LegacyAbsSignAlignmentRequiredNativeOwners(t *testing.T) {
	const name = "x86_legacy_abs_sign_alignment_native_linux_amd64_test.go"
	for _, target := range []struct{ os, arch string }{
		{"linux", "amd64"}, {"linux", "386"}, {"linux", "arm64"},
		{"darwin", "amd64"}, {"darwin", "arm64"}, {"windows", "amd64"},
	} {
		context := build.Default
		context.GOOS, context.GOARCH = target.os, target.arch
		selected, err := context.MatchFile(".", name)
		if err != nil || selected != (target.os == "linux" && target.arch == "amd64") {
			t.Fatal("required actual-native fault owner selected on the wrong host", target, selected, err)
		}
	}
	syntax, err := parser.ParseFile(token.NewFileSet(), name, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	wanted := map[string]bool{
		"TestCrossLinuxRuntimeMatrixX86LegacyAbsSignOriginalAlignmentFaultsAMD64": false,
		"TestCrossLinuxRuntimeMatrixX86LegacyAbsSignOriginalAlignmentFaults386":   false,
	}
	for _, declaration := range syntax.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		ast.Inspect(function, func(node ast.Node) bool {
			if selector, ok := node.(*ast.SelectorExpr); ok && strings.HasPrefix(selector.Sel.Name, "Skip") {
				t.Fatal("required native ABS/sign execution became an optional skip", function.Name.Name)
			}
			return true
		})
		if _, found := wanted[function.Name.Name]; !found {
			continue
		}
		host, full := 0, 0
		ast.Inspect(function.Body, func(node ast.Node) bool {
			if call, ok := node.(*ast.CallExpr); ok {
				if callee, ok := call.Fun.(*ast.Ident); ok {
					if callee.Name == "x87IA32RequireNativeHost" {
						host++
					}
					if callee.Name == "x86LegacyAbsSignNativeAlignment" {
						full++
					}
				}
			}
			return true
		})
		if host != 2 || full != 1 || wanted[function.Name.Name] {
			t.Fatal("required native owner lost host brackets or its entire fault matrix", function.Name.Name, host, full)
		}
		wanted[function.Name.Name] = true
	}
	for root, found := range wanted {
		if !found {
			t.Fatal("required native owner is absent", root)
		}
	}
}

// Real LLVM22 machine objects, not IR spelling or align16 metadata, must
// retain each legacy m128 opcode in both live and dead-result functions.
// Original-process signal behavior is a separate required native owner.
func TestX86LegacyAbsSignAlignmentActualNativeObjects(t *testing.T) {
	llc, disassembler := findLLVM22Tool("llc"), findLLVM22Tool("llvm-objdump")
	if llc == "" || disassembler == "" {
		t.Fatal("LLVM22 llc and llvm-objdump required for original native alignment objects")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			source, declarations := x86AbsSignAlignmentProgram(target.arch, false)
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, source, declarations, target.arch, target.triple, api)
					if err != nil || ir == "" {
						t.Fatal("initialized original legacy source rejected before object examination", err)
					}
					for _, pipeline := range []struct{ name, body, level string }{
						{"backend_O0", ir, "-O0"}, {"backend_O2", ir, "-O2"},
						{"IR_O2_backend_O2", x86DwordProductIROptimize(t, ir), "-O2"},
					} {
						t.Run(pipeline.name, func(t *testing.T) {
							directory := t.TempDir()
							input, object := filepath.Join(directory, "alignment.ll"), filepath.Join(directory, "alignment.o")
							if err := os.WriteFile(input, []byte(pipeline.body), 0600); err != nil {
								t.Fatal(err)
							}
							stage, cancel := runtimeOracleStageContext(t)
							output, err := runtimeOracleCombinedOutput(stage, exec.CommandContext(stage, llc,
								pipeline.level, "-mtriple="+target.triple, "-filetype=obj", input, "-o", object))
							cancel()
							if err != nil {
								t.Fatalf("actual native alignment object: %v\n%s", err, output)
							}
							stage, cancel = runtimeOracleStageContext(t)
							listing, err := runtimeOracleCombinedOutput(stage, exec.CommandContext(stage, disassembler, "-d", object))
							cancel()
							if err != nil {
								t.Fatalf("independent LLVM22 disassembly: %v\n%s", err, listing)
							}
							sections := regexp.MustCompile(`(?m)^[0-9a-f]+ <_?(AbsAlignment[0-9]+_(?:live|dead))>:\r?\n`).FindAllSubmatchIndex(listing, -1)
							if len(sections) != 12 {
								t.Fatal("actual object lost a complete live/dead original function", len(sections))
							}
							for index, section := range sections {
								name := string(listing[section[2]:section[3]])
								end := len(listing)
								if index+1 < len(sections) {
									end = sections[index+1][0]
								}
								body := listing[section[1]:end]
								t.Run(name, func(t *testing.T) {
									var row int
									if _, err := fmt.Sscanf(name, "AbsAlignment%d_", &row); err != nil || row < 0 || row >= 6 {
										t.Fatal("unexpected physical object function", name, err)
									}
									pattern := `(?m)\b` + strings.ToLower(x86AbsSignAlignmentRows()[row].opcode) + `[ \t]+[^\r\n]*\([^\r\n]+\),[ \t]*%xmm[0-9]+`
									if count := len(regexp.MustCompile(pattern).FindAll(body, -1)); count != 1 {
										t.Fatalf("original faulting m128 opcode not retained exactly once: %d\n%s", count, body)
									}
								})
							}
						})
					}
				})
			}
		})
	}
}
