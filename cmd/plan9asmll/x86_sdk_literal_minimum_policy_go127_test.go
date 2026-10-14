//go:build go1.27

package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"testing"
)

func TestMapsLiteralMinimumTargetsHaveIndependentCompleteOwners(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "x86_sdk_literal_minimum_go127_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	owners := map[string]string{
		"TestCompileOneX86SDKOriginalMapsLiteralMinimumSourceLinux":   "linux",
		"TestCompileOneX86SDKOriginalMapsLiteralMinimumSourceDarwin":  "darwin",
		"TestCompileOneX86SDKOriginalMapsLiteralMinimumSourceWindows": "windows",
	}
	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		target, required := owners[fn.Name.Name]
		if !required {
			continue
		}
		if fn.Body == nil || len(fn.Body.List) != 1 {
			t.Fatal("target owner must unconditionally invoke its complete fixture", fn.Name.Name)
		}
		expression, ok := fn.Body.List[0].(*ast.ExprStmt)
		if !ok {
			t.Fatal("target owner contains a conditional or early return", fn.Name.Name)
		}
		call, ok := expression.X.(*ast.CallExpr)
		if !ok || len(call.Args) != 2 {
			t.Fatal("complete target fixture call missing", fn.Name.Name)
		}
		helper, ok := call.Fun.(*ast.Ident)
		literal, literalOK := call.Args[1].(*ast.BasicLit)
		if !ok || helper.Name != "compileOneX86SDKOriginalMapsLiteralMinimumSource" ||
			!literalOK || literal.Value != strconv.Quote(target) {
			t.Fatal("target owner lost its exact complete platform", fn.Name.Name)
		}
		delete(owners, fn.Name.Name)
	}
	if len(owners) != 0 {
		t.Fatal("target owners are missing; one shared test deadline cannot cover all targets", owners)
	}
}
