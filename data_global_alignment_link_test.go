package plan9asm

import (
	"debug/elf"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

func dataGlobalAlignmentCommand(t *testing.T, directory string, environment []string, executable string, arguments ...string) []byte {
	t.Helper()
	ctx, cancel := runtimeOracleStageContext(t)
	defer cancel()
	command := exec.CommandContext(ctx, executable, arguments...)
	command.Dir = directory
	command.Env = append(os.Environ(), environment...)
	output, err := runtimeOracleCombinedOutput(ctx, command)
	if err != nil {
		t.Fatalf("%s %q: %v\n%s", filepath.Base(executable), arguments, err, output)
	}
	return output
}

func dataGlobalAlignmentGoProgram(goarch string, maximum int) (string, string) {
	var assembly, declarations, checks strings.Builder
	assembly.WriteString(dataGlobalAlignmentSource())
	declarations.WriteString("package main\nimport \"unsafe\"\n")
	ordinal := 0
	for _, row := range dataGlobalAlignmentSizes {
		alignment := row.alignment
		if alignment > maximum {
			alignment = maximum
		}
		for _, role := range []string{"rw", "ro", "local"} {
			name := fmt.Sprintf("asm_align_%s_n%d", role, row.size)
			if role == "local" {
				name += "<>"
			}
			word := 8
			if goarch == "386" || goarch == "arm" {
				word = 4
			}
			fmt.Fprintf(&assembly, "TEXT ·Address%d(SB),4,$0-%d\n", ordinal, word)
			switch goarch {
			case "amd64":
				fmt.Fprintf(&assembly, "LEAQ %s(SB),AX\nMOVQ AX,ret+0(FP)\n", name)
			case "386":
				fmt.Fprintf(&assembly, "MOVL $%s(SB),AX\nMOVL AX,ret+0(FP)\n", name)
			case "arm":
				fmt.Fprintf(&assembly, "MOVW $%s(SB),R0\nMOVW R0,ret+0(FP)\n", name)
			case "arm64":
				fmt.Fprintf(&assembly, "MOVD $%s(SB),R0\nMOVD R0,ret+0(FP)\n", name)
			case "wasm":
				fmt.Fprintf(&assembly, "Get SP\nI64Const $%s(SB)\nI64Store ret+0(FP)\n", name)
			}
			assembly.WriteString("RET\n")
			fmt.Fprintf(&declarations, "func Address%d() uintptr\n", ordinal)
			fmt.Fprintf(&checks, `{
    address := Address%d()
    if address == 0 || address%%%d != 0 { panic("%s alignment") }
    bytes := (*[%d]byte)(unsafe.Pointer(address))
    if bytes[0] != 0xa5 { panic("%s first byte") }
    for _, value := range bytes[1:] { if value != 0 { panic("%s zero tail") } }
}
`, ordinal, alignment, name, row.size, name, name)
			ordinal++
		}
	}
	declarations.WriteString("func main() {\n" + checks.String() + "println(\"actual Go assembly alignment and bytes PASS\")\n}\n")
	return assembly.String(), declarations.String()
}

func dataGlobalAlignmentGoBuild(t *testing.T, goarch, goos string, maximum int) string {
	t.Helper()
	directory := t.TempDir()
	assembly, program := dataGlobalAlignmentGoProgram(goarch, maximum)
	for name, source := range map[string]string{
		"go.mod": "module alignment_oracle\n\ngo 1.20\n", "main.go": program, "data.s": assembly,
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	executable := filepath.Join(directory, "original")
	environment := []string{"GOWORK=off", "GOENV=off", "GOTOOLCHAIN=local", "CGO_ENABLED=0", "GOARCH=" + goarch, "GOOS=" + goos}
	dataGlobalAlignmentCommand(t, directory, environment, "go", "build", "-p=2", "-o", executable, ".")
	return executable
}

func dataGlobalAlignmentELF(t *testing.T, path string, maximum int, relocated bool) {
	t.Helper()
	image, err := elf.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer image.Close()
	symbols, err := image.Symbols()
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range dataGlobalAlignmentSizes {
		alignment := row.alignment
		if alignment > maximum {
			alignment = maximum
		}
		for _, role := range []string{"rw", "ro", "local"} {
			name := fmt.Sprintf("asm_align_%s_n%d", role, row.size)
			matches := 0
			for _, symbol := range symbols {
				if symbol.Name != name {
					continue
				}
				matches++
				if symbol.Section == elf.SHN_UNDEF || int(symbol.Section) >= len(image.Sections) || symbol.Size != uint64(row.size) {
					t.Fatalf("missing complete native definition: %+v", symbol)
				}
				section := image.Sections[symbol.Section]
				if symbol.Value%uint64(alignment) != 0 || relocated && symbol.Value == 0 ||
					!relocated && section.Addralign < uint64(alignment) {
					t.Errorf("%s: symbol address=%#x section alignment=%d, Go requires %d", name, symbol.Value, section.Addralign, alignment)
				}
				bytes, err := section.Data()
				start := symbol.Value - section.Addr
				if err != nil || start > uint64(len(bytes)) || symbol.Size > uint64(len(bytes))-start {
					t.Fatal("native symbol has no complete section bytes", symbol, err)
				}
				for index, value := range bytes[start : start+symbol.Size] {
					want := byte(0)
					if index == 0 {
						want = 0xa5
					}
					if value != want {
						t.Fatalf("%s byte%d=%#x, want %#x", name, index, value, want)
					}
				}
			}
			if matches != 1 {
				t.Fatalf("%s has %d native definitions, want one", name, matches)
			}
		}
	}
}

func TestDataGlobalAlignmentActualGoLinkedFiveArchitectures(t *testing.T) {
	for _, target := range dataGlobalAlignmentTargets {
		t.Run(target.goarch, func(t *testing.T) {
			executable := dataGlobalAlignmentGoBuild(t, target.goarch, target.goos, target.maximum)
			if target.arch != ArchWASM {
				dataGlobalAlignmentELF(t, executable, target.maximum, true)
				return
			}
			ctx, cancel := runtimeOracleStageContext(t)
			defer cancel()
			command, err := wasmPackedGoRuntimeCommand(ctx, testGOROOT(t), wasmPackedNode(t), executable)
			if err != nil {
				t.Fatal(err)
			}
			command.Env = append(os.Environ(), "GOMAXPROCS=1")
			output, err := runtimeOracleCombinedOutput(ctx, command)
			if err != nil || strings.TrimSpace(string(output)) != "actual Go assembly alignment and bytes PASS" {
				t.Fatalf("actual Go wasm linked addresses/bytes: %v\n%s", err, output)
			}
		})
	}
}

func dataGlobalAlignmentModule(t *testing.T, arch Arch, triple, route string) string {
	t.Helper()
	file, err := Parse(arch, dataGlobalAlignmentSource())
	if err != nil {
		t.Fatal(err)
	}
	options := Options{TargetTriple: triple, ResolveSym: dataGlobalAlignmentResolve}
	if route == "text" {
		ir, err := Translate(file, options)
		if err != nil {
			t.Fatal(err)
		}
		return ir
	}
	ctx := llvm.NewContext()
	defer ctx.Dispose()
	module, err := translateModuleDirectInContext(ctx, file, options)
	if !module.IsNil() {
		defer module.Dispose()
	}
	if err != nil {
		t.Fatal("direct emitter must not fall back", err)
	}
	return module.String()
}

func TestDataGlobalAlignmentLLVMObjectsAndLinkedAddresses(t *testing.T) {
	llc, linker, wasmLinker := findLLVM22Tool("llc"), findLLVM22Tool("ld.lld"), findLLVM22Tool("wasm-ld")
	if llc == "" || linker == "" || wasmLinker == "" {
		t.Fatal("LLVM22 llc, ld.lld and wasm-ld required")
	}
	for _, target := range dataGlobalAlignmentTargets {
		for _, route := range []string{"text", "direct"} {
			t.Run(target.goarch+"/"+route, func(t *testing.T) {
				ir := dataGlobalAlignmentModule(t, target.arch, target.triple, route)
				directory := t.TempDir()
				input, object, linked := filepath.Join(directory, "data.ll"), filepath.Join(directory, "data.o"), filepath.Join(directory, "linked")
				if err := os.WriteFile(input, []byte(ir), 0600); err != nil {
					t.Fatal(err)
				}
				dataGlobalAlignmentCommand(t, directory, nil, llc, "-O0", "-filetype=obj", input, "-o", object)
				if target.arch != ArchWASM {
					dataGlobalAlignmentELF(t, object, target.maximum, false)
					dataGlobalAlignmentCommand(t, directory, nil, linker, "-shared", object, "-o", linked)
					dataGlobalAlignmentELF(t, linked, target.maximum, true)
					return
				}
				dataGlobalAlignmentCommand(t, directory, nil, wasmLinker, "--no-entry", "--export-all", object, "-o", linked)
				var checks strings.Builder
				for _, row := range dataGlobalAlignmentSizes {
					for _, role := range []string{"rw", "ro"} {
						name := fmt.Sprintf("asm_align_%s_n%d", role, row.size)
						fmt.Fprintf(&checks, `{
  const address = e[%q].value;
  if (!address || address%%%d) throw Error(%q + " alignment=" + address);
  for (let i=0;i<%d;i++) if (bytes[address+i] !== (i ? 0 : 0xa5)) throw Error(%q + " bytes");
}
`, name, row.alignment, name, row.size, name)
					}
				}
				script := `const fs=require("fs");
WebAssembly.instantiate(fs.readFileSync(process.argv[1])).then(({instance})=>{
const e=instance.exports, bytes=new Uint8Array(e.memory.buffer);
` + checks.String() + `console.log("LLVM linked wasm alignment and bytes PASS");
}).catch(error=>{ console.error(error); process.exit(1); });`
				output := dataGlobalAlignmentCommand(t, directory, nil, wasmPackedNode(t), "-e", script, linked)
				if strings.TrimSpace(string(output)) != "LLVM linked wasm alignment and bytes PASS" {
					t.Fatal("linked wasm checks did not execute", string(output))
				}
			})
		}
	}
}
