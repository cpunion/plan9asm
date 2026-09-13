package main

import (
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/xgo-dev/plan9asm"
	"golang.org/x/tools/go/packages"
)

func TestAsmFilesOfPkgSkipsCommentOnlyAssembly(t *testing.T) {
	dir := t.TempDir()
	comments := filepath.Join(dir, "comments.s")
	code := filepath.Join(dir, "code.s")
	include := filepath.Join(dir, "include.s")
	for path, contents := range map[string]string{
		comments: "//go:build amd64\n\n/* license only */\n",
		code:     "// comment\nTEXT ·f(SB),0,$0-0\n",
		include:  "#include \"textflag.h\"\n",
	} {
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	missing := filepath.Join(dir, "missing.s")
	pkg := &packages.Package{OtherFiles: []string{comments, code, include, missing}}
	want := []string{code, include, missing}
	if got := asmFilesOfPkg(pkg); !reflect.DeepEqual(got, want) {
		t.Fatalf("asmFilesOfPkg() = %#v, want %#v", got, want)
	}
}

func TestFilterPackagesByModuleExcludesNestedModules(t *testing.T) {
	pkgs := []*packages.Package{
		{PkgPath: "example.com/root/pkg", Module: &packages.Module{Path: "example.com/root"}},
		{PkgPath: "example.com/root/v2", Module: &packages.Module{Path: "example.com/root/v2"}},
		{PkgPath: "example.com/root/vendorless", Module: nil},
	}
	want := []*packages.Package{pkgs[0]}
	if got := filterPackagesByModule(pkgs, "example.com/root"); !reflect.DeepEqual(got, want) {
		t.Fatalf("filterPackagesByModule() = %#v, want %#v", got, want)
	}
}

func TestDefaultMatrixTargetsCoversEveryPlan9Architecture(t *testing.T) {
	want := []targetSpec{
		{Goos: "darwin", Goarch: "amd64"},
		{Goos: "darwin", Goarch: "arm64"},
		{Goos: "linux", Goarch: "386"},
		{Goos: "linux", Goarch: "amd64"},
		{Goos: "linux", Goarch: "arm"},
		{Goos: "linux", Goarch: "arm64"},
		{Goos: "windows", Goarch: "386"},
		{Goos: "windows", Goarch: "amd64"},
		{Goos: "windows", Goarch: "arm64"},
		{Goos: "js", Goarch: "wasm"},
		{Goos: "wasip1", Goarch: "wasm"},
	}
	if got := defaultMatrixTargets(); !reflect.DeepEqual(got, want) {
		t.Fatalf("defaultMatrixTargets() = %#v, want %#v", got, want)
	}
}

func TestExternalCorpusTargetArchitectureAndTriple(t *testing.T) {
	tests := []struct {
		goos       string
		goarch     string
		wantArch   plan9asm.Arch
		wantTriple string
	}{
		{goos: "linux", goarch: "arm", wantArch: plan9asm.ArchARM, wantTriple: "armv7-unknown-linux-gnueabihf"},
		{goos: "js", goarch: "wasm", wantArch: plan9asm.ArchWASM, wantTriple: "wasm32-unknown-unknown"},
		{goos: "wasip1", goarch: "wasm", wantArch: plan9asm.ArchWASM, wantTriple: "wasm32-wasi"},
	}
	for _, test := range tests {
		t.Run(test.goos+"/"+test.goarch, func(t *testing.T) {
			arch, err := toPlan9Arch(test.goarch)
			if err != nil {
				t.Fatal(err)
			}
			if arch != test.wantArch {
				t.Fatalf("toPlan9Arch(%q) = %q, want %q", test.goarch, arch, test.wantArch)
			}
			if got := targetTriple(test.goos, test.goarch); got != test.wantTriple {
				t.Fatalf("targetTriple(%q, %q) = %q, want %q", test.goos, test.goarch, got, test.wantTriple)
			}
		})
	}
}

func TestLLVMArgsAndFrameSlotsForTupleSliceParam(t *testing.T) {
	tup := types.NewTuple(types.NewVar(token.NoPos, nil, "b", types.NewSlice(types.Typ[types.Byte])))
	sz := types.SizesFor("gc", "amd64")
	args, slots, nextOff, err := llvmArgsAndFrameSlotsForTuple(tup, "amd64", sz, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(args, []plan9asm.LLVMType{"{ ptr, i64, i64 }"}) {
		t.Fatalf("args mismatch: %#v", args)
	}
	wantSlots := []plan9asm.FrameSlot{
		{Offset: 0, Type: plan9asm.Ptr, Index: 0, Field: 0},
		{Offset: 8, Type: plan9asm.I64, Index: 0, Field: 1},
		{Offset: 16, Type: plan9asm.I64, Index: 0, Field: 2},
	}
	if !reflect.DeepEqual(slots, wantSlots) {
		t.Fatalf("slots mismatch: got=%#v want=%#v", slots, wantSlots)
	}
	if nextOff != 24 {
		t.Fatalf("nextOff mismatch: got=%d want=24", nextOff)
	}
}

func TestLLVMArgsAndFrameSlotsForTupleSliceResultFlatten(t *testing.T) {
	tup := types.NewTuple(types.NewVar(token.NoPos, nil, "r", types.NewSlice(types.Typ[types.Byte])))
	sz := types.SizesFor("gc", "amd64")
	args, slots, nextOff, err := llvmArgsAndFrameSlotsForTuple(tup, "amd64", sz, 0, true)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(args, []plan9asm.LLVMType{plan9asm.Ptr, plan9asm.I64, plan9asm.I64}) {
		t.Fatalf("args mismatch: %#v", args)
	}
	wantSlots := []plan9asm.FrameSlot{
		{Offset: 0, Type: plan9asm.Ptr, Index: 0, Field: -1},
		{Offset: 8, Type: plan9asm.I64, Index: 1, Field: -1},
		{Offset: 16, Type: plan9asm.I64, Index: 2, Field: -1},
	}
	if !reflect.DeepEqual(slots, wantSlots) {
		t.Fatalf("slots mismatch: got=%#v want=%#v", slots, wantSlots)
	}
	if nextOff != 24 {
		t.Fatalf("nextOff mismatch: got=%d want=24", nextOff)
	}
}

func TestLLVMArgsAndFrameSlotsForTupleInterfaceParam(t *testing.T) {
	iface := types.NewInterfaceType(nil, nil)
	iface.Complete()
	tup := types.NewTuple(types.NewVar(token.NoPos, nil, "v", iface))
	sz := types.SizesFor("gc", "amd64")
	args, slots, nextOff, err := llvmArgsAndFrameSlotsForTuple(tup, "amd64", sz, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(args, []plan9asm.LLVMType{"{ ptr, ptr }"}) {
		t.Fatalf("args mismatch: %#v", args)
	}
	wantSlots := []plan9asm.FrameSlot{
		{Offset: 0, Type: plan9asm.Ptr, Index: 0, Field: 0},
		{Offset: 8, Type: plan9asm.Ptr, Index: 0, Field: 1},
	}
	if !reflect.DeepEqual(slots, wantSlots) {
		t.Fatalf("slots mismatch: got=%#v want=%#v", slots, wantSlots)
	}
	if nextOff != 16 {
		t.Fatalf("nextOff mismatch: got=%d want=16", nextOff)
	}
}
