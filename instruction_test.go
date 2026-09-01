package plan9asm

import (
	"errors"
	"testing"
)

func TestDescribeInstructionDistinguishesOperandForms(t *testing.T) {
	reg := DescribeInstruction(ArchAMD64, "amd64", Instr{
		Op:   "XORB",
		Args: []Operand{{Kind: OpReg, Reg: SI}, {Kind: OpReg, Reg: AX}},
	})
	mem := DescribeInstruction(ArchAMD64, "amd64", Instr{
		Op:   "XORB",
		Args: []Operand{{Kind: OpReg, Reg: SI}, {Kind: OpMem, Mem: MemRef{Base: AX}}},
	})
	if reg.Form == mem.Form {
		t.Fatalf("register and memory forms collapsed to %q", reg.Form)
	}
	if mem.Family != "logical" {
		t.Fatalf("XORB family = %q, want logical", mem.Family)
	}
}

func TestInstructionFamiliesAreArchitectureAware(t *testing.T) {
	cases := []struct {
		arch Arch
		op   string
		want string
	}{
		{ArchAMD64, "PUSHQ", "stack"},
		{ArchAMD64, "PUNPCKLQDQ", "sse-mmx"},
		{ArchAMD64, "POPCNTQ", "bit-shift"},
		{ArchAMD64, "PREFETCHT0", "cache-memory"},
		{ArchAMD64, "VGF2P8AFFINEQB", "crypto-carryless"},
		{ArchARM64, "VLD1", "neon"},
		{ArchARM64, "ZADD", "sve"},
		{ArchARM64, "CASAL", "atomic-memory"},
		{ArchARM64, "BIC", "logical"},
		{ArchARM64, "BEQ", "control-flow"},
		{ArchARM, "MRC", "system"},
		{ArchARM, "BIC", "logical"},
	}
	for _, tc := range cases {
		if got := InstructionFamily(tc.arch, tc.op); got != tc.want {
			t.Errorf("InstructionFamily(%s, %s) = %q, want %q", tc.arch, tc.op, got, tc.want)
		}
	}
}

func TestProbeInstructionChecksFormsThroughLowerer(t *testing.T) {
	cases := []Instr{
		{Op: "XORB", Raw: "XORB SI, (AX)", Args: []Operand{{Kind: OpReg, Reg: SI}, {Kind: OpMem, Mem: MemRef{Base: AX}}}},
		{Op: "PUNPCKLQDQ", Raw: "PUNPCKLQDQ X0, X0", Args: []Operand{{Kind: OpReg, Reg: "X0"}, {Kind: OpReg, Reg: "X0"}}},
	}
	for _, ins := range cases {
		if err := ProbeInstruction(ArchAMD64, "amd64", ins); err != nil {
			t.Errorf("ProbeInstruction(%q) error = %v", ins.Raw, err)
		}
	}

	fp := Instr{Op: "MOVQ", Raw: "MOVQ x+0(FP), AX", Args: []Operand{{Kind: OpFP, FPName: "x"}, {Kind: OpReg, Reg: AX}}}
	if err := ProbeInstruction(ArchAMD64, "amd64", fp); !errors.Is(err, ErrProbeNeedsContext) {
		t.Fatalf("FP probe error = %v, want ErrProbeNeedsContext", err)
	}
}
