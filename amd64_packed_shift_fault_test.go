package plan9asm

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func TestCrossLinuxRuntimeMatrixX86PackedShiftNativeReadFaults(t *testing.T) {
	crossRosetta := runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" && rosettaAvailable()
	if runtime.GOARCH != "amd64" && !crossRosetta {
		t.Skip("requires Linux/amd64 memory-fault counterpart")
	}
	llc, clang, ok := findLlcAndClang(t)
	if !ok {
		t.Fatal("LLVM22 tools required")
	}
	var source, declarations, checks strings.Builder
	sigs := map[string]FuncSig{}
	for i, row := range []string{
		"VPSRLW.Z (CX),Y0,K1,Y1", "VPSRLD.Z (CX),Y0,K1,Y1", "VPSRLQ.Z (CX),Y0,K1,Y1",
		"VPSRLW.Z $1,(CX),K1,Y1", "VPSRLDQ $1,(CX),Y1",
		"VPSRLD.Z $1,(CX),K1,Y1", "VPSRLQ.BCST.Z $1,(CX),K1,Y1",
		"VPSRLVW.Z (CX),Y0,K1,Y1", "VPSRLVD.BCST.Z (CX),Y0,K1,Y1",
		"VPSRLD $255,(CX),Y1", "VPSRLVD (CX),Y0,Y1", "VPSRLVD.Z (CX),Y0,K1,Y1",
	} {
		name := fmt.Sprintf("faultshift%d", i)
		mask := 0
		if i == 11 {
			mask = 1
		}
		fmt.Fprintf(&source, "TEXT %s(SB),4,$0-24\nMOVQ out+0(FP),AX\nMOVQ data+8(FP),BX\nMOVQ count+16(FP),CX\nVMOVUPS (BX),Y0\nVPXOR Y0,Y0,Y0\nMOVQ $%d,DX\nKMOVQ DX,K1\n%s\nVMOVUPS Y1,(AX)\nRET\n", name, mask, row)
		sigs[name] = FuncSig{Name: name, Args: []LLVMType{Ptr, Ptr, Ptr}, Ret: Void, Frame: FrameLayout{Params: []FrameSlot{
			{Offset: 0, Type: Ptr, Index: 0, Field: -1}, {Offset: 8, Type: Ptr, Index: 1, Field: -1}, {Offset: 16, Type: Ptr, Index: 2, Field: -1},
		}}}
		fmt.Fprintf(&declarations, "extern void %s(uint8_t*,const uint8_t*,const uint8_t*);\n", name)
		accessible := 8
		if i == 11 {
			accessible = 3
		}
		fmt.Fprintf(&checks, "if(check_fault(%s,%d,%d,page)) return %d;\n", name, x86ShiftBoolean(i < 5 || i >= 9), accessible, i+10)
	}
	goAssemblerOracleOutput(t, "amd64", source.String(), "", true)
	file, err := Parse(ArchAMD64, source.String())
	if err != nil {
		t.Fatal(err)
	}
	triple := testTargetTriple(runtime.GOOS, "amd64")
	ir, err := Translate(file, Options{Goarch: "amd64", TargetTriple: triple, Sigs: sigs})
	if err != nil {
		t.Fatal(err)
	}
	const helper = `
#include <stdint.h>
#include <string.h>
#include <sys/mman.h>
#include <sys/resource.h>
#include <sys/wait.h>
#include <unistd.h>
#include <signal.h>
#include <stdio.h>
typedef void (*fault_fn)(uint8_t*,const uint8_t*,const uint8_t*);
static int check_fault(fault_fn fn,int should_fault,int accessible,size_t page) {
 uint8_t *area=mmap(0,page*2,PROT_READ|PROT_WRITE,MAP_ANON|MAP_PRIVATE,-1,0);
 if(area==MAP_FAILED || mprotect(area+page,page,PROT_NONE)) return 1;
 uint8_t data[32],out[34]; memset(data,0x71,32); memset(out,0xa7,34);
 memset(area+page-8,0,8);
 pid_t child=fork(); if(child<0) return 2;
 if(!child) {
  struct rlimit no_core={0,0};
  if(setrlimit(RLIMIT_CORE,&no_core)) _exit(120);

  const uint8_t *address=should_fault?area+page-accessible:area+page;
  fn(out+1,data,address);
  if(out[0]!=0xa7 || out[33]!=0xa7) _exit(3);
  for(int i=1;i<=32;++i) if(out[i]) _exit(4);
  _exit(0);
 }
 int status=0; if(waitpid(child,&status,0)!=child) return 5;
 if(munmap(area,page*2)) return 6;
 if(should_fault) {
  int memory_signal=WIFSIGNALED(status) && WTERMSIG(status)==SIGSEGV;
#if defined(__APPLE__)
  /* Rosetta's guarded-page diagnostic can report the host's SIGBUS. */
  memory_signal=memory_signal || (WIFSIGNALED(status) && WTERMSIG(status)==SIGBUS);
#endif
  if(!memory_signal) { fprintf(stderr,"expected full-read memory fault, status=%x\n",status); return 7; }
 } else if(!WIFEXITED(status) || WEXITSTATUS(status)!=0) { fprintf(stderr,"expected masked fault suppression, status=%x\n",status); return 8; }
 return 0;
}
`
	var prefix []string
	if crossRosetta {
		prefix = []string{"/usr/bin/arch", "-x86_64"}
	}
	driver := helper + declarations.String() + "int main(void) { size_t page=(size_t)sysconf(_SC_PAGESIZE);\n" + checks.String() + "return 0; }\n"
	compileAndRunRuntimeTestForTarget(t, llc, clang, "packed_shift_memory_faults", triple, ir, driver, prefix)
	t.Logf("8 full/active native reads and 4 lane-mask suppression cases; Rosetta=%t", crossRosetta)
}
