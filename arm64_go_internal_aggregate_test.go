package plan9asm

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// Go permits explicit ABI selectors only to its trusted runtime assembly.
// Use a private package and fully qualified main symbols with the assembler's
// runtime identity. The compiler still receives the real gensymabis output
// and emits its ordinary Go ABIInternal calls/wrappers from the declaration.
func runARM64GoInternalOracle(t *testing.T, source, main string, cross bool) {
	t.Helper()
	dir := t.TempDir()
	for name, data := range map[string]string{
		"go.mod":  "module internalregisteroracle\n\ngo 1.27\n",
		"main.go": main, "oracle_arm64.s": strings.ReplaceAll(source, "·", "main·"),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	args := []string{"run", "-p=2", "-asmflags=internalregisteroracle=-p=runtime"}
	env := append(os.Environ(), "GOARCH=arm64", "CGO_ENABLED=0")
	if cross {
		args = append(args, "-exec=qemu-aarch64")
		env = append(env, "GOOS=linux")
	}
	args = append(args, ".")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir, cmd.Env = dir, env
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("actual Go ABIInternal oracle: %v\n%s", err, out)
	}
	t.Log("actual native Go ABIInternal entry/call oracle passed")
}

func TestCrossLinuxRuntimeMatrixARM64GoInternalAggregates(t *testing.T) {
	llc, triple, compiler, runner := arm64FPPairRuntimeTools(t)
	for _, tc := range []struct {
		name, decl, source, goCallee, goCheck, llvmShim, cCheck string
	}{
		{
			name:     "caller_save_redefined",
			decl:     "func Pass(uint64) uint64\nfunc Compute(uint64) uint64",
			source:   "TEXT ·Pass<ABIInternal>(SB),4,$0-16\nCALL ·Compute<ABIInternal>(SB)\nMOVD $99,R9\nEOR R9,R0,R0\nRET\n",
			goCallee: "func Compute(a uint64) uint64 { return a*7+9 }",
			goCheck:  "if Pass(a)!=99^(a*7+9) { panic(\"redefined caller-save mismatch\") }",
			llvmShim: `define i64 @Compute(i64 %a) {
  %x = mul i64 %a, 7
  %r = add i64 %x, 9
  ret i64 %r
}
define i64 @invoke(i64 %a) {
  %r = call i64 @Pass(i64 %a)
  ret i64 %r
}
`,
			cCheck: "99^(inputs[i]*7+9)",
		},
		{
			name:     "caller_save_defined_on_both_paths",
			decl:     "func Pass(uint64) uint64\nfunc Compute(uint64) uint64",
			source:   "TEXT ·Pass<ABIInternal>(SB),4,$0-16\nCALL ·Compute<ABIInternal>(SB)\nCBZ R0,left\nMOVD $99,R9\nB done\nleft:\nMOVD $3,R9\ndone:\nEOR R9,R0,R0\nRET\n",
			goCallee: "func Compute(a uint64) uint64 { return a }",
			goCheck:  "want:=a^99; if a==0 { want=3 }; if Pass(a)!=want { panic(\"caller-save branch join mismatch\") }",
			llvmShim: `define i64 @Compute(i64 %a) {
  ret i64 %a
}
define i64 @invoke(i64 %a) {
  %r = call i64 @Pass(i64 %a)
  ret i64 %r
}
`,
			cCheck: "inputs[i]==0?3:inputs[i]^99",
		},
		{
			name:     "flags_redefined_after_call",
			decl:     "func Pass(uint64) uint64\nfunc Compute(uint64) uint64",
			source:   "TEXT ·Pass<ABIInternal>(SB),4,$0-16\nCALL ·Compute<ABIInternal>(SB)\nCMP $0,R0\nCINC NE,R0,R0\nRET\n",
			goCallee: "func Compute(a uint64) uint64 { return a }",
			goCheck:  "want:=a; if a!=0 { want++ }; if Pass(a)!=want { panic(\"redefined flags mismatch\") }",
			llvmShim: `define i64 @Compute(i64 %a) {
  ret i64 %a
}
define i64 @invoke(i64 %a) {
  %r = call i64 @Pass(i64 %a)
  ret i64 %r
}
`,
			cCheck: "inputs[i]+(inputs[i]!=0)",
		},
		{
			name:     "float32_result_width",
			decl:     "func Pass(uint64) uint64\nfunc Compute(uint64) float32",
			source:   "TEXT ·Pass<ABIInternal>(SB),4,$0-16\nCALL ·Compute<ABIInternal>(SB)\nFMOVS F0,R0\nRET\n",
			goCallee: "func Compute(a uint64) float32 { return 91.5 }",
			goCheck:  "if Pass(a)!=0x42b70000 { panic(\"float32 register result width mismatch\") }",
			llvmShim: `define float @Compute(i64 %a) {
  ret float 9.150000e+01
}
define i64 @invoke(i64 %a) {
  %r = call i64 @Pass(i64 %a)
  ret i64 %r
}
`,
			cCheck: "UINT64_C(0x42b70000)",
		},
		{
			name:     "scalar_call",
			decl:     "func Pass(uint64) uint64\nfunc Compute(uint64) uint64",
			source:   "TEXT ·Pass<ABIInternal>(SB),4,$0-16\nCALL ·Compute<ABIInternal>(SB)\nRET\n",
			goCallee: "func Compute(a uint64) uint64 { return a*7+9 }",
			goCheck:  "if Pass(a)!=a*7+9 { panic(\"scalar ABIInternal call mismatch\") }",
			llvmShim: `define i64 @Compute(i64 %a) {
  %x = mul i64 %a, 7
  %r = add i64 %x, 9
  ret i64 %r
}
define i64 @invoke(i64 %a) {
  %r = call i64 @Pass(i64 %a)
  ret i64 %r
}
`,
			cCheck: "inputs[i]*7+9",
		},
		{
			name:    "nested_integer_entry",
			decl:    "type Pair struct { Lo,Hi uint64 }; type Nested struct { A uint64; B Pair }; func Pass(Nested) uint64",
			source:  "TEXT ·Pass<ABIInternal>(SB),4,$0-32\nEOR R1,R0,R0\nEOR R2,R0,R0\nRET\n",
			goCheck: "if Pass(Nested{a,Pair{a*7+9,a*11+3}})!=a^(a*7+9)^(a*11+3) { panic(\"nested ABIInternal entry mismatch\") }",
			llvmShim: `define i64 @invoke(i64 %a) {
  %x0 = mul i64 %a, 7
  %x = add i64 %x0, 9
  %y0 = mul i64 %a, 11
  %y = add i64 %y0, 3
  %v0 = insertvalue { i64, { i64, i64 } } undef, i64 %a, 0
  %v1 = insertvalue { i64, { i64, i64 } } %v0, i64 %x, 1, 0
  %v = insertvalue { i64, { i64, i64 } } %v1, i64 %y, 1, 1
  %r = call i64 @Pass({ i64, { i64, i64 } } %v)
  ret i64 %r
}
`,
			cCheck: "inputs[i]^(inputs[i]*7+9)^(inputs[i]*11+3)",
		},
		{
			name:    "mixed_banks_entry_result",
			decl:    "type Mixed struct { S float32; D float64; I uint64 }; type Result struct { D float64; I uint64 }; func Pass(Mixed) Result",
			source:  "TEXT ·Pass<ABIInternal>(SB),4,$0-40\nFMOVD F1,F0\nADD $9,R0\nRET\n",
			goCheck: "r:=Pass(Mixed{-13.25,91.5,a}); if r.D!=91.5 || r.I!=a+9 { panic(\"mixed ABIInternal banks mismatch\") }",
			llvmShim: `define i64 @invoke(i64 %a) {
  %v0 = insertvalue { float, double, i64 } undef, float -1.325000e+01, 0
  %v1 = insertvalue { float, double, i64 } %v0, double 9.150000e+01, 1
  %v = insertvalue { float, double, i64 } %v1, i64 %a, 2
  %r = call { double, i64 } @Pass({ float, double, i64 } %v)
  %d = extractvalue { double, i64 } %r, 0
  %i = extractvalue { double, i64 } %r, 1
  %ok = fcmp oeq double %d, 9.150000e+01
  %got = select i1 %ok, i64 %i, i64 0
  ret i64 %got
}
`,
			cCheck: "inputs[i]+9",
		},
		{
			name:    "one_element_array",
			decl:    "func Pass([1]uint64) [1]uint64",
			source:  "TEXT ·Pass<ABIInternal>(SB),4,$0-16\nADD $9,R0\nRET\n",
			goCheck: "if Pass([1]uint64{a})[0]!=a+9 { panic(\"array ABIInternal mismatch\") }",
			llvmShim: `define i64 @invoke(i64 %a) {
  %v = insertvalue [1 x i64] undef, i64 %a, 0
  %r = call i64 @Pass([1 x i64] %v)
  ret i64 %r
}
`,
			cCheck: "inputs[i]+9",
		},
		{
			name:     "nested_call_mixed_results",
			decl:     "type Pair struct { Lo,Hi uint64 }; type Nested struct { A uint64; B Pair }; func Pass(Nested) uint64\nfunc Compute(Nested) (uint64,float64,uint64)",
			source:   "TEXT ·Pass<ABIInternal>(SB),4,$0-32\nCALL ·Compute<ABIInternal>(SB)\nFMOVD F0,R9\nEOR R1,R0,R0\nEOR R9,R0,R0\nRET\n",
			goCallee: "func Compute(a Nested) (uint64,float64,uint64) { return a.A*7+9,91.5,a.B.Hi*11+3 }",
			goCheck:  "if Pass(Nested{a,Pair{a*7+9,a*11+3}})!=(a*7+9)^0x4056e00000000000^((a*11+3)*11+3) { panic(\"nested ABIInternal result banks mismatch\") }",
			llvmShim: `define { i64, double, i64 } @Compute({ i64, { i64, i64 } } %v) {
  %a = extractvalue { i64, { i64, i64 } } %v, 0
  %hi = extractvalue { i64, { i64, i64 } } %v, 1, 1
  %x0 = mul i64 %a, 7
  %x = add i64 %x0, 9
  %y0 = mul i64 %hi, 11
  %y = add i64 %y0, 3
  %r0 = insertvalue { i64, double, i64 } undef, i64 %x, 0
  %r1 = insertvalue { i64, double, i64 } %r0, double 9.150000e+01, 1
  %r = insertvalue { i64, double, i64 } %r1, i64 %y, 2
  ret { i64, double, i64 } %r
}
define i64 @invoke(i64 %a) {
  %x0 = mul i64 %a, 7
  %x = add i64 %x0, 9
  %y0 = mul i64 %a, 11
  %y = add i64 %y0, 3
  %v0 = insertvalue { i64, { i64, i64 } } undef, i64 %a, 0
  %v1 = insertvalue { i64, { i64, i64 } } %v0, i64 %x, 1, 0
  %v = insertvalue { i64, { i64, i64 } } %v1, i64 %y, 1, 1
  %r = call i64 @Pass({ i64, { i64, i64 } } %v)
  ret i64 %r
}
`,
			cCheck: "(inputs[i]*7+9)^UINT64_C(0x4056e00000000000)^((inputs[i]*11+3)*11+3)",
		},
		{
			name:     "zero_argument_call",
			decl:     "func Pass() uint64\nfunc Compute() uint64",
			source:   "TEXT ·Pass<ABIInternal>(SB),4,$0-8\nCALL ·Compute<ABIInternal>(SB)\nRET\n",
			goCallee: "func Compute() uint64 { return 0x123456789abcdef0 }",
			goCheck:  "if Pass()!=0x123456789abcdef0 { panic(\"zero-argument ABIInternal mismatch\") }",
			llvmShim: `define i64 @Compute() {
  ret i64 1311768467463790320
}
define i64 @invoke(i64 %a) {
  %r = call i64 @Pass()
  ret i64 %r
}
`,
			cCheck: "UINT64_C(0x123456789abcdef0)",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			goDecl := tc.decl
			if tc.goCallee != "" {
				goDecl = strings.Split(goDecl, "\nfunc Compute")[0] + "\n" + tc.goCallee
			}
			runARM64GoInternalOracle(t, tc.source, "package main\n"+goDecl+"\nfunc main() { for _,a:=range []uint64{0,1,123,0x8000000000000000,^uint64(0)} { _=a; "+tc.goCheck+" } }\n", len(runner) != 0)
			pkg := mustGoPackage(t, "test/internalaggregate", "package internalaggregate\n"+tc.decl+"\n")
			translate := func(target string) string {
				tr, err := TranslateGoModule(pkg, []byte(tc.source), GoModuleOptions{
					GOARCH: "arm64", TargetTriple: target,
					ResolveSym: func(sym string) string { return strings.TrimPrefix(goStripABISuffix(sym), "·") },
				})
				if err != nil {
					t.Fatal(err)
				}
				defer tr.Module.Dispose()
				ir := tr.Module.String()
				// The independent LLVM callee definition replaces only its exact
				// external declaration; it does not rewrite translated instructions.
				ir = strings.ReplaceAll(ir, "declare i64 @Compute(i64)\n", "")
				ir = strings.ReplaceAll(ir, "declare i64 @Compute()\n", "")
				ir = strings.ReplaceAll(ir, "declare float @Compute(i64)\n", "")
				ir = strings.ReplaceAll(ir, "declare { i64, double, i64 } @Compute({ i64, { i64, i64 } })\n", "")
				return ir + "\n" + tc.llvmShim
			}
			for _, target := range []string{
				"aarch64-unknown-linux-gnu", "aarch64-unknown-linux-musl",
				"aarch64-apple-darwin", "aarch64-pc-windows-msvc",
			} {
				compileLLVMToObject(t, llc, target, "go-internal.ll", "go-internal.o", translate(target))
			}
			main := fmt.Sprintf(`#include <stdint.h>
extern uint64_t invoke(uint64_t);
int main(void) {
  uint64_t inputs[]={0,1,123,UINT64_C(0x8000000000000000),UINT64_MAX};
  for (unsigned i=0;i<5;i++) {
    uint64_t want=%s;
    if (invoke(inputs[i])!=want) return 1;
  }
  return 0;
}
`, tc.cCheck)
			compileAndRunRuntimeTestWithCompiler(t, llc, compiler, "go_internal_aggregate", triple, translate(triple), main, runner)
		})
	}
}

func TestCrossLinuxRuntimeMatrixARM64GoInternalBytealgCount(t *testing.T) {
	llc, triple, compiler, runner := arm64FPPairRuntimeTools(t)
	source, err := os.ReadFile(filepath.Join(runtime.GOROOT(), "src", "internal", "bytealg", "count_arm64.s"))
	if err != nil {
		t.Fatal(err)
	}
	const decl = "func Count([]byte,byte) int\nfunc CountString(string,byte) int\n"
	const goMain = `package main
import "bytes"
func Count([]byte,byte) int
func CountString(string,byte) int
func main() {
  var data [641]byte
  for i:=range data { data[i]=byte(i%19) }
  for _,n:=range []int{0,1,15,16,17,31,32,33,255,256,257,640} {
    for _,offset:=range []int{0,1} {
      b:=data[offset:offset+n]
      for _,c:=range []byte{0,7,18,255} {
        want:=bytes.Count(b,[]byte{c})
        if Count(b,c)!=want || CountString(string(b),c)!=want { panic("actual bytealg count register contract mismatch") }
      }
    }
  }
}
`
	runARM64GoInternalOracle(t, string(source), goMain, len(runner) != 0)
	pkg := mustGoPackage(t, "test/bytealgcount", "package bytealgcount\n"+decl)
	const shim = `define i64 @invokeCount(ptr %p, i64 %n, i8 %c) {
  %b0 = insertvalue { ptr, i64, i64 } undef, ptr %p, 0
  %b1 = insertvalue { ptr, i64, i64 } %b0, i64 %n, 1
  %b = insertvalue { ptr, i64, i64 } %b1, i64 %n, 2
  %r = call i64 @Count({ ptr, i64, i64 } %b, i8 %c)
  ret i64 %r
}
define i64 @invokeCountString(ptr %p, i64 %n, i8 %c) {
  %s0 = insertvalue { ptr, i64 } undef, ptr %p, 0
  %s = insertvalue { ptr, i64 } %s0, i64 %n, 1
  %r = call i64 @CountString({ ptr, i64 } %s, i8 %c)
  ret i64 %r
}
`
	translate := func(target string) string {
		tr, err := TranslateGoModule(pkg, source, GoModuleOptions{
			GOARCH: "arm64", TargetTriple: target,
			ResolveSym: func(sym string) string { return strings.TrimPrefix(goStripABISuffix(sym), "·") },
		})
		if err != nil {
			t.Fatal(err)
		}
		defer tr.Module.Dispose()
		return tr.Module.String() + "\n" + shim
	}
	for _, target := range []string{
		"aarch64-unknown-linux-gnu", "aarch64-unknown-linux-musl",
		"aarch64-apple-darwin", "aarch64-pc-windows-msvc",
	} {
		compileLLVMToObject(t, llc, target, "bytealg-count.ll", "bytealg-count.o", translate(target))
	}
	const main = `#include <stdint.h>
extern int64_t invokeCount(uint8_t *,uint64_t,uint8_t);
extern int64_t invokeCountString(uint8_t *,uint64_t,uint8_t);
int main(void) {
  uint8_t data[641], search[]={0,7,18,255};
  unsigned sizes[]={0,1,15,16,17,31,32,33,255,256,257,640};
  for (unsigned i=0;i<641;i++) data[i]=i%19;
  for (unsigned j=0;j<12;j++) for (unsigned off=0;off<2;off++) {
    for (unsigned k=0;k<4;k++) {
      int64_t want=0;
      for(unsigned i=0;i<sizes[j];i++) want+=data[off+i]==search[k];
      if (invokeCount(data+off,sizes[j],search[k])!=want) return 1;
      if (invokeCountString(data+off,sizes[j],search[k])!=want) return 2;
    }
  }
  return 0;
}
`
	compileAndRunRuntimeTestWithCompiler(t, llc, compiler, "go_internal_count", triple, translate(triple), main, runner)
}
