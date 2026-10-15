package plan9asm

import (
	"testing"

	"github.com/xgo-dev/llvm"
)

// These independent IR fixtures exercise only footprint observation. They
// cannot produce an input-pointer, source, raw or native-memory permission.
func TestX86X87ENVFootprintReaderHandlesOptimizedPointerValues(t *testing.T) {
	for _, row := range []struct {
		name, body string
		good       bool
	}{
		{"pointer-argument", "entry:\n%v = load volatile i224, ptr %a, align 1\nret void", true},
		{"selected-arguments", "entry:\n%p = select i1 %gate, ptr %a, ptr %b\n%v = load volatile i224, ptr %p, align 1\nret void", true},
		{"joined-arguments", "entry:\nbr i1 %gate, label %left, label %right\nleft:\nbr label %join\nright:\nbr label %join\njoin:\n%p = phi ptr [%a, %left], [%b, %right]\n%v = load volatile i224, ptr %p, align 1\nret void", true},
		{"selected-different-offsets", "entry:\n%offset = getelementptr i8, ptr %b, i64 1\n%p = select i1 %gate, ptr %a, ptr %offset\n%v = load volatile i224, ptr %p, align 1\nret void", false},
		{"selected-null", "entry:\n%p = select i1 %gate, ptr %a, ptr null\n%v = load volatile i224, ptr %p, align 1\nret void", false},
		{"joined-different-offsets", "entry:\n%offset = getelementptr i8, ptr %b, i64 1\nbr i1 %gate, label %left, label %right\nleft:\nbr label %join\nright:\nbr label %join\njoin:\n%p = phi ptr [%a, %left], [%offset, %right]\n%v = load volatile i224, ptr %p, align 1\nret void", false},
	} {
		t.Run(row.name, func(t *testing.T) {
			ir := "define void @EnvRead(ptr %a, ptr %b, i1 %gate) {\n" + row.body + "\n}\n"
			context := llvm.NewContext()
			defer context.Dispose()
			module, err := parseIRModuleInContext(context, ir)
			if !module.IsNil() {
				defer module.Dispose()
			}
			if err != nil {
				t.Fatal("independent reader fixture must parse", err)
			}
			if err := llvm.VerifyModule(module, llvm.ReturnStatusAction); err != nil {
				t.Fatal("independent reader fixture must be valid LLVM IR", err)
			}
			if err := x86X87ENVReadFootprint(ir, 0); (err == nil) != row.good {
				t.Fatal("footprint reader confused pointer normalization with a complete read", err)
			}
		})
	}
}
