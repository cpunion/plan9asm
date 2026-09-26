package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

func testARM64RawSVEExtraShiftRuntime(t *testing.T, llc string) {
	var source, declarations, checks strings.Builder
	declarations.WriteString(arm64SVEExtraShiftScalarReference)
	sigs := make(map[string]FuncSig)
	for operation, form := range arm64RawSVEExtraShiftCases() {
		for size := 0; size < 4; size++ {
			bits := 8 << size
			shifts := []int{1, bits / 2, bits - 1, bits}
			if form.left {
				shifts = []int{0, 1, bits / 2, bits - 1}
			}
			if form.reverse {
				shifts = []int{0}
			}
			for _, shift := range shifts {
				for _, destination := range []int{31, 30} {
					name := fmt.Sprintf("extra_shift_%d", len(sigs))
					native := []string{
						"ptrue p0.b", "ld1b { z29.b }, p0/z, [x2]", "cmpne p7.b, p0/z, z29.b, #0",
						"ld1b { z30.b }, p0/z, [x0]", "ld1b { z31.b }, p0/z, [x1]",
						form.assembly(size, destination, 30, 7, shift),
						fmt.Sprintf("st1b { z%d.b }, p0, [x3]", destination),
					}
					fmt.Fprintf(&source, "TEXT %s(SB),$0-32\nMOVD a+0(FP),R0\nMOVD b+8(FP),R1\nMOVD mask+16(FP),R2\nMOVD out+24(FP),R3\n", name)
					for _, word := range assembleARM64LLVMWords(t, native, "+sve2") {
						fmt.Fprintf(&source, "WORD $%#08x\n", word)
					}
					source.WriteString("RET\n")
					sigs[name] = FuncSig{Name: name, Args: []LLVMType{Ptr, Ptr, Ptr, Ptr}, Ret: Void, Frame: FrameLayout{Params: []FrameSlot{
						{Offset: 0, Type: Ptr, Index: 0, Field: -1}, {Offset: 8, Type: Ptr, Index: 1, Field: -1},
						{Offset: 16, Type: Ptr, Index: 2, Field: -1}, {Offset: 24, Type: Ptr, Index: 3, Field: -1},
					}}}
					fmt.Fprintf(&declarations, "extern void %s(const void *, const void *, const void *, void *);\n", name)
					assembly := strings.NewReplacer("[x0]", "[%[a]]", "[x1]", "[%[b]]", "[x2]", "[%[mask]]", "[x3]", "[%[out]]").Replace(strings.Join(native, "\\n\\t"))
					old := "b"
					if destination == 30 {
						old = "a"
					}
					input := "a"
					if form.predicated && !form.reverse {
						input = old
					}
					active := "1"
					if form.predicated {
						active = "mask[i]"
					}
					fmt.Fprintf(&checks, `    {
      unsigned char a[256], b[256], mask[256], got[256], native[256], scalar[256];
      const unsigned bits = %d, bytes = bits / 8;
      uint64_t sign = UINT64_C(1) << (bits - 1), maximum = UINT64_MAX >> (64 - bits);
      uint64_t values[] = {0, 1, maximum, sign, sign - 1, sign + 1, bits - 1, bits, bits + 1,
        maximum >> 1, 2, 3, 0x5555555555555555ULL, 0xaaaaaaaaaaaaaaaaULL};
      for (unsigned phase = 0; phase < 18; phase++) {
        for (unsigned i = 0; i < sizeof(a); i += bytes) {
          unsigned lane = i / bytes;
          uint64_t x = values[(lane + phase) %% 14], y = values[(lane * 3 + phase * 5) %% 14];
          if (phase >= 14) {
            x = (0x0123456789abcdefULL * (lane + phase)) ^ maximum;
            y = 0xfedcba9876543211ULL * (lane * 7 + phase);
          }
          memcpy(a + i, &x, bytes);
          memcpy(b + i, &y, bytes);
          for (unsigned j = 0; j < bytes; j++) mask[i + j] = phase < 2 ? phase : (lane + phase) %% 3 != 0;
        }
        memset(got, 0x5a, sizeof(got));
        memset(native, 0x5a, sizeof(native));
        memset(scalar, 0x5a, sizeof(scalar));
        for (unsigned i = 0; i < vl; i += bytes) {
          uint64_t x = 0, old = 0;
          memcpy(&x, %s + i, bytes);
          memcpy(&old, %s + i, bytes);
          uint64_t value = old;
          if (%s) value = extra_shift_reference(%d, bits, x, old, %d);
          memcpy(scalar + i, &value, bytes);
        }
        __asm__ volatile("%s" :: [a]"r"(a), [b]"r"(b), [mask]"r"(mask), [out]"r"(native)
          : "p0", "p7", "z29", "z30", "z31", "memory");
        %s(a, b, mask, got);
        if (memcmp(got, native, sizeof(got)) || memcmp(got, scalar, sizeof(got))) {
          fprintf(stderr, "%s (%s): vl=%%u phase=%%u native=%%d scalar=%%d\n", vl, phase,
            memcmp(got, native, sizeof(got)), memcmp(got, scalar, sizeof(got)));
          return 1;
        }
      }
    }
`, bits, input, old, active, operation, shift, assembly, name, name, form.op)
				}
			}
		}
	}
	file, err := Parse(ArchARM64, source.String())
	if err != nil {
		t.Fatal(err)
	}
	const triple = "aarch64-unknown-linux-gnu"
	ir, err := Translate(file, Options{Goarch: "arm64", TargetTriple: triple, Sigs: sigs})
	if err != nil {
		t.Fatal(err)
	}
	main := arm64SVEVectorLengthMain(declarations.String(), checks.String())
	compileAndRunRuntimeTestWithCompiler(t, llc, []string{"aarch64-linux-gnu-gcc", "-march=armv8.2-a+sve2"}, "raw_extra_shift", triple, ir, main,
		[]string{"qemu-aarch64", "-cpu", "max,sve-max-vq=16", "-L", "/usr/aarch64-linux-gnu"})
}

// Widen intermediate arithmetic so rounding, sign extension and 64-bit shifts
// never overflow or shift by the C type width. Operation order matches the
// independent native-case list, not the implementation's intrinsic table.
const arm64SVEExtraShiftScalarReference = `
#include <stdlib.h>
static uint64_t extra_shift_reference(unsigned op, unsigned bits, uint64_t x, uint64_t old, unsigned shift) {
  unsigned __int128 modulus = (unsigned __int128)1 << bits;
  uint64_t maximum = (uint64_t)(modulus - 1);
  __int128 signed_x = (x & (UINT64_C(1) << (bits - 1))) ? (__int128)x - (__int128)modulus : x;
  unsigned __int128 value = x, rounding = shift ? (unsigned __int128)1 << (shift - 1) : 0;
  switch (op) {
  case 0: return signed_x / ((__int128)1 << shift); // ASRD rounds toward zero.
  case 1: return (old & ((UINT64_C(1) << shift) - 1)) | (value << shift);
  case 2:
    if (signed_x < 0) return 0;
    value <<= shift;
    return value > maximum ? maximum : value;
  case 3: return (old & (maximum ^ ((uint64_t)(modulus >> shift) - 1))) | (value >> shift);
  case 4: return (signed_x + (__int128)rounding) >> shift;
  case 5: return old + ((signed_x + (__int128)rounding) >> shift);
  case 6: return old + (signed_x >> shift);
  case 7: return (value + rounding) >> shift;
  case 8: return old + ((value + rounding) >> shift);
  case 9: return old + (value >> shift);
  case 10: return old >= bits ? (signed_x < 0 ? maximum : 0) : signed_x >> old;
  case 11: return old >= bits ? 0 : x << old;
  case 12: return old >= bits ? 0 : x >> old;
  default: abort();
  }
}
`
