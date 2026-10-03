package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

func TestCrossLinuxRuntimeMatrixX86PackedUnpackZeroMaskStillFaults(t *testing.T) {
	llc, clang, triple, prefix := x86FloatAddSubRuntimeTools(t)
	var assembly, declarations, calls strings.Builder
	sigs := map[string]FuncSig{}
	for index, row := range []struct {
		instruction string
		bytes       int
	}{
		{"VPUNPCKLBW.Z (BX),Z0,K1,Z1", 64},
		{"VPUNPCKHWD.Z (BX),Y0,K1,Y1", 32},
		{"VUNPCKHPS.BCST.Z (BX),X0,K1,X1", 4},
		{"VUNPCKHPD.BCST.Z (BX),Z0,K1,Z1", 8},
	} {
		name := fmt.Sprintf("unpackfault%d", index)
		fmt.Fprintf(&assembly, "TEXT %s(SB),4,$0-16\nMOVQ input+0(FP),BX\nMOVQ out+8(FP),AX\nVPXORD Z0,Z0,Z0\nMOVL $0,CX\nKMOVQ CX,K1\n%s\nVMOVUPS Z1,(AX)\nRET\n", name, row.instruction)
		sigs[name] = FuncSig{Name: name, Args: []LLVMType{Ptr, Ptr}, Ret: Void, Frame: FrameLayout{Params: []FrameSlot{
			{Offset: 0, Type: Ptr, Index: 0, Field: -1}, {Offset: 8, Type: Ptr, Index: 1, Field: -1},
		}}}
		fmt.Fprintf(&declarations, "extern void %s(const uint8_t*,uint8_t*);\n", name)
		fmt.Fprintf(&calls, "if(check_fault(%s,0) || check_fault(%s,boundary-%d+1)) return 1;\n", name, name, row.bytes)
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
#include <stdio.h>
#include <signal.h>
#include <sys/mman.h>
#include <sys/resource.h>
#include <sys/wait.h>
#include <unistd.h>
` + declarations.String() + `
static int check_fault(void (*program)(const uint8_t*,uint8_t*),const uint8_t *input) {
 pid_t child=fork();
 if(child<0) return 1;
 if(child==0) {
  struct rlimit limit={0,0};
  if(setrlimit(RLIMIT_CORE,&limit)) _exit(2);
  uint8_t output[64];
  program(input,output);
  _exit(0);
 }
 int status=0;
 if(waitpid(child,&status,0)!=child) return 1;
 if(!WIFSIGNALED(status) || (WTERMSIG(status)!=SIGSEGV && WTERMSIG(status)!=SIGBUS)) {
  fprintf(stderr,"E4NF zero-mask input %p failed to fault: status=%d\n",input,status);
  return 1;
 }
 return 0;
}
int main(void) {
 long page=sysconf(_SC_PAGESIZE);
 if(page<=0) return 1;
 uint8_t *memory=mmap(0,(size_t)page*2,PROT_READ|PROT_WRITE,MAP_PRIVATE|MAP_ANON,-1,0);
 if(memory==MAP_FAILED || mprotect(memory+page,(size_t)page,PROT_NONE)) return 1;
 uint8_t *boundary=memory+page;
` + calls.String() + `
 if(munmap(memory,(size_t)page*2)) return 1;
 return 0;
}
`
	compileAndRunRuntimeTestForTarget(t, llc, clang, "unpack-e4nf-fault", triple, ir, main, prefix)
}
