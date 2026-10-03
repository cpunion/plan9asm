package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

func TestCrossLinuxRuntimeMatrixX86PackedUnpackMMXNativeMemoryAndFP(t *testing.T) {
	llc, clang, triple, prefix := x86FloatAddSubRuntimeTools(t)
	var assembly, declarations, calls strings.Builder
	sigs := map[string]FuncSig{}
	count := 0
	for _, opcode := range x86CallerUnpackOpcodes() {
		if !x86UnpackTestHasMMX(opcode) {
			continue
		}
		name, fpName := fmt.Sprintf("unpackmmx%d", count), fmt.Sprintf("unpackmmxfp%d", count)
		count++
		fmt.Fprintf(&assembly, "TEXT %s(SB),4,$0-24\nMOVQ input+0(FP),AX\nMOVQ second+8(FP),M0\n%s (AX),M0\nMOVQ M0,ret+16(FP)\nEMMS\nRET\n", name, opcode)
		sigs[name] = FuncSig{Name: name, Args: []LLVMType{Ptr, I64}, Ret: I64, Frame: FrameLayout{
			Params:  []FrameSlot{{Offset: 0, Type: Ptr, Index: 0, Field: -1}, {Offset: 8, Type: I64, Index: 1, Field: -1}},
			Results: []FrameSlot{{Offset: 16, Type: I64, Index: 0, Field: -1}},
		}}
		fmt.Fprintf(&assembly, "TEXT %s(SB),4,$0-16\nMOVQ $0,AX\nMOVQ AX,M0\n%s a+0(FP),M0\nMOVQ M0,ret+8(FP)\nEMMS\nRET\n", fpName, opcode)
		sigs[fpName] = FuncSig{Name: fpName, Args: []LLVMType{I32, I32}, Ret: I64, Frame: FrameLayout{
			Params:  []FrameSlot{{Offset: 0, Type: I32, Index: 0, Field: -1}, {Offset: 4, Type: I32, Index: 1, Field: -1}},
			Results: []FrameSlot{{Offset: 8, Type: I64, Index: 0, Field: -1}},
		}}
		fmt.Fprintf(&declarations, "extern uint64_t %s(const uint8_t*,uint64_t);\nextern uint64_t %s(uint32_t,uint32_t);\n", name, fpName)
		high, extent := 0, 4
		if strings.Contains(opcode, "UNPCKH") {
			high, extent = 4, 8
		}
		fmt.Fprintf(&calls, "if(checkMMX(%s,%s,%d,%d,%d,begin,end)) return 1;\n", name, fpName, x86UnpackTestLaneBytes(opcode), high, extent)
	}
	requireX86GoAssemblerResult(t, "amd64", assembly.String(), true)
	file, err := Parse(ArchAMD64, assembly.String())
	if err != nil {
		t.Fatal(err)
	}
	ir, err := Translate(file, Options{Goarch: "amd64", TargetTriple: triple, Sigs: sigs})
	if err != nil {
		t.Fatal(err)
	}
	main := `#include <stdint.h>
#include <string.h>
#include <signal.h>
#include <sys/mman.h>
#include <sys/resource.h>
#include <sys/wait.h>
#include <unistd.h>
` + declarations.String() + `
static uint64_t expectedMMX(const uint8_t *first,uint64_t second,int lane,int high) {
 uint8_t old[8],out[8];
 memcpy(old,&second,8);
 for(int i=0;i<8;++i) {
  int source=high+(i/(2*lane))*lane+i%lane;
  out[i]=(i/lane)%2 ? first[source] : old[source];
 }
 uint64_t value;
 memcpy(&value,out,8);
 return value;
}
static int checkMMX(uint64_t (*program)(const uint8_t*,uint64_t),
 uint64_t (*fp)(uint32_t,uint32_t),int lane,int high,int extent,uint8_t *begin,uint8_t *end) {
 for(int seed=0;seed<16;++seed) {
  uint8_t *first=end-extent,source[8];
  for(int i=0;i<8;++i) source[i]=(uint8_t)(seed*73+i*19+7);
  memcpy(first,source,extent);
  uint64_t second=UINT64_C(0xcafe11223344ba5e)^((uint64_t)seed*UINT64_C(0x1234567890));
  if(program(first,second)!=expectedMMX(source,second,lane,high) || memcmp(first,source,extent)) return 1;
  uint32_t a,b;
  memcpy(&a,source,4);
  memcpy(&b,source+4,4);
  if(fp(a,b)!=expectedMMX(source,0,lane,high)) return 1;
 }
 const uint8_t *bad[]={end-extent+1,high ? begin-4 : end-3};
 for(int i=0;i<2;++i) {
  pid_t child=fork();
  if(child<0) return 1;
  if(child==0) {
   struct rlimit limit={0,0};
   if(setrlimit(RLIMIT_CORE,&limit)) _exit(2);
   volatile uint64_t observed=program(bad[i],UINT64_C(0xabcdef1234567890));
   (void)observed;
   _exit(0);
  }
  int status=0;
  if(waitpid(child,&status,0)!=child || !WIFSIGNALED(status) ||
   (WTERMSIG(status)!=SIGSEGV && WTERMSIG(status)!=SIGBUS)) return 1;
 }
 return 0;
}
int main(void) {
 long page=sysconf(_SC_PAGESIZE);
 if(page<=0) return 1;
 uint8_t *memory=mmap(0,(size_t)page*3,PROT_READ|PROT_WRITE,MAP_PRIVATE|MAP_ANON,-1,0);
 if(memory==MAP_FAILED || mprotect(memory,(size_t)page,PROT_NONE) ||
  mprotect(memory+2*page,(size_t)page,PROT_NONE)) return 1;
 uint8_t *begin=memory+page,*end=memory+2*page;
` + calls.String() + `
 if(munmap(memory,(size_t)page*3)) return 1;
 return 0;
}
`
	compileAndRunRuntimeTestForTarget(t, llc, clang, "unpack-mmx-native-memory-fp", triple, ir, main, prefix)
}
