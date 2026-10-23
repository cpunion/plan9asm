package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/types"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
	"github.com/xgo-dev/plan9asm"
	"golang.org/x/tools/go/packages"
)

const simpleImportedStatePath = "example.com/clisource/state"

// Load the real selected files/types once through the CLI's existing loader.
// No synthetic AST/type package, SDK capture or translation-time reload supplies
// the imported declarations. The other-OS file must remain unselected.
func simpleImportedGlobalPackage(t *testing.T, arch string) (*loadedPackage, string, []byte) {
	t.Helper()
	directory := t.TempDir()
	symbol := strings.NewReplacer("/", "∕", ".", "·").Replace(simpleImportedStatePath)
	var source strings.Builder
	for _, row := range []struct{ name, instruction string }{
		{"CompareByte", "CMPB " + symbol + "·Flags(SB),$1"},
		{"CompareWord", "CMPL " + symbol + "·Word(SB),$0x12345678"},
		{"CompareLocal", "CMPB ·Flags(SB),$0"},
	} {
		fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-4\n%s\nJNE different\nMOVL $1,ret+0(FP)\nRET\ndifferent:\nMOVL $0,ret+0(FP)\nRET\n", row.name, row.instruction)
	}
	inputs := map[string][]byte{
		"go.mod":                          []byte("module example.com/clisource\n\ngo 1.24\n"),
		"state/state.go":                  []byte("package state\ntype Byte uint8\nvar Flags Byte = 1\n"),
		"state/word_linux.go":             []byte("//go:build linux\n\npackage state\nvar Word uint32 = 0x12345678\n"),
		"state/word_other.go":             []byte("//go:build !linux\n\npackage state\nvar Word uint64\n"),
		"consumer/consumer.go":            []byte("package consumer\nimport state \"" + simpleImportedStatePath + "\"\nvar _ = state.Flags\nvar _ = state.Word\nvar Flags uint8\nfunc CompareByte() uint32\nfunc CompareWord() uint32\nfunc CompareLocal() uint32\n"),
		"consumer/compare_" + arch + ".s": []byte(source.String()),
	}
	for name, input := range inputs {
		file := filepath.Join(directory, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, input, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		for name, expected := range inputs {
			actual, err := os.ReadFile(filepath.Join(directory, filepath.FromSlash(name)))
			if err != nil || !bytes.Equal(actual, expected) {
				t.Errorf("original loaded input changed: %s: %v", name, err)
			}
		}
	})
	for key, value := range map[string]string{
		"GOROOT": runtime.GOROOT(), "GOENV": "off", "GOWORK": "off", "GOTOOLCHAIN": "local",
		"GOFLAGS": "", "GOPACKAGESDRIVER": "off", "CGO_ENABLED": "0", "GOEXPERIMENT": "",
		"GOAMD64": "v1", "GO386": "sse2", "GOCACHE": t.TempDir(),
	} {
		t.Setenv(key, value)
	}
	t.Chdir(filepath.Join(directory, "consumer"))
	loaded, err := loadPackage(".", "linux", arch)
	if err != nil || loaded == nil {
		t.Fatalf("actual CLI package selection failed before translation: %v", err)
	}
	pkg := loaded.Package
	dependency := pkg.Imports[simpleImportedStatePath]
	if pkg.PkgPath != "example.com/clisource/consumer" || pkg.Types == nil || pkg.TypesSizes == nil ||
		len(pkg.Syntax) != 1 || dependency == nil || dependency.Types == nil || len(dependency.Syntax) != 2 ||
		len(dependency.CompiledGoFiles) != 2 || dependency.IllTyped || len(dependency.Errors) != 0 {
		t.Fatal("actual target package or complete imported source/type selection is missing")
	}
	if word, ok := dependency.Types.Scope().Lookup("Word").(*types.Var); !ok ||
		!types.Identical(word.Type(), types.Typ[types.Uint32]) || pkg.TypesSizes.Sizeof(word.Type()) != 4 {
		t.Fatal("the original Linux source selection did not establish the real uint32 variable")
	}
	for _, name := range dependency.CompiledGoFiles {
		if filepath.Base(name) != "state.go" && filepath.Base(name) != "word_linux.go" {
			t.Fatal("unexpected imported compiled source", name)
		}
	}
	assembly := filepath.Join(directory, "consumer", "compare_"+arch+".s")
	return loaded, assembly, []byte(source.String())
}

func TestSimpleCLILoadedImportedGlobalSyntaxForwarding(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		t.Run(arch, func(t *testing.T) {
			loaded, _, _ := simpleImportedGlobalPackage(t, arch)
			pkg := loaded.Package
			view := goPackageView(pkg)
			dependency := pkg.Imports[simpleImportedStatePath]
			if view.Types != pkg.Types || view.Path != pkg.PkgPath || !reflect.DeepEqual(view.Syntax, pkg.Syntax) ||
				view.Imports[simpleImportedStatePath] != dependency.Types ||
				!reflect.DeepEqual(view.ImportSyntax[simpleImportedStatePath], dependency.Syntax) {
				t.Fatal("CLI view discarded the actual complete imported declaration syntax or typed identity")
			}
		})
	}
}

func simpleImportedGlobalModule(t *testing.T, pkg *packages.Package, source []byte, arch string) (string, error) {
	t.Helper()
	result, err := plan9asm.TranslateGoModule(goPackageView(pkg), source, plan9asm.GoModuleOptions{
		GOOS: "linux", GOARCH: arch, TargetTriple: targetTriple("linux", arch),
		Sizes: pkg.TypesSizes, FrameSizes: pkg.TypesSizes, ResolveSym: resolveSymFunc(pkg.PkgPath),
		SourcePreprocessed: true,
	})
	if result == nil {
		return "", err
	}
	defer result.Module.Dispose()
	if err == nil {
		if verify := llvm.VerifyModule(result.Module, llvm.ReturnStatusAction); verify != nil {
			t.Fatal("actual loaded declaration module does not verify", verify)
		}
		if len(result.Functions) != 3 {
			t.Fatal("complete original three-function source was cropped")
		}
	}
	return result.Module.String(), err
}

func simpleImportedGlobalRequireReads(t *testing.T, ir, pkgPath string) {
	t.Helper()
	for _, name := range []string{simpleImportedStatePath + ".Flags", simpleImportedStatePath + ".Word", pkgPath + ".Flags"} {
		if !strings.Contains(ir, `@"`+name+`" = external global `) {
			t.Fatal("actual mutable Go storage lost its external namespace or became a definition/constant", name)
		}
	}
	if strings.Count(ir, "load volatile i8") != 2 || strings.Count(ir, "load volatile i32") != 1 ||
		!strings.Contains(ir, "br i1") {
		t.Fatal("complete imported/local mutable reads lost exact widths or observable predicates")
	}
	for _, name := range []string{"CompareByte", "CompareWord", "CompareLocal"} {
		if !strings.Contains(ir, `@"`+pkgPath+"."+name+`"(`) {
			t.Fatal("original complete source entry missing", name)
		}
	}
}

func TestSimpleCLILoadedImportedGlobalWholeGoModule(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		t.Run(arch, func(t *testing.T) {
			loaded, _, source := simpleImportedGlobalPackage(t, arch)
			ir, err := simpleImportedGlobalModule(t, loaded.Package, source, arch)
			if err != nil || ir == "" {
				t.Fatal("ordinary GoModule lost real CLI-loaded imported declarations", err)
			}
			simpleImportedGlobalRequireReads(t, ir, loaded.PkgPath)
		})
	}
}

func TestSimpleCLITranslateLoadedImportedGlobalFile(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		t.Run(arch, func(t *testing.T) {
			loaded, assembly, _ := simpleImportedGlobalPackage(t, arch)
			translated, text, err := translateAsmForPackage(loaded.Package, assembly, "linux", arch, false, loaded.selected)
			if err != nil || !text {
				t.Fatal("actual CLI whole-file path omitted imported declaration binding", err)
			}
			if len(translated.Functions) != 3 {
				t.Fatal("CLI omitted an original source entry")
			}
			simpleImportedGlobalRequireReads(t, translated.LLVMIR, loaded.PkgPath)
		})
	}
}

func TestSimpleCLIImportedTypeOnlyViewHasNoDeclarationAuthority(t *testing.T) {
	loaded, _, source := simpleImportedGlobalPackage(t, "amd64")
	pkg := *loaded.Package
	dependency := *pkg.Imports[simpleImportedStatePath]
	dependency.Syntax = nil // Preserve the real export/type identity, but no declaration source.
	pkg.Imports = map[string]*packages.Package{simpleImportedStatePath: &dependency}
	ir, err := simpleImportedGlobalModule(t, &pkg, source, "amd64")
	if ir != "" || !errors.Is(err, plan9asm.ErrProbeNeedsContext) {
		t.Fatal("a real type-only import was incorrectly treated as complete declaration evidence", err)
	}
}
