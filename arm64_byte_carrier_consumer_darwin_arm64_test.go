//go:build go1.27

package plan9asm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

// This first RED keeps the real original caller's initialized native cell and
// adjacent canary. Typed-result-only equality cannot detect a dropped tail.
func TestARM64ByteCarrierMOVDUint32SameCarrier(t *testing.T) {
	directory := t.TempDir()
	var err error
	directory, err = filepath.EvalSymlinks(directory)
	if err != nil {
		t.Fatal(err)
	}
	packagePath := "example.invalid/nativecarrier"
	width := arm64NativeResultTailCase{op: "MOVD", resultBits: 32, storeBits: 64}
	callee := strings.ToLower(width.name())
	goSource := "package main\n" + strings.ReplaceAll(width.declaration(), width.name(), callee) +
		"func observe_" + callee + "(input uint64) (value,native,guard uint64)\n" +
		`func main() {
    value, native, guard := observe_movd_u32(0x123456789abcdef0)
    if value != 0x9abcdef0 || native != 0x123456789abcdef0 || guard != 0x7ac35ef012345678 {
        panic(1)
    }
    println("BYTE_CARRIER_MOVD32", value, native, guard)
}
`
	assembly := strings.ReplaceAll(width.source()+arm64NativeResultTailObserver(width), width.name(), callee)
	for name, data := range map[string]string{
		"go.mod":  "module " + packagePath + "\ngo 1.20\n",
		"decl.go": goSource, "original_arm64.s": assembly,
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	request := arm64ByteCarrierRequest(t, directory, packagePath)
	request.Target = "darwin/arm64"
	for index, pair := range request.Environment {
		if strings.HasPrefix(pair, "GOOS=") {
			request.Environment[index] = "GOOS=darwin"
		}
	}
	ctx, cancel := runtimeOracleStageContext(t)
	defer cancel()
	command := exec.CommandContext(ctx, request.GoBinary, "run", "-p=1", "-mod=readonly", ".")
	command.Dir, command.Env = request.Directory, request.Environment
	output, err := runtimeOracleCombinedOutput(ctx, command)
	want := fmt.Sprintf("BYTE_CARRIER_MOVD32 %d %d %d\n", uint64(0x9abcdef0), uint64(0x123456789abcdef0), uint64(0x7ac35ef012345678))
	if err != nil || string(output) != want {
		t.Fatalf("unchanged original Go whole-byte/canary witness: %v\n%s", err, output)
	}
	t.Logf("actual original Go same-carrier witness: %s", strings.TrimSpace(string(output)))
	source, err := CaptureGoARM64ByteCarrierSource(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	if len(source.report.Calls) != 1 || len(source.report.Calls[0].SPWrites) != 1 || source.report.Calls[0].SPWrites[0].Bytes != 8 {
		t.Fatal("original native-width same-call observation is incomplete")
	}
	pkg := mustGoPackage(t, packagePath, goSource)
	resolve := testResolveSym(pkg.Path)
	triple := "aarch64-apple-darwin"
	llc, clang := findLLVM22Tool("llc"), findLLVM22Tool("clang")
	if llc == "" || clang == "" {
		t.Fatal("LLVM 22 llc and clang are required")
	}
	for _, api := range []string{"text", "module", "Go"} {
		t.Run(api, func(t *testing.T) {
			llvmContext := llvm.NewContext()
			defer llvmContext.Dispose()
			var ir string
			var failure error
			if api == "Go" {
				translated, err := translateGoModuleInContext(llvmContext, pkg, []byte(assembly), GoModuleOptions{
					GOOS: "darwin", GOARCH: "arm64", TargetTriple: triple, ResolveSym: resolve,
					ARM64ByteCarrierSource: source,
				})
				if translated != nil && !translated.Module.IsNil() {
					ir = translated.Module.String()
					translated.Module.Dispose()
				}
				failure = err
			} else {
				file, err := Parse(ArchARM64, assembly)
				if err != nil {
					t.Fatal(err)
				}
				signatures, err := goSigsForAsmFile(pkg, file, resolve, "arm64", nil, nil, nil)
				if err != nil {
					t.Fatal(err)
				}
				options := Options{Goarch: "arm64", TargetTriple: triple, ResolveSym: resolve, Sigs: signatures,
					ARM64ByteCarrierSource: source}
				if api == "text" {
					ir, failure = Translate(file, options)
				} else {
					module, err := TranslateModuleInContext(llvmContext, file, options)
					if !module.IsNil() {
						ir = module.String()
						module.Dispose()
					}
					failure = err
				}
			}
			if failure != nil {
				t.Fatalf("live original same-carrier MOVD uint32 still lacks its bounded native transport: %v", failure)
			}
			ir += fmt.Sprintf(`
define i32 @byte_carrier_check(i64 %%input) {
    %%result = call { i64, i64, i64 } %s(i64 %%input)
    %%value = extractvalue { i64, i64, i64 } %%result, 0
    %%native = extractvalue { i64, i64, i64 } %%result, 1
    %%guard = extractvalue { i64, i64, i64 } %%result, 2
    %%low = and i64 %%input, 4294967295
    %%typed_ok = icmp eq i64 %%value, %%low
    %%native_ok = icmp eq i64 %%native, %%input
    %%guard_ok = icmp eq i64 %%guard, %d
    %%both = and i1 %%typed_ok, %%native_ok
    %%all = and i1 %%both, %%guard_ok
    %%status = select i1 %%all, i32 0, i32 1
    ret i32 %%status
}
`, llvmGlobal(packagePath+".observe_"+callee), uint64(0x7ac35ef012345678))
			mainC := `#include <stdint.h>
extern int32_t byte_carrier_check(uint64_t input);
int main(void) {
    return byte_carrier_check(UINT64_C(0x123456789abcdef0));
}
`
			compileAndRunRuntimeTest(t, llc, clang, "native_carrier_movd32", ir, mainC)
		})
	}
}
