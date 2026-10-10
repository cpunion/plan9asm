package plan9asm

import (
	"fmt"
	"runtime"
	"testing"
)

// Independent first-party counterexample: integer CMP sets source NZCV=0110.
// MRS must read that source state, not whichever flags LLVM's own bookkeeping
// last left in the physical processor.
const arm64NZCVReadSource = `TEXT compareNZCV(SB),4,$0-16
MOVD in+0(FP),R0
MOVD out+8(FP),R1
MOVD (R0),R2
CMP R2,R2
MRS NZCV,R3
MOVD R3,(R1)
RET
`

func arm64NZCVReadRuntime(t *testing.T, triple, goos string, compiler, runner []string) {
	t.Helper()
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc not found")
	}
	original := `package main
import "fmt"
func compareNZCV(*uint64,*uint64)
func main() {
 for _, input := range []uint64{0,1,0x8000000000000000,0xffffffffffffffff} {
  var output uint64
  compareNZCV(&input,&output)
  if output != 0x60000000 { panic(fmt.Sprintf("original NZCV=%x",output)) }
 }
 fmt.Println("ARM64_COMPARE_ORIGINAL_PASS NZCV=0x60000000")
}
`
	arm64ScalarCompareActualGo(t, arm64NZCVReadSource, original, goos, runner)
	file, err := Parse(ArchARM64, arm64NZCVReadSource)
	if err != nil {
		t.Fatal(err)
	}
	ir, err := Translate(file, Options{Goarch: "arm64", TargetTriple: triple,
		Sigs: map[string]FuncSig{"compareNZCV": arm64SM3Signature("compareNZCV")}})
	if err != nil {
		t.Fatal(err)
	}
	driver := `#include <stdint.h>
#include <stdio.h>
extern void compareNZCV(uint64_t*,uint64_t*);
int main(void) {
 uint64_t inputs[]={0,1,UINT64_C(0x8000000000000000),UINT64_MAX};
 for(unsigned i=0;i<4;i++) {
  uint64_t output=0;
  compareNZCV(&inputs[i],&output);
  if(output!=0x60000000) {
   fprintf(stderr,"source NZCV input=%u got=%llx want=60000000\n",i,(unsigned long long)output);
   return 1;
  }
 }
 return 0;
}
`
	for _, variant := range []struct{ name, ir string }{{"none", ir}, {"O2", armFloatEffectOptimizedIR(t, ir)}} {
		for _, level := range []int{0, 2} {
			t.Run(fmt.Sprintf("IR_%s/backend_O%d", variant.name, level), func(t *testing.T) {
				compileAndRunRuntimeTestWithCompiler(t, llc, compiler, "nzcv_read", triple, variant.ir, driver, runner, level)
			})
		}
	}
}

func TestARM64NZCVReadOriginalGoNativeRuntime(t *testing.T) {
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		t.Skip("Darwin host counterpart; required Linux/QEMU owner runs the same original-Go oracle")
	}
	clang := findLLVM22Tool("clang")
	if clang == "" {
		t.Fatal("LLVM 22 clang not found")
	}
	arm64NZCVReadRuntime(t, "aarch64-apple-darwin", "darwin", []string{clang}, nil)
}
