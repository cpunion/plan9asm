package plan9asm

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// The source-role check uses actual SDK selection, not a runtime Skip or a
// filename convention inferred by this test. Native execution remains required
// on Linux; portable object/transport observations receive no native credit.
func TestX86StackFlagsIDOriginalGoBuildRoles(t *testing.T) {
	for _, target := range []struct {
		goos, arch string
		native     bool
		cross      bool
	}{
		{"darwin", "arm64", false, false},
		{"windows", "amd64", false, false},
		{"linux", "amd64", true, true},
		{"linux", "386", true, false},
	} {
		t.Run(target.goos+"/"+target.arch, func(t *testing.T) {
			command := exec.Command(filepath.Join(runtime.GOROOT(), "bin/go"), "list", "-json", ".")
			command.Env = append(os.Environ(), "GOOS="+target.goos, "GOARCH="+target.arch,
				"CGO_ENABLED=0", "GOTOOLCHAIN=local", "GOENV=off", "GOWORK=off", "GOFLAGS=", "GOEXPERIMENT=")
			output := x86FloatCompareCaptureCommand(t, command)
			var selected struct {
				Dir         string
				TestGoFiles []string
			}
			if err := json.Unmarshal(output, &selected); err != nil {
				t.Fatal(err)
			}
			functions := make(map[string]bool)
			for _, name := range selected.TestGoFiles {
				file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(selected.Dir, name), nil, 0)
				if err != nil {
					t.Fatal(err)
				}
				for _, declaration := range file.Decls {
					if function, ok := declaration.(*ast.FuncDecl); ok {
						functions[function.Name.Name] = true
					}
				}
			}
			for _, root := range []struct {
				name string
				want bool
			}{
				{"TestX86StackFlagsIDNativeGo", target.native},
				{"TestCrossLinuxRuntimeMatrixX86StackFlagsIDNativeGo", target.cross},
				{"TestX86StackFlagsIDOriginalGoPortableTransport", true},
			} {
				if functions[root.name] != root.want {
					t.Errorf("actual SDK source role %s: selected=%v want=%v", root.name, functions[root.name], root.want)
				}
			}
		})
	}
}

func x86StackIDOriginalExecutable(t *testing.T, goos, arch string) string {
	t.Helper()
	source := strings.Replace(x86StackIDSource(arch), "idroundtrip(SB)", "·idroundtrip(SB)", 1)
	if arch == "386" {
		source = strings.Replace(source, "$0-8", "$0-4", 1)
	}
	const declarations = "package main\nfunc idroundtrip(out *[4]uint32)\n"
	const driver = `package main

import "fmt"

func main() {
    var out [4]uint32
    idroundtrip(&out)
    if out != [4]uint32{0x200000, 0x200000, 0x200000, 1} {
        panic(fmt.Sprintf("original Go flags roundtrip: %x", out))
    }
    fmt.Println("original Go stack ID observations 4")
}
`
	return x86OrderingHintBuildOriginal(t, t.TempDir(), goos, arch, declarations, source, driver)
}

func TestX86StackFlagsIDOriginalGoPortableTransport(t *testing.T) {
	execution := x86FloatCompareExecutionTools(t)
	original := x86StackIDOriginalExecutable(t, runtime.GOOS, "amd64")
	output := execution.run(t, original)
	if string(output) != "original Go stack ID observations 4\n" {
		t.Fatalf("complete original Go ID/status observations: %q", output)
	}
	t.Logf("four original Go observations via %q; transport observation, not native Linux evidence", execution.prefix)
}

func TestX86StackFlagsIDOriginal386StaticELFCompileOnly(t *testing.T) {
	original := x86StackIDOriginalExecutable(t, "linux", "386")
	sealed, err := x87IA32StaticExecutable(original)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("original Go IA32 static ELF bytes=%d SHA256=%x; compile only, required native execution is separate",
		sealed.size, sealed.hash)
}
