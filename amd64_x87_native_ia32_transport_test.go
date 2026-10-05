package plan9asm

import (
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"testing"
)

func TestX87IA32RequiredNativeTransportSource(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "amd64_x87_flag_compare_linux_amd64_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"x87IA32RequireNativeHost":   false,
		"x87IA32NativePreflight":     false,
		"x87IA32OriginalGoDirect":    false,
		"x87IA32CompileAndRunDirect": false,
	}
	ast.Inspect(file, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.BasicLit:
			if value.Kind == token.STRING {
				text, err := strconv.Unquote(value.Value)
				if err != nil {
					t.Fatal(err)
				}
				if text == "qemu-i386" || text == "-exec" {
					t.Errorf("required native IA32 oracle still selects emulation: %s", text)
				}
			}
		case *ast.CallExpr:
			if function, ok := value.Fun.(*ast.Ident); ok {
				if _, required := want[function.Name]; required {
					want[function.Name] = true
				}
			}
			if method, ok := value.Fun.(*ast.SelectorExpr); ok && method.Sel.Name == "Skip" {
				t.Error("required native IA32 oracle cannot skip failed preflight")
			}
		}
		return true
	})
	for function, found := range want {
		if !found {
			t.Errorf("required native IA32 route lacks %s", function)
		}
	}
}

func TestX87IA32OriginalReferenceMatrixPreserved(t *testing.T) {
	for _, fixture := range []struct {
		name, source, hash string
	}{
		{"386 assembly", x87FlagCompareAssembly("386"), "09ae2f421502ee1f22367d019e98f83767319bf03e414fdf305ad9afb940dea3"},
		{"amd64 assembly", x87FlagCompareAssembly("amd64"), "254191f4443a8d3c24c372c5aa5beb820f77e2a7d96a13179ca7fb2b37799664"},
		{"Go reference", x87FlagCompareOriginalReference, "62784b45ccd0cf0f6c04ec3cb498c744c48f5ba82354a8e2a6041a0067b7ade6"},
		{"C reference", x87FlagCompareCReference, "ea4efe97b695fe4a398422ba3b32148e2456ed6dac3a4c8c88194080dcec4466"},
	} {
		if got := fmt.Sprintf("%x", sha256.Sum256([]byte(fixture.source))); got != fixture.hash {
			t.Errorf("%s changed: %s", fixture.name, got)
		}
	}
	if got := len(x87FlagCompareCases()); got != 32 {
		t.Fatalf("original 32-kernel matrix changed: %d", got)
	}
}
