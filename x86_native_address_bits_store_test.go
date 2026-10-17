package plan9asm

import (
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func x86NativeAddressBitsStoreSource(arch, operation string) (string, string) {
	word, move := 8, "MOVQ"
	if arch == "386" {
		word, move = 4, "MOVL"
	}
	source := fmt.Sprintf("TEXT ·CopyBits(SB),4,$0-%d\n%s dst+0(FP),DI\n%s src+%d(FP),SI\n%s SI,1(DI)\nMOVL $7,ret+%d(FP)\nRET\n",
		2*word+4, move, move, word, operation, 2*word)
	return source, "func CopyBits(dst,src *byte) uint32"
}

// Observe a complete native word's actual entry-address bits, not numerical
// GP initialization, pointee validity, a pointer result or GC/escape safety.
func TestX86NativeAddressBitsStoreFiveTargetThreeAPIObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		operations := []string{"MOVQ", "MOVD", "MOVDQ2Q"}
		if target.arch == "386" {
			operations = []string{"MOVL"}
		}
		for _, operation := range operations {
			t.Run(target.triple+"/"+operation, func(t *testing.T) {
				source, declaration := x86NativeAddressBitsStoreSource(target.arch, operation)
				x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
				for _, api := range []string{"text", "module", "Go"} {
					ir, err := x86ScalarFPBitViewAPI(t, source, declaration, target.arch, target.triple, api)
					if err != nil {
						t.Fatal("closed leaf cannot copy its actual full-word address bits", api, err)
					}
					for _, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
						for _, level := range []string{"-O0", "-O2"} {
							x86PackedStringCompileObject(t, llc, target.triple, text, level)
						}
					}
				}
			})
		}
	}
}

func TestX86NativeAddressBitsStorePortableValueAndCanaries(t *testing.T) {
	llc, clang, found := findLlcAndClang(t)
	if !found {
		t.Fatal("LLVM22 llc/clang required")
	}
	arch, move := "amd64", "MOVQ"
	if strconv.IntSize == 32 {
		arch, move = "386", "MOVL"
	}
	triple := testTargetTriple(runtime.GOOS, runtime.GOARCH)
	source, declaration := x86NativeAddressBitsStoreSource(arch, move)
	const driver = `#include <stdint.h>
#include <stdio.h>
#include <string.h>
extern uint32_t CopyBits(uint8_t*,const uint8_t*);
int main(void) {
    const uint8_t input[4]={17,23,31,47};
    const uint8_t *values[]={input,input+3,NULL};
    for (unsigned index=0;index<sizeof(values)/sizeof(values[0]);++index) {
        uint8_t output[sizeof(uintptr_t)+2];
        memset(output,0xa5,sizeof(output));
        if (CopyBits(output,values[index])!=7) return 1;
        uintptr_t observed;
        memcpy(&observed,output+1,sizeof(observed));
        if (observed!=(uintptr_t)values[index] || output[0]!=0xa5 || output[sizeof(output)-1]!=0xa5 ||
            input[0]!=17 || input[1]!=23 || input[2]!=31 || input[3]!=47) {
            fprintf(stderr,"opaque address-bit copy mismatch row=%u\n",index);
            return 1;
        }
    }
    return 0;
}

func TestX86NativeAddressBitsStoreVoidCannotHideUnknownInputs(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		word, move := 8, "MOVQ"
		if target.arch == "386" {
			word, move = 4, "MOVL"
		}
		for _, unusedBool := range []bool{false, true} {
			declaration, extent := "func CopyBits(dst,src *byte)", 2*word
			if unusedBool {
				declaration, extent = "func CopyBits(dst,src *byte,unused bool)", 2*word+1
			}
			body := fmt.Sprintf("%s dst+0(FP),DI\n%s src+%d(FP),SI\n%s SI,1(DI)\nRET\n", move, move, word, move)
			for _, row := range []struct{ name, body string }{
				{"unknown-source", strings.Replace(body, fmt.Sprintf("%s src+%d(FP),SI\n", move, word), "", 1)},
				{"unknown-destination", strings.Replace(body, move+" dst+0(FP),DI\n", "", 1)},
				{"partial-source", strings.Replace(body, move+" SI,1(DI)", "MOVW $0,SI\n"+move+" SI,1(DI)", 1)},
			} {
				source := fmt.Sprintf("TEXT ·CopyBits(SB),4,$0-%d\n%s", extent, row.body)
				for _, api := range []string{"text", "module", "Go"} {
					t.Run(fmt.Sprintf("%s/%s/unused-bool=%t/%s", target.triple, row.name, unusedBool, api), func(t *testing.T) {
						ir, err := x86ScalarFPBitViewAPI(t, source, declaration, target.arch, target.triple, api)
						if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
							t.Fatal("Void or an unused bool supplied missing native store input facts", err)
						}
					})
				}
			}
		}
	}
}
`
	for _, api := range []string{"text", "module", "Go"} {
		t.Run(api, func(t *testing.T) {
			ir, err := x86ScalarFPBitViewAPI(t, source, declaration, arch, triple, api)
			if err != nil {
				t.Fatal(err)
			}
			for _, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
				for _, level := range []int{0, 2} {
					compileAndRunRuntimeTestWithCompiler(t, llc, []string{clang}, "opaque-address-bit-copy", triple, text, driver, nil, level)
				}
			}
		})
	}
}
