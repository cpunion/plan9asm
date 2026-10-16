package plan9asm

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"
)

const x86Legacy386OriginalStackAdjust = `TEXT stackAdjust(SB),NOSPLIT,$0-0
	PUSHFL
	ADJSP $156
	MOVL AX, 152(SP)
	ADJSP $-156
	POPFL
	RET
`

const x86Legacy386OriginalIDProbe = `TEXT cpuidProbe(SB),NOSPLIT,$0-0
	PUSHFL
	PUSHFL
	XORL $0x200000, 0(SP)
	POPFL
	PUSHFL
	POPL AX
	XORL 0(SP), AX
	POPFL
	TESTL $0x200000, AX
	JNE cpuidAvailable
	MOVL $0, AX
	RET
cpuidAvailable:
	MOVL $1, AX
	RET
`

// These separate programs have explicit inputs, initial FLAGS and result
// slots. Neither supplies native entry FLAGS or physical-SP backing to an
// original fixture. The original ADJSP and physical-memory forms stay in the
// complete emitter-only object below, not in these public execution claims.
const x86Legacy386InitializedStack = `TEXT stackAdjust(SB),4,$156-8
	PUSHL $2
	POPFL
	MOVL value+0(FP),AX
	MOVL AX,spill-4(SP)
	STC
	PUSHFL
	CLC
	POPFL
	SETCS CL
	MOVBLZX CL,CX
	SHLL $8,CX
	MOVL spill-4(SP),AX
	ANDL $255,AX
	ORL CX,AX
	MOVL AX,ret+4(FP)
	RET
`

const x86Legacy386InitializedIDProbe = `TEXT cpuidProbe(SB),4,$0-4
	PUSHL $2
	POPFL
	PUSHFL
	POPL AX
	MOVL AX,CX
	XORL $0x200000,AX
	PUSHL AX
	POPFL
	PUSHFL
	POPL AX
	PUSHL CX
	POPFL
	XORL CX,AX
	TESTL $0x200000,AX
	JNE cpuidAvailable
	MOVL $0,AX
	JMP cpuidDone
cpuidAvailable:
	MOVL $1,AX
cpuidDone:
	MOVL AX,ret+0(FP)
	RET
`

func x86Legacy386StackCounterparts(t *testing.T, original string) string {
	t.Helper()
	for _, row := range []struct{ original, initialized string }{
		{x86Legacy386OriginalStackAdjust, x86Legacy386InitializedStack},
		{x86Legacy386OriginalIDProbe, x86Legacy386InitializedIDProbe},
	} {
		if strings.Count(original, row.original) != 1 {
			t.Fatal("complete original stack fixture anchor changed")
		}
		original = strings.Replace(original, row.original, row.initialized, 1)
	}
	return original
}

// The full original functions, including ADJSP, both returns and the ID branch,
// are emitted as format-only objects. This bypasses no public source proof:
// the public APIs independently require Context for the originals below.
func x86Legacy386StackFormatObject(t *testing.T, triple string) string {
	t.Helper()
	file, err := Parse(ArchAMD64, x86Legacy386OriginalStackAdjust+x86Legacy386OriginalIDProbe)
	if err != nil || len(file.Funcs) != 2 {
		t.Fatal("complete original stack/ID format inventory changed", err)
	}
	var ir strings.Builder
	fmt.Fprintf(&ir, "target triple = %q\n", triple)
	emitAMD64Prelude(&ir, "386", file)
	for _, fn := range file.Funcs {
		ret := Void
		if fn.Sym == "cpuidProbe" {
			ret = I32
		}
		sig := FuncSig{Name: fn.Sym, Ret: ret}
		fmt.Fprintf(&ir, "define %s %s() {\n", ret, llvmGlobal(fn.Sym))
		consumer := newX86Ctx(&ir, fn, sig, func(name string) string { return name }, nil, "386", triple, false)
		if err := consumer.emitEntryAllocas(); err != nil {
			t.Fatal(err)
		}
		if err := consumer.lowerBlocks(); err != nil {
			t.Fatal("original source format was not emitted completely", fn.Sym, err)
		}
		ir.WriteString("}\n")
	}
	return ir.String()
}

func TestX86Legacy386StackFixturesKeepOriginalContextAndObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, triple := range []string{"i386-unknown-linux-gnu", "i686-pc-windows-msvc"} {
		t.Run(triple, func(t *testing.T) {
			for _, row := range []struct {
				name, source string
				ret          LLVMType
			}{
				{"stackAdjust", x86Legacy386OriginalStackAdjust, Void},
				{"cpuidProbe", x86Legacy386OriginalIDProbe, I32},
			} {
				x86LEAActualTargetGoObject(t, "386", triple, strings.ReplaceAll(row.source, ",NOSPLIT,", ",4,"))
				pkg := mustGoPackage(t, "legacy386", "package legacy386")
				for _, api := range []string{"text", "module", "Go"} {
					ir, err := x86ScalarFPBitViewPackageAPIWithManual(t, row.source, pkg, "386", triple, api,
						func(name string) (FuncSig, bool) {
							return FuncSig{Name: row.name, Ret: row.ret}, name == row.name
						})
					if !errors.Is(err, ErrProbeNeedsContext) || ir != "" {
						t.Fatal("original physical stack acquired a public entry/backing contract", row.name, api, err)
					}
				}
			}
			ir := x86Legacy386StackFormatObject(t, triple)
			for _, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
				for _, level := range []string{"-O0", "-O2"} {
					x86PackedStringCompileObject(t, llc, triple, text, level)
				}
			}
		})
	}
}

func TestX86Legacy386InitializedStackFixturesThreeAPIs(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	const declarations = "func stackAdjust(value uint32) uint32\nfunc cpuidProbe() uint32\n"
	const source = x86Legacy386InitializedStack + x86Legacy386InitializedIDProbe
	for _, triple := range []string{"i386-unknown-linux-gnu", "i686-pc-windows-msvc"} {
		t.Run(triple, func(t *testing.T) {
			x86LEAActualTargetGoObject(t, "386", triple, source)
			for _, api := range []string{"text", "module", "Go"} {
				ir, err := x86ScalarFPBitViewAPI(t, source, declarations, "386", triple, api)
				if err != nil || ir == "" {
					t.Fatal("initialized stack/ID program lost its complete source proof", api, err)
				}
				for _, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
					for _, level := range []string{"-O0", "-O2"} {
						x86PackedStringCompileObject(t, llc, triple, text, level)
					}
				}
			}
		})
	}
}

func TestX86Legacy386InitializedStackFixturesPortableValues(t *testing.T) {
	llc, clang, found := findLlcAndClang(t)
	if !found {
		t.Fatal("LLVM22 llc and clang required")
	}
	const driver = `#include <stdint.h>
#include <stdio.h>
extern uint32_t stackAdjust(uint32_t);
extern uint32_t cpuidProbe(void);
int main(void) {
    const uint32_t values[]={0,1,7,127,128,255,256,65535,0x12345678,UINT32_MAX};
    for(unsigned i=0;i<sizeof(values)/sizeof(values[0]);i++) {
        uint32_t v=values[i],want=(v&255)|256;
        uint32_t got=stackAdjust(v),id=cpuidProbe();
        if(got!=want||id!=1) {
            fprintf(stderr,"legacy386 case=%u value=%x got=%x want=%x id=%x\n",i,v,got,want,id);
            return 1;
        }
    }
    return 0;
}
`
	triple := testTargetTriple(runtime.GOOS, runtime.GOARCH)
	for _, api := range []string{"text", "module", "Go"} {
		t.Run(api, func(t *testing.T) {
			ir, err := x86ScalarFPBitViewAPI(t, x86Legacy386InitializedStack+x86Legacy386InitializedIDProbe,
				"func stackAdjust(value uint32) uint32\nfunc cpuidProbe() uint32\n", "386", triple, api)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(ir, "asm ") {
				t.Fatal("portable numerical stack test acquired native ISA execution")
			}
			for _, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
				for _, level := range []int{0, 2} {
					compileAndRunRuntimeTestWithCompiler(t, llc, []string{clang}, "legacy386-stack-values", triple, text, driver, nil, level)
				}
			}
		})
	}
}
