package plan9asm

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func x86NativeLEAAddressBitsProgram(arch string) (string, string) {
	word, move, old := 8, "MOVQ", "0x1122334455667788"
	if arch == "386" {
		word, move, old = 4, "MOVL", "0x55667788"
	}
	var operations string
	for index, width := range []string{"W", "L", map[string]string{"amd64": "Q", "386": "L"}[arch]} {
		operations += fmt.Sprintf("%s $%s,AX\nLEA%s -7(SI)(DX*8),AX\n%s AX,%d(DI)\n", move, old, width, move, index*word)
	}
	source := fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\n%s out+0(FP),DI\n%s base+%d(FP),SI\n"+
		"MOVL index+%d(FP),DX\nMOVL $0x7fffffff,R8\nADDL $1,R8\nSTC\n%s"+
		"SETCS %d(DI)\nSETOS %d(DI)\nSETEQ %d(DI)\nSETMI %d(DI)\nRET\n",
		2*word+4, move, move, word, 2*word, operations, 3*word, 3*word+1, 3*word+2, 3*word+3)
	if arch == "386" {
		// The scratch is an ordinary physical GP in both source modes.
		source = strings.ReplaceAll(source, "R8", "BX")
	}
	return source, "func Value(out *[40]byte,base *byte,index uint32)\n"
}

func TestX86NativeLEAAddressBitsCompleteWidthsAndTargets(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("native LEA address-bit objects require LLVM22")
	}
	for _, target := range x86RawReturnTargets {
		source, declaration := x86NativeLEAAddressBitsProgram(target.arch)
		x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
		for _, api := range []string{"text", "module", "Go"} {
			t.Run(target.triple+"/"+api, func(t *testing.T) {
				ir, err := x86ScalarFPBitViewAPI(t, source, declaration, target.arch, target.triple, api)
				if err != nil || ir == "" {
					t.Fatal("initialized original LEA address observations rejected", err)
				}
				for _, pipeline := range []struct{ input, level string }{
					{ir, "-O0"}, {ir, "-O2"}, {x86DwordProductIROptimize(t, ir), "-O2"},
				} {
					x86PackedStringCompileObject(t, llc, target.triple, pipeline.input, pipeline.level)
				}
			})
		}
	}
}

func TestX86NativeLEAAddressBitsPortableValuesAndStatus(t *testing.T) {
	llc, clang, found := findLlcAndClang(t)
	if !found {
		t.Fatal("LLVM22 llc/clang required")
	}
	const driver = `#include <stdint.h>
#include <string.h>
extern void Value(uint8_t *,const uint8_t *,uint32_t);
int main(void) {
    uint8_t storage[8]={0};
    const uintptr_t bases[]={(uintptr_t)storage,0,UINTPTR_MAX-3,UINT32_MAX-3};
    const uint32_t indexes[]={0,1,UINT32_MAX};
    for (unsigned b=0;b<4;b++) for (unsigned i=0;i<3;i++) {
        uint8_t out[40]; memset(out,0xa5,sizeof(out));
        Value(out,(const uint8_t*)bases[b],indexes[i]);
        uintptr_t address=bases[b]+(uintptr_t)indexes[i]*8-7;
        uintptr_t actual[3]; memcpy(actual,out,sizeof(actual));
        uintptr_t old=sizeof(uintptr_t)==8 ? UINT64_C(0x1122334455667788) : UINT32_C(0x55667788);
        uintptr_t want16=(old&~(uintptr_t)0xffff)|(address&0xffff);
        if (actual[0]!=want16 || actual[1]!=(uint32_t)address || actual[2]!=address) return 1;
        unsigned flags=3*sizeof(uintptr_t);
        if (out[flags]!=1 || out[flags+1]!=1 || out[flags+2]!=0 || out[flags+3]!=1) return 2;
        for (unsigned j=flags+4;j<sizeof(out);j++) if (out[j]!=0xa5) return 3;
    }
    return 0;
}
`
	source, declaration := x86NativeLEAAddressBitsProgram("amd64")
	triple := testTargetTriple(runtime.GOOS, runtime.GOARCH)
	for _, api := range []string{"text", "module", "Go"} {
		t.Run(api, func(t *testing.T) {
			ir, err := x86ScalarFPBitViewAPI(t, source, declaration, "amd64", triple, api)
			if err != nil || ir == "" {
				t.Fatal("complete native-address LEA program rejected before portable value execution", err)
			}
			for _, pipeline := range []struct {
				input string
				level int
			}{
				{ir, 0}, {ir, 2}, {x86DwordProductIROptimize(t, ir), 2},
			} {
				compileAndRunRuntimeTestWithCompiler(t, llc, []string{clang}, "lea-address-bits", triple, pipeline.input, driver, nil, pipeline.level)
			}
		})
	}
}
