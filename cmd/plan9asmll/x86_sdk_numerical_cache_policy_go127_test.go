//go:build go1.27

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"testing"
)

// This standalone standard-library policy can run without loading the CLI's
// LLVM/SDK producers. It protects only tool-cache configuration, never proof.
func x86SDKNumericalCLICachePolicy(file *ast.File) error {
	var profile *ast.FuncDecl
	for _, declaration := range file.Decls {
		if fn, ok := declaration.(*ast.FuncDecl); ok && fn.Name.Name == "setX86SDKNumericalCLIProfile" {
			if profile != nil {
				return fmt.Errorf("duplicate actual SDK CLI profile")
			}
			profile = fn
		}
	}
	if profile == nil || profile.Body == nil {
		return fmt.Errorf("actual SDK CLI profile missing")
	}
	cacheSelections, cacheBindings, privateLoops := 0, 0, 0
	var failure error
	ast.Inspect(profile.Body, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.CallExpr:
			if fn, ok := value.Fun.(*ast.Ident); ok && fn.Name == "x86SDKNumericalCLIBuildCache" {
				cacheSelections++
			}
			if fn, ok := value.Fun.(*ast.SelectorExpr); ok && fn.Sel.Name == "Setenv" && len(value.Args) == 2 {
				if key, ok := value.Args[0].(*ast.BasicLit); ok && key.Value == strconv.Quote("GOCACHE") {
					cacheBindings++
					if selected, ok := value.Args[1].(*ast.Ident); !ok || selected.Name != "cache" {
						failure = fmt.Errorf("selected build cache was replaced")
					}
				}
			}
		case *ast.RangeStmt:
			list, ok := value.X.(*ast.CompositeLit)
			if !ok {
				break
			}
			if _, stringsList := list.Type.(*ast.ArrayType); !stringsList {
				break
			}
			keys := make(map[string]bool)
			for _, element := range list.Elts {
				if literal, ok := element.(*ast.BasicLit); ok && literal.Kind == token.STRING {
					key, _ := strconv.Unquote(literal.Value)
					keys[key] = true
				}
			}
			private, unsets := false, false
			ast.Inspect(value.Body, func(child ast.Node) bool {
				if call, ok := child.(*ast.CallExpr); ok {
					if fn, ok := call.Fun.(*ast.SelectorExpr); ok {
						private = private || fn.Sel.Name == "TempDir"
						unsets = unsets || fn.Sel.Name == "Unsetenv"
					}
				}
				return true
			})
			if private {
				privateLoops++
				if len(keys) != 4 || !keys["GOPATH"] || !keys["GOMODCACHE"] || !keys["TMPDIR"] || !keys["GOTMPDIR"] {
					failure = fmt.Errorf("build cache was recreated or private command roots changed: %v", keys)
				}
			}
			if unsets && keys["GOCACHE"] {
				failure = fmt.Errorf("effective defaults discarded the selected build cache")
			}
		}
		return true
	})
	if failure != nil {
		return failure
	}
	if cacheSelections != 1 || cacheBindings != 1 || privateLoops != 1 {
		return fmt.Errorf("actual profile needs one selected cache and independent private roots: select=%d bind=%d private=%d", cacheSelections, cacheBindings, privateLoops)
	}
	return nil
}

func TestX86SDKCLIProfileRetainsSelectedBuildCache(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "x86_sdk_numerical_source_go127_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := x86SDKNumericalCLICachePolicy(file); err != nil {
		t.Fatal(err)
	}
}
