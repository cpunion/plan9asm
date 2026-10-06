package plan9asm

import (
	"fmt"
	"go/types"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

// The independent oracle describes the ISA, not production properties. Go's
// reversed operands put its second source in the low half of each 128-bit lane.
type x86NarrowOracleFamily struct {
	legacy, vector string
	inputBytes     int
	unsigned       bool
}

var x86NarrowOracleFamilies = []x86NarrowOracleFamily{
	{"PACKSSLW", "VPACKSSDW", 4, false},
	{"PACKSSWB", "VPACKSSWB", 2, false},
	{"PACKUSDW", "VPACKUSDW", 4, true},
	{"PACKUSWB", "VPACKUSWB", 2, true},
}

func x86NarrowOracleIR(t *testing.T, source, declarations, triple string) string {
	t.Helper()
	file, err := Parse(ArchAMD64, source)
	if err != nil {
		t.Fatal(err)
	}
	pkg := mustGoPackage(t, "narroworacle", declarations)
	sigs := make(map[string]FuncSig)
	for _, function := range file.Funcs {
		name := strings.TrimPrefix(function.Sym, "·")
		object, ok := pkg.Types.Scope().Lookup(name).(*types.Func)
		if !ok {
			t.Fatal("independent source lacks its declaration", name)
		}
		sig, err := DeriveGoFuncABI0Signature(name, object, "amd64", types.SizesFor("gc", "amd64"))
		if err != nil {
			t.Fatal(err)
		}
		sigs[name] = sig
	}
	context := llvm.NewContext()
	defer context.Dispose()
	module, err := TranslateModuleInContext(context, file, Options{
		Goarch: "amd64", TargetTriple: triple, Sigs: sigs,
		ResolveSym: func(symbol string) string { return strings.TrimPrefix(symbol, "·") },
	})
	if !module.IsNil() {
		defer module.Dispose()
	}
	if err != nil {
		t.Fatal(err)
	}
	return module.String()
}

// Intel SDM Vol. 2A, Table 2-50 and the PACK chapters classify these complete
// memory tuples as E4NF/E4NF.nb: a zero output mask does not suppress the read.
// No AVX/AVX512 original instruction is executed by this software LLVM oracle.
func TestCrossLinuxRuntimeMatrixX86PackedNarrowIndependentE4NF(t *testing.T) {
	execution := x86FloatCompareExecutionTools(t)
	var assembly, declarations, table strings.Builder
	declarations.WriteString("package narroworacle\n")
	count := 0
	for _, family := range x86NarrowOracleFamilies {
		for widthIndex, bank := range []string{"X", "Y", "Z"} {
			for _, broadcast := range []bool{false, true} {
				if broadcast && family.inputBytes != 4 {
					continue
				}
				for _, zero := range []bool{false, true} {
					name := fmt.Sprintf("fault%d", count)
					count++
					op := family.vector
					bytes := 16 << widthIndex
					if broadcast {
						op += ".BCST"
						bytes = 4
					}
					if zero {
						op += ".Z"
					}
					fmt.Fprintf(&declarations, "func %s(input *byte)\n", name)
					fmt.Fprintf(&assembly, "TEXT ·%s(SB),4,$0-8\nMOVQ input+0(FP),AX\nVPXORD %s0,%s0,%s0\nMOVL $0,CX\nKMOVQ CX,K1\n%s (AX),%s0,K1,%s1\nRET\n",
						name, bank, bank, bank, op, bank, bank)
					fmt.Fprintf(&table, "{%s,%d,\"%s %s\"},\n", name, bytes, op, bank)
				}
			}
		}
	}
	goAssemblerOracleOutput(t, "amd64", assembly.String(), "", true)
	ir := x86NarrowOracleIR(t, assembly.String(), declarations.String(), execution.triple)
	var main strings.Builder
	main.WriteString(`#include <stdint.h>
#include <stdio.h>
#include <signal.h>
#include <sys/mman.h>
#include <sys/resource.h>
#include <sys/wait.h>
#include <unistd.h>
`)
	for index := 0; index < count; index++ {
		fmt.Fprintf(&main, "extern void fault%d(const uint8_t*);\n", index)
	}
	main.WriteString("static const struct {void (*run)(const uint8_t*); int bytes; const char *name;} cases[]={\n" + table.String() + "};\n")
	main.WriteString(`
static int observe(void (*run)(const uint8_t*),const uint8_t *input,int fault,const char *name) {
 pid_t child=fork();
 if(child<0) return 1;
 if(child==0) {
  struct rlimit limit={0,0};
  if(setrlimit(RLIMIT_CORE,&limit)) _exit(2);
  run(input);
  _exit(0);
 }
 int status=0;
 if(waitpid(child,&status,0)!=child) return 1;
 int correct=fault ? WIFSIGNALED(status)&&(WTERMSIG(status)==SIGSEGV||WTERMSIG(status)==SIGBUS)
                   : WIFEXITED(status)&&WEXITSTATUS(status)==0;
 if(!correct) fprintf(stderr,"E4NF %s input=%p requiredFault=%d waitStatus=%d\n",name,input,fault,status);
 return !correct;
}
int main(void) {
 long page=sysconf(_SC_PAGESIZE);
 if(page<64) return 2;
 uint8_t *memory=mmap(0,(size_t)page*2,PROT_READ|PROT_WRITE,MAP_PRIVATE|MAP_ANON,-1,0);
 if(memory==MAP_FAILED) return 3;
 for(long i=0;i<page;++i) memory[i]=(uint8_t)(i*13+17);
 if(mprotect(memory+page,(size_t)page,PROT_NONE)) return 4;
 unsigned failures=0,valid=0,faults=0;
 for(unsigned row=0;row<sizeof(cases)/sizeof(cases[0]);++row) {
  failures+=observe(cases[row].run,memory+page-cases[row].bytes,0,cases[row].name); ++valid;
  failures+=observe(cases[row].run,0,1,cases[row].name); ++faults;
  failures+=observe(cases[row].run,memory+page-cases[row].bytes+1,1,cases[row].name); ++faults;
 }
 if(munmap(memory,(size_t)page*2)) return 5;
 printf("E4NF valid=%u faultProbes=%u mismatches=%u\n",valid,faults,failures);
 return failures?1:0;
}
`)
	for _, level := range []string{"-O0", "-O2"} {
		t.Run(level, func(t *testing.T) {
			output := execution.compileAndRunLLVM(t, "narrow-independent-e4nf", ir, main.String(), level)
			want := fmt.Sprintf("E4NF valid=%d faultProbes=%d mismatches=0\n", count, count*2)
			if string(output) != want {
				t.Fatalf("incomplete E4NF observer: %q want %q", output, want)
			}
			t.Log(strings.TrimSpace(string(output)))
		})
	}
}
