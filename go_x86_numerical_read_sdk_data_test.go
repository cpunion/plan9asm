//go:build go1.27

package plan9asm

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// The source, generated header, assembly, object, CPU variable and initializer
// are actual independently captured SDK inputs. No instruction, DATA, feature
// offset, package declaration or original byte is replaced for this witness.
func numericalSDKLiteralFixture(t *testing.T) (context.Context, GoX86NumericalReadCaptureRequest, *GoX86SDKASMDataSource) {
	t.Helper()
	ctx, request, _ := numericalSDKCRC32Request(t)
	source, err := CaptureGoX86SDKASMDataSource(ctx, GoX86SDKASMDataCaptureRequest{
		GoBinary: request.GoBinary, Directory: request.Directory, PackagePath: request.PackagePath,
		Target: request.Target, Environment: request.Environment,
	})
	if err != nil {
		t.Fatal("actual complete SDK source/header/CPP/object precondition", err)
	}
	t.Cleanup(func() {
		if err := source.Close(); err != nil {
			t.Error(err)
		}
	})
	owner, err := CaptureGoX86SDKReadonlyLiteralDataOwner(ctx, source)
	if err != nil {
		t.Fatal("actual complete SDK literal initializer precondition", err)
	}
	t.Cleanup(func() {
		if err := owner.Close(); err != nil {
			t.Error(err)
		}
	})
	request.ReadonlyLiterals = owner
	return ctx, request, source
}

func numericalSDKLiteralPackage(t *testing.T, source *GoX86SDKASMDataSource, name string) (GoPackage, *File, Options) {
	t.Helper()
	fset := token.NewFileSet()
	var syntax []*ast.File
	for _, name := range source.guard.selection.GoFiles {
		filename := filepath.Join(source.request.Directory, name)
		file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
		if err != nil {
			t.Fatal("original complete selected SDK Go source", err)
		}
		syntax = append(syntax, file)
	}
	imports := importer.ForCompiler(fset, "gc", func(path string) (io.ReadCloser, error) {
		if mapped, found := source.guard.selection.ImportMap[path]; found {
			path = mapped
		}
		selected, found := source.guard.imports[path]
		if !found || selected.Export == "" {
			return nil, fmt.Errorf("missing actual SDK dependency export: %s", path)
		}
		return os.Open(selected.Export)
	})
	arch := strings.TrimPrefix(source.request.Target, "linux/")
	sizes := types.SizesFor("gc", arch)
	typed, err := (&types.Config{Sizes: sizes, Importer: imports, GoVersion: "go1.27"}).Check(source.request.PackagePath, fset, syntax, nil)
	if err != nil {
		t.Fatal("independently typed original complete SDK package", err)
	}
	pkg := GoPackage{Path: typed.Path(), Types: typed, Syntax: syntax, Imports: make(map[string]*types.Package)}
	for _, dependency := range typed.Imports() {
		pkg.Imports[dependency.Path()] = dependency
	}
	ordinal := -1
	for index, selected := range source.guard.selection.SFiles {
		if selected == name {
			if ordinal != -1 {
				t.Fatal("ambiguous original selected SDK assembly")
			}
			ordinal = index
		}
	}
	if ordinal < 0 || ordinal >= len(source.observation.CPP) {
		t.Fatal("missing complete SDK assembly CPP observation")
	}
	file, err := Parse(ArchAMD64, source.observation.CPP[ordinal].Output)
	if err != nil {
		t.Fatal(err)
	}
	options := Options{Goarch: arch, TargetTriple: "x86_64-unknown-linux-gnu",
		Sigs: make(map[string]FuncSig), ResolveSym: func(symbol string) string {
			return numericalNativeName(symbol, pkg.Path)
		}}
	if arch == "386" {
		options.TargetTriple = "i386-unknown-linux-gnu"
	}
	for _, fn := range file.Funcs {
		name := strings.TrimPrefix(fn.Sym, "·")
		declaration, ok := typed.Scope().Lookup(name).(*types.Func)
		if !ok {
			t.Fatal("original TEXT lacks its actual Go function declaration", fn.Sym)
		}
		name = numericalNativeName(fn.Sym, pkg.Path)
		sig, err := DeriveGoFuncABI0Signature(name, declaration, arch, sizes)
		if err != nil {
			t.Fatal(err)
		}
		options.Sigs[name] = sig
	}
	return pkg, file, options
}

func TestGoX86SDKNumericalReadLiteralOriginalCRC32LLVM22(t *testing.T) {
	ctx, request, observed := numericalSDKLiteralFixture(t)
	pkg, file, options := numericalSDKLiteralPackage(t, observed, request.AssemblyFile)
	if len(file.Funcs) != 3 || len(file.Globl) != 4 {
		t.Fatal("original complete CRC assembly changed; update the independent inventory")
	}
	source, err := CaptureGoX86SDKNumericalReadSource(ctx, request)
	if err != nil {
		t.Fatal("live initializer and CPU owner do not compose through the original SDK source", err)
	}
	t.Cleanup(func() { closeNumericalConsumerSource(t, source) })
	bound, err := source.BindNumericalReads(ctx, pkg, file, options)
	if err != nil {
		t.Fatal("complete original SDK numerical/data binding", err)
	}
	if len(source.reads) != 1 || source.reads[0].variable.NativeName != "internal/cpu.X86" || source.reads[0].width != 1 {
		t.Fatal("private literals acquired numerical Go variable authority")
	}
	for _, api := range []string{"text", "module", "go-module"} {
		t.Run(api, func(t *testing.T) {
			ir := numericalFamilyTranslate(t, api, pkg, file, bound)
			if !strings.Contains(ir, "load volatile i8") || !strings.Contains(ir, "external global [") ||
				strings.Count(ir, "internal constant [") != 4 || !strings.Contains(ir, "align 32") {
				t.Fatal("original mutable CPU read/private literal roles differ", ir)
			}
			numericalLLVM22Axes(t, ir, options.TargetTriple, nil)
		})
	}
}

func TestGoX86SDKNumericalReadLiteralRequestBoundaries(t *testing.T) {
	ctx, request, observed := numericalSDKLiteralFixture(t)
	for _, test := range []struct {
		name   string
		mutate func(*GoX86NumericalReadCaptureRequest)
	}{
		{"different package", func(r *GoX86NumericalReadCaptureRequest) { r.PackagePath = "internal/chacha8rand" }},
		{"different directory", func(r *GoX86NumericalReadCaptureRequest) { r.Directory = filepath.Dir(r.Directory) }},
		{"different target", func(r *GoX86NumericalReadCaptureRequest) { r.Target = "linux/386" }},
		{"unselected assembly", func(r *GoX86NumericalReadCaptureRequest) { r.AssemblyFile = "unselected_amd64.s" }},
		{"different tool", func(r *GoX86NumericalReadCaptureRequest) { r.GoBinary += ".foreign" }},
		{"different profile", func(r *GoX86NumericalReadCaptureRequest) {
			r.Environment = append(append([]string(nil), r.Environment...), "GOFLAGS=-N")
		}},
		{"copied literal owner", func(r *GoX86NumericalReadCaptureRequest) {
			copy := reflect.New(reflect.TypeOf(r.ReadonlyLiterals).Elem())
			copy.Elem().Set(reflect.ValueOf(r.ReadonlyLiterals).Elem())
			r.ReadonlyLiterals = copy.Interface().(*GoX86SDKReadonlyLiteralDataOwner)
		}},
		{"uncaptured literal owner", func(r *GoX86NumericalReadCaptureRequest) {
			r.ReadonlyLiterals = new(GoX86SDKReadonlyLiteralDataOwner)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			changed := request
			test.mutate(&changed)
			source, err := CaptureGoX86SDKNumericalReadSource(ctx, changed)
			if source != nil {
				closeNumericalConsumerSource(t, source)
			}
			if source != nil || !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatal("foreign source/profile/owner published a numerical token", err)
			}
			if err := request.ReadonlyLiterals.Verify(ctx); err != nil {
				t.Fatal("failed join revoked the independent original initializer", err)
			}
			if err := observed.Verify(ctx); err != nil {
				t.Fatal("failed join changed the independent SDK source", err)
			}
		})
	}
	if source, err := CaptureGoX86NumericalReadSource(ctx, request); source != nil || !errors.Is(err, ErrProbeNeedsContext) {
		if source != nil {
			closeNumericalConsumerSource(t, source)
		}
		t.Fatal("ordinary owned-main factory borrowed an SDK literal role", err)
	}
}

func TestGoX86SDKNumericalReadLiteralCompleteBindingBoundaries(t *testing.T) {
	ctx, request, observed := numericalSDKLiteralFixture(t)
	pkg, original, originalOptions := numericalSDKLiteralPackage(t, observed, request.AssemblyFile)
	for _, test := range []struct {
		name   string
		mutate func(*File, *Options)
	}{
		{"missing DATA", func(f *File, _ *Options) { f.Data = f.Data[1:] }},
		{"missing GLOBL", func(f *File, _ *Options) { f.Globl = f.Globl[1:] }},
		{"extra DATA", func(f *File, _ *Options) { f.Data = append(f.Data, f.Data[0]) }},
		{"extra GLOBL", func(f *File, _ *Options) { f.Globl = append(f.Globl, f.Globl[0]) }},
		{"DATA bytes", func(f *File, _ *Options) { f.Data[0].Value ^= 1 }},
		{"DATA offset", func(f *File, _ *Options) { f.Data[0].Off++ }},
		{"DATA width", func(f *File, _ *Options) { f.Data[0].Width-- }},
		{"DATA pointer", func(f *File, _ *Options) { f.Data[0].Addr = "internal∕cpu·X86(SB)" }},
		{"DATA namespace", func(f *File, _ *Options) { f.Data[0].Sym = strings.TrimSuffix(f.Data[0].Sym, "<>") }},
		{"GLOBL extent", func(f *File, _ *Options) { f.Globl[0].Size++ }},
		{"GLOBL mutability", func(f *File, _ *Options) { f.Globl[0].Flags = "0" }},
		{"missing TEXT", func(f *File, _ *Options) { f.Funcs = f.Funcs[1:] }},
		{"DATA function alias", func(_ *File, o *Options) {
			resolve := o.ResolveSym
			o.ResolveSym = func(symbol string) string {
				if strings.HasSuffix(symbol, "<>") {
					return "hash/crc32.ieeeCLMUL"
				}
				return resolve(symbol)
			}
		}},
		{"DATA numerical alias", func(_ *File, o *Options) {
			resolve := o.ResolveSym
			o.ResolveSym = func(symbol string) string {
				if strings.HasSuffix(symbol, "<>") {
					return "internal/cpu.X86"
				}
				return resolve(symbol)
			}
		}},
		{"DATA identity collapse", func(_ *File, o *Options) {
			resolve := o.ResolveSym
			o.ResolveSym = func(symbol string) string {
				if strings.HasSuffix(symbol, "<>") {
					return "hash/crc32.collapsed<>"
				}
				return resolve(symbol)
			}
		}},
		{"function DATA alias", func(_ *File, o *Options) {
			resolve := o.ResolveSym
			o.ResolveSym = func(symbol string) string {
				if symbol == "·ieeeCLMUL" {
					return "hash/crc32.r2r1<>"
				}
				return resolve(symbol)
			}
		}},
		{"numerical DATA alias", func(_ *File, o *Options) {
			resolve := o.ResolveSym
			o.ResolveSym = func(symbol string) string {
				if numericalNativeName(symbol, "hash/crc32") == "internal/cpu.X86" {
					return "hash/crc32.r2r1<>"
				}
				return resolve(symbol)
			}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			source, err := CaptureGoX86SDKNumericalReadSource(ctx, request)
			if err != nil {
				t.Fatal("actual joined SDK source precondition", err)
			}
			defer closeNumericalConsumerSource(t, source)
			file := *original
			file.Data = append([]DataStmt(nil), original.Data...)
			file.Globl = append([]GloblStmt(nil), original.Globl...)
			file.Funcs = append([]Func(nil), original.Funcs...)
			options := originalOptions
			test.mutate(&file, &options)
			bound, err := source.BindNumericalReads(ctx, pkg, &file, options)
			if bound.NumericalReads != nil || !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatal("changed complete source/initializer/namespace acquired a binding", err)
			}
		})
	}
}

func TestGoX86SDKNumericalReadLiteralOwnerLifetime(t *testing.T) {
	ctx, request, observed := numericalSDKLiteralFixture(t)
	pkg, file, options := numericalSDKLiteralPackage(t, observed, request.AssemblyFile)
	for _, change := range []string{"initializer close", "initializer report", "source owner replacement", "source owner removed"} {
		t.Run(change, func(t *testing.T) {
			owner, err := CaptureGoX86SDKReadonlyLiteralDataOwner(ctx, observed)
			if err != nil {
				t.Fatal(err)
			}
			defer owner.Close()
			local := request
			local.ReadonlyLiterals = owner
			source, err := CaptureGoX86SDKNumericalReadSource(ctx, local)
			if err != nil {
				t.Fatal("actual joined SDK source precondition", err)
			}
			defer closeNumericalConsumerSource(t, source)
			bound, err := source.BindNumericalReads(ctx, pkg, file, options)
			if err != nil {
				t.Fatal("actual complete SDK binding precondition", err)
			}
			switch change {
			case "initializer close":
				if err := owner.Close(); err != nil {
					t.Fatal(err)
				}
			case "initializer report":
				owner.report[0] ^= 1
			case "source owner replacement":
				source.request.ReadonlyLiterals = request.ReadonlyLiterals
			case "source owner removed":
				source.request.ReadonlyLiterals = nil
			}
			if ir, err := Translate(file, bound); ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatal("closed or foreign initializer published numerical IR", err)
			}
			source.request.ReadonlyLiterals = owner
			if !errors.Is(source.Verify(ctx), ErrProbeNeedsContext) {
				t.Fatal("restoring the initializer revived a revoked numerical source")
			}
			if err := observed.Verify(ctx); err != nil {
				t.Fatal("derived-token failure revoked the independent original SDK observation", err)
			}
		})
	}
}
