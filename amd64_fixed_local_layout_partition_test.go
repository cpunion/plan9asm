package plan9asm

import (
	"errors"
	"fmt"
	"reflect"
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

func TestGoAMD64FixedLocalLayoutPartitionResidueAliases(t *testing.T) {
	body := strings.Replace(fixedLocalLayoutPingPong, "LEAQ 16(SP),R14",
		"MOVQ $77,16(SP)\nLEAQ 16(SP),R14\nANDQ $-16,R14", 1)
	body = strings.Replace(body, "MOVQ AX,ret+16(FP)",
		"ADDQ 16(SP),AX\nMOVQ AX,ret+16(FP)", 1)
	_, _, sig, source := x86ResultInitializationFixture(t, "amd64",
		"func Value(p *uint64, n uint64) uint64", "MOVQ n+8(FP),CX\n"+body)
	source = strings.Replace(source, "$0-24", "$64-24", 1)
	goAssemblerOracleOutput(t, "amd64", source, "", true)
	file, err := Parse(ArchAMD64, source)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := proveX86FixedLocal(file.Funcs[0], sig); !errors.Is(err, ErrProbeNeedsContext) ||
		!strings.Contains(err.Error(), "byte identity depends") {
		t.Fatal("layout alternatives lost the canary's residue-dependent alias", err)
	}
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
	var layouts x86FixedLocalLayoutGraph
	err := replayX86FixedLocal(&fn, sig, proof, 0, &budget, nil, &layouts)
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

func TestGoAMD64FixedLocalLayoutPartitionFrozenEdges(t *testing.T) {
	assembly, _ := fixedLocalLayoutPartitionProgram()
	file, err := Parse(ArchAMD64, assembly)
	if err != nil {
		t.Fatal(err)
	}
	_, sig := fixedLocalReplayFixture(t, "")
	fn := file.Funcs[0]
	proof := &x86FixedLocalProof{frame: fn.FrameSize}
	budget := x86FixedLocalBudget{steps: 1 << 18, bytes: 64 << 20}
	var layouts x86FixedLocalLayoutGraph
	if err := replayX86FixedLocal(&fn, sig, proof, 0, &budget, nil, &layouts); err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []string{"missing edge", "different block", "extra node", "other layout"} {
		t.Run(mutation, func(t *testing.T) {
			changed := layouts
			changed.nodes = append([]x86FixedLocalLayoutNode(nil), layouts.nodes...)
			switch mutation {
			case "missing edge":
				changed.nodes[0].targets = [2]int{-1, -1}
			case "different block":
				changed.nodes[0].end++
			case "extra node":
				changed.nodes = append(changed.nodes, x86FixedLocalLayoutNode{targets: [2]int{-1, -1}})
			case "other layout":
				changedEdge := false
				for index, node := range changed.nodes {
					for edge, target := range node.targets {
						for other, candidate := range changed.nodes {
							if !changedEdge && target >= 0 && target != other && candidate.pc == changed.nodes[target].pc {
								changed.nodes[index].targets[edge], changedEdge = other, true
							}
						}
					}
				}
				if !changedEdge {
					t.Fatal("fixture has no separate private layouts at the same source PC")
				}
			}
			before := append([]Instr(nil), fn.Instrs...)
			for index := range before {
				before[index].x86FixedLocalMemory = cloneX86ReplayFixedLocalMemory(before[index].x86FixedLocalMemory)
			}
			budget := x86FixedLocalBudget{steps: 1 << 18, bytes: 64 << 20}
			if err := replayX86FixedLocal(&fn, sig, proof, 1, &budget, nil, &changed); !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatal("a different residue acquired changed layout correspondence", err)
			}
			if !reflect.DeepEqual(before, fn.Instrs) {
				t.Fatal("failed layout replay published a provisional observation")
			}
		})
	}
}
