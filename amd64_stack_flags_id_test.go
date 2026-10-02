package plan9asm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func x86StackIDSource(goarch string) string {
	// Go 1.27's asm6.go gives PUSHF/POPF six zero-operand W/L/Q forms.
	// Intel SDM Volume 2B's POPF description preserves the upper flags on
	// a 16-bit operand; the CPUID ID bit belongs to EFLAGS/RFLAGS bit 21.
	// https://www.intel.com/content/dam/www/public/us/en/documents/manuals/64-ia-32-architectures-software-developer-vol-2b-manual.pdf
	width, flags := "Q", "Q"
	if goarch == "386" {
		width, flags = "L", "L"
	}
	return fmt.Sprintf(`TEXT idroundtrip(SB),4,$0-8
	MOV%s out+0(FP), DI
	PUSHF%s
	POP%s AX
	MOV%s AX, CX
	XOR%s $0x200000, AX
	PUSH%s AX
	POPF%s
	PUSHF%s
	POP%s AX
	PUSH%s CX
	POPF%s
	XOR%s CX, AX
	AND%s $0x200000, AX
	MOVL AX, 0(DI)
	PUSHF%s
	POP%s AX
	OR%s $0x200000, AX
	PUSH%s AX
	POPF%s
	PUSHW $2
	POPFW
	PUSHF%s
	POP%s AX
	AND%s $0x200000, AX
	MOVL AX, 4(DI)
	LAHF
	SAHF
	STC
	PUSHFW
	CLC
	POPFW
	SETCS 12(DI)
	PUSHF%s
	POP%s AX
	AND%s $0x200000, AX
	MOVL AX, 8(DI)
	PUSH%s CX
	POPF%s
	RET
`, width, flags, width, width, width, width, flags, flags, width,
		width, flags, width, width, flags, width, width, width, flags,
		flags, width, width, flags, width, width, width, flags)
}

func TestX86StackFlagsIDObjects(t *testing.T) {
	llc, _, ok := findLlcAndClang(t)
	if !ok {
		t.Fatal("LLVM 22 llc/clang not found")
	}
	for _, target := range []struct{ arch, triple string }{
		{"amd64", "x86_64-apple-darwin"},
		{"amd64", "x86_64-unknown-linux-gnu"},
		{"amd64", "x86_64-pc-windows-msvc"},
		{"386", "i386-unknown-linux-gnu"},
		{"386", "i686-pc-windows-msvc"},
	} {
		t.Run(target.triple, func(t *testing.T) {
			source := x86StackIDSource(target.arch)
			if target.arch == "386" {
				source = strings.Replace(source, "$0-8", "$0-4", 1)
			}
			requireX86GoAssemblerResult(t, target.arch, source, true)
			file, err := Parse(ArchAMD64, source)
			if err != nil {
				t.Fatal(err)
			}
			ir, err := Translate(file, x86StackIDOptions(target.arch, target.triple))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(ir, "%flags_id = alloca i1") {
				t.Fatal("native stack flags cannot silently omit the CPUID ID bit")
			}
			dir := t.TempDir()
			path := filepath.Join(dir, "flags.ll")
			if err := os.WriteFile(path, []byte(ir), 0600); err != nil {
				t.Fatal(err)
			}
			if output, err := exec.Command(llc, "-filetype=obj", "-mtriple="+target.triple, path, "-o", filepath.Join(dir, "flags.o")).CombinedOutput(); err != nil {
				t.Fatalf("LLVM 22 flags object: %v\n%s", err, output)
			}
		})
	}
}

func x86StackIDOptions(arch, triple string) Options {
	return Options{Goarch: arch, TargetTriple: triple,
		Sigs: map[string]FuncSig{"idroundtrip": {
			Name: "idroundtrip", Args: []LLVMType{Ptr}, Ret: Void,
			Frame: FrameLayout{Params: []FrameSlot{{Offset: 0, Type: Ptr, Index: 0, Field: -1}}},
		}},
	}
}

func TestX86StackFlagsIDNativeGo(t *testing.T) {
	if runtime.GOARCH != "amd64" && runtime.GOARCH != "386" {
		t.Skip("native x86 Go execution belongs to the required x86 runtime job")
	}
	source := strings.Replace(x86StackIDSource(runtime.GOARCH), "idroundtrip(SB)", "·idroundtrip(SB)", 1)
	if runtime.GOARCH == "386" {
		source = strings.Replace(source, "$0-8", "$0-4", 1)
	}
	dir := t.TempDir()
	files := map[string]string{
		"go.mod":                         "module flagsidnative\n\ngo 1.20\n",
		"flags_" + runtime.GOARCH + ".s": source,
		"flags_test.go": `package flagsidnative
import "testing"
func idroundtrip(out *[4]uint32)
func TestOriginalGoAssembly(t *testing.T) {
  var out [4]uint32
  idroundtrip(&out)
  if out != [4]uint32{0x200000, 0x200000, 0x200000, 1} {
    t.Fatalf("original Go flags roundtrip: %x", out)
  }
}
`,
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	command := exec.Command("go", "test", "-count=1", "-v")
	command.Dir = dir
	command.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("original Go flags runtime: %v\n%s", err, output)
	}
}

func TestAMD64StackFlagsIDRuntime(t *testing.T) {
	crossRosetta := runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" && rosettaAvailable()
	if runtime.GOARCH != "amd64" && !crossRosetta {
		t.Skip("requires native amd64 or the required cross-runtime counterpart")
	}
	llc, clang, ok := findLlcAndClang(t)
	if !ok {
		t.Fatal("LLVM 22 llc/clang not found")
	}
	triple := testTargetTriple(runtime.GOOS, runtime.GOARCH)
	var prefix []string
	if crossRosetta {
		triple = "x86_64-apple-macosx"
		prefix = []string{"/usr/bin/arch", "-x86_64"}
	}
	file, err := Parse(ArchAMD64, x86StackIDSource("amd64"))
	if err != nil {
		t.Fatal(err)
	}
	ir, err := Translate(file, x86StackIDOptions("amd64", triple))
	if err != nil {
		t.Fatal(err)
	}
	const oracle = `
#include <stdint.h>
extern void idroundtrip(uint32_t *);
int main(void) {
  uint32_t out[4] = {0};
  idroundtrip(out);
  if (out[0] != 0x200000) return 10;
  if (out[1] != 0x200000) return 11;
  if (out[2] != 0x200000) return 12;
  if (out[3] != 1) return 13;
  return 0;
}
`
	compileAndRunRuntimeTestForTarget(t, llc, clang, "stack_flags_id", triple, ir, oracle, prefix)
}
