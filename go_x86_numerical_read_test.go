//go:build go1.27

package plan9asm

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
	"github.com/xgo-dev/plan9asm/internal/goasmcmd"
)

func numericalSBFixture(t *testing.T, arch string) (GoPackage, GoAMD64FrameCaptureRequest, *File, Options) {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	goSource := []byte("package numericalfixture\nfunc Feature() byte\n")
	assembly := `#include "textflag.h"
TEXT ·Feature(SB),NOSPLIT,$0-1
MOVL $0,AX
CMPB internal∕cpu·X86+64(SB),$1
SETEQ AL
PINSRD $0,AX,X0
PEXTRD $0,X0,AX
MOVB AL,ret+0(FP)
RET
`
	for name, data := range map[string][]byte{
		"go.mod":               []byte("module example.invalid/numericalfixture\ngo 1.27\n"),
		"value.go":             goSource,
		"value_" + arch + ".s": []byte(assembly),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	fset := token.NewFileSet()
	syntax, err := parser.ParseFile(fset, filepath.Join(dir, "value.go"), goSource, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	typed, err := (&types.Config{Sizes: types.SizesFor("gc", arch)}).Check("example.invalid/numericalfixture", fset, []*ast.File{syntax}, nil)
	if err != nil {
		t.Fatal(err)
	}
	binary, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	binary, err = filepath.EvalSymlinks(binary)
	if err != nil {
		t.Fatal(err)
	}
	env := make(map[string]string)
	for _, pair := range os.Environ() {
		key, value, _ := strings.Cut(pair, "=")
		env[key] = value
	}
	for _, key := range []string{"GOCACHE", "GOMODCACHE", "GOPATH", "TMPDIR", "GOTMPDIR"} {
		if !filepath.IsAbs(env[key]) {
			env[key] = t.TempDir()
		}
		env[key], err = filepath.EvalSymlinks(env[key])
		if err != nil {
			t.Fatal(err)
		}
	}
	for key, value := range map[string]string{
		"GOOS": "linux", "GOARCH": arch, "GOAMD64": "v1", "GO386": "sse2",
		"GOFLAGS": "", "GOEXPERIMENT": "", "GOTOOLCHAIN": "local", "GOENV": "off",
		"GOWORK": "off", "CGO_ENABLED": "0", "GOPROXY": "off",
	} {
		env[key] = value
	}
	var environment []string
	for key, value := range env {
		environment = append(environment, key+"="+value)
	}
	request := GoAMD64FrameCaptureRequest{GoBinary: binary, Directory: dir, AssemblyFile: "value_" + arch + ".s", Target: "linux/" + arch, Environment: environment}
	file, err := Parse(ArchAMD64, strings.Replace(assembly, "#include \"textflag.h\"", "#define NOSPLIT 4", 1))
	if err != nil {
		t.Fatal(err)
	}
	declaration := typed.Scope().Lookup("Feature").(*types.Func)
	sig, err := DeriveGoFuncABI0Signature(typed.Path()+".Feature", declaration, arch, types.SizesFor("gc", arch))
	if err != nil {
		t.Fatal(err)
	}
	triple := "x86_64-unknown-linux-gnu"
	if arch == "386" {
		triple = "i386-unknown-linux-gnu"
	}
	opt := Options{Goarch: arch, TargetTriple: triple, ResolveSym: func(sym string) string {
		if strings.HasPrefix(sym, "·") {
			return typed.Path() + "." + strings.TrimPrefix(sym, "·")
		}
		return strings.ReplaceAll(strings.ReplaceAll(sym, "∕", "/"), "·", ".")
	}, Sigs: map[string]FuncSig{sig.Name: sig}}
	return GoPackage{Path: typed.Path(), Types: typed, Syntax: []*ast.File{syntax}}, request, file, opt
}

func TestGoX86NumericalSBReadOriginalContext(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		t.Run(arch, func(t *testing.T) {
			_, request, file, opt := numericalSBFixture(t, arch)
			ctx, cancel := runtimeOracleStageContext(t)
			defer cancel()
			selection, err := goasmcmd.Run(ctx, request.Directory, request.Environment, request.GoBinary, "list", "-json", "-compiled", "-export", "-mod=readonly", ".")
			if err != nil {
				t.Fatalf("actual original Go export failed: %v\n%s", err, selection)
			}
			listing, err := goasmcmd.Run(ctx, request.Directory, request.Environment, request.GoBinary, "tool", "asm", "-S", "-p", "example.invalid/numericalfixture", "-I", filepath.Join(filepath.Dir(filepath.Dir(request.GoBinary)), "pkg", "include"), "-o", filepath.Join(t.TempDir(), "original.o"), request.AssemblyFile)
			if err != nil {
				t.Fatalf("actual original Go asm failed: %v\n%s", err, listing)
			}
			t.Logf("actual %s original Go export succeeded; source64/read1 listing:\n%s", arch, listing)
			for _, api := range []string{"text", "module", "context"} {
				t.Run(api, func(t *testing.T) {
					var err error
					if api == "text" {
						_, err = Translate(file, opt)
					} else if api == "module" {
						mod, failure := TranslateModule(file, opt)
						err = failure
						if !mod.IsNil() {
							mod.Dispose()
						}
					} else {
						llvmContext := llvm.NewContext()
						mod, failure := TranslateModuleInContext(llvmContext, file, opt)
						err = failure
						if !mod.IsNil() {
							mod.Dispose()
						}
						llvmContext.Dispose()
					}
					if os.Getenv("PLAN9ASM_NUMERICAL_SB_EXPECT_ACCEPT") == "1" {
						if err != nil {
							t.Fatalf("actual-Go accepted numerical SB read not closed: %v", err)
						}
					} else if !errors.Is(err, ErrProbeNeedsContext) {
						t.Fatalf("declaration-only SB read acquired authority: %v", err)
					}
				})
			}
		})
	}
}
