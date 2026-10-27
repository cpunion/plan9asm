package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func fixedLocalLayoutPartitionBodies() []string {
	three := strings.Replace(fixedLocalLayoutPingPong,
		"MOVQ $1,(R14)", "LEAQ 32(SP),R13\nMOVQ R14,R12\nMOVQ $1,(R14)", 1)
	three = strings.Replace(three, "XCHGQ R14,R15",
		"MOVQ R14,R12\nMOVQ R15,R14\nMOVQ R13,R15\nMOVQ R12,R13", 1)
	spill := strings.Replace(fixedLocalLayoutPingPong, "MOVQ $1,(R14)",
		"MOVQ R14,32(SP)\nMOVQ R15,40(SP)\nMOVQ $1,(R14)", 1)
	spill = strings.Replace(spill, "loop:\n", "loop:\nMOVQ 32(SP),R14\nMOVQ 40(SP),R15\n", 1)
	spill = strings.Replace(spill, "XCHGQ R14,R15",
		"MOVQ R15,32(SP)\nMOVQ R14,40(SP)", 1)
	return []string{fixedLocalLayoutPingPong, three, spill}
}

func fixedLocalLayoutPartitionProgram() (assembly, declarations string) {
	var asm, decl strings.Builder
	for index, body := range fixedLocalLayoutPartitionBodies() {
		body = strings.ReplaceAll(body, "ret+16(FP)", "ret+8(FP)")
		fmt.Fprintf(&asm, "TEXT ·Value%d(SB),4,$64-16\nMOVQ a+0(FP),CX\n%s", index, body)
		fmt.Fprintf(&decl, "func Value%d(a uint64) uint64\n", index)
	}
	return asm.String(), decl.String()
}

func TestGoAMD64FixedLocalLayoutPartitionPrograms(t *testing.T) {
	assembly, _ := fixedLocalLayoutPartitionProgram()
	goAssemblerOracleOutput(t, "amd64", assembly, "", true)
	file, err := Parse(ArchAMD64, assembly)
	if err != nil || len(file.Funcs) != 3 {
		t.Fatal("complete private-layout source inventory changed", err)
	}
	_, sig := fixedLocalReplayFixture(t, "")
	for index, fn := range file.Funcs {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			if _, err := proveX86FixedLocal(fn, sig); err != nil {
				t.Fatal("current initialized buffer lost its matching private layout", err)
			}
		})
	}
}

func TestGoAMD64FixedLocalLayoutPartitionObjects(t *testing.T) {
	assembly, declarations := fixedLocalLayoutPartitionProgram()
	fixedLocalCompleteVectorObjects(t, assembly, declarations, 3)
}

func TestGoAMD64FixedLocalLayoutPartitionStorageBound(t *testing.T) {
	var body strings.Builder
	for index := 0; index < 8; index++ {
		fmt.Fprintf(&body, "LEAQ %d(SP),R%d\n", index*8, index+8)
	}
	body.WriteString("loop:\nTESTQ $1,AX\nJZ rotate\nXCHGQ R8,R9\nJMP count\nrotate:\n")
	for index := 8; index < 15; index++ {
		fmt.Fprintf(&body, "XCHGQ R%d,R%d\n", index, index+1)
	}
	body.WriteString("count:\nSUBQ $1,AX\nJNZ loop\n")
	fn, sig := fixedLocalReplayFixture(t, body.String())
	proof := &x86FixedLocalProof{frame: fn.FrameSize}
	budget := x86FixedLocalBudget{steps: 1 << 18, bytes: 64 << 20}
	err := replayX86FixedLocal(&fn, sig, proof, 0, &budget, nil)
	if !errors.Is(err, ErrProbeNeedsContext) || !strings.Contains(err.Error(), "state storage budget") {
		t.Fatal("small frames with many private layouts exceeded retained storage", err)
	}
	if budget.steps <= 0 || budget.bytes <= 0 {
		t.Fatal("layout storage must remain independent of instruction/work bounds")
	}
	for _, ins := range fn.Instrs {
		if len(ins.x86FixedLocalMemory) != 0 {
			t.Fatal("layout storage failure published provisional accesses")
		}
	}
}
