package plan9asm

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// The first minimum-position instruction fully defines its output. It must
// not manufacture a physical high-bank identity for the second instruction
// on mode32 merely because Go's named encoder accepts that spelling.
func x86BlendMinposBankProgram(arch string, registers []int) (string, string) {
	word, move := goWordSize(arch), "MOVQ"
	if arch == "386" {
		move = "MOVL"
	}
	var source, declarations strings.Builder
	for _, register := range registers {
		fmt.Fprintf(&source, "TEXT ·Bank%d(SB),4,$0-%d\n%s out+0(FP),DI\nPXOR X0,X0\n"+
			"VPHMINPOSUW X0,X%d\nVPHMINPOSUW X%d,X0\nMOVOU X0,(DI)\nRET\n",
			register, word, move, register, register)
		fmt.Fprintf(&declarations, "func Bank%d(out *[16]byte)\n", register)
	}
	return source.String(), declarations.String()
}

func TestX86BlendMinposMode32HighNamedNeedsPhysicalBinding(t *testing.T) {
	registers := []int{8, 9, 10, 11, 12, 13, 14, 15}
	complete, _ := x86BlendMinposBankProgram("386", registers)
	// A successful real Go object is only a frontend premise, not evidence
	// that physical mode32 has distinct X8-X15 registers.
	x86LEAActualTargetGoObject(t, "386", "i386-unknown-linux-gnu", complete)
	for _, register := range registers {
		t.Run(fmt.Sprintf("X%d", register), func(t *testing.T) {
			source, declarations := x86BlendMinposBankProgram("386", []int{register})
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, source, declarations, "386", "i386-unknown-linux-gnu", api)
					if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
						t.Fatal("initialized numerical bytes acquired an unbound physical mode32 vector identity", err)
					}
				})
			}
			for _, instruction := range []string{fmt.Sprintf("VPHMINPOSUW X0,X%d", register), fmt.Sprintf("VPHMINPOSUW X%d,X0", register)} {
				t.Run(instruction, func(t *testing.T) {
					ins := x86ScalarLaneInstruction(t, instruction)
					if _, known, err := x86BlendMinposProperties(ins.Op, ins, "386"); !known || !errors.Is(err, ErrProbeNeedsContext) {
						t.Fatal("typed grammar confused Go-accepted spelling with a physical mode32 binding", err)
					}
				})
			}
		})
	}
}

func TestX86BlendMinposSameHighNamedAMD64StillWorks(t *testing.T) {
	source, declarations := x86BlendMinposBankProgram("amd64", []int{8, 9, 10, 11, 12, 13, 14, 15})
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("same initialized AMD64 high bank requires LLVM22 object compilation")
	}
	for _, target := range x86AddressMemoryTargets("amd64") {
		t.Run(target, func(t *testing.T) {
			x86LEAActualTargetGoObject(t, "amd64", target, source)
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, source, declarations, "amd64", target, api)
					if err != nil || ir == "" {
						t.Fatal("physical AMD64 high bank regressed to mode32 Context", err)
					}
					compileLLVMToObject(t, llc, target, "high-bank.ll", "high-bank.o", ir)
				})
			}
		})
	}
}

func TestX86BlendMinposLowBankPhysicalAndInitializedFiveTargets(t *testing.T) {
	registers := []int{0, 1, 2, 3, 4, 5, 6, 7}
	var body strings.Builder
	for _, register := range registers {
		fmt.Fprintf(&body, "VPHMINPOSUW X0,X%d\nVPHMINPOSUW X%d,X0\n", register, register)
	}
	code, pcs := x86NumericalBoundaryGoBytes(t, "386", body.String()+"RET")
	if len(pcs) != 17 || len(code) == 0 || code[len(code)-1] != 0xc3 {
		t.Fatal("actual Go mode32 low-bank directory is incomplete")
	}
	for index, register := range registers {
		for direction, modrm := range []byte{0xc0 | byte(register)<<3, 0xc0 | byte(register)} {
			at := 2*index + direction
			want := []byte{0xc4, 0xe2, 0x79, 0x41, modrm}
			if !bytes.Equal(code[pcs[at]:pcs[at+1]], want) {
				t.Fatal("Go bytes differ from independent VEX.128.66.0F38.W0 NOVSR low-bank fields", register, direction)
			}
		}
	}
	x86HandoffOriginalBytes(t, "386", code)
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("initialized physical low-bank objects require LLVM22")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			source, declarations := x86BlendMinposBankProgram(target.arch, registers)
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, source, declarations, target.arch, target.triple, api)
					if err != nil || ir == "" {
						t.Fatal("initialized physically bound low-bank producer failed", err)
					}
					compileLLVMToObject(t, llc, target.triple, "low-bank.ll", "low-bank.o", ir)
				})
			}
		})
	}
}
