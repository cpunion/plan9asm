package plan9asm

import (
	"bytes"
	"fmt"
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
