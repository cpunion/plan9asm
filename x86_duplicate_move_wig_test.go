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

// Intel SDM Vol. 2B specifies VEX.128/256 WIG for all three opcodes.
// Go's named optab emits W0, which does not make a raw W1 instruction invalid.
// These byte witnesses do not use the production or existing raw encoders.
func x86DuplicateWIGWitnesses(visit func(name, bank string, memory, w bool, code []byte)) {
	for _, row := range []struct {
		name           string
		prefix, opcode byte
	}{{"VMOVDDUP", 3, 0x12}, {"VMOVSHDUP", 2, 0x16}, {"VMOVSLDUP", 2, 0x12}} {
		for _, bank := range []string{"X", "Y"} {
			for _, memory := range []bool{false, true} {
				for _, w := range []bool{false, true} {
					bits, operand := byte(0x78|row.prefix), byte(0xd0) // vector0 -> vector2
					if bank == "Y" {
						bits |= 4
					}
					if memory {
						operand = 0x13 // (BX) -> vector2
					}
					if w {
						bits |= 0x80
					}
					visit(row.name, bank, memory, w, []byte{0xc4, 0xe1, bits, row.opcode, operand})
				}
			}
		}
	}
}

func TestDecodeX86DuplicateMoveVEXWIGCompleteShapes(t *testing.T) {
	for _, mode := range []int{32, 64} {
		count := 0
		x86DuplicateWIGWitnesses(func(name, bank string, memory, w bool, code []byte) {
			t.Run(fmt.Sprintf("mode%d/%s/%s/memory=%t/W=%t", mode, name, bank, memory, w), func(t *testing.T) {
				original := append([]byte(nil), code...)
				decoded, size, known, err := decodedX86DuplicateMoveInstruction(code, mode)
				if err != nil || !known || size != len(code) || decoded.Op != Op(name) || len(decoded.Args) != 2 ||
					decoded.Args[1].Kind != OpReg || decoded.Args[1].Reg != Reg(bank+"2") || !bytes.Equal(code, original) {
					t.Fatal("official WIG axis changed decoded operation/width/identity", decoded, size, known, err)
				}
				if memory {
					if decoded.Args[0].Kind != OpMem || decoded.Args[0].Mem != (MemRef{Base: BX}) {
						t.Fatal("WIG memory source changed its exact original address", decoded.Args[0])
					}
				} else if decoded.Args[0].Kind != OpReg || decoded.Args[0].Reg != Reg(bank+"0") {
					t.Fatal("WIG register source changed its original bank", decoded.Args[0])
				}
			})
			count++
		})
		if count != 24 {
			t.Fatal("complete VEX WIG width/address axis changed", count)
		}
	}
}

func x86DuplicateWIGOriginalSource() string {
	var source strings.Builder
	index := 0
	x86DuplicateWIGWitnesses(func(_, _ string, _, _ bool, code []byte) {
		fmt.Fprintf(&source, "TEXT ·WIG%d(SB),4,$0-0\n", index)
		for _, value := range code {
			fmt.Fprintf(&source, "BYTE $%d\n", value)
		}
		source.WriteString("RET\n")
		index++
	})
	return source.String()
}

func TestTranslateX86DuplicateMoveVEXWIGOriginalFormatsKeepContext(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			source := x86DuplicateWIGOriginalSource()
			file, signatures := x86UnboundVoidFormatSources(t, source, target.arch, target.triple)
			normalized, err := normalizeX86RawFile(file, target.arch)
			if err != nil {
				t.Fatal("valid WIG bytes could not be normalized without granting raw inputs", err)
			}
			ir := x86DuplicateMoveFormatObjectUnit(t, normalized, signatures, target.arch, target.triple)
			for _, body := range []string{ir, x86DwordProductIROptimize(t, ir)} {
				compileLLVMToObject(t, llc, target.triple, "duplicate-wig.ll", "duplicate-wig.o", body)
			}
		})
	}
}

func TestX86DuplicateMoveVEXWIGIndependentLLVM22Disassembly(t *testing.T) {
	llc, disassembler := findLLVM22Tool("llc"), findLLVM22Tool("llvm-objdump")
	if llc == "" || disassembler == "" {
		t.Fatal("LLVM22 llc and llvm-objdump required")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			var ir strings.Builder
			fmt.Fprintf(&ir, "target triple = %q\n", target.triple)
			index := 0
			x86DuplicateWIGWitnesses(func(_, _ string, _, _ bool, code []byte) {
				var operands []string
				for _, value := range code {
					operands = append(operands, fmt.Sprintf("%d", value))
				}
				fmt.Fprintf(&ir, "define void @WIG%d() {\n call void asm sideeffect %q, \"\"()\n ret void\n}\n",
					index, ".byte "+strings.Join(operands, ","))
				index++
			})
			directory := t.TempDir()
			input, object := filepath.Join(directory, "bytes.ll"), filepath.Join(directory, "bytes.o")
			if err := os.WriteFile(input, []byte(ir.String()), 0600); err != nil {
				t.Fatal(err)
			}
			stage, cancel := runtimeOracleStageContext(t)
			output, err := runtimeOracleCombinedOutput(stage, exec.CommandContext(stage, llc,
				"-O2", "-mtriple="+target.triple, "-filetype=obj", input, "-o", object))
			cancel()
			if err != nil {
				t.Fatalf("exact WIG bytes object failed: %v\n%s", err, output)
			}
			stage, cancel = runtimeOracleStageContext(t)
			output, err = runtimeOracleCombinedOutput(stage, exec.CommandContext(stage, disassembler, "-d", object))
			cancel()
			if err != nil {
				t.Fatalf("independent WIG disassembly failed: %v\n%s", err, output)
			}
			for _, name := range []string{"vmovddup", "vmovshdup", "vmovsldup"} {
				if count := len(regexp.MustCompile(`(?m)\b`+name+`[ \t]+`).FindAll(output, -1)); count != 8 {
					t.Fatalf("LLVM22 did not independently decode all eight width/address/W forms for %s: %d\n%s", name, count, output)
				}
			}
		})
	}
}
