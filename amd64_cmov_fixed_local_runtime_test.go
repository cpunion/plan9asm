package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

func TestCrossLinuxRuntimeMatrixX86CMOVFixedLocalAUTOWholeBytes(t *testing.T) {
	execution := x86FloatCompareExecutionTools(t)
	target := "linux/amd64"
	triple := "x86_64-unknown-linux-gnu"
	if strings.Contains(execution.triple, "apple") {
		target = "darwin/amd64"
		triple = "x86_64-apple-darwin"
	}
	var assembly, declarations, prototypes, rows, goRows strings.Builder
	count := 0
	prefix := ""
	if target == "darwin/amd64" {
		prefix = "_"
	}
	for _, width := range []string{"W", "L", "Q"} {
		for condition, row := range x86CMOVOracleConditions {
			for _, source := range append([]string{"AX", "8(SP)"}, x86CMOVAUTOSources...) {
				fn := fmt.Sprintf("F%d", count)
				count++
				assembly.WriteString(x86CMOVFixedLocalAssembly(fn, "CMOV"+width+row.name, source))
				fmt.Fprintf(&declarations, "func %s(a,b uint64) uint64\n", fn)
				fmt.Fprintf(&prototypes, "extern uint64_t %s(uint64_t,uint64_t) __asm__(\"%sexample.invalid/framefixture.%s\");\n", fn, prefix, fn)
				fmt.Fprintf(&rows, "{%s,%d,%d,0},\n", fn, map[string]int{"W": 2, "L": 4, "Q": 8}[width], condition)
				fmt.Fprintf(&goRows, "{%s,%d,%d,0},\n", fn, map[string]int{"W": 2, "L": 4, "Q": 8}[width], condition)
			}
		}
	}
	for mode, operation := range []string{"MOVQ tmp-8(SP),BX", "MOVQ AX,tmp-8(SP)\nMOVQ tmp-8(SP),BX", "MOVOU tmp-16(SP),X0\nMOVQ X0,BX"} {
		fn := fmt.Sprintf("F%d", count)
		count++
		fmt.Fprintf(&assembly, "TEXT ·%s(SB),4,$16-24\nMOVQ a+0(FP),AX\nMOVQ b+8(FP),BX\nMOVQ AX,0(SP)\nMOVQ BX,8(SP)\n%s\nMOVQ BX,ret+16(FP)\nRET\n", fn, operation)
		fmt.Fprintf(&declarations, "func %s(a,b uint64) uint64\n", fn)
		fmt.Fprintf(&prototypes, "extern uint64_t %s(uint64_t,uint64_t) __asm__(\"%sexample.invalid/framefixture.%s\");\n", fn, prefix, fn)
		fmt.Fprintf(&rows, "{%s,8,0,%d},\n", fn, mode+1)
		fmt.Fprintf(&goRows, "{%s,8,0,%d},\n", fn, mode+1)
	}
	pkg, request, file, options := x86CMOVFixedLocalFixture(t, assembly.String(), declarations.String(), target, triple)
	observer := requireGoAMD64FrameCapture(t, pkg, request)
	ctx, cancel := runtimeOracleStageContext(t)
	defer cancel()
	bound, err := observer.BindFixedLocal(ctx, pkg, file, options)
	if err != nil {
		t.Fatal(err)
	}
	ir, err := Translate(file, bound)
	if err != nil {
		t.Fatal(err)
	}
	reference := fmt.Sprintf(x86CMOVFixedLocalWholeBytesC, prototypes.String(), rows.String())
	for _, pipeline := range []struct{ name, backend, ir string }{
		{"backend_O0", "-O0", ir}, {"backend_O2", "-O2", ir}, {"IR_O2/backend_O2", "-O2", x86DwordProductIROptimize(t, ir)},
	} {
		t.Run(pipeline.name, func(t *testing.T) {
			output := execution.compileAndRunLLVM(t, "cmov-fixed-local-auto", pipeline.ir, reference, pipeline.backend)
			want := fmt.Sprintf("cmov/private GP/SIMD AUTO kernels %d whole-byte outcomes %d failures 0\n", count, count*64)
			if string(output) != want {
				t.Fatal("actual native AUTO displacement/predicate/writeback mismatch", string(output))
			}
			t.Log(strings.TrimSpace(string(output)))
		})
	}
	x86WordMultiplyRunOriginalGo(t, execution, "package oracle\n"+declarations.String(), assembly.String(), fmt.Sprintf(x86CMOVFixedLocalOriginalGo, goRows.String()))
	t.Logf("%d independently executed actual Go ABI0 local CMOV/GP/SIMD kernels ×64 full-width input pairs; host native/Rosetta is not strict Linux evidence", count)
}

const x86CMOVFixedLocalWholeBytesC = `
#include <stdint.h>
#include <stdio.h>
` + x86CMOVConditionC + `
%s
typedef struct {uint64_t(*fn)(uint64_t,uint64_t);unsigned width,condition,mode;} Kernel;
static Kernel kernels[]={%s};
int main(void) {
 const uint64_t edges[]={0,1,UINT64_MAX,UINT64_C(0x8000000000000000),UINT64_C(0x7fffffffffffffff),UINT64_C(0x87654321fedcba98),UINT64_C(0x80000000),UINT64_C(0x7fc123457f800001)};
 unsigned count=0;
 for(unsigned k=0;k<sizeof(kernels)/sizeof(kernels[0]);k++) for(unsigned x=0;x<8;x++) for(unsigned y=0;y<8;y++) {
  Kernel kernel=kernels[k];uint64_t a=edges[x],b=edges[y],difference=a-b;
  unsigned ones=0;for(unsigned bit=0;bit<8;bit++) ones+=(difference>>bit)&1;
  uint64_t flags=(a<b)|((uint64_t)(!(ones&1))<<2)|((uint64_t)(difference==0)<<6)|((difference>>63)<<7)|((((a^b)&(a^difference))>>63)<<11);
  uint64_t selected=predicate(kernel.condition,flags)?a:b;
  uint64_t mask=kernel.width==8?UINT64_MAX:((UINT64_C(1)<<(kernel.width*8))-1);
  uint64_t want=kernel.width==2?((b&~mask)|(selected&mask)):(selected&mask);
  if(kernel.mode) want=kernel.mode==1?b:a;
  uint64_t got=kernel.fn(a,b);
  if(got!=want) {fprintf(stderr,"kernel%%u a%%llx b%%llx got%%llx want%%llx\n",k,(unsigned long long)a,(unsigned long long)b,(unsigned long long)got,(unsigned long long)want);return 1;}
  count++;
 }
 printf("cmov/private GP/SIMD AUTO kernels %%zu whole-byte outcomes %%u failures 0\n",sizeof(kernels)/sizeof(kernels[0]),count);
 return 0;
}
`

const x86CMOVFixedLocalOriginalGo = `package oracle
import "testing"
func TestOriginal(t *testing.T) {
 kernels:=[]struct{fn func(uint64,uint64)uint64;width,condition,mode int}{%s}
 edges:=[]uint64{0,1,0xffffffffffffffff,0x8000000000000000,0x7fffffffffffffff,0x87654321fedcba98,0x80000000,0x7fc123457f800001}
 count:=0
 for k,kernel:=range kernels {for _,a:=range edges {for _,b:=range edges {
  d:=a-b;ones:=0;for bit:=uint(0);bit<8;bit++ {ones+=int(d>>bit&1)}
  cf,pf,zf,sf,of:=a<b,ones&1==0,d==0,d>>63!=0,((a^b)&(a^d))>>63!=0
  predicates:=[]bool{!cf,cf,zf,sf==of,!zf&&sf==of,!cf&&!zf,zf||sf!=of,cf||zf,sf!=of,sf,!zf,!of,of,!pf,!sf,pf}
  selected:=b;if predicates[kernel.condition] {selected=a}
  mask:=uint64(0xffffffffffffffff);if kernel.width<8 {mask=(uint64(1)<<uint(kernel.width*8))-1}
  want:=selected&mask;if kernel.width==2 {want|=b&^mask};if kernel.mode!=0 {want=a;if kernel.mode==1 {want=b}}
  if got:=kernel.fn(a,b);got!=want {t.Fatalf("kernel%%d a%%x b%%x got%%x want%%x",k,a,b,got,want)};count++
 }}}
 t.Logf("actual Go local AUTO whole-byte outcomes %%d failures 0",count)
}
`
