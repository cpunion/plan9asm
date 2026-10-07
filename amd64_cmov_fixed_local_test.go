package plan9asm

import (
	"errors"
	"fmt"
	"go/types"
	"strings"
	"testing"

	"golang.org/x/arch/x86/x86asm"
)

var x86CMOVAUTOSources = []string{"tmp(SP)", "UNKNOWN(SP)", "tmp+0(SP)", "tmp+(0*4)(SP)", "tmp-8(SP)", "tmp-(2*4)(SP)"}

func TestX86CMOVFamilyAUTOActualGoPhysicalRoles(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		for _, width := range []string{"W", "L", "Q"} {
			if arch == "386" && width == "Q" {
				continue
			}
			for _, source := range x86CMOVAUTOSources {
				t.Run(arch+"/"+width+"/"+source, func(t *testing.T) {
					code := x86OrderingHintGoBytes(t, arch, "TEXT original(SB),4,$16-0\nCMOV"+width+"LE "+source+",AX\nRET\n")
					mode := 64
					if arch == "386" {
						mode = 32
					}
					found := 0
					for len(code) > 0 {
						ins, err := x86asm.Decode(code, mode)
						if err != nil || ins.Len == 0 {
							t.Fatal("independent actual Go object decode", err)
						}
						if ins.Op == x86asm.CMOVLE {
							memory, ok := ins.Args[1].(x86asm.Mem)
							want := int64(16)
							if strings.Contains(source, "-") {
								want = 8
							}
							if !ok || memory.Disp != want || memory.Index != 0 || ins.MemBytes != map[string]int{"W": 2, "L": 4, "Q": 8}[width] {
								t.Fatal("NAME_AUTO is frame+offset, not virtual/native zero", source, ins)
							}
							found++
						}
						code = code[ins.Len:]
					}
					if found != 1 {
						t.Fatal("actual Go did not emit one CMOV", found)
					}
				})
			}
		}
	}
}

func TestX86CMOVFamilyAUTOGrammarThreeAPIFiveTargets(t *testing.T) {
	for _, target := range x86OrderingHintTargets {
		for _, source := range x86CMOVAUTOSources {
			var assembly strings.Builder
			for _, name := range x86CMOVOracleNames(target.arch) {
				fmt.Fprintf(&assembly, "%s %s,AX\n", name, source)
				file, err := Parse(ArchAMD64, "TEXT Value(SB),4,$16-0\n"+name+" "+source+",AX\nRET\n")
				if err != nil {
					t.Fatal(err)
				}
				ins := file.Funcs[0].Instrs[1]
				t.Run(target.triple+"/"+name+"/"+source, func(t *testing.T) {
					if _, known, err := x86ConditionalMoveProperties(ins.Op, ins, target.arch); !known || err != nil {
						t.Fatal("accepted NAME_AUTO descriptive offset is not an undefined macro", err)
					}
				})
			}
			original := "TEXT ·Value(SB),4,$16-0\n" + assembly.String() + "RET\n"
			goAssemblerOracleOutput(t, target.arch, original, "", true)
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(target.triple+"/"+source+"/"+api, func(t *testing.T) {
					_, err := x86OrderingHintTranslateAPI(t, api, original, "package ordering\nfunc Value()\n", target.arch, target.triple)
					if !errors.Is(err, ErrProbeNeedsContext) {
						t.Fatal("legal grammar is not original native storage authority", err)
					}
				})
			}
		}
	}
}

func x86CMOVFixedLocalAssembly(name, op, source string) string {
	shift, restore := "", ""
	if source != "AX" && source != "8(SP)" {
		shift, restore = "MOVQ SP,R11\nADDQ $-16,SP\n", "MOVQ R11,SP\n"
	}
	return fmt.Sprintf("TEXT ·%s(SB),4,$16-24\nMOVQ a+0(FP),AX\nMOVQ b+8(FP),BX\nMOVQ AX,0(SP)\nMOVQ AX,8(SP)\n%sCMPQ AX,BX\n%s %s,BX\n%sMOVQ BX,ret+16(FP)\nRET\n", name, shift, op, source, restore)
}

func x86CMOVFixedLocalFixture(t *testing.T, assembly, declarations, target, triple string) (GoPackage, GoAMD64FrameCaptureRequest, *File, Options) {
	t.Helper()
	pkg, request := goAMD64FrameTypedFixture(t, declarations, assembly)
	request.Target = target
	file, err := Parse(ArchAMD64, assembly)
	if err != nil {
		t.Fatal(err)
	}
	sigs := map[string]FuncSig{}
	for _, name := range pkg.Types.Scope().Names() {
		sig, err := DeriveGoFuncABI0Signature(pkg.Path+"."+name, pkg.Types.Scope().Lookup(name).(*types.Func), "amd64", types.SizesFor("gc", "amd64"))
		if err != nil {
			t.Fatal(err)
		}
		sigs[sig.Name] = sig
	}
	return pkg, request, file, Options{Goarch: "amd64", TargetTriple: triple, Sigs: sigs,
		ResolveSym: func(symbol string) string { return strings.Replace(symbol, "·", pkg.Path+".", 1) }}
}

func TestX86CMOVFamilyFixedLocalAUTOActualBoundWholeFamily(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 required")
	}
	var assembly, declarations strings.Builder
	count := 0
	for _, op := range x86CMOVOracleNames("amd64") {
		for _, source := range append([]string{"AX", "8(SP)"}, x86CMOVAUTOSources...) {
			name := fmt.Sprintf("F%d", count)
			count++
			assembly.WriteString(x86CMOVFixedLocalAssembly(name, op, source))
			fmt.Fprintf(&declarations, "func %s(a,b uint64) uint64\n", name)
		}
	}
	for _, target := range []struct{ name, triple string }{{"linux/amd64", "x86_64-unknown-linux-gnu"}, {"darwin/amd64", "x86_64-apple-darwin"}} {
		t.Run(target.name, func(t *testing.T) {
			pkg, request, file, options := x86CMOVFixedLocalFixture(t, assembly.String(), declarations.String(), target.name, target.triple)
			observer := requireGoAMD64FrameCapture(t, pkg, request)
			ctx, cancel := runtimeOracleStageContext(t)
			defer cancel()
			bound, err := observer.BindFixedLocal(ctx, pkg, file, options)
			if err != nil {
				t.Fatal("actual numeric CMOV source retains the existing bounded private local contract", err)
			}
			fixedLocalAPIs(t, pkg, file, bound, func(t *testing.T, ir string, err error) {
				if err != nil || !strings.Contains(ir, "fixed_local_native") {
					t.Fatal("exact sealed original local storage", err)
				}
				compileLLVMToObject(t, llc, target.triple, "cmov-fixed-local.ll", "cmov-fixed-local.o", ir)
			})
			t.Logf("%d complete actual-object condition/width/register/physical/AUTO kernels; NAME_AUTO source bias is not caller or saved-BP storage", count)
		})
	}
}

func TestX86CMOVFamilyFixedLocalEveryReadAndAuthorityNegatives(t *testing.T) {
	for _, op := range x86CMOVOracleNames("amd64") {
		base := x86CMOVFixedLocalAssembly("Value", op, "tmp(SP)")
		pkg := mustGoPackage(t, "cmovlocal", "package cmovlocal\nfunc Value(a,b uint64) uint64\n")
		sig := x86WordMultiplySignature(t, pkg, "Value")
		width := map[byte]int{'W': 2, 'L': 4, 'Q': 8}[op[4]]
		mutations := map[string]string{
			"unknown flags":          strings.Replace(base, "CMPQ AX,BX", "MOVQ AX,DX", 1),
			"old target pointer":     strings.Replace(base, "CMPQ AX,BX", "CMPQ AX,BX\nMOVQ SP,BX", 1),
			"g target":               strings.Replace(base, "tmp(SP),BX", "tmp(SP),R14", 1),
			"BP target":              strings.Replace(base, "tmp(SP),BX", "tmp(SP),BP", 1),
			"AUTO saved BP":          strings.Replace(base, "ADDQ $-16,SP", "ADDQ $0,SP", 1),
			"AUTO return":            strings.Replace(base, "tmp(SP)", "tmp+8(SP)", 1),
			"undefined expression":   strings.Replace(base, "tmp(SP)", "slot+UNKNOWN(SP)", 1),
			"AUTO address input":     strings.Replace(base, op+" tmp(SP),BX", "MOVQ $tmp(SP),AX\n"+op+" AX,BX", 1),
			"FP pointer permission":  strings.Replace(base, "tmp(SP)", "p+0(FP)", 1),
			"source private pointer": strings.Replace(base, "MOVQ AX,0(SP)", "MOVQ SP,0(SP)", 1),
		}
		for missing := 0; missing < width; missing++ {
			var stores, old strings.Builder
			for byteIndex := 0; byteIndex < width; byteIndex++ {
				if byteIndex != missing {
					fmt.Fprintf(&stores, "MOVB $1,%d(SP)\n", byteIndex)
				}
			}
			// Old destination reads are potential even when every flag seed
			// makes the predicate true. Byte-state is checked independently.
			old.WriteString("MOVQ b+8(FP),DX\n")
			mutations[fmt.Sprintf("source byte %d", missing)] = strings.Replace(base, "MOVQ AX,0(SP)\n", stores.String(), 1)
			changed, err := Parse(ArchAMD64, strings.Replace(base, "MOVQ b+8(FP),BX\n", old.String(), 1))
			if err != nil {
				t.Fatal(err)
			}
			t.Run(op+fmt.Sprintf("/old byte %d", missing), func(t *testing.T) {
				state := x86FixedLocalState{pointers: map[Reg]int64{SP: -16}, known: map[Reg]uint8{AX: 255, BX: 255 &^ (1 << uint(missing))}, bytes: make([]bool, 16), flags: true}
				for index := range state.bytes {
					state.bytes[index] = true
				}
				for _, ins := range changed.Funcs[0].Instrs {
					if ins.Op == Op(op) {
						if err := fixedLocalStep(&ins, &state, sig, &x86FixedLocalProof{frame: 16}, 0); !errors.Is(err, ErrProbeNeedsContext) {
							t.Fatal("missing old byte became a defined potential input", err)
						}
					}
				}
			})
		}
		for name, source := range mutations {
			t.Run(op+"/"+name, func(t *testing.T) {
				file, err := Parse(ArchAMD64, source)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := proveX86FixedLocal(file.Funcs[0], sig); !errors.Is(err, ErrProbeNeedsContext) {
					t.Fatal("fixed local memory/flags/pointer/frame authority grew", err)
				}
			})
		}
	}
}
