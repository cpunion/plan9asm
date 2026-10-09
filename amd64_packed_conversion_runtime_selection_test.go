package plan9asm

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

func TestX86PackedConversionRequiredNativeOwners(t *testing.T) {
	const file = "amd64_packed_conversion_runtime_linux_amd64_test.go"
	for _, target := range []struct{ os, arch string }{
		{"linux", "amd64"}, {"linux", "386"}, {"linux", "arm"}, {"linux", "arm64"},
		{"js", "wasm"}, {"darwin", "amd64"}, {"darwin", "arm64"}, {"windows", "amd64"},
	} {
		context := build.Default
		context.GOOS, context.GOARCH = target.os, target.arch
		selected, err := context.MatchFile(".", file)
		if err != nil || selected != (target.os == "linux" && target.arch == "amd64") {
			t.Fatal("required native execution owner selection", target, selected, err)
		}
	}
	syntax, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]string{
		"TestCrossLinuxRuntimeMatrixX86PackedConversionNativeAVX512Status": {"x86PackedConversionRuntimeMatrix", "t", "true"},
		"TestCrossLinuxRuntimeMatrixX86PackedConversionNativeIndependentC": {"amd64PackedNumericConversionNativeRuntimeSemantics", "t"},
		"TestCrossLinuxRuntimeMatrixX86PackedConversionNativeDeadFaults":   {"x86PackedConversionDeadFaults", "t", "true", "true"},
	}
	for _, declaration := range syntax.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		expected, found := want[function.Name.Name]
		if !found || function.Body == nil || len(function.Body.List) != 1 {
			t.Fatal("mandatory native entry changed or became conditional", function.Name.Name)
		}
		statement, ok := function.Body.List[0].(*ast.ExprStmt)
		if !ok {
			t.Fatal("required native helper not executed directly")
		}
		call, ok := statement.X.(*ast.CallExpr)
		if !ok || len(call.Args)+1 != len(expected) {
			t.Fatal("native helper lost an execution axis")
		}
		for index, expression := range append([]ast.Expr{call.Fun}, call.Args...) {
			identifier, ok := expression.(*ast.Ident)
			if !ok || identifier.Name != expected[index] {
				t.Fatal("native helper or required full-family argument changed")
			}
		}
		delete(want, function.Name.Name)
	}
	if len(want) != 0 {
		t.Fatal("missing required native owners", want)
	}
	workflow, err := os.ReadFile(".github/workflows/go-ci.yml")
	if err != nil || !strings.Contains(string(workflow), "-ci-selector '^(TestCrossLinuxRuntimeMatrix|") {
		t.Fatal("required native runtime roots not selected by CI", err)
	}
}
