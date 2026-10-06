package plan9asm

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm/internal/goasmrules"
)

func x86OfficialPackedBinaryWitnesses(t *testing.T) map[string][]string {
	t.Helper()
	rules, err := goasmrules.LoadX86AVXRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	byOpcode := map[string]map[string]bool{}
	for _, rule := range rules {
		if rule.Ytab != "_yvandnpd" {
			continue
		}
		witnesses, err := goasmrules.X86PackedBinarySyntaxWitnesses(rule)
		if err != nil {
			t.Fatal(err)
		}
		if byOpcode[rule.Opcode] == nil {
			byOpcode[rule.Opcode] = map[string]bool{}
		}
		for _, witness := range witnesses {
			byOpcode[rule.Opcode][witness] = true
		}
	}
	result := map[string][]string{}
	for opcode, witnesses := range byOpcode {
		for witness := range witnesses {
			result[opcode] = append(result[opcode], witness)
		}
		sort.Strings(result[opcode])
	}
	return result
}

func TestX86OfficialPackedBinarySourceWitnessesCompileWithGo(t *testing.T) {
	byOpcode := x86OfficialPackedBinaryWitnesses(t)
	if len(byOpcode) < 60 {
		t.Fatal("official packed-binary table inventory is incomplete")
	}
	opcodes := make([]string, 0, len(byOpcode))
	for opcode := range byOpcode {
		opcodes = append(opcodes, opcode)
	}
	sort.Strings(opcodes)
	for _, arch := range []string{"amd64", "386"} {
		t.Run(arch, func(t *testing.T) {
			var source strings.Builder
			accepted, maskedFrontendRows := 0, 0
			for _, opcode := range opcodes {
				for _, witness := range byOpcode[opcode] {
					if arch == "386" && strings.Count(witness, ",") == 3 {
						// This is a separate frontend limitation, not a missing
						// encoder row or a successful translated object. Its real
						// Go rejection is checked below for each family opcode.
						maskedFrontendRows++
						continue
					}
					fmt.Fprintf(&source, "TEXT probe%d(SB),4,$0-0\n%s\nRET\n", accepted, witness)
					accepted++
				}
			}
			goAssemblerOracleOutput(t, arch, source.String(), "", true)
			t.Logf("actual Go assembled %d complete-table syntax witnesses/%d opcodes; %d masked mode32 candidates have separate frontend refusals; not LLVM/runtime coverage", accepted, len(opcodes), maskedFrontendRows)
		})
	}
	for _, opcode := range opcodes {
		t.Run("386_masked_frontend/"+opcode, func(t *testing.T) {
			goAssemblerOracleOutput(t, "386", "TEXT f(SB),4,$0-0\n"+opcode+" Z0,Z1,K1,Z2\nRET\n", "", false)
		})
	}
}
