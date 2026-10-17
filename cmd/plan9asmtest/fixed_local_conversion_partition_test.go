package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestFixedLocalConversionObjectsHaveIndependentTargetOwners(t *testing.T) {
	syntax, err := parser.ParseFile(token.NewFileSet(),
		filepath.Join("..", "..", "amd64_fixed_local_conversion_test.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][2]string{
		"TestGoAMD64FixedLocalConversionActualObjectsLinux":   {"linux/amd64", "x86_64-unknown-linux-gnu"},
		"TestGoAMD64FixedLocalConversionActualObjectsDarwin":  {"darwin/amd64", "x86_64-apple-darwin"},
		"TestGoAMD64FixedLocalConversionActualObjectsWindows": {"windows/amd64", "x86_64-pc-windows-msvc"},
	}
	var roots []string
	for _, declaration := range syntax.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, "TestGoAMD64FixedLocalConversionActualObjects") {
			continue
		}
		target, found := want[function.Name.Name]
		if !found {
			t.Fatal("unknown or monolithic conversion-object owner", function.Name.Name)
		}
		if len(function.Body.List) != 1 {
			t.Fatal("target owner must call the complete shared object plan", function.Name.Name)
		}
		statement, ok := function.Body.List[0].(*ast.ExprStmt)
		if !ok {
			t.Fatal("target owner has no direct complete-plan call", function.Name.Name)
		}
		call, ok := statement.X.(*ast.CallExpr)
		if !ok || len(call.Args) != 3 {
			t.Fatal("target owner changed its complete-plan arguments", function.Name.Name)
		}
		helper, ok := call.Fun.(*ast.Ident)
		if !ok || helper.Name != "fixedLocalConversionActualObjects" {
			t.Fatal("target owner substituted its complete-plan helper", function.Name.Name)
		}
		for index, expected := range target {
			literal, ok := call.Args[index+1].(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				t.Fatal("target binding must retain its literal platform", function.Name.Name)
			}
			actual, err := strconv.Unquote(literal.Value)
			if err != nil || actual != expected {
				t.Fatal("target owner changed platform or LLVM triple", function.Name.Name, actual, expected)
			}
		}
		delete(want, function.Name.Name)
		roots = append(roots, function.Name.Name)
	}
	if len(want) != 0 {
		t.Fatal("missing complete target owners", want)
	}
	shards, err := partition(roots, 32)
	if err != nil {
		t.Fatal(err)
	}
	for index, roots := range shards {
		if len(roots) > 1 {
			t.Fatalf("expensive conversion targets share shard %d: %v", index, roots)
		}
	}
}

func TestFixedLocalConversionObjectCacheIsOwnedByTheWholeTarget(t *testing.T) {
	syntax, err := parser.ParseFile(token.NewFileSet(),
		filepath.Join("..", "..", "amd64_fixed_local_conversion_test.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var helper *ast.FuncDecl
	for _, declaration := range syntax.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Name.Name == "fixedLocalConversionActualObjects" {
			helper = function
		}
	}
	if helper == nil {
		t.Fatal("missing complete conversion object helper")
	}
	isCache := func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || len(call.Args) != 2 {
			return false
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "Setenv" {
			return false
		}
		receiver, ok := selector.X.(*ast.Ident)
		key, literal := call.Args[0].(*ast.BasicLit)
		return ok && receiver.Name == "t" && literal && key.Value == `"GOCACHE"`
	}
	all, parent := 0, 0
	ast.Inspect(helper.Body, func(node ast.Node) bool {
		if isCache(node) {
			all++
		}
		return true
	})
	for _, statement := range helper.Body.List {
		if expression, ok := statement.(*ast.ExprStmt); ok && isCache(expression.X) {
			call := expression.X.(*ast.CallExpr)
			cache, ok := call.Args[1].(*ast.CallExpr)
			if !ok || len(cache.Args) != 0 {
				t.Fatal("build cache must be created by the parent test")
			}
			selector, ok := cache.Fun.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "TempDir" {
				t.Fatal("build cache must retain test-owned cleanup")
			}
			parent++
		}
	}
	if all != 1 || parent != 1 {
		t.Fatal("conversion batches must share exactly one parent-owned build cache", all, parent)
	}
}
