package plan9asm

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

func runtime386CoreOriginalGoObject(t *testing.T) {
	t.Helper()
	driver, driverHash := x86LEAActualSDKDriver(t)
	include := filepath.Join(runtime.GOROOT(), "pkg", "include")
	header := filepath.Join(include, "textflag.h")
	before, err := os.ReadFile(header)
	if err != nil {
		t.Fatal(err)
	}
	// The original TEXT bytes remain unchanged. Only the actual selected
	// SDK's header supplies their existing NOSPLIT spelling to Go's assembler.
	source := "#include \"textflag.h\"\n" + runtime386CoreAssembly
	directory := t.TempDir()
	input, object := filepath.Join(directory, "core.s"), filepath.Join(directory, "core.o")
	if err := os.WriteFile(input, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := runtimeOracleStageContext(t)
	defer cancel()
	command := exec.CommandContext(ctx, driver, "tool", "asm", "-I", include, "-p", "runtimecore", "-o", object, input)
	command.Env = append(os.Environ(), "GOOS=windows", "GOARCH=386", "CGO_ENABLED=0", "GOTOOLCHAIN=local", "GOENV=off", "GOWORK=off")
	output, err := runtimeOracleCombinedOutput(ctx, command)
	if failure := goAssemblerOracleOutcome(input, object, output, err, ctx.Err(), true, "386"); failure != nil {
		t.Fatalf("complete original Windows386 Go assembly: %v\n%s", failure, output)
	}
	if after, err := os.ReadFile(header); err != nil || !bytes.Equal(before, after) {
		t.Fatal("selected SDK's original header changed during assembly", err)
	}
	x86LEAReadObjectReceipt(t, "386", "windows", source, input, driver, driverHash, object)
}

// This portable owner observes all original source and manual transport
// signatures. It does not execute the Windows C driver or invent a GoModule
// ABI0 declaration for cas64's existing SI/AX/DX/BX/CX register convention.
func TestRuntimeExec386CoreCompleteSourceGoAndLLVMObjects(t *testing.T) {
	if got := fmt.Sprintf("%x", sha256.Sum256([]byte(runtime386CoreAssembly))); got != "8cfb20eb17f5105cb92fc184bffa1333007c80fcfe19865795ed15b0c6d1fbb2" {
		t.Fatal("original complete Windows386 source changed", got)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256([]byte(runtime386CoreCSource))); got != "50eb6e546b44eff9347c96721d22f49eb70e0be57b2be06b75c2324cae12916a" {
		t.Fatal("original Windows native C driver changed", got)
	}
	want := []string{"addCarry", "cpuidProbe", "frameAddress", "copyForward", "copyBackward", "copyBackwardRestoredDF",
		"fillWords", "findByte", "zeroLengthScanPreservesZF", "negateHasCarry", "floorValue", "uint32ToFloat64",
		"cas64", "xadd32", "identity64", "identityFloat64", "readFS", "roundNearest"}
	file, err := Parse(ArchAMD64, runtime386CoreAssembly)
	if err != nil {
		t.Fatal(err)
	}
	options := runtime386CoreOptions()
	if len(file.Funcs) != 18 || len(options.Sigs) != 18 || options.Goarch != "386" || options.TargetTriple != "i686-pc-windows-msvc" {
		t.Fatal("complete original source/signature/target inventory changed")
	}
	for index, function := range file.Funcs {
		if function.Sym != want[index] || options.Sigs[function.Sym].Name != want[index] {
			t.Fatal("original TEXT order or manual signature identity changed", function.Sym)
		}
	}
	if !reflect.DeepEqual(options.Sigs["cas64"].ArgRegs, []Reg{SI, AX, DX, BX, CX}) || options.Sigs["cas64"].Ret != I1 {
		t.Fatal("original custom-register cas64 transport was replaced by a fabricated Go ABI")
	}
	t.Run("original-Go-windows386", runtime386CoreOriginalGoObject)
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("complete original Windows386 objects require LLVM22 llc")
	}
	for _, api := range []string{"text", "module"} {
		t.Run(api, func(t *testing.T) {
			file, err := Parse(ArchAMD64, runtime386CoreAssembly)
			if err != nil {
				t.Fatal(err)
			}
			options := runtime386CoreOptions()
			var ir string
			if api == "text" {
				ir, err = Translate(file, options)
			} else {
				context := llvm.NewContext()
				defer context.Dispose()
				module, moduleErr := TranslateModuleInContext(context, file, options)
				if !module.IsNil() {
					defer module.Dispose()
					ir = module.String()
				}
				err = moduleErr
			}
			if err != nil || ir == "" {
				t.Fatal("complete unchanged Windows386 source has an unclosed contract", err)
			}
			for _, name := range want {
				x86StringMemoryFunction(t, ir, name)
			}
			if !strings.Contains(ir, "target triple = \"i686-pc-windows-msvc\"") {
				t.Fatal("portable observation changed the original Windows386 target")
			}
			for _, level := range []string{"0", "2"} {
				input := ir
				if level == "2" {
					input = x86DwordProductIROptimize(t, input)
				}
				for _, backend := range []string{"0", "2"} {
					t.Run("IR"+level+"/backend"+backend, func(t *testing.T) {
						x86PackedStringCompileObject(t, llc, options.TargetTriple, input, "-O"+backend)
					})
				}
			}
		})
	}
}
