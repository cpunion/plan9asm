package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

func TestFixedLocalConversionObjectsHaveIndependentTargetOwners(t *testing.T) {
	syntax, err := parser.ParseFile(token.NewFileSet(),
		filepath.Join("..", "..", "amd64_fixed_local_conversion_test.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"TestGoAMD64FixedLocalConversionActualObjectsLinux":   true,
		"TestGoAMD64FixedLocalConversionActualObjectsDarwin":  true,
		"TestGoAMD64FixedLocalConversionActualObjectsWindows": true,
	}
	var roots []string
	for _, declaration := range syntax.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, "TestGoAMD64FixedLocalConversionActualObjects") {
			continue
		}
		if !want[function.Name.Name] {
			t.Fatal("unknown or monolithic conversion-object owner", function.Name.Name)
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
