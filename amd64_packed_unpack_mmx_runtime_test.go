package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

func x86PackedUnpackMMXNativeProgram() (string, map[string]FuncSig, string, string, int) {
	var assembly, declarations, rows strings.Builder
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
		fmt.Fprintf(&rows, "{%s, %s, %d, %d, %d},\n", name, fpName, x86UnpackTestLaneBytes(opcode), high, extent)
	}
	return assembly.String(), sigs, declarations.String(), rows.String(), count
}

func TestCrossLinuxRuntimeMatrixX86PackedUnpackMMXNativeMemoryAndFP(t *testing.T) {
	execution := x86FloatCompareExecutionTools(t)
	assembly, sigs, declarations, rows, count := x86PackedUnpackMMXNativeProgram()
	requireX86GoAssemblerResult(t, "amd64", assembly, true)
	file, err := Parse(ArchAMD64, assembly)
	if err != nil {
		t.Fatal(err)
	}
	ir, err := Translate(file, Options{Goarch: "amd64", TargetTriple: execution.triple, Sigs: sigs})
	if err != nil {
		t.Fatal(err)
	}
	main := x86MemoryFaultSupportC + declarations + `
static const struct {
 uint64_t (*run)(const uint8_t*,uint64_t);
 uint64_t (*fp)(uint32_t,uint32_t);
 int lane, high, extent;
} cases[] = {
` + rows + `};
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
 uint64_t (*fp)(uint32_t,uint32_t),int lane,int high,int extent,uint8_t *end) {
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
 return 0;
}
int main(int argc, char **argv) {
    if (argc != 4) return 80;
    unsigned row = (unsigned)atoi(argv[1]);
    unsigned probe = (unsigned)atoi(argv[2]);
    unsigned original = (unsigned)atoi(argv[3]);
    if (row >= sizeof(cases)/sizeof(cases[0]) || probe > 3 || original > 1) return 80;
    int setup = memory_fault_setup(1);
    if (setup) return setup;
    uint8_t *end = guard + page;
    const uint8_t *input = NULL;
    if (probe == 1) input = end - cases[row].extent + 1;
    if (probe == 2) input = end - cases[row].extent;
    if (probe == 3) input = cases[row].high ? guard - 4 : end - 3;

    printf("memory enter %u %u %u\n", row, probe, original);
    if (fflush(stdout)) return 83;
    if (original) {
        const volatile uint8_t *source = input;
        volatile uint8_t sink = 0;
        for (int i = 0; i < cases[row].extent; ++i) sink ^= source[i];
    } else if (probe == 2) {
        if (checkMMX(cases[row].run, cases[row].fp, cases[row].lane,
                cases[row].high, cases[row].extent, end)) return 91;
    } else {
        volatile uint64_t observed = cases[row].run(input, UINT64_C(0xabcdef1234567890));
        (void)observed;
    }
    if (free_guard()) return 84;
    puts("memory complete");
    return fflush(stdout) ? 85 : 0;
}
`
	for _, level := range []string{"-O0", "-O2"} {
		t.Run(level, func(t *testing.T) {
			executable := execution.compileLLVM(t, "unpack-mmx-native-memory-fp", ir, main, level)
			x86CheckMemoryFaultMatrix(t, execution, executable, count)
			for row := 0; row < count; row++ {
				want := x86ObserveMemoryFault(t, execution, executable, row, 3, 1)
				if want != "segv" && want != "bus" {
					t.Fatalf("row %d: independent leading/short read did not fault: %s", row, want)
				}
				if got := x86ObserveMemoryFault(t, execution, executable, row, 3, 0); got != want {
					t.Fatalf("row %d: translated leading/short read %s, want %s", row, got, want)
				}
			}
		})
	}
}
