package plan9asm

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestARM64ScalarMinMaxBoundedRuntimeRoots(t *testing.T) {
	syntax, err := parser.ParseFile(token.NewFileSet(), "arm64_scalar_minmax_runtime_linux_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"TestCrossLinuxRuntimeMatrixARM64ScalarMinMaxFMAX":   "FMAX",
		"TestCrossLinuxRuntimeMatrixARM64ScalarMinMaxFMIN":   "FMIN",
		"TestCrossLinuxRuntimeMatrixARM64ScalarMinMaxFMAXNM": "FMAXNM",
		"TestCrossLinuxRuntimeMatrixARM64ScalarMinMaxFMINNM": "FMINNM",
	}
	for _, declaration := range syntax.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		family, required := want[function.Name.Name]
		if !required {
			continue
		}
		calls := 0
		ast.Inspect(function.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			name, ok := call.Fun.(*ast.Ident)
			if !ok || name.Name != "arm64MinMaxLinuxFamilyRuntime" {
				return true
			}
			calls++
			if len(call.Args) != 2 {
				t.Fatal("partition owner lost explicit family", function.Name.Name)
			}
			value, ok := call.Args[1].(*ast.BasicLit)
			if !ok || value.Kind != token.STRING || value.Value != "\""+family+"\"" {
				t.Fatal("partition owner changed family", function.Name.Name)
			}
			return true
		})
		if calls != 1 {
			t.Fatal("partition owner lost exact runtime helper", function.Name.Name)
		}
		delete(want, function.Name.Name)
	}
	if len(want) != 0 {
		t.Fatal("45-minute root needs exhaustive independently schedulable families", want)
	}
}

func TestARM64ScalarMinMaxFamilyPartitionsComplete(t *testing.T) {
	total := 0
	for _, banks := range []bool{true, false} {
		seen := make(map[arm64MinMaxStateCase]bool)
		for _, family := range []string{"FMAX", "FMIN", "FMAXNM", "FMINNM"} {
			rows := arm64MinMaxFamilyStateCases(banks, family)
			expected := 205
			if !banks {
				expected = 10
			}
			if len(rows) != expected {
				t.Fatal("changed full family denominator", family, banks, len(rows))
			}
			for _, row := range rows {
				if seen[row] {
					t.Fatal("duplicate cross-partition runtime form", row)
				}
				seen[row] = true
				total += len(arm64MinMaxSamples(row.bits, banks))
			}
		}
		original := arm64MinMaxStateCases(banks)
		if len(seen) != len(original) {
			t.Fatal("partition union differs from original complete grammar")
		}
		for _, row := range original {
			if !seen[row] {
				t.Fatal("original runtime form was omitted", row)
			}
		}
	}
	if total != 14600 {
		t.Fatal("partitioning removed original Go/LLVM state comparisons", total)
	}
	if len(arm64MinMaxFamilyStateCases(true, "unknown")) != 0 {
		t.Fatal("unknown family borrowed another partition")
	}
}

func TestARM64ScalarMinMaxAFPPartitionWitnesses(t *testing.T) {
	// The original oracle increments AH only for the ordinary min/max signed-
	// zero witness. NM partitions retain their full AH input matrix but cannot
	// acquire that counter from a different partition's instruction family.
	for _, family := range []string{"", "FMAX", "FMIN", "FMAXNM", "FMINNM"} {
		numeric := family == "FMAXNM" || family == "FMINNM"
		for _, counts := range [][2]int{{0, 0}, {0, 40}, {1025, 0}, {1025, 40}} {
			nep, ah := counts[0], counts[1]
			want := nep > 0 && (numeric || ah > 0)
			err := arm64MinMaxAFPObservationError(family, nep, ah)
			if (err == nil) != want {
				t.Errorf("family=%q NEP=%d AH=%d: error=%v; want accepted=%t", family, nep, ah, err, want)
			}
		}
	}
}
