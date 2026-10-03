//go:build linux && (amd64 || arm64)

package plan9asm

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func x86PackedShiftLinuxExecution(t *testing.T) []string {
	t.Helper()
	if runtime.GOARCH == "amd64" {
		return nil // Required native Linux/amd64 counterpart, not Darwin Rosetta.
	}
	qemu, err := exec.LookPath("qemu-x86_64")
	if err != nil {
		t.Fatal("required explicit QEMU x86 execution:", err)
	}
	ctx, cancel := runtimeOracleStageContext(t)
	output, err := runtimeOracleCombinedOutput(ctx, exec.CommandContext(ctx, qemu, "--version"))
	cancel()
	if err != nil || !strings.HasPrefix(string(output), "qemu-x86_64 version 10.2.3 ") {
		t.Fatalf("required QEMU10.2.3: %v\n%s", err, output)
	}
	t.Logf("explicit executable transport: %s -L /usr/x86_64-linux-gnu", qemu)
	return []string{qemu, "-L", "/usr/x86_64-linux-gnu"}
}

func x86PackedShiftLegacyMemoryAssembly() (string, map[string]FuncSig) {
	var source strings.Builder
	sigs := map[string]FuncSig{}
	for i, operation := range x86LegacyShiftMemoryOpcodes {
		name := fmt.Sprintf("legacycount%d", i)
		fmt.Fprintf(&source, `TEXT %s(SB),4,$0-32
MOVQ out+0(FP),AX
MOVQ data+8(FP),BX
MOVQ count+16(FP),CX
MOVQ seed+24(FP),DX
MOVOU (BX),X0
CMPQ DX,$0
JE carry
CMPQ DX,$1
JE overflow
MOVQ $1,DX
CMPQ DX,$0
JMP shifted
carry:
MOVQ $-1,DX
ADDQ $1,DX
JMP shifted
overflow:
MOVQ $0x7fffffffffffffff,DX
ADDQ $1,DX
shifted:
%s (CX),X0
MOVOU X0,(AX)
SETCS 16(AX)
SETOS 17(AX)
SETEQ 18(AX)
SETMI 19(AX)
SETPS 20(AX)
RET
`, name, operation.name)
		sigs[name] = FuncSig{Name: name, Args: []LLVMType{Ptr, Ptr, Ptr, I64}, Ret: Void,
			Frame: FrameLayout{Params: []FrameSlot{
				{Offset: 0, Type: Ptr, Index: 0, Field: -1}, {Offset: 8, Type: Ptr, Index: 1, Field: -1},
				{Offset: 16, Type: Ptr, Index: 2, Field: -1}, {Offset: 24, Type: I64, Index: 3, Field: -1},
			}}}
	}
	return source.String(), sigs
}

const x86PackedShiftLegacyMemoryGoDriver = `package main
import (
 "os"
 "strconv"
 "unsafe"
)
func numeric(functions []func(unsafe.Pointer,unsafe.Pointer,unsafe.Pointer,uint64)) {
 limits:=[]uint64{0,1,7,15,16,17,31,32,33,63,64,65,255,256,1<<32,^uint64(0)}
 flags:=[3][5]byte{{1,0,1,0,1},{0,1,0,1,1},{0,0,0,0,0}}
 for kernel,fn:=range functions {
  for round,count:=range limits {
   for seed:=0;seed<3;seed++ {
    var data [18]byte
    var storage [96]byte
    var out [23]byte
    for i:=range data { data[i]=byte(i*71+round*97+13) }
    for i:=range storage { storage[i]=byte(i*31+7) }
    for i:=range out { out[i]=0xc3 }
    start:=uintptr(unsafe.Pointer(&storage[0]))
    offset:=int((16-start%16)%16)
    for i:=0;i<8;i++ { storage[offset+i]=byte(count>>(8*i)) }
    beforeData,beforeStorage:=data,storage
    fn(unsafe.Pointer(&out[1]),unsafe.Pointer(&data[1]),unsafe.Pointer(&storage[offset]),uint64(seed))
    if out[0]!=0xc3 || out[22]!=0xc3 || data!=beforeData || storage!=beforeStorage {
     panic("canary/input mutation")
    }
    for i:=0;i<5;i++ {
     if out[17+i]!=flags[seed][i] { panic("flags") }
    }
    bytes:=[]int{2,4,8,2,4,8}[kernel]
    for i:=0;i<16;i+=bytes {
     var value,want uint64
     for b:=0;b<bytes;b++ { value|=uint64(data[i+b+1])<<(8*b) }
     if count<uint64(bytes*8) {
      if kernel<3 { want=value<<count } else { want=value>>count }
     }
     for b:=0;b<bytes;b++ {
      if out[i+b+1]!=byte(want>>(8*b)) { panic("numeric") }
     }
    }
   }
  }
 }
}
func main() {
 functions:=[]func(unsafe.Pointer,unsafe.Pointer,unsafe.Pointer,uint64){legacycount0,legacycount1,legacycount2,legacycount3,legacycount4,legacycount5}
 if os.Args[1]=="numeric" { numeric(functions); return }
 index,err:=strconv.Atoi(os.Args[1])
 if err!=nil { panic(err) }
 var data,count [64]byte
 var out [23]byte
 for i:=range data {
  data[i]=byte(i*71+13)
  count[i]=byte(i*31+7)
 }
 start:=uintptr(unsafe.Pointer(&count[0]))
 offset:=(16-start%16)%16+1
 functions[index](unsafe.Pointer(&out[1]),unsafe.Pointer(&data[1]),unsafe.Pointer(&count[offset]),0)
 os.Exit(90)
}
`

func TestCrossLinuxRuntimeMatrixX86PackedShiftOriginalLegacyMemory(t *testing.T) {
	prefix := x86PackedShiftLinuxExecution(t)
	assembly, _ := x86PackedShiftLegacyMemoryAssembly()
	var declarations strings.Builder
	for i := range x86LegacyShiftMemoryOpcodes {
		fmt.Fprintf(&declarations, "func legacycount%d(out,data,count unsafe.Pointer,seed uint64)\n", i)
	}
	dir := t.TempDir()
	for name, source := range map[string]string{
		"go.mod":           "module originalcount\n\ngo 1.20\n",
		"main.go":          x86PackedShiftLegacyMemoryGoDriver + declarations.String(),
		"original_amd64.s": strings.ReplaceAll(assembly, "TEXT legacycount", "TEXT ·legacycount"),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	env := append(os.Environ(), "GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0", "GOWORK=off", "GOENV=off", "GOTOOLCHAIN=local")
	ctx, cancel := runtimeOracleStageContext(t)
	arguments := []string{"run", "-p=2"}
	if len(prefix) != 0 {
		arguments = append(arguments, "-exec="+strings.Join(prefix, " "))
	}
	arguments = append(arguments, ".", "numeric")
	command := exec.CommandContext(ctx, "go", arguments...)
	command.Dir, command.Env = dir, env
	output, err := runtimeOracleCombinedOutput(ctx, command)
	cancel()
	if err != nil {
		t.Fatalf("original Go aligned numeric/flags: %v\n%s", err, output)
	}
	executable := filepath.Join(dir, "original")
	ctx, cancel = runtimeOracleStageContext(t)
	command = exec.CommandContext(ctx, "go", "build", "-p=2", "-o", executable, ".")
	command.Dir, command.Env = dir, env
	output, err = runtimeOracleCombinedOutput(ctx, command)
	cancel()
	if err != nil {
		t.Fatalf("original Go body/link: %v\n%s", err, output)
	}
	for i := range x86LegacyShiftMemoryOpcodes {
		ctx, cancel = runtimeOracleStageContext(t)
		argv := append(append([]string(nil), prefix...), executable, strconv.Itoa(i))
		command = exec.CommandContext(ctx, argv[0], argv[1:]...)
		output, err = runtimeOracleCombinedOutput(ctx, command)
		cancel()
		var exit *exec.ExitError
		text := string(output)
		if !errors.As(err, &exit) || exit.ExitCode() != 2 || !strings.Contains(text, "main.legacycount") ||
			(!strings.Contains(text, "unexpected fault address") && !strings.Contains(text, "invalid memory address")) {
			t.Fatalf("original Go unaligned m128 count %d: %v\n%s", i, err, output)
		}
	}
	t.Log("6 original Go m128 kernels ×48 numeric/flags/canary inputs and 6 unaligned faults")
}

func TestCrossLinuxRuntimeMatrixX86PackedShiftLLVMAlignedMemory(t *testing.T) {
	prefix := x86PackedShiftLinuxExecution(t)
	llc, clang, ok := findLlcAndClang(t)
	if !ok {
		t.Fatal("LLVM22 tools required")
	}
	assembly, sigs := x86PackedShiftLegacyMemoryAssembly()
	goAssemblerOracleOutput(t, "amd64", assembly, "", true)
	file, err := Parse(ArchAMD64, assembly)
	if err != nil {
		t.Fatal(err)
	}
	const triple = "x86_64-unknown-linux-gnu"
	ir, err := Translate(file, Options{Goarch: "amd64", TargetTriple: triple, Sigs: sigs})
	if err != nil {
		t.Fatal(err)
	}
	var registrations strings.Builder
	for i := range x86LegacyShiftMemoryOpcodes {
		fmt.Fprintf(&registrations, "extern void legacycount%d(uint8_t*,const uint8_t*,const uint8_t*,uint64_t);\n", i)
	}
	registrations.WriteString("static legacy_fn kernels[]={legacycount0,legacycount1,legacycount2,legacycount3,legacycount4,legacycount5};\n")
	driver := x86PackedShiftLegacyCountCHeader + registrations.String() + x86PackedShiftLegacyCountCMain
	compileAndRunRuntimeTestWithCompiler(t, llc, x86PackedShiftLinuxCCompiler(t, clang), "legacy_count_contract", triple, ir, driver, prefix)
	t.Log("6 LLVM m128 kernels ×48 numeric/flags/canary inputs, 6 alignment and 6 guard-page faults")
}

func x86PackedShiftLinuxCCompiler(t *testing.T, clang string) []string {
	t.Helper()
	if runtime.GOARCH == "amd64" {
		return []string{clang}
	}
	gcc, err := exec.LookPath("x86_64-linux-gnu-gcc")
	if err != nil {
		t.Fatal("required AMD64 C/link driver:", err)
	}
	return []string{gcc, "-no-pie"}
}

func TestCrossLinuxRuntimeMatrixX86PackedShiftCompleteNumerical(t *testing.T) {
	prefix := x86PackedShiftLinuxExecution(t)
	llc, clang, ok := findLlcAndClang(t)
	if !ok {
		t.Fatal("LLVM22 tools required")
	}
	cases := x86PackedShiftNumericalCases()
	assembly, sigs := x86PackedShiftNumericalAssembly(cases)
	goAssemblerOracleOutput(t, "amd64", assembly, "", true)
	file, err := Parse(ArchAMD64, assembly)
	if err != nil {
		t.Fatal(err)
	}
	const triple = "x86_64-unknown-linux-gnu"
	ir, err := Translate(file, Options{Goarch: "amd64", TargetTriple: triple, Sigs: sigs})
	if err != nil {
		t.Fatal(err)
	}
	var driver strings.Builder
	driver.WriteString(x86PackedShiftIndependentC)
	for _, c := range cases {
		fmt.Fprintf(&driver, "extern void %s(uint8_t*,const uint8_t*,const uint8_t*,uint64_t,uint64_t);\n", c.name)
	}
	driver.WriteString("static const struct shift_case cases[]={\n")
	for _, c := range cases {
		fmt.Fprintf(&driver, "{%s,%q,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d},\n", c.name, c.row,
			c.width, c.bits, c.countMode, c.immediate, c.memory, c.maskMode, c.destination,
			x86ShiftBoolean(c.left), x86ShiftBoolean(c.legacy), x86ShiftBoolean(c.broadcast))
	}
	driver.WriteString(`};
int main(void) {
 for(unsigned i=0;i<sizeof(cases)/sizeof(cases[0]);i++) {
  if(check_case(&cases[i],(int)i)) return 1;
 }
 return 0;
}
`)
	compileAndRunRuntimeTestWithCompiler(t, llc, x86PackedShiftLinuxCCompiler(t, clang),
		"complete_packed_shift", triple, ir, driver.String(), prefix)
	t.Logf("%d exact LLVM programs ×192 independent whole-byte/flags/canary inputs", len(cases))
}

const x86PackedShiftLegacyCountCHeader = `
#include <stdint.h>
#include <string.h>
#include <stdio.h>
#include <sys/mman.h>
#include <sys/wait.h>
#include <unistd.h>
#include <signal.h>
typedef void (*legacy_fn)(uint8_t*,const uint8_t*,const uint8_t*,uint64_t);
`

const x86PackedShiftLegacyCountCMain = `
static uint64_t read_le(const uint8_t *p,int bytes) {
 uint64_t value=0;
 for(int i=0;i<bytes;i++) value|=(uint64_t)p[i]<<(8*i);
 return value;
}
int main(void) {
 const uint64_t counts[]={0,1,7,15,16,17,31,32,33,63,64,65,255,256,UINT64_C(1)<<32,UINT64_MAX};
 const uint8_t flags[3][5]={{1,0,1,0,1},{0,1,0,1,1},{0,0,0,0,0}};
 for(int k=0;k<6;k++) for(int round=0;round<16;round++) for(int seed=0;seed<3;seed++) {
  uint8_t data[18],out[23],beforeData[18],beforeCount[32];
  _Alignas(16) uint8_t count[32];
  for(int i=0;i<18;i++) data[i]=(uint8_t)(i*71+round*97+13);
  for(int i=0;i<32;i++) count[i]=(uint8_t)(i*31+7);
  for(int i=0;i<8;i++) count[i]=(uint8_t)(counts[round]>>(8*i));
  memcpy(beforeData,data,18); memcpy(beforeCount,count,32); memset(out,0xc3,23);
  kernels[k](out+1,data+1,count,seed);
  if(out[0]!=0xc3 || out[22]!=0xc3 || memcmp(data,beforeData,18) || memcmp(count,beforeCount,32) || memcmp(out+17,flags[seed],5)) return 10+k;
  int bytes=(int[]){2,4,8,2,4,8}[k];
  for(int i=0;i<16;i+=bytes) {
   uint64_t value=read_le(data+i+1,bytes),want=0;
   if(counts[round]<(uint64_t)(bytes*8)) want=k<3?value<<counts[round]:value>>counts[round];
   for(int b=0;b<bytes;b++) if(out[i+b+1]!=(uint8_t)(want>>(8*b))) return 20+k;
  }
 }
 size_t page=(size_t)sysconf(_SC_PAGESIZE);
 for(int k=0;k<6;k++) for(int edge=0;edge<2;edge++) {
  uint8_t *area=mmap(0,page*2,PROT_READ|PROT_WRITE,MAP_PRIVATE|MAP_ANON,-1,0);
  if(area==MAP_FAILED || mprotect(area+page,page,PROT_NONE)) return 30;
  memset(area,7,page);
  pid_t child=fork(); if(child<0) return 31;
  if(!child) {
   uint8_t data[18],out[23]; memset(data,13,18); memset(out,0xc3,23);
   kernels[k](out+1,data+1,edge?area+page-8:area+33,0);
   _exit(90);
  }
  int status=0;
  if(waitpid(child,&status,0)!=child || munmap(area,page*2)) return 32;
  if(!WIFSIGNALED(status) || WTERMSIG(status)!=SIGSEGV) {
   fprintf(stderr,"legacy kernel=%d edge=%d expected SIGSEGV, actual=%x\n",k,edge,status);
   return 40+k;
  }
 }
 return 0;
}
`
