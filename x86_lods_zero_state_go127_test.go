//go:build go1.27

package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

// This test-only LLVM caller observes the actual translated void function's
// four output cells. Its fresh alloca is not an ASM frame/backing certificate.
// It compares the opaque source pointer by identity, never as numerical data.
func x86LODSZeroStateObserver(row x86LODSStateCase) string {
	typ := "i64"
	if row.arch == "386" {
		typ = "i32"
	}
	result := fmt.Sprintf("{ %s, i1, %s, %s }", typ, typ, typ)
	return fmt.Sprintf(`
define %s @observe_%s(ptr %%source) {
entry:
  %%out = alloca [4 x %s]
  call void @%s(ptr %%out, ptr %%source)
  %%si_cell = getelementptr [4 x %s], ptr %%out, i32 0, i32 1
  %%cx_cell = getelementptr [4 x %s], ptr %%out, i32 0, i32 2
  %%flags_cell = getelementptr [4 x %s], ptr %%out, i32 0, i32 3
  %%ax = load %s, ptr %%out
  %%si = load ptr, ptr %%si_cell
  %%same_si = icmp eq ptr %%si, %%source
  %%cx = load %s, ptr %%cx_cell
  %%flags = load %s, ptr %%flags_cell
  %%r0 = insertvalue %s undef, %s %%ax, 0
  %%r1 = insertvalue %s %%r0, i1 %%same_si, 1
  %%r2 = insertvalue %s %%r1, %s %%cx, 2
  %%r3 = insertvalue %s %%r2, %s %%flags, 3
  ret %s %%r3
}
`, result, row.name, typ, row.name, typ, typ, typ, typ, typ, typ,
		result, typ, result, result, typ, result, typ, result)
}

func TestX86LODSZeroCountOriginalGoAndLLVMWholeState(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		source, declarations := x86LODSStateFixture(arch)
		goAssemblerOracleOutput(t, arch, source, "", true)
		triple, typ := "x86_64-unknown-linux-gnu", "i64"
		if arch == "386" {
			triple, typ = "i386-unknown-linux-gnu", "i32"
		}
		for _, api := range []string{"text", "module", "Go"} {
			t.Run(arch+"/"+api, func(t *testing.T) {
				ir, err := x86ScalarFPBitViewAPI(t, source, declarations, arch, triple, api)
				if err != nil {
					t.Fatal("complete original fixture did not reach its state observer", err)
				}
				for _, row := range x86LODSStateCases(arch) {
					if row.prefix != "" && row.count == 0 {
						ir += x86LODSZeroStateObserver(row)
					}
				}
				optimized := x86UndefinedTrapOptimizedIR(t, ir)
				for _, row := range x86LODSStateCases(arch) {
					// Keep the complete original matrix in this same module. The
					// nonzero/bare cases are not silently replaced by zero cases.
					x86StringMemoryFunction(t, optimized, row.name)
					if row.prefix == "" || row.count != 0 {
						continue
					}
					body := x86StringMemoryFunction(t, optimized, "observe_"+row.name)
					expected := x86LODSStateExpected(row)
					want := fmt.Sprintf("ret { %s, i1, %s, %s } { %s %d, i1 true, %s 0, %s %d }",
						typ, typ, typ, typ, expected.ax, typ, typ, expected.flags)
					if !strings.Contains(body, want) {
						t.Errorf("zero count changed complete AX, source SI, CX or initialized FLAGS: %s/%s/count0/backward=%t\nwant %s\n%s",
							arch, row.op, row.backward, want, body)
					}
				}
			})
		}
	}
}
