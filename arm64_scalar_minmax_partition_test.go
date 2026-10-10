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
