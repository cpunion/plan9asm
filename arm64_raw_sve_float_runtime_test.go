package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

func testARM64RawSVEFloatRuntime(t *testing.T, llc string) {
	var source, declarations, checks strings.Builder
	sigs := make(map[string]FuncSig)
	for _, width := range []struct {
		suffix, scalar, ctype string
		bytes                 int
	}{
		{"h", "h", "_Float16", 2}, {"s", "s", "float", 4}, {"d", "d", "double", 8},
	} {
		for _, op := range []string{"fadd", "fsub", "fmul", "faddv", "fadda"} {
			name := op + "_" + width.suffix
			loadStore := map[int]string{2: "h", 4: "w", 8: "d"}[width.bytes]
			native := []string{
				"ptrue p7.b",
				fmt.Sprintf("ptrue p0.%s, vl3", width.suffix),
				fmt.Sprintf("ld1%s { z0.%s }, p7/z, [x0]", loadStore, width.suffix),
				fmt.Sprintf("ld1%s { z1.%s }, p7/z, [x1]", loadStore, width.suffix),
				fmt.Sprintf("ld1%s { z5.%s }, p7/z, [x0]", loadStore, width.suffix),
			}
			count := "vl"
			if op == "faddv" || op == "fadda" {
				native[2] = fmt.Sprintf("ld1%s { z0.%s }, p7/z, [x1]", loadStore, width.suffix)
				instruction := fmt.Sprintf("faddv %s0, p0, z5.%s", width.scalar, width.suffix)
				if op == "fadda" {
					instruction = fmt.Sprintf("fadda %s0, p0, %s0, z5.%s", width.scalar, width.scalar, width.suffix)
				}
				native = append(native, instruction, "st1b { z0.b }, p7, [x2]")
			} else {
				native = append(native, fmt.Sprintf("%s z0.%s, p0/m, z0.%s, z1.%s", op, width.suffix, width.suffix, width.suffix),
					"st1b { z0.b }, p7, [x2]")
			}
			words := assembleARM64LLVMWords(t, native, "+sve")
			fmt.Fprintf(&source, "TEXT %s(SB),$0-24\nMOVD a+0(FP),R0\nMOVD b+8(FP),R1\nMOVD out+16(FP),R2\n", name)
			for _, word := range words {
				fmt.Fprintf(&source, "WORD $%#08x\n", word)
			}
			source.WriteString("RET\n")
			sigs[name] = FuncSig{Name: name, Args: []LLVMType{Ptr, Ptr, Ptr}, Ret: Void, Frame: FrameLayout{Params: []FrameSlot{
				{Offset: 0, Type: Ptr, Index: 0, Field: -1},
				{Offset: 8, Type: Ptr, Index: 1, Field: -1},
				{Offset: 16, Type: Ptr, Index: 2, Field: -1},
			}}}
			fmt.Fprintf(&declarations, "extern void %s(const void *, const void *, void *);\n", name)
			assembly := strings.NewReplacer("[x0]", "[%[a]]", "[x1]", "[%[b]]", "[x2]", "[%[out]]").Replace(strings.Join(native, "\\n\\t"))
			fmt.Fprintf(&checks, `    {
      %s a[128], b[128];
      const double samples[] = {65504, -65504, 0.125, -0.0, 0.00001, -1.0};
      for (unsigned i = 0; i < vl / %d; i++) {
        a[i] = (%s)samples[i %% 6];
        b[i] = (%s)(i + 2);
      }
      unsigned char got[256] = {0}, want[256] = {0};
      __asm__ volatile("%s" :: [a]"r"(a), [b]"r"(b), [out]"r"(want)
        : "p0", "p7", "z0", "z1", "z5", "memory");
      %s(a, b, got);
      if (memcmp(got, want, %s) != 0) return %d;
    }
`, width.ctype, width.bytes, width.ctype, width.ctype, assembly, name, count, len(sigs))
		}
	}
	main := "#include <stdint.h>\n#include <string.h>\n#include <sys/prctl.h>\n" + declarations.String() + `
int main(void) {
  const unsigned lengths[] = {16, 32, 48, 64, 128, 256};
  for (unsigned i = 0; i < sizeof(lengths)/sizeof(lengths[0]); i++) {
    unsigned vl = lengths[i];
    if (prctl(PR_SVE_SET_VL, vl) != (int)vl) return 99;
` + checks.String() + "  }\n  return 0;\n}\n"
	file, err := Parse(ArchARM64, source.String())
	if err != nil {
		t.Fatal(err)
	}
	const triple = "aarch64-unknown-linux-gnu"
	ir, err := Translate(file, Options{Goarch: "arm64", TargetTriple: triple, Sigs: sigs})
	if err != nil {
		t.Fatal(err)
	}
	compileAndRunRuntimeTestWithCompiler(t, llc, []string{"aarch64-linux-gnu-gcc", "-march=armv8.2-a+sve"}, "raw_float_sve", triple, ir, main,
		[]string{"qemu-aarch64", "-cpu", "max,sve-max-vq=16", "-L", "/usr/aarch64-linux-gnu"})
}
