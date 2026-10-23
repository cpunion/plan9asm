//go:build go1.27

package main

import (
	"crypto/sha256"
	"go/types"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm"
)

// This is the complete actual current SDK file, not a hasADX instruction
// prototype. Its ordinary declaration route must precede an expensive SDK
// reload/capture: one attempt suffices without changing any native ABI rules.
func TestCompileOneOrdinaryGlobalDeclarationsSDKMathBigWholeFile(t *testing.T) {
	setX86SDKNumericalCLIProfile(t, "explicit")
	t.Setenv("GOROOT", runtime.GOROOT())
	directory := filepath.Join(runtime.GOROOT(), "src", "math", "big")
	t.Chdir(directory)
	ctx := arm64PrivateCLIContext(t)
	selected, environment, err := loadPkgsInEnvironment("linux", "amd64", []string{"."}, nil, "", true, false, nil)
	if err != nil || len(selected) != 1 || selected[0].PkgPath != "math/big" {
		t.Fatalf("actual selected SDK math/big load failed: %v %+v", err, selected)
	}
	pkg := selected[0]
	variable, ok := pkg.Types.Scope().Lookup("hasADX").(*types.Var)
	if !ok || variable.Pkg() != pkg.Types || variable.IsField() ||
		!types.Identical(variable.Type(), types.Typ[types.Bool]) || pkg.TypesSizes.Sizeof(variable.Type()) != 1 {
		t.Fatal("the selected SDK lacks its real same-package mutable byte feature flag")
	}
	input := ""
	for _, name := range pkg.OtherFiles {
		if filepath.Base(name) == "arith_amd64.s" {
			if input != "" {
				t.Fatal("duplicate selected SDK arithmetic file")
			}
			input = name
		}
	}
	if input == "" || filepath.Dir(input) != pkg.Dir {
		t.Fatal("complete selected SDK arithmetic file is absent")
	}
	original, err := os.ReadFile(input)
	if err != nil || !strings.Contains(string(original), "CMPB") || !strings.Contains(string(original), "hasADX") {
		t.Fatal("original SDK counterexample is absent", err)
	}
	source, reads, err := plan9asm.ReadGoAssemblySourceWithReport(input, pkg.Dir)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := plan9asm.ParseWithDefines(plan9asm.ArchAMD64, string(source), plan9asm.GoAssemblerDefines("linux", "amd64"))
	if err != nil {
		t.Fatal("the complete original SDK source did not parse", err)
	}
	wanted := map[string]bool{"·addVV": true, "·subVV": true, "·lshVU": true, "·rshVU": true, "·mulAddVWW": true, "·addMulVVWW": true}
	if len(parsed.Funcs) != len(wanted) {
		t.Fatal("the original six-function SDK source inventory changed", len(parsed.Funcs))
	}
	seen := make(map[string]bool)
	for _, function := range parsed.Funcs {
		if !wanted[function.Sym] || seen[function.Sym] {
			t.Fatal("unexpected or repeated original SDK TEXT", function.Sym)
		}
		seen[function.Sym] = true
	}
	config, err := resolveCompileConfig(true, "", true, 2)
	if err != nil {
		t.Fatal(err)
	}
	config.Context = ctx
	config.x86CallerSelections = captureX86CallerSelectionsInEnvironment(selected, "linux/amd64", environment)
	output := filepath.Join(t.TempDir(), "arith.ll")
	if err := compileOneAttempt(pkg, plan9asm.ArchAMD64, "linux", "amd64", "x86_64-unknown-linux-gnu",
		asmTask{PkgPath: pkg.PkgPath, AsmFile: input, OutLL: output}, false, config); err != nil {
		t.Fatal("ordinary complete SDK declarations fell through to SDK capture/retranslation", err)
	}
	ir, err := os.ReadFile(output)
	if err != nil || !strings.Contains(string(ir), `load volatile i8, ptr @"math/big.hasADX"`) ||
		!strings.Contains(string(ir), `@"math/big.hasADX" = external global `) {
		t.Fatal("the real mutable flag lost its exact volatile external byte read", err)
	}
	for symbol := range wanted {
		found := false
		for _, line := range strings.Split(string(ir), "\n") {
			found = found || strings.HasPrefix(line, "define ") && strings.Contains(line, `@"math/big.`+strings.TrimPrefix(symbol, "·")+`"(`)
		}
		if !found {
			t.Fatal("an original SDK function was omitted from its LLVM object", symbol)
		}
	}
	if info, err := os.Lstat(strings.TrimSuffix(output, ".ll") + ".o"); err != nil || !info.Mode().IsRegular() || info.Size() == 0 {
		t.Fatal("the complete original SDK LLVM22 object is missing", err)
	}
	for name, expected := range reads.Inputs {
		actual, err := os.ReadFile(name)
		if err != nil || expected != featureBytesSHA256(actual) {
			t.Fatal("original SDK source/header bytes changed", name, err)
		}
	}
	if after, err := os.ReadFile(input); err != nil || sha256.Sum256(after) != sha256.Sum256(original) {
		t.Fatal("the original SDK arithmetic source changed", err)
	}
}

// A complete GoModule consumes private tail TEXTs as source graph nodes, not
// callable/public metadata entries. All four real files formerly passed the
// raw CLI path; declaration forwarding must preserve their actual objects.
func TestCompileOneOrdinaryGlobalDeclarationsSDKPrivateTailWholeFiles(t *testing.T) {
	setX86SDKNumericalCLIProfile(t, "explicit")
	t.Setenv("GOROOT", runtime.GOROOT())
	t.Chdir(filepath.Join(runtime.GOROOT(), "src", "internal", "bytealg"))
	ctx := arm64PrivateCLIContext(t)
	selected, environment, err := loadPkgsInEnvironment("linux", "amd64", []string{"."}, nil, "", true, false, nil)
	if err != nil || len(selected) != 1 || selected[0].PkgPath != "internal/bytealg" {
		t.Fatalf("actual selected SDK bytealg load failed: %v %+v", err, selected)
	}
	pkg := selected[0]
	config, err := resolveCompileConfig(true, "", true, 2)
	if err != nil {
		t.Fatal(err)
	}
	config.Context = ctx
	config.x86CallerSelections = captureX86CallerSelectionsInEnvironment(selected, "linux/amd64", environment)
	for _, test := range []struct {
		file, helper string
		public       []string
	}{
		{"compare_amd64.s", "cmpbody<>", []string{"internal/bytealg.Compare", "runtime.cmpstring"}},
		{"count_amd64.s", "countbody<>", []string{"internal/bytealg.Count", "internal/bytealg.CountString"}},
		{"index_amd64.s", "indexbody<>", []string{"internal/bytealg.Index", "internal/bytealg.IndexString"}},
		{"indexbyte_amd64.s", "indexbytebody<>", []string{"internal/bytealg.IndexByte", "internal/bytealg.IndexByteString"}},
	} {
		t.Run(test.file, func(t *testing.T) {
			input := ""
			for _, name := range pkg.OtherFiles {
				if filepath.Base(name) == test.file {
					if input != "" {
						t.Fatal("duplicate original SDK assembly member")
					}
					input = name
				}
			}
			if input == "" {
				t.Fatal("selected SDK assembly file missing")
			}
			source, reads, err := plan9asm.ReadGoAssemblySourceWithReport(input, pkg.Dir)
			if err != nil || !strings.Contains(string(source), "TEXT "+test.helper) &&
				!strings.Contains(string(source), "TEXT\t"+test.helper) {
				t.Fatal("original complete source lost its private tail TEXT", err)
			}
			output := filepath.Join(t.TempDir(), "complete.ll")
			if err := compileOneAttempt(pkg, plan9asm.ArchAMD64, "linux", "amd64", "x86_64-unknown-linux-gnu",
				asmTask{PkgPath: pkg.PkgPath, AsmFile: input, OutLL: output}, false, config); err != nil {
				t.Fatal("complete original source rejected its authenticated private tail", err)
			}
			ir, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			definitions := 0
			for _, line := range strings.Split(string(ir), "\n") {
				if !strings.HasPrefix(line, "define ") {
					continue
				}
				definitions++
				matched := false
				for _, name := range test.public {
					matched = matched || strings.Contains(line, `@"`+name+`"(`)
				}
				if !matched {
					t.Fatal("private helper became a callable entry or a foreign public entry appeared", line)
				}
			}
			if definitions != len(test.public) {
				t.Fatal("the original public root inventory changed", definitions)
			}
			if info, err := os.Lstat(strings.TrimSuffix(output, ".ll") + ".o"); err != nil || !info.Mode().IsRegular() || info.Size() == 0 {
				t.Fatal("complete source LLVM22 object is absent", err)
			}
			for name, digest := range reads.Inputs {
				data, err := os.ReadFile(name)
				if err != nil || featureBytesSHA256(data) != digest {
					t.Fatal("original SDK source/header changed", name, err)
				}
			}
		})
	}
}
