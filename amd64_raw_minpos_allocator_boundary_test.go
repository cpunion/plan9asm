package plan9asm

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

// These are explicit observations of an unresolved synthetic allocation, not
// successful stack semantics, source rejection, native bounds or runtime tests.
// Both original Go sources are legal. Public translation currently emits a
// >4 GiB allocation. The original object's llc O0 gate rejects its frame offsets;
// default O2 instead only warns. Neither grants native stack/return permission.
func TestX86RawMinimumPositionHugeSPAllocatorBoundaryObservation(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc not found")
	}
	// Keep the complete mode32 address inventory in both source forms. The
	// original encoded RET remains a negative stack/continuation control;
	// the raw compile-only probe replaces only that RET with encoded UD2.
	source, _, count, _ := x86MinimumPositionAddressSources("386", "giant", "UD2")
	if count != 240 {
		t.Fatalf("complete mode32 address denominator changed: %d", count)
	}
	namedCode := x86MinimumPositionOriginalGoBytes(t, "386", source)
	namedFile, err := Parse(ArchAMD64, source)
	if err != nil {
		t.Fatal(err)
	}
	rawSource, _, _, _ := x86MinimumPositionAddressSources("386", "giant", "RET")
	rawCode := x86MinimumPositionOriginalGoBytes(t, "386", rawSource)
	rawFile := x86RawStackObservingFormProbeFile(t, "386", "giant", rawCode)
	for _, fixture := range []struct {
		name   string
		file   *File
		source string
		code   []byte
	}{
		{"named", namedFile, source, namedCode},
		{"original-byte-probe", rawFile, rawSource, rawCode},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			file := fixture.file
			before := fmt.Sprintf("%#v", file.Funcs)
			for _, triple := range []string{"i386-unknown-linux-gnu", "i686-pc-windows-msvc"} {
				for _, api := range []string{"text", "module", "context"} {
					t.Run(triple+"/"+api, func(t *testing.T) {
						options := Options{Goarch: "386", TargetTriple: triple, Sigs: map[string]FuncSig{"giant": {Name: "giant", Ret: Void}}}
						var ir string
						var err error
						switch api {
						case "text":
							ir, err = Translate(file, options)
						case "module":
							module, callErr := TranslateModule(file, options)
							err = callErr
							if !module.IsNil() {
								defer module.Dispose()
								ir = module.String()
							}
						case "context":
							ctx := llvm.NewContext()
							defer ctx.Dispose()
							module, callErr := TranslateModuleInContext(ctx, file, options)
							err = callErr
							if !module.IsNil() {
								defer module.Dispose()
								ir = module.String()
							}
						}
						// The exact source envelope is [-2^31, 2^31-1+16), with
						// two existing 64-byte synthetic guards: 4,294,967,439 bytes.
						if err != nil || !strings.Contains(ir, "%local_stack = alloca [4294967439 x i8]") || before != fmt.Sprintf("%#v", file.Funcs) {
							t.Fatalf("allocator observation changed: %v; complete source unchanged=%t", err, before == fmt.Sprintf("%#v", file.Funcs))
						}
						dir := t.TempDir()
						input, object := filepath.Join(dir, "giant.ll"), filepath.Join(dir, "giant.o")
						if err := os.WriteFile(input, []byte(ir), 0600); err != nil {
							t.Fatal(err)
						}
						ctx, cancel := runtimeOracleStageContext(t)
						defer cancel()
						command := exec.CommandContext(ctx, llc, "-O0", "-filetype=obj", "-mtriple="+triple, "-o", object, input)
						output, err := runtimeOracleCombinedOutput(ctx, command)
						if ctx.Err() != nil {
							t.Fatalf("bounded compiler observation expired: %v: %s", err, output)
						}
						exit, ordinaryExit := err.(*exec.ExitError)
						if !ordinaryExit || exit.ExitCode() != 1 {
							t.Fatalf("expected plain LLVM32 O0 offset failure, got %v: %s", err, output)
						}
						diagnostic := "error: <unknown>:0:0: 64-bit offset calculated but target is 32-bit"
						warning := regexp.MustCompile(`^warning: <unknown>:0:0: stack frame size \(([0-9]+)\) exceeds limit \(4294967295\) in function 'giant'$`)
						errors, warnings := 0, 0
						text := strings.ReplaceAll(string(output), "\r\n", "\n")
						if strings.Contains(text, "\r") {
							t.Fatalf("malformed diagnostic line endings: %q", output)
						}
						for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
							if line == diagnostic {
								errors++
								continue
							}
							match := warning.FindStringSubmatch(line)
							if len(match) != 2 {
								t.Fatalf("unknown diagnostic is not the observed allocator boundary: %s", output)
							}
							frameBytes, err := strconv.ParseUint(match[1], 10, 64)
							if err != nil || frameBytes <= 4294967295 {
								t.Fatalf("compiler warning did not observe an oversized frame: %s", line)
							}
							warnings++
						}
						if warnings != 1 || errors == 0 {
							t.Fatalf("missing/changed exact allocator diagnostics: errors=%d warnings=%d", errors, warnings)
						}
						t.Logf("LLVM_ALLOCATOR_BOUNDARY_OBSERVATION: source SHA=%x; original Go code SHA=%x; %s %s emitted local_stack=4294967439 bytes; compiler O0 offset errors=%d, frame warnings=%d; no object/stack correctness/runtime or source-N/A credit", sha256.Sum256([]byte(fixture.source)), sha256.Sum256(fixture.code), fixture.name, api, errors, warnings)
					})
				}
			}
		})
	}
}
