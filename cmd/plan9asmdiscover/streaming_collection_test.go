package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestDiscoveryCollectionTransferCLIDirectoryCannotUseLegacyReader(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "main.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	registered := 0
	ast.Inspect(file, func(node ast.Node) bool {
		branch, ok := node.(*ast.IfStmt)
		if !ok {
			return true
		}
		isDirectory, isConversion := false, false
		ast.Inspect(branch.Cond, func(node ast.Node) bool {
			if call, ok := node.(*ast.CallExpr); ok {
				if selector, ok := call.Fun.(*ast.SelectorExpr); ok && selector.Sel.Name == "IsDir" {
					isDirectory = true
				}
			}
			return true
		})
		if branch.Init != nil {
			ast.Inspect(branch.Init, func(node ast.Node) bool {
				if name, ok := node.(*ast.Ident); ok && name.Name == "convertReport" {
					isConversion = true
				}
				return true
			})
		}
		if !isDirectory || !isConversion {
			return true
		}
		streamCalls := 0
		ast.Inspect(branch.Body, func(node ast.Node) bool {
			if call, ok := node.(*ast.CallExpr); ok {
				if name, ok := call.Fun.(*ast.Ident); ok {
					if name.Name == "mergeLedgerStreaming" {
						streamCalls++
					}
					if name.Name == "readSeenReport" || name.Name == "readShardedReports" {
						t.Errorf("directory conversion calls full reader %s", name.Name)
					}
				}
			}
			return true
		})
		if _, ok := branch.Body.List[len(branch.Body.List)-1].(*ast.ReturnStmt); !ok {
			t.Error("collection directory falls through to legacy full reader")
		}
		if streamCalls != 1 {
			t.Errorf("directory has %d streaming calls, want one mandatory call", streamCalls)
		}
		registered++
		return true
	})
	if registered != 1 {
		t.Fatalf("actual CLI has %d directory conversion branches, want one", registered)
	}
}
