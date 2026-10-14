//go:build go1.27

package main

import (
	"crypto/sha256"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm"
	"github.com/xgo-dev/plan9asm/gotoolprofile"
	"golang.org/x/tools/go/packages"
)

func TestCompileOneX86SDKOriginalMapsLiteralMinimumSource(t *testing.T) {
	for _, goos := range []string{"linux", "darwin", "windows"} {
		t.Run(goos, func(t *testing.T) {
			setX86SDKNumericalCLIProfile(t, "effective-defaults")
			t.Setenv("GOOS", goos)
			ctx := arm64PrivateCLIContext(t)
			directory := filepath.Join(runtime.GOROOT(), "src", "internal", "runtime", "maps")
			input := filepath.Join(directory, "memhash_amd64.s")
			original, err := os.ReadFile(input)
			if err != nil {
				t.Fatal(err)
			}
			actual, err := packages.Load(&packages.Config{Context: ctx, Dir: directory, Env: os.Environ(), Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedModule | packages.NeedDeps | packages.NeedImports | packages.NeedTypes | packages.NeedTypesSizes}, ".")
			if err != nil || len(actual) != 1 || len(actual[0].Errors) != 0 {
				t.Fatalf("actual complete SDK load: %v %+v", err, actual)
			}
			config, err := resolveCompileConfig(true, "", true, 2)
			if err != nil {
				t.Fatal(err)
			}
			config.Context = ctx
			config.x86CallerSelections = captureX86CallerSelections(actual, goos+"/amd64", nil)
			output := filepath.Join(t.TempDir(), "maps.ll")
			if err := compileOne(actual[0], plan9asm.ArchAMD64, goos, "amd64", targetTriple(goos, "amd64"), asmTask{PkgPath: actual[0].PkgPath, AsmFile: input, OutLL: output}, false, config); err != nil {
				t.Fatal("whole original SDK source/minimum/LLVM object", err)
			}
			ir, err := os.ReadFile(output)
			if err != nil || strings.Count(string(ir), "define ") != 2 || !strings.Contains(string(ir), "external global [128 x i8]") || strings.Count(string(ir), "internal constant [256 x i8]") != 2 {
				t.Fatal("whole original TEXT/DATA/runtime mutable source publication differs", err)
			}
			if info, err := os.Stat(strings.TrimSuffix(output, ".ll") + ".o"); err != nil || info.Size() == 0 {
				t.Fatal("LLVM22 complete object absent", err)
			}
			if goos == "linux" {
				folder, report := t.TempDir(), filepath.Join(t.TempDir(), "report.json")
				repository, err := filepath.Abs("../..")
				if err != nil {
					t.Fatal(err)
				}
				args, err := json.Marshal([]string{"-targets=linux/amd64", "-patterns=.", "-compile", "-keep-obj", "-strict-load", "-out=" + folder, "-report=" + report, "-repo-root=" + repository})
				if err != nil {
					t.Fatal(err)
				}
				stdout, stderr, err := gotoolprofile.RunBounded(ctx, directory, append(os.Environ(), "PLAN9ASM_CPUCLI_ARGS="+string(args)), os.Args[0], "-test.run=^TestFeatureConsumerCLIChild$")
				if err != nil {
					t.Fatalf("genuine original package public CLI: %v\n%s%s", err, stdout, stderr)
				}
				raw, err := os.ReadFile(report)
				var decoded matrixReport
				if err != nil || json.Unmarshal(raw, &decoded) != nil || decoded.Success != 2 || decoded.Failed != 0 || decoded.NotApplicable != 0 || decoded.TotalAsm != 2 || decoded.TotalTargets != 1 {
					t.Fatalf("genuine complete package denominator/result differs: %v\n%s", err, raw)
				}
			}
			after, err := os.ReadFile(input)
			if err != nil || sha256.Sum256(after) != sha256.Sum256(original) {
				t.Fatal("original SDK source changed", err)
			}
			t.Logf("complete SDK source SHA256=%x; 2 TEXT/all DATA logical object/public source route only, not native SDK entry or execution", sha256.Sum256(original))
		})
	}
}
