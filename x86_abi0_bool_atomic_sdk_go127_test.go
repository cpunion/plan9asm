//go:build go1.27

package plan9asm

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

type x86ABI0BoolUnsafeImporter struct{}

func (x86ABI0BoolUnsafeImporter) Import(path string) (*types.Package, error) {
	if path == "unsafe" {
		return types.Unsafe, nil
	}
	return nil, fmt.Errorf("unexpected actual declaration fixture import %q", path)
}

// Complete untouched Go files supply actual types without executing cmd/go or
// manufacturing an SDK source/object observation token. This is pure source/
// declaration/IR evidence; actual selected graph/native observations are separate.
func x86ABI0BoolActualAtomicPackage(t *testing.T, path string, files []string) GoPackage {
	t.Helper()
	fset := token.NewFileSet()
	var syntax []*ast.File
	for _, name := range files {
		data, err := os.ReadFile(filepath.Join(runtime.GOROOT(), "src", filepath.FromSlash(path), name))
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("actual Go declaration source %s/%s bytes=%d sha256=%x", path, name, len(data), sha256.Sum256(data))
		file, err := parser.ParseFile(fset, name, data, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		syntax = append(syntax, file)
	}
	config := types.Config{Importer: x86ABI0BoolUnsafeImporter{}, Sizes: types.SizesFor("gc", "386")}
	pkg, err := config.Check(path, fset, syntax, nil)
	if err != nil {
		t.Fatal(err)
	}
	return GoPackage{Path: path, Types: pkg, Syntax: syntax}
}

func TestX86ABI0Mode32CanonicalBoolActualAtomicCompleteSourceNextContext(t *testing.T) {
	for _, row := range []struct {
		path, member, boundary string
		declarations           []string
		texts                  int
	}{
		{"internal/runtime/atomic", "atomic_386.s", "instruction 3: LOCK", []string{"stubs.go", "atomic_386.go", "unaligned.go"}, 42},
		{"sync/atomic", "asm.s", "bool byte requires", []string{"doc.go", "doc_32.go"}, 36},
	} {
		t.Run(row.path, func(t *testing.T) {
			pkg := x86ABI0BoolActualAtomicPackage(t, row.path, row.declarations)
			source, err := os.ReadFile(filepath.Join(runtime.GOROOT(), "src", filepath.FromSlash(row.path), row.member))
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("complete original assembly %s/%s bytes=%d sha256=%x", row.path, row.member, len(source), sha256.Sum256(source))
			file, err := ParseWithDefines(ArchAMD64, string(source), GoAssemblerDefines("linux", "386"))
			if err != nil {
				t.Fatal(err)
			}
			if len(file.Funcs) != row.texts || len(file.Data) != 0 || len(file.Globl) != 0 {
				t.Fatal("complete actual assembly inventory changed", len(file.Funcs), len(file.Data), len(file.Globl))
			}
			resolve := testResolveSym(row.path)
			sigs, err := goSigsForAsmFile(pkg, file, resolve, "386", nil, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			if row.path == "internal/runtime/atomic" {
				plan, err := prepareX86ABI0BoolCalls(file, Options{Goarch: "386", TargetTriple: "i386-unknown-linux-gnu", ResolveSym: resolve, Sigs: sigs})
				if err != nil || plan == nil || len(plan.callers["·Cas64"].bytes) != 0 {
					t.Fatal("actual noarg void call acquired a bool-return/outgoing contract", err)
				}
			}
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(api, func(t *testing.T) {
					ctx := llvm.NewContext()
					defer ctx.Dispose()
					published := false
					var err error
					if api == "text" {
						var ir string
						ir, err = Translate(file, Options{Goarch: "386", TargetTriple: "i386-unknown-linux-gnu", ResolveSym: resolve, Sigs: sigs})
						published = ir != ""
					} else if api == "module" {
						var module llvm.Module
						module, err = TranslateModuleInContext(ctx, file, Options{Goarch: "386", TargetTriple: "i386-unknown-linux-gnu", ResolveSym: resolve, Sigs: sigs})
						published = !module.IsNil()
						if published {
							defer module.Dispose()
						}
					} else {
						module, failure := translateGoModuleInContext(ctx, pkg, source, GoModuleOptions{GOARCH: "386", GOOS: "linux", TargetTriple: "i386-unknown-linux-gnu", ResolveSym: resolve})
						err = failure
						published = module != nil
						if module != nil && !module.Module.IsNil() {
							defer module.Module.Dispose()
						}
					}
					if published || !errors.Is(err, ErrProbeNeedsContext) || !strings.Contains(err.Error(), row.boundary) {
						t.Fatal("complete source changed its genuine next unsupported contract", published, err)
					}
					t.Log("complete original source still Context, not SDK/object/runtime success:", err)
				})
			}
		})
	}
}
