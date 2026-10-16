package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func x86PackedFaultProgramGoDeclarations(signatures map[string]FuncSig, parameters string) string {
	var declarations strings.Builder
	for name := range signatures {
		fmt.Fprintf(&declarations, "func %s(%s)\n", name, parameters)
	}
	return declarations.String()
}

func x86PackedFaultProgramObjects(t *testing.T, source, declarations string) {
	t.Helper()
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		if target.arch != "amd64" {
			continue // These exact native source programs have AMD64-only named masks and ABI0 offsets.
		}
		t.Run(target.triple, func(t *testing.T) {
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, source, declarations, target.arch, target.triple, api)
					if err != nil {
						t.Fatal("original native fault program lacks a complete paired source proof", err)
					}
					for _, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
						for _, level := range []string{"-O0", "-O2"} {
							x86PackedStringCompileObject(t, llc, target.triple, text, level)
						}
					}
				})
			}
		})
	}
}

func TestX86PackedShiftOriginalFaultSourceThreeAPIsThreeTargetObjects(t *testing.T) {
	source, signatures, _, _, _ := x86PackedShiftFaultProgram()
	x86PackedFaultProgramObjects(t, source, x86PackedFaultProgramGoDeclarations(signatures, "out,data,count *byte"))
}

func TestX86PackedUnpackOriginalFaultSourceThreeAPIsThreeTargetObjects(t *testing.T) {
	source, signatures, _, _ := x86PackedUnpackFaultProgram()
	x86PackedFaultProgramObjects(t, source, x86PackedFaultProgramGoDeclarations(signatures, "input,out *byte"))
}

func TestX86PackedShiftUnpackVoidNativeInputsCannotBorrowShadowState(t *testing.T) {
	for _, row := range []string{
		"VPSRLW (AX),Y0,Y1", "VPSRLD $1,(AX),Y1",
		"VPSRLVD.BCST (AX),Y0,Y1", "VPSRLDQ $1,(AX),Y1",
		"VPUNPCKLBW (AX),Z0,Z1", "VUNPCKHPD.BCST (AX),Z0,Z1",
	} {
		for _, target := range x86RawReturnTargets {
			word, move := 8, "MOVQ"
			if target.arch == "386" {
				word, move = 4, "MOVL"
			}
			source := fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\n%s p+0(FP),AX\nVPXORD Z0,Z0,Z0\n%s\nRET\n", word, move, row)
			t.Run(target.triple+"/"+row, func(t *testing.T) {
				x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
				for _, api := range []string{"text", "module", "Go"} {
					if ir, err := x86ScalarFPBitViewAPI(t, source, "func Value(p *[64]byte)", target.arch, target.triple, api); err != nil || ir == "" {
						t.Fatal("complete independent native Void control rejected", api, err)
					}
					unknown := strings.Replace(source, move+" p+0(FP),AX\n", "", 1)
					if ir, err := x86ScalarFPBitViewAPI(t, unknown, "func Value(p *[64]byte)", target.arch, target.triple, api); ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
						t.Error("unobserved native input acquired an implementation address", api, err)
					}
				}
			})
		}
	}
}

func TestX86PackedUnpackLegacyMemoryRequiresOriginalNativeOpcode(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			var source strings.Builder
			signatures := map[string]FuncSig{}
			for _, op := range x86CallerUnpackOpcodes() {
				if strings.HasPrefix(op, "V") {
					continue
				}
				fmt.Fprintf(&source, "TEXT %s(SB),4,$0-0\n%s (AX),X1\nRET\n", op, op)
				signatures[op] = FuncSig{Name: op, Ret: Void}
			}
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source.String())
			file := mustParseX86RawReturn(t, source.String())
			ir := x86InstructionFormatObjectUnit(t, file, signatures, target.arch, target.triple, 0,
				func(c *amd64Ctx, ins Instr) (bool, bool, error) { return c.lowerPackedUnpack(ins.Op, ins) })
			for _, op := range x86CallerUnpackOpcodes() {
				if strings.HasPrefix(op, "V") {
					continue
				}
				mnemonic := strings.ToLower(op)
				mnemonic = strings.NewReplacer("lwl", "lwd", "hwl", "hwd", "llq", "ldq", "hlq", "hdq").Replace(mnemonic)
				if strings.Count(ir, `asm sideeffect "`+mnemonic+` ($2), $0", "=x,0,r,~{memory}"`) != 1 {
					t.Error("legacy m128 alignment/fault opcode became an align-1 load", op)
				}
			}
			for _, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
				for _, level := range []string{"-O0", "-O2"} {
					x86PackedStringCompileObject(t, llc, target.triple, text, level)
				}
			}
		})
	}
}
