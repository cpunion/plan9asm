package plan9asm

import (
	"fmt"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

type x86VectorUpperOracle struct {
	name      string
	binary    bool
	immediate bool
}

func x86VectorUpperMembers() []x86VectorUpperOracle {
	var members []x86VectorUpperOracle
	for _, op := range x86CompactExpandResultMembers {
		members = append(members, x86VectorUpperOracle{name: op})
	}
	for _, op := range x86IndexedPermuteResultNames {
		members = append(members, x86VectorUpperOracle{name: op, binary: true})
	}
	for _, row := range x86GFNIResultOracle {
		members = append(members, x86VectorUpperOracle{name: row.name, binary: true, immediate: row.immediate})
	}
	for _, row := range x86BitCountResultOracle {
		members = append(members, x86VectorUpperOracle{name: row.name})
	}
	return members
}

func x86VectorUpperFixture(arch string) (string, string, []string) {
	var source, declarations strings.Builder
	var names []string
	for _, member := range x86VectorUpperMembers() {
		for _, bank := range []string{"X", "Y"} {
			for _, mode := range []string{"plain", "merge", "zero"} {
				arity := 2
				if member.binary {
					arity++
				}
				if member.immediate {
					arity++
				}
				if mode != "plain" {
					arity++
				}
				if arch == "386" && arity > 3 {
					continue // Original named frontend limit, not an execution skip.
				}
				op, mask, second, immediate := member.name, "", "", ""
				if mode != "plain" {
					mask = "K1,"
				}
				if mode == "zero" {
					op += ".Z"
				}
				if member.binary {
					second = bank + "1,"
				}
				if member.immediate {
					immediate = "$0,"
				}
				name := "Value" + member.name + bank + mode
				names = append(names, name)
				fmt.Fprintf(&declarations, "func %s() uint32\n", name)
				fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-4\n"+
					"VPXORD Z0,Z0,Z0\nVPXORD Z1,Z1,Z1\nVPXORD Z2,Z2,Z2\nVPXORD Z7,Z7,Z7\n"+
					"MOVL $-1,AX\nKMOVD AX,K2\nVPMOVM2D K2,Z2\nMOVL $0,AX\nKMOVD AX,K1\nVPCOMPRESSQ Z0,K1,Z7\n"+
					"%s %s%s0,%s%s%s2\nVEXTRACTI32X4 $3,Z2,X3\nPEXTRD $0,X3,AX\nMOVL AX,ret+0(FP)\nRET\n",
					name, op, immediate, bank, second, mask, bank)
			}
		}
	}
	return source.String(), declarations.String(), names
}

func x86VectorUpperRequireZeroReturns(t *testing.T, ir string, names []string) {
	t.Helper()
	context := llvm.NewContext()
	defer context.Dispose()
	module, err := parseIRModuleInContext(context, ir)
	if !module.IsNil() {
		defer module.Dispose()
	}
	if err != nil {
		t.Fatal(err)
	}
	if err := llvm.VerifyModule(module, llvm.ReturnStatusAction); err != nil {
		t.Fatal(err)
	}
	options := llvm.NewPassBuilderOptions()
	defer options.Dispose()
	if err := module.RunPasses("default<O2>", llvm.TargetMachine{}, options); err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			function := module.NamedFunction(name)
			if function.IsNil() || function.FirstBasicBlock().IsNil() {
				t.Fatal("original nonzero-upper source body disappeared")
			}
			returns := 0
			for block := function.FirstBasicBlock(); !block.IsNil(); block = llvm.NextBasicBlock(block) {
				terminator := block.LastInstruction()
				if terminator.IsNil() || terminator.InstructionOpcode() != llvm.Ret {
					continue
				}
				returns++
				if terminator.OperandsCount() != 1 || terminator.Operand(0).IsAConstantInt().IsNil() || terminator.Operand(0).ZExtValue() != 0 {
					t.Fatal("VEX/EVEX write retained old nonzero upper bytes", terminator.String())
				}
			}
			if returns != 1 {
				t.Fatal("missing exact independent upper-quarter result", returns)
			}
		})
	}
}

func TestX86VectorValueFamiliesClearNonzeroUpperAliases(t *testing.T) {
	if len(x86VectorUpperMembers()) != 35 {
		t.Fatal("complete independently inventoried family changed")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			source, declarations, names := x86VectorUpperFixture(target.arch)
			want := 210
			if target.arch == "386" {
				want = 138
			}
			if len(names) != want {
				t.Fatal("whole-family write-width/mask denominator changed", len(names), want)
			}
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, source, declarations, target.arch, target.triple, api)
					if err != nil || ir == "" {
						t.Fatal("initialized nonzero-upper source rejected", err)
					}
					x86VectorUpperRequireZeroReturns(t, ir, names)
				})
			}
		})
	}
}
