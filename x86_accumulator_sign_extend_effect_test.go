package plan9asm

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var x86AccumulatorSignExtendOracle = []struct {
	op, prefix, opcode string
	read, write        int64
	destination        Reg
}{
	{"CBW", "Pe", "0x98", 1, 2, AX},
	{"CWDE", "Px", "0x98", 2, 4, AX},
	{"CDQE", "Pw", "0x98", 4, 8, AX},
	{"CWD", "Pe", "0x99", 2, 2, DX},
	{"CDQ", "Px", "0x99", 4, 4, DX},
	{"CQO", "Pw", "0x99", 8, 8, DX},
}

func TestX86AccumulatorSignExtendCompleteEffectsAndGoRows(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(runtime.GOROOT(), "src/cmd/internal/obj/x86/asm6.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "var ynone = []ytab{\n\t{Zlit, 1, argList{}},\n}") {
		t.Fatal("complete original zero-operand encoder row changed")
	}
	for _, row := range x86AccumulatorSignExtendOracle {
		want := fmt.Sprintf("{A%s, ynone, %s, opBytes{%s}}", row.op, row.prefix, row.opcode)
		if !strings.Contains(string(data), want) {
			t.Fatal("independent Go opcode attributes changed", want)
		}
		for _, arch := range []string{"amd64", "386"} {
			t.Run(arch+"/"+row.op, func(t *testing.T) {
				ins := Instr{Op: Op(row.op), Raw: row.op}
				effect, known, err := x86TypedInstructionByteEffect(ins.Op, ins, arch)
				if arch == "386" && row.write == 8 {
					if !known || err == nil {
						t.Fatal("Go-rejected physical qword acquired a result effect", known, err)
					}
					return
				}
				if !known || err != nil || len(effect.operands) != 2 || !effect.requiresResultProof ||
					effect.flagsRead != 0 || effect.flagsDefined != 0 || effect.flagsKilled != 0 {
					t.Fatal("complete implicit AX dependency or preserved FLAGS is absent", effect, known, err)
				}
				read, write := effect.operands[0], effect.operands[1]
				if read.operand.Reg != AX || read.width != row.read || !read.read || read.write ||
					write.operand.Reg != row.destination || write.width != row.write || write.read || !write.write {
					t.Fatal("implicit source/destination byte roles diverged", effect)
				}
				for missing := int64(-1); missing < row.read; missing++ {
					state := x86ResultState{flags: x86ByteFlagsAll}
					state.gp[0] = uint8(1<<uint(row.read) - 1)
					if missing >= 0 {
						state.gp[0] &^= 1 << uint(missing)
					}
					if got := x86ResultTransfer(ins, &state, nil, 0, arch); got != (missing < 0) {
						t.Fatal("implicit AX bytes borrowed implementation initialization", missing, got)
					}
					if state.flags != x86ByteFlagsAll {
						t.Fatal("sign extension changed source FLAGS")
					}
				}
			})
		}
	}
}

func TestX86AccumulatorSignExtendVoidNeedsSource(t *testing.T) {
	for _, row := range x86AccumulatorSignExtendOracle {
		for _, target := range x86RawReturnTargets {
			if target.arch == "386" && row.write == 8 {
				continue // Actual Go mode rejection is checked separately above.
			}
			source := "TEXT ·Value(SB),4,$0-0\n" + row.op + "\nRET\n"
			for _, api := range []string{"text", "module", "Go"} {
				ir, err := x86ScalarFPBitViewAPI(t, source, "func Value()", target.arch, target.triple, api)
				if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
					t.Fatal("dead implicit input borrowed a zeroed register", row.op, target.triple, api, err)
				}
			}
		}
	}
}
