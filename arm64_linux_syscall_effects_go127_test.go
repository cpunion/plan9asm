//go:build go1.27

package plan9asm

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
	"github.com/xgo-dev/plan9asm/internal/gostaticstorage"
)

// These are independent original-kernel oracles, not a source capability or a
// sleep-borrow certificate. Pointer-free syscalls isolate GP/NZCV from memory.
type arm64LinuxSyscallEffectCase struct {
	name        string
	number      int
	registers   []int
	left, right uint64
	flags       bool
}

func arm64LinuxSyscallEffectCases() []arm64LinuxSyscallEffectCase {
	var registers []int
	for register := 1; register <= 27; register++ {
		if register != 18 { // Never replace the platform register or Go g/FP/LR.
			registers = append(registers, register)
		}
	}
	var cases []arm64LinuxSyscallEffectCase
	for _, syscall := range []struct {
		name   string
		number int
	}{{"getpid", 172}, {"close_invalid", 57}} {
		cases = append(cases, arm64LinuxSyscallEffectCase{
			name: "gp_" + syscall.name, number: syscall.number, registers: registers,
		})
		for _, comparison := range []struct {
			name        string
			left, right uint64
		}{
			{"equal", 7, 7},
			{"negative", 0, 1},
			{"positive_overflow", 0x8000000000000000, 1},
			{"negative_overflow", 0x7fffffffffffffff, ^uint64(0)},
		} {
			cases = append(cases, arm64LinuxSyscallEffectCase{
				name: "nzcv_" + syscall.name + "_" + comparison.name, number: syscall.number,
				flags: true, left: comparison.left, right: comparison.right,
			})
		}
	}
	return cases
}

func (test arm64LinuxSyscallEffectCase) resultCount() int {
	if test.flags {
		return 5 // Raw R0 and independent N/Z/C/V observations.
	}
	return 1 + len(test.registers)
}

func (test arm64LinuxSyscallEffectCase) fields() []string {
	fields := make([]string, test.resultCount())
	for index := range fields {
		fields[index] = fmt.Sprintf("r%d", index)
	}
	return fields
}

func (test arm64LinuxSyscallEffectCase) assembly() string {
	var source strings.Builder
	// Numeric NOSPLIT|NOFRAME: no automatic Go saved-frame cells to borrow.
	fmt.Fprintf(&source, "TEXT ·%s(SB),516,$0-%d\n", test.name, 8+8*test.resultCount())
	if test.flags {
		for register := 1; register <= 5; register++ {
			fmt.Fprintf(&source, "\tMOVD $0, R%d\n", register)
		}
		fmt.Fprintf(&source, "\tMOVD $0x%x, R6\n\tMOVD $0x%x, R7\n", test.left, test.right)
	} else {
		// Single-16-bit literals and small scaled stores do not need Go's R27
		// scratch expansion; the original complete object/runtime still checks it.
		for _, register := range test.registers {
			if register != 8 {
				fmt.Fprintf(&source, "\tMOVD $0x%x, R%d\n", 0x5a00+register, register)
			}
		}
	}
	if test.number == 57 {
		source.WriteString("\tMOVD $-1, R0\n")
	} else {
		source.WriteString("\tMOVD input+0(FP), R0\n")
	}
	fmt.Fprintf(&source, "\tMOVD $%d, R8\n", test.number)
	if test.flags {
		source.WriteString("\tCMP R7, R6\n")
	}
	source.WriteString("\tSVC $0\n")
	if test.flags {
		// No intervening flag-changing arithmetic can hide the kernel state.
		source.WriteString("\tCSET MI, R2\n\tCSET EQ, R3\n\tCSET CS, R4\n\tCSET VS, R5\n")
	}
	source.WriteString("\tMOVD R0, r0+8(FP)\n")
	registers := test.registers
	if test.flags {
		registers = []int{2, 3, 4, 5}
	}
	for index, register := range registers {
		fmt.Fprintf(&source, "\tMOVD R%d, r%d+%d(FP)\n", register, index+1, 16+8*index)
	}
	source.WriteString("\tRET\n\n")
	return source.String()
}

func (test arm64LinuxSyscallEffectCase) numericalExpected() []uint64 {
	if test.flags {
		// Independent numerical SUB equations, not translator flag helpers.
		difference := test.left - test.right
		values := []bool{difference>>63 != 0, difference == 0, test.left >= test.right,
			(test.left^test.right)&(test.left^difference)&(uint64(1)<<63) != 0}
		result := make([]uint64, len(values))
		for index, value := range values {
			if value {
				result[index] = 1
			}
		}
		return result
	}
	result := make([]uint64, len(test.registers))
	for index, register := range test.registers {
		result[index] = uint64(0x5a00 + register)
		if register == 8 {
			result[index] = uint64(test.number)
		}
	}
	return result
}

func arm64LinuxSyscallEffectFixture(t *testing.T) (directory, packagePath, declarations, assembly string) {
	t.Helper()
	directory, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	packagePath = "example.invalid/kerneloracle"
	var goSource, asmSource, original strings.Builder
	goSource.WriteString("package kerneloracle\n\n")
	original.WriteString("package kerneloracle\nimport (\"os\"; \"syscall\"; \"testing\")\nfunc TestOriginalLinuxSyscallEffects(t *testing.T) {\n")
	for _, test := range arm64LinuxSyscallEffectCases() {
		fields := strings.Join(test.fields(), ",")
		fmt.Fprintf(&goSource, "func %s(input uint64) (%s uint64)\n", test.name, fields)
		fmt.Fprintf(&goSource, "func invoke_%s(input uint64) (%s uint64) { return %s(input) }\n", test.name, fields, test.name)
		asmSource.WriteString(test.assembly())
		fmt.Fprintf(&original, "t.Run(%q,func(t *testing.T) { %s := %s(0x45); got := []uint64{%s}\n", test.name, fields, test.name, fields)
		if test.number == 172 {
			original.WriteString("if got[0] == 0 || got[0] != uint64(os.Getpid()) { t.Fatal(\"original getpid is not the actual process PID\", got[0],os.Getpid()) }\n")
		} else {
			original.WriteString("if got[0] != ^(uint64(syscall.EBADF)-1) { t.Fatal(\"original close(-1) must return raw -EBADF\",got[0]) }\n")
		}
		for index, value := range test.numericalExpected() {
			fmt.Fprintf(&original, "if got[%d] != 0x%x { t.Fatal(\"original GP/NZCV axis %d\",got[%d],uint64(0x%x)) }\n", index+1, value, index+1, index+1, value)
		}
		fmt.Fprintf(&original, "t.Log(%q)\n})\n", "ORIGINAL_KERNEL_AXIS_PASS "+test.name)
	}
	original.WriteString("}\n")
	for name, data := range map[string]string{
		"go.mod": "module " + packagePath + "\ngo 1.20\n", "decl.go": goSource.String(),
		"original_arm64.s": asmSource.String(), "original_test.go": original.String(),
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return directory, packagePath, goSource.String(), asmSource.String()
}

func arm64LinuxSyscallEffectOriginalObjects(t *testing.T, directory, packagePath string) {
	t.Helper()
	archive, object, builtins := arm64ByteCarrierOriginal(t, directory, packagePath)
	members, err := gostaticstorage.ObserveARM64AssemblyArchive(archive, builtins, "linux/arm64", []string{"original_arm64.s"})
	if err != nil || len(members) != 3 {
		t.Fatalf("ORIGINAL_OBJECT_PRECONDITION: complete archive: %v", err)
	}
	texts, err := gostaticstorage.ObserveARM64AssemblyText(object, builtins, "linux/arm64")
	cases := arm64LinuxSyscallEffectCases()
	if err != nil || len(texts) != len(cases) || len(members[2].Texts) != len(cases) {
		t.Fatalf("ORIGINAL_OBJECT_PRECONDITION: complete TEXT inventory: %v", err)
	}
	for index, test := range cases {
		text, prior := texts[index], members[2].Texts[index]
		if text.Name != packagePath+"."+test.name || text.ABI != 0 || text.Name != prior.Name ||
			!bytes.Equal(text.Code, prior.Code) || len(text.Relocations) != 0 || len(prior.Relocations) != 0 ||
			len(text.Positions) == 0 || len(prior.Positions) == 0 || len(text.Code)%4 != 0 {
			t.Fatalf("ORIGINAL_OBJECT_PRECONDITION: complete standalone/archive join differs: %s", test.name)
		}
		count := 0
		for pc := 0; pc < len(text.Code); pc += 4 {
			if binary.LittleEndian.Uint32(text.Code[pc:]) == 0xd4000001 {
				count++
			}
		}
		if count != 1 {
			t.Fatalf("ORIGINAL_OBJECT_PRECONDITION: %s has %d actual SVC0 words", test.name, count)
		}
		t.Logf("ORIGINAL_KERNEL_OBJECT name=%s bytes=%d codeSHA256=%x SVC0=%d", test.name, len(text.Code), sha256.Sum256(text.Code), count)
	}
}

func TestARM64LinuxSyscallEffectsOriginalObjects(t *testing.T) {
	directory, packagePath, _, assembly := arm64LinuxSyscallEffectFixture(t)
	file, err := Parse(ArchARM64, assembly)
	if err != nil {
		t.Fatal(err)
	}
	for _, function := range file.Funcs {
		if function.FrameSize != 0 || arm64SourceGoFrame(function).present {
			t.Fatal("kernel oracle is not a zero-frame independent control")
		}
	}
	arm64LinuxSyscallEffectOriginalObjects(t, directory, packagePath)
}

func arm64LinuxSyscallEffectTranslate(t *testing.T, api string, pkg GoPackage, assembly string) (string, error) {
	t.Helper()
	context := llvm.NewContext()
	defer context.Dispose()
	file, err := Parse(ArchARM64, assembly)
	if err != nil {
		t.Fatal(err)
	}
	resolve := testResolveSym(pkg.Path)
	signatures, err := goSigsForAsmFile(pkg, file, resolve, "arm64", nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	options := Options{Goarch: "arm64", TargetTriple: "aarch64-unknown-linux-gnu", ResolveSym: resolve, Sigs: signatures}
	switch api {
	case "text":
		return Translate(file, options)
	case "module":
		module, err := TranslateModuleInContext(context, file, options)
		if !module.IsNil() {
			defer module.Dispose()
			return module.String(), err
		}
		return "", err
	case "Go":
		translated, err := translateGoModuleInContext(context, pkg, []byte(assembly), GoModuleOptions{
			GOOS: "linux", GOARCH: "arm64", TargetTriple: options.TargetTriple, ResolveSym: resolve,
		})
		if translated != nil && !translated.Module.IsNil() {
			defer translated.Module.Dispose()
			return translated.Module.String(), err
		}
		return "", err
	default:
		t.Fatalf("unknown syscall oracle API %q", api)
		return "", nil
	}
}

func arm64LinuxSyscallEffectRuntimeChecks(packagePath string, test arm64LinuxSyscallEffectCase) (string, string) {
	var shim, driver strings.Builder
	count := test.resultCount()
	types := make([]string, count)
	for index := range types {
		types[index] = "i64"
	}
	aggregate := "{ " + strings.Join(types, ", ") + " }"
	// LLVM aggregate calls stay internal. C sees only an explicit output pointer.
	fmt.Fprintf(&shim, "\ndefine void @kernel_check(ptr %%out, i64 %%input) {\n  %%result = call %s %s(i64 %%input)\n", aggregate, llvmGlobal(packagePath+"."+test.name))
	for index := 0; index < count; index++ {
		fmt.Fprintf(&shim, "  %%v%d = extractvalue %s %%result, %d\n  %%p%d = getelementptr i64, ptr %%out, i64 %d\n  store i64 %%v%d, ptr %%p%d, align 8\n", index, aggregate, index, index, index, index, index)
	}
	shim.WriteString("  ret void\n}\n")
	driver.WriteString("#include <stdint.h>\n#include <stdio.h>\n#include <unistd.h>\n#include <errno.h>\nint32_t cliteErrno(void) { return errno; }\nextern void kernel_check(uint64_t *, uint64_t);\nint main(void) {\n")
	fmt.Fprintf(&driver, "uint64_t out[%d] = {0}; kernel_check(out, UINT64_C(0x45));\n", count)
	if test.number == 172 {
		driver.WriteString("if (!out[0] || out[0] != (uint64_t)getpid()) { fprintf(stderr,\"actual process PID mismatch\\n\"); return 1; }\n")
	} else {
		driver.WriteString("if (out[0] != (uint64_t)-(int64_t)EBADF) { fprintf(stderr,\"raw -EBADF mismatch\\n\"); return 1; }\n")
	}
	for index, value := range test.numericalExpected() {
		fmt.Fprintf(&driver, "if (out[%d] != UINT64_C(0x%x)) { fprintf(stderr,\"GP/NZCV axis %d got=%%llx expected=%x\\n\",(unsigned long long)out[%d]); return 1; }\n", index+1, value, index+1, value, index+1)
	}
	fmt.Fprintf(&driver, "puts(\"LLVM_KERNEL_AXIS_PASS %s\"); return 0;\n}\n", test.name)
	return shim.String(), driver.String()
}
