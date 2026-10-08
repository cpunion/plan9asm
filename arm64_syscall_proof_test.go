package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func arm64SyscallProofFixture(t *testing.T) (*arm64Ctx, int, Instr) {
	t.Helper()
	source := arm64SyscallEffectSource("CMP R2,R1\nSVC\nCSET CS,R2")
	requireARM64GoAssemblerResult(t, source, true)
	file, err := Parse(ArchARM64, source)
	if err != nil || len(file.Funcs) != 1 {
		t.Fatalf("complete first-party syscall fixture: %v", err)
	}
	var output strings.Builder
	sig := FuncSig{Name: "example.com/syscalleffects.trap", Ret: I64,
		Frame: FrameLayout{Results: []FrameSlot{{Offset: 0, Type: I64, Index: 0, Field: -1}}}}
	c := newARM64CtxWithTarget(&output, file.Funcs[0], sig, testResolveSym("example.com/syscalleffects"), nil,
		"aarch64-unknown-linux-gnu", false)
	if err := c.prepareLocalControl(); err != nil {
		t.Fatal("actual whole-source continuation/frame prerequisite", err)
	}
	c.emitEntryAllocasAndArgInit()
	c.machineAvailability = newARM64MachineAvailability(c)
	if err := c.lowerBlocks(); err != nil {
		t.Fatal("actual source lowering prerequisite", err)
	}
	if err := c.machineAvailability.validate(); err != nil {
		t.Fatal("actual numerical effect prerequisite", err)
	}
	start := -1
	var original Instr
	count := 0
	for offset, native := range c.machineAvailability.nativeIR {
		if native.syscall == nil {
			continue
		}
		count++
		if strings.Contains(native.line, "call i64 @syscall(") {
			start, original = offset, native.syscall.source
			c.currentBlock, c.currentInstruction = native.syscall.block, native.syscall.instruction
			c.machineAvailability.current = c.currentBlock
		}
	}
	if count != 2 || start < 0 {
		t.Fatalf("actual typed bridge must retain two distinct call effects: count=%d", count)
	}
	c.machineAvailability.blocks[c.currentBlock].effects = nil
	return c, start, original
}

func TestARM64SyscallBridgeProofExactIdentity(t *testing.T) {
	for _, change := range []struct {
		name  string
		apply func(*arm64Ctx, int, *Instr) *arm64Ctx
	}{
		{"unaltered", func(c *arm64Ctx, _ int, _ *Instr) *arm64Ctx { return c }},
		{"context-copy", func(c *arm64Ctx, _ int, _ *Instr) *arm64Ctx { copy := *c; return &copy }},
		{"proof-copy", func(c *arm64Ctx, offset int, _ *Instr) *arm64Ctx {
			native := c.machineAvailability.nativeIR[offset]
			copy := *native.syscall
			native.syscall = &copy
			c.machineAvailability.nativeIR[offset] = native
			return c
		}},
		{"descriptor", func(c *arm64Ctx, offset int, _ *Instr) *arm64Ctx {
			c.machineAvailability.nativeIR[offset].syscall.form.gpWrites = 3
			return c
		}},
		{"target", func(c *arm64Ctx, _ int, _ *Instr) *arm64Ctx { c.targetTriple = "aarch64-unknown-linux-musl"; return c }},
		{"line", func(c *arm64Ctx, offset int, _ *Instr) *arm64Ctx {
			native := c.machineAvailability.nativeIR[offset]
			native.line = strings.Replace(native.line, "i64", "i32", 1)
			c.machineAvailability.nativeIR[offset] = native
			return c
		}},
		{"offset", func(c *arm64Ctx, offset int, _ *Instr) *arm64Ctx {
			c.machineAvailability.nativeIR[offset+1] = c.machineAvailability.nativeIR[offset]
			delete(c.machineAvailability.nativeIR, offset)
			return c
		}},
		{"ordinal", func(c *arm64Ctx, offset int, _ *Instr) *arm64Ctx {
			c.machineAvailability.nativeIR[offset].syscall.instruction++
			return c
		}},
		{"block", func(c *arm64Ctx, offset int, _ *Instr) *arm64Ctx {
			c.machineAvailability.nativeIR[offset].syscall.block++
			return c
		}},
		{"source", func(c *arm64Ctx, _ int, ins *Instr) *arm64Ctx { ins.Raw += " foreign"; return c }},
		{"source-op", func(c *arm64Ctx, _ int, ins *Instr) *arm64Ctx { ins.Op = "NOP"; return c }},
		{"source-operand", func(c *arm64Ctx, _ int, ins *Instr) *arm64Ctx {
			ins.Args = []Operand{{Kind: OpImm, Imm: 1}}
			return c
		}},
		{"current-source", func(c *arm64Ctx, _ int, _ *Instr) *arm64Ctx {
			c.blocks[c.currentBlock].instrs[c.currentInstruction].Raw += " changed"
			return c
		}},
		{"store-effect", func(c *arm64Ctx, offset int, _ *Instr) *arm64Ctx {
			native := c.machineAvailability.nativeIR[offset]
			native.store = false
			c.machineAvailability.nativeIR[offset] = native
			return c
		}},
		{"missing-CFG", func(c *arm64Ctx, _ int, _ *Instr) *arm64Ctx { c.localControl = nil; return c }},
		{"replaced-CFG", func(c *arm64Ctx, _ int, _ *Instr) *arm64Ctx {
			copy := *c.localControl
			c.localControl = &copy
			return c
		}},
		{"same-callee-unregistered", func(c *arm64Ctx, _ int, _ *Instr) *arm64Ctx {
			c.b.WriteString("  %foreign = call i64 @syscall(i64 0, i64 0, i64 0, i64 0, i64 0, i64 0, i64 0)\n")
			return c
		}},
	} {
		t.Run(change.name, func(t *testing.T) {
			c, offset, original := arm64SyscallProofFixture(t)
			c = change.apply(c, offset, &original)
			c.recordMachineOpaqueIR(offset, original)
			stores, opaque := 0, 0
			for _, effect := range c.machineAvailability.blocks[c.machineAvailability.current].effects {
				if effect.store {
					stores++
				}
				if effect.opaque {
					opaque++
				}
			}
			if change.name == "unaltered" {
				if stores != 2 || opaque != 0 {
					t.Fatalf("actual ordinary-return bridge erased stores or gained opaque effects: stores=%d opaque=%d", stores, opaque)
				}
			} else if opaque == 0 {
				t.Fatal("copied/changed/private-unregistered bridge acquired typed effects")
			}
		})
	}
}

func TestARM64SyscallUnmodeledFlagsCannotBorrowCMP(t *testing.T) {
	for _, api := range []string{"text", "module", "Go"} {
		t.Run(api, func(t *testing.T) {
			_, err := arm64SyscallEffectAPI(t, api, "aarch64-unknown-linux-gnu",
				arm64SyscallEffectSource("CMP R2,R1\nMSR R3,NZCV\nSVC\nCSET CS,R2"))
			if !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatalf("unregistered native NZCV cannot retain old CMP witness through SVC: %v", err)
			}
		})
	}
}

func TestARM64FlagFlowOrderedBitEffects(t *testing.T) {
	for _, bit := range []arm64FlagMask{arm64FlagN, arm64FlagZ, arm64FlagC, arm64FlagV} {
		for _, test := range []struct {
			name    string
			effects []arm64FlagEffect
			valid   bool
		}{
			{"write-read", []arm64FlagEffect{{write: bit}, {read: bit}}, true},
			{"read-write", []arm64FlagEffect{{read: bit}, {write: bit}}, false},
			{"other-bit", []arm64FlagEffect{{write: arm64FlagAll &^ bit}, {read: bit}}, false},
			{"write-invalidate-read", []arm64FlagEffect{{write: arm64FlagAll}, {invalidate: bit}, {read: bit}}, false},
			{"invalidate-write-read", []arm64FlagEffect{{invalidate: bit}, {write: bit}, {read: bit}}, true},
			{"unread-other-invalid", []arm64FlagEffect{{write: arm64FlagAll}, {invalidate: arm64FlagAll &^ bit}, {read: bit}}, true},
		} {
			t.Run(fmt.Sprintf("%x/%s", bit, test.name), func(t *testing.T) {
				flow := &arm64FlagFlow{blocks: []arm64FlagFlowBlock{{name: "entry", effects: test.effects}}, indices: map[string]int{"entry": 0}}
				err := flow.validate()
				if test.valid && err != nil || !test.valid && !errors.Is(err, ErrProbeNeedsContext) {
					t.Fatalf("actual ordered bit effects valid=%v: %v", test.valid, err)
				}
			})
		}
		for _, invalidate := range []bool{false, true} {
			t.Run(fmt.Sprintf("%x/loop-invalidate-%v", bit, invalidate), func(t *testing.T) {
				loop := []arm64FlagEffect{{read: bit}}
				if invalidate {
					loop = append(loop, arm64FlagEffect{invalidate: bit})
				}
				flow := &arm64FlagFlow{blocks: []arm64FlagFlowBlock{
					{name: "entry", effects: []arm64FlagEffect{{write: bit}}, successors: []string{"loop", "join"}},
					{name: "loop", effects: loop, successors: []string{"loop", "join"}},
					{name: "join", effects: []arm64FlagEffect{{read: bit}}},
				}, indices: map[string]int{"entry": 0, "loop": 1, "join": 2}}
				err := flow.validate()
				if !invalidate && err != nil || invalidate && !errors.Is(err, ErrProbeNeedsContext) {
					t.Fatalf("unknown loop predecessor was lost: %v", err)
				}
			})
		}
	}
}
