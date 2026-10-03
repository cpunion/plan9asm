package plan9asm

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

func x86SignedIntegerFloatOptimize(t *testing.T, ir string) string {
	t.Helper()
	ctx := llvm.NewContext()
	defer ctx.Dispose()
	module, err := parseIRModuleInContext(ctx, ir)
	if !module.IsNil() {
		defer module.Dispose()
	}
	if err != nil {
		t.Fatal(err)
	}
	options := llvm.NewPassBuilderOptions()
	defer options.Dispose()
	if err := module.RunPasses("default<O2>", llvm.TargetMachine{}, options); err != nil {
		t.Fatalf("LLVM22 O2 signed conversion: %v", err)
	}
	return module.String()
}

// Independent hardware SSE: no production map, LLVM conversion or language
// cast computes the expected bits or exception flags. Save/restore prevents
// each oracle sample from changing the tested caller's ambient control word.
const x86SignedIntegerFloatSSEOracle = `#include <stdint.h>
#include <string.h>
#include <stdio.h>
static uint32_t sse_csr;
__attribute__((noinline)) static uint64_t sse_convert(uint64_t input,int sourceBits,int resultBits,unsigned csr) {
 unsigned saved;__asm__ volatile("stmxcsr %0":"=m"(saved));
 __asm__ volatile("ldmxcsr %0"::"m"(csr):"memory");
 uint64_t result=0;
 if(resultBits==32) {
  float value;
  if(sourceBits==32){int32_t integer=(int32_t)input;__asm__ volatile("cvtsi2ssl %1,%0":"=x"(value):"r"(integer));}
  else {int64_t integer=(int64_t)input;__asm__ volatile("cvtsi2ssq %1,%0":"=x"(value):"r"(integer));}
  uint32_t bits;memcpy(&bits,&value,4);result=bits;
 } else {
  double value;
  if(sourceBits==32){int32_t integer=(int32_t)input;__asm__ volatile("cvtsi2sdl %1,%0":"=x"(value):"r"(integer));}
  else {int64_t integer=(int64_t)input;__asm__ volatile("cvtsi2sdq %1,%0":"=x"(value):"r"(integer));}
  memcpy(&result,&value,8);
 }
 __asm__ volatile("stmxcsr %0":"=m"(sse_csr));
 __asm__ volatile("ldmxcsr %0"::"m"(saved):"memory");return result;
}
`

func TestCrossLinuxRuntimeMatrixX86SignedIntegerScalarFloatCompleteSemantics(t *testing.T) {
	llc, clang, triple, prefix := x86FloatAddSubRuntimeTools(t)
	rows := []string{}
	for _, row := range x86SignedIntegerFloatRows("amd64") {
		if strings.Contains(row, "(SI*") {
			continue
		}
		row = strings.Replace(row, "AX,", "R8,", 1)
		row = strings.Replace(row, "8(BX)", "128(BX)", 1)
		rows = append(rows, row)
		if strings.HasPrefix(row, "V") && !strings.Contains(row, "(BX)") {
			rows = append(rows, strings.Replace(row, "X0,X1", "X0,X0", 1), strings.Replace(row, "X0,X1", "X1,X0", 1))
		}
	}
	var source, declarations, checks strings.Builder
	sigs := map[string]FuncSig{}
	for index, row := range rows {
		file, _ := Parse(ArchAMD64, "TEXT f(SB),4,$0-0\n"+row+"\nRET\n")
		ins := file.Funcs[0].Instrs[1]
		f, _, err := x86SignedIntegerFloatProperties(ins, "amd64")
		if err != nil {
			t.Fatal(err)
		}
		mode := -1
		for i, suffix := range []string{"RN_SAE", "RD_SAE", "RU_SAE", "RZ_SAE"} {
			if f.rounding == suffix {
				mode = i
			}
		}
		name := fmt.Sprintf("convert%d", index)
		fmt.Fprintf(&source, "TEXT %s(SB),4,$0-28\nMOVQ out+0(FP),AX\nMOVQ value+8(FP),R8\nMOVQ pass+16(FP),BX\nLDMXCSR csr+24(FP)\nVMOVUPS (BX),Z0\nVMOVUPS 64(BX),Z1\nSTC\n%s\nVMOVUPS %s,(AX)\nSTMXCSR 64(AX)\nSETCS 68(AX)\nRET\n", name, row, x86CallerVectorRoot(ins.Args[f.destination].Reg))
		sigs[name] = FuncSig{Args: []LLVMType{Ptr, I64, Ptr, I32}, Ret: Void, Frame: FrameLayout{Params: []FrameSlot{
			{Offset: 0, Type: Ptr, Index: 0, Field: -1}, {Offset: 8, Type: I64, Index: 1, Field: -1}, {Offset: 16, Type: Ptr, Index: 2, Field: -1}, {Offset: 24, Type: I32, Index: 3, Field: -1},
		}}}
		fmt.Fprintf(&declarations, "extern void %s(void*,uint64_t,const void*,uint32_t);\n", name)
		passBank := 0
		if ins.Args[f.upper].Reg == "X1" {
			passBank = 64
		}
		fmt.Fprintf(&checks, "if(check(%s,%d,%d,%d,%d,%d))return %d;\n", name, f.sourceBits, f.spec.resultBits, x86FloatAddSubBoolInt(f.spec.legacy), passBank, mode, index+10)
	}
	goAssemblerOracleOutput(t, "amd64", source.String(), "", true)
	file, err := Parse(ArchAMD64, source.String())
	if err != nil {
		t.Fatal(err)
	}
	ir, err := Translate(file, Options{Goarch: "amd64", TargetTriple: triple, Sigs: sigs})
	if err != nil {
		t.Fatal(err)
	}
	const check = `typedef void(*convert)(void*,uint64_t,const void*,uint32_t);
__attribute__((noinline)) static int check(convert fn,int sourceBits,int resultBits,int legacy,int bank,int mode) {
 const uint64_t values[]={0,1,UINT64_MAX,2,UINT64_C(0xffffffff80000000),UINT64_C(0x7fffffff),
 UINT64_C(0x80000000),UINT64_C(0x100000001),UINT64_C(0x1000001),UINT64_C(0xfffffffffeffffff),
 UINT64_C(0x8000000000000000),UINT64_C(0x7fffffffffffffff),UINT64_C(0x20000000000001),
 UINT64_C(0xffdfffffffffffff),UINT64_C(0x4000002000000001)};
 const uint32_t controls[]={0x1f80,0x1fc0,0x9f80,0x9fc0,0x1f85,0x9fe0};
 for(unsigned sample=0;sample<256;sample++)for(unsigned ambient=0;ambient<4;ambient++)for(unsigned c=0;c<6;c++) {
  uint64_t input=sample<15?values[sample]:UINT64_C(0xd31733a6287fbbc3)*sample+UINT64_C(0x92c481352)
   +((uint64_t)sample<<52);
  uint8_t pass[138],copy[138],out[71],want[64];
  memset(pass,0xa7,sizeof(pass));for(unsigned i=0;i<128;i++)pass[i+1]=(uint8_t)(i*37+sample*71);
  memcpy(pass+129,&input,8);memcpy(copy,pass,sizeof(pass));memset(out,0xc3,sizeof(out));memset(want,0,sizeof(want));
  unsigned csr=controls[c]|(ambient<<13),operationCSR=(csr&~0x6000u)|((mode<0?ambient:(unsigned)mode)<<13);
  uint64_t result=sse_convert(input,sourceBits,resultBits,operationCSR);
  unsigned expectedCSR=mode<0?sse_csr:csr;
  memcpy(want,pass+1+bank,legacy?64:16);memcpy(want,&result,resultBits/8);
  fn(out+1,input,pass+1,csr);
  unsigned after,observed;__asm__ volatile("stmxcsr %0":"=m"(after));memcpy(&observed,out+65,4);
  if(memcmp(out+1,want,64)||observed!=expectedCSR||after!=expectedCSR) {
   fprintf(stderr,"signed source/result/legacy/bank/mode=%d/%d/%d/%d/%d input=%llx ambient=%u csr=%x observed=%x want=%x\n",
    sourceBits,resultBits,legacy,bank,mode,(unsigned long long)input,ambient,csr,observed,expectedCSR);return 1;
  }
  if(out[0]!=0xc3||out[69]!=1||out[70]!=0xc3||memcmp(pass,copy,sizeof(pass)))return 2;
 }
 return 0;
}
`
	main := x86SignedIntegerFloatSSEOracle + check + declarations.String() + "int main(void){\n" + checks.String() + "unsigned csr=0x1f80;__asm__ volatile(\"ldmxcsr %0\"::\"m\"(csr));return 0;}\n"
	compileAndRunRuntimeTestForTarget(t, llc, clang, "signed-integer-float-full", triple, x86SignedIntegerFloatOptimize(t, ir), main, prefix)
	t.Logf("%d original opcode/rounding/native-memory/alias programs x6144 signed/MXCSR/canary samples", len(rows))
}

func TestCrossLinuxRuntimeMatrixX86SignedIntegerScalarFloatNativeReadExtentAndPrecision(t *testing.T) {
	x86SignedIntegerScalarFloatFaultFixture(t, false)
}

func TestCrossLinuxRuntimeMatrixX86SignedIntegerScalarFloatNativeLinuxPrecisionSignal(t *testing.T) {
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		t.Skip("strict native Linux/amd64 precision-signal counterpart; Rosetta observation is not hardware-trap proof")
	}
	x86SignedIntegerScalarFloatFaultFixture(t, true)
}

func x86SignedIntegerScalarFloatFaultFixture(t *testing.T, requireTrap bool) {
	t.Helper()
	llc, clang, triple, prefix := x86FloatAddSubRuntimeTools(t)
	var source, declarations, checks strings.Builder
	sigs := map[string]FuncSig{}
	for index, op := range []string{"CVTSL2SS", "CVTSL2SD", "CVTSQ2SS", "CVTSQ2SD", "VCVTSI2SSL", "VCVTSI2SDL", "VCVTSI2SSQ", "VCVTSI2SDQ"} {
		name, args := fmt.Sprintf("read%d", index), "(BX),X0"
		if strings.HasPrefix(op, "V") {
			args = "(BX),X0,X1"
		}
		// Overwrite all results: the native memory read and precision trap
		// remain observable and cannot be dead-code-eliminated by LLVM O2.
		fmt.Fprintf(&source, "TEXT %s(SB),4,$0-8\nMOVQ input+0(FP),BX\n%s %s\nRET\n", name, op, args)
		sigs[name] = FuncSig{Args: []LLVMType{Ptr}, Ret: Void, Frame: FrameLayout{Params: []FrameSlot{{Offset: 0, Type: Ptr, Index: 0, Field: -1}}}}
		fmt.Fprintf(&declarations, "extern void %s(const void*);\n", name)
		width, resultBits := 4, 32
		if strings.Contains(op, "Q") {
			width = 8
		}
		if strings.Contains(op, "SD") {
			resultBits = 64
		}
		fmt.Fprintf(&checks, "if(check(%s,0,%d,%d,0x1f80)||check(%s,page+size-%d,%d,%d,0x1f80))return %d;\n", name, width, resultBits, name, width/2, width, resultBits, index+1)
		// Only the signed32→double conversion is always exact.
		fmt.Fprintf(&checks, "if(check(%s,&inexact,%d,%d,0x0f80))return %d;\n", name, width, resultBits, index+20)
	}
	goAssemblerOracleOutput(t, "amd64", source.String(), "", true)
	file, err := Parse(ArchAMD64, source.String())
	if err != nil {
		t.Fatal(err)
	}
	ir, err := Translate(file, Options{Goarch: "amd64", TargetTriple: triple, Sigs: sigs})
	if err != nil {
		t.Fatal(err)
	}
	const reference = `#include <stdint.h>
#include <stdio.h>
#include <signal.h>
#include <unistd.h>
#include <sys/mman.h>
#include <sys/wait.h>
typedef void(*readfn)(const void*);
static int check(readfn fn,const void*input,int width,int bits,unsigned csr){
 int statuses[2];
 for(unsigned original=0;original<2;original++){
  pid_t child=fork();if(child<0)return 1;
  if(child==0){__asm__ volatile("ldmxcsr %0"::"m"(csr):"memory");
   if(!original)fn(input);
   else if(bits==32){
    if(width==4)__asm__ volatile("cvtsi2ssl (%0),%%xmm0"::"r"(input):"xmm0","memory");
    else __asm__ volatile("cvtsi2ssq (%0),%%xmm0"::"r"(input):"xmm0","memory");
   } else {
    if(width==4)__asm__ volatile("cvtsi2sdl (%0),%%xmm0"::"r"(input):"xmm0","memory");
    else __asm__ volatile("cvtsi2sdq (%0),%%xmm0"::"r"(input):"xmm0","memory");
   }
   _exit(0);
  }
  if(waitpid(child,&statuses[original],0)!=child)return 2;
 }
 if(WIFSIGNALED(statuses[0])!=WIFSIGNALED(statuses[1]) ||
  (WIFSIGNALED(statuses[1])?WTERMSIG(statuses[0])!=WTERMSIG(statuses[1]):WEXITSTATUS(statuses[0])!=WEXITSTATUS(statuses[1]))) {
  fprintf(stderr,"native signed conversion input=%p width=%d result=%d csr=%x statuses=%x/%x\n",input,width,bits,csr,statuses[0],statuses[1]);return 1;
 }
 if(!input||csr==0x1f80){if(!WIFSIGNALED(statuses[1])||(WTERMSIG(statuses[1])!=SIGSEGV&&WTERMSIG(statuses[1])!=SIGBUS))return 3;}
 if(REQUIRE_TRAP&&csr==0x0f80&&(width==8||bits==32)){
  if(!WIFSIGNALED(statuses[1])||WTERMSIG(statuses[1])!=SIGFPE){fprintf(stderr,"original SSE missing precision SIGFPE status=%x\n",statuses[1]);return 4;}
 }
 return 0;
}
`
	main := fmt.Sprintf("#define REQUIRE_TRAP %d\n", x86FloatAddSubBoolInt(requireTrap)) + reference + declarations.String() + `int main(void){long size=sysconf(_SC_PAGESIZE);if(size<=0)return 4;
 char *page=mmap(0,size*2,PROT_READ|PROT_WRITE,MAP_PRIVATE|MAP_ANON,-1,0);
 if(page==MAP_FAILED||mprotect(page+size,size,PROT_NONE))return 5;
 uint64_t inexact=UINT64_C(0x20000001000001);
` + checks.String() + "munmap(page,size*2);return 0;}\n"
	compileAndRunRuntimeTestForTarget(t, llc, clang, "signed-native-fault", triple, x86SignedIntegerFloatOptimize(t, ir), main, prefix)
}
