package plan9asm

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestX86WaitPKGCompleteGoRowsAndNUMInputs(t *testing.T) {
	table, err := os.ReadFile(filepath.Join(runtime.GOROOT(), "src/cmd/internal/obj/x86/asm6.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range []string{
		"var ywrfsbase = []ytab{\n\t{Zm_o, 2, argList{Yrl}},\n}",
		"{ATPAUSE, ywrfsbase, Pq, opBytes{0xae, 06}}",
		"{AUMWAIT, ywrfsbase, Pf2, opBytes{0xae, 06}}",
	} {
		if !strings.Contains(string(table), row) {
			t.Fatal("complete Go WAITPKG encoding changed", row)
		}
	}
	for _, arch := range []string{"386", "amd64"} {
		registers := []Reg{AX, CX, DX, BX, BP, SI, DI}
		if arch == "amd64" {
			registers = append(registers, "R8", "R9", "R10", "R11", "R12", "R13", "R14", "R15")
		}
		for _, op := range []string{"UMWAIT", "TPAUSE"} {
			for _, register := range registers {
				t.Run(arch+"/"+op+"/"+string(register), func(t *testing.T) {
					ins := x86ScalarLaneInstruction(t, op+" "+string(register))
					requireX86GoAssemblerResult(t, arch, "TEXT wait(SB),4,$0-0\n"+ins.Raw+"\nRET\n", true)
					before := x86ResultState{direction: x86ResultDirectionBackward}
					before.xmm[2], before.masks[1] = 0x1234, true
					before.valueStack = &x86ResultValueStack{bytes: []uint8{1, 2}, id: true}
					inputs := []int{0, 2}
					index, _, valid := x86ResultGP(register, 4, arch)
					if !valid {
						// AMD64 BP is a physical-frame role, not an ordinary NUM input.
						state := before.clone()
						if x86ResultTransfer(ins, &state, nil, 0, arch) {
							t.Fatal("physical frame acquired ordinary numerical inputs")
						}
						return
					}
					inputs = append(inputs, index)
					for _, input := range inputs {
						before.gp[input] = 0x0f
					}
					state, want := before.clone(), before.clone()
					want.flags = x86ByteFlagsAll
					if !x86ResultTransfer(ins, &state, nil, 0, arch) || !reflect.DeepEqual(state, want) {
						t.Error("wait must read EDX:EAX/control, define status and preserve unrelated state")
					}
					effect, known, err := x86TypedInstructionByteEffect(ins.Op, ins, arch)
					if !known || err != nil || !effect.requiresResultProof {
						t.Error("missing mandatory wait input descriptor", err)
					}
					for _, input := range inputs {
						for missing := uint(0); missing < 4; missing++ {
							state := before.clone()
							state.gp[input] &^= 1 << missing
							unchanged := state.clone()
							if x86ResultTransfer(ins, &state, nil, 0, arch) || !reflect.DeepEqual(state, unchanged) {
								t.Error("wait accepted missing data or partially updated status", input, missing)
							}
						}
					}
				})
			}
		}
	}
}

func TestX86WaitPKGVoidInputsThreeAPIs(t *testing.T) {
	x86WaitPKGVoidInputsThreeAPIs(t, false)
}

func TestX86RawWaitPKGVoidInputsThreeAPIs(t *testing.T) {
	x86WaitPKGVoidInputsThreeAPIs(t, true)
}

func x86WaitPKGBody(op string, raw bool) string {
	if !raw {
		return op + " BX\n"
	}
	prefix := byte(0xf2)
	if op == "TPAUSE" {
		prefix = 0x66
	}
	return x86RawDisplacementDirectives([]byte{prefix, 0x0f, 0xae, 0xf3})
}

func x86WaitPKGVoidInputsThreeAPIs(t *testing.T, raw bool) {
	t.Helper()
	for _, target := range x86RawReturnTargets {
		for _, op := range []string{"UMWAIT", "TPAUSE"} {
			for _, setup := range []struct {
				name, source string
				valid        bool
			}{
				{"none", "", false},
				{"deadline-only", "MOVL $0,AX\nMOVL $0,DX\n", false},
				{"no-ax", "MOVL $0,DX\nMOVL $0,BX\n", false},
				{"no-dx", "MOVL $0,AX\nMOVL $0,BX\n", false},
				{"partial-control", "MOVL $0,AX\nMOVL $0,DX\nMOVB $0,BL\n", false},
				{"complete", "MOVL $0,AX\nMOVL $0,DX\nMOVL $0,BX\n", true},
			} {
				t.Run(target.triple+"/"+op+"/"+setup.name, func(t *testing.T) {
					source := "TEXT ·Wait(SB),4,$0-0\n" + setup.source + x86WaitPKGBody(op, raw) + "RET\n"
					x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
					for _, api := range []string{"text", "module", "Go"} {
						ir, err := x86ScalarFPBitViewAPI(t, source, "func Wait()", target.arch, target.triple, api)
						if setup.valid {
							if err != nil || !strings.Contains(ir, fmt.Sprintf("asm sideeffect %q", strings.ToLower(op)+" $1; setc $0")) {
								t.Error("initialized wait lost its native opcode", api, err)
							}
						} else if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
							t.Error("Void wait borrowed uninitialized inputs", api, err)
						}
					}
				})
			}
		}
	}
}

// Execute no WAITPKG instruction on the host: these objects retain actual
// native operations, their flags consumers and all original argument/result
// bytes. Native CPU availability and execution are separate requirements.
func TestX86WaitPKGInitializedStatusFiveTargets(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		word, value, offset := "L", "uint32", 12
		if target.arch == "amd64" {
			word, value, offset = "Q", "uint64", 16
		}
		for _, raw := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/raw=%t", target.triple, raw), func(t *testing.T) {
				var source, declarations strings.Builder
				for _, op := range []string{"UMWAIT", "TPAUSE"} {
					fmt.Fprintf(&declarations, "func %s(control,low,high uint32) %s\n", op, value)
					fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-%d\nMOVL control+0(FP),BX\nMOVL low+4(FP),AX\nMOVL high+8(FP),DX\nPUSH%s $0x8d7\nPOPF%s\n%sPUSHF%s\nPOP%s CX\nMOV%s CX,ret+%d(FP)\nRET\n",
						op, offset+int(x86ResultPointerWord(target.arch)), word, word, x86WaitPKGBody(op, raw), word, word, word, offset)
				}
				x86LEAActualTargetGoObject(t, target.arch, target.triple, source.String())
				for _, api := range []string{"text", "module", "Go"} {
					ir, err := x86ScalarFPBitViewAPI(t, source.String(), declarations.String(), target.arch, target.triple, api)
					if err != nil || ir == "" {
						t.Error("initialized wait/status program rejected", api, err)
						continue
					}
					for _, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
						for _, level := range []string{"-O0", "-O2"} {
							x86PackedStringCompileObject(t, llc, target.triple, text, level)
						}
					}
				}
			})
		}
	}
}
