package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

func testARM64SVEMemoryOffsetRuntime(t *testing.T, llc string, nonFaulting bool) {
	var source, declarations, checks strings.Builder
	sigs := make(map[string]FuncSig)
	ops := []string{"ld1b", "ld1h", "ld1w", "ld1d", "ld1sb", "ld1sh", "ld1sw", "st1b", "st1h", "st1w", "st1d"}
	if nonFaulting {
		ops = []string{"ldnf1b", "ldnf1h", "ldnf1w", "ldnf1d", "ldnf1sb", "ldnf1sh", "ldnf1sw"}
	}
	for _, op := range ops {
		memoryBytes := map[byte]int{'b': 1, 'h': 2, 'w': 4, 'd': 8}[op[len(op)-1]]
		load, signed := strings.HasPrefix(op, "ld"), strings.Contains(op, "1s")
		for size := 0; size < 5; size++ {
			elementBytes := 1 << size
			if elementBytes < memoryBytes || signed && elementBytes == memoryBytes || size == 4 && (nonFaulting || signed || memoryBytes < 4) {
				continue
			}
			// Named tests cover all currently accepted single-vector forms;
			// existing raw B/H decoders share their address implementation.
			forms := []bool{false}
			if !nonFaulting && !signed && memoryBytes <= 2 {
				forms = append(forms, true)
			}
			for _, raw := range forms {
				for _, offset := range []int{-8, 1, 7} {
					name := fmt.Sprintf("ordinary_offset_%d", len(sigs))
					mode := ""
					if load {
						mode = "/z"
					}
					native := []string{
						"ptrue p0.b", "ld1b { z20.b }, p0/z, [x2]", "cmpne p7.b, p0/z, z20.b, #0",
						"ld1b { z31.b }, p0/z, [x1]",
						fmt.Sprintf("%s { z31.%c }, p7%s, [x0, #%d, mul vl]", op, "bhsdq"[size], mode, offset),
						"st1b { z31.b }, p0, [x3]",
					}
					words := assembleARM64LLVMWords(t, native, "+sve2p1")
					fmt.Fprintf(&source, "TEXT %s(SB),$0-32\nMOVD base+0(FP),R0\nMOVD data+8(FP),R1\nMOVD mask+16(FP),R2\nMOVD out+24(FP),R3\n", name)
					if nonFaulting {
						source.WriteString("SETFFR\n")
					}
					for i, word := range words {
						if i != 4 || raw {
							fmt.Fprintf(&source, "WORD $%#08x\n", word)
							continue
						}
						address := fmt.Sprintf("(VL*%d)(R0)", offset)
						if offset < 0 {
							address = fmt.Sprintf("(-VL*%d)(R0)", -offset)
						}
						if load {
							fmt.Fprintf(&source, "Z%s %s, P7.Z, [Z31.%c]\n", strings.ToUpper(op), address, "BHSDQ"[size])
						} else {
							fmt.Fprintf(&source, "Z%s [Z31.%c], P7, %s\n", strings.ToUpper(op), "BHSDQ"[size], address)
						}
					}
					source.WriteString("RET\n")
					sigs[name] = FuncSig{Name: name, Args: []LLVMType{Ptr, Ptr, Ptr, Ptr}, Ret: Void, Frame: FrameLayout{Params: []FrameSlot{
						{Offset: 0, Type: Ptr, Index: 0, Field: -1}, {Offset: 8, Type: Ptr, Index: 1, Field: -1},
						{Offset: 16, Type: Ptr, Index: 2, Field: -1}, {Offset: 24, Type: Ptr, Index: 3, Field: -1},
					}}}
					fmt.Fprintf(&declarations, "extern void %s(void *, const void *, const void *, void *);\n", name)
					// SVE2.1 Q forms can predate the native cross assembler. The
					// independent LLVM encoding still addresses X0, loaded below.
					native[4] = fmt.Sprintf("mov x0, %%[base]\\n\\t.inst %#08x", words[4])
					assembly := strings.NewReplacer("[x1]", "[%[data]]", "[x2]", "[%[mask]]", "[x3]", "[%[out]]").Replace(strings.Join(native, "\\n\\t"))
					if nonFaulting {
						assembly = "setffr\\n\\t" + assembly
					}
					isLoad, isSigned := 0, 0
					if load {
						isLoad = 1
					}
					if signed {
						isSigned = 1
					}
					fmt.Fprintf(&checks, `    {
      unsigned char memory[8192], reference[8192], scalar_memory[8192], data[256], mask[256];
      unsigned char got[256], native[256], scalar[256];
      const unsigned element = %d, access = %d;
      const int load = %d, signed_load = %d;
      const int displacement = %d * (int)(vl / element * access);
      for (unsigned phase = 0; phase < 5; phase++) {
        for (unsigned i = 0; i < sizeof(memory); i++) memory[i] = i * 53 + (i >> 8) + phase * 71;
        memcpy(reference, memory, sizeof(memory));
        memcpy(scalar_memory, memory, sizeof(memory));
        for (unsigned i = 0; i < sizeof(data); i++) {
          data[i] = i * 83 + phase * 23;
          mask[i] = phase < 2 ? phase : (i / element + phase * 3) %% 7 != 0;
        }
        memset(got, 0x5a, sizeof(got));
        memset(native, 0x5a, sizeof(native));
        memset(scalar, 0x5a, sizeof(scalar));
        if (load) memset(scalar, 0, vl);
        else memcpy(scalar, data, vl);
        for (unsigned i = 0; i < vl / element; i++) {
          if (!mask[i * element]) continue;
          unsigned char *location = scalar_memory + 4096 + displacement + i * access;
          if (load) {
            uint64_t value = 0;
            memcpy(&value, location, access);
            if (signed_load && (value >> (access * 8 - 1))) value |= ~(UINT64_MAX >> (64 - access * 8));
            memcpy(scalar + i * element, &value, element < 8 ? element : 8);
          } else {
            memcpy(location, data + i * element, access);
          }
        }
        void *base = phase ? memory + 4096 : 0;
        void *native_base = phase ? reference + 4096 : 0;
        __asm__ volatile("%s" :: [base]"r"(native_base), [data]"r"(data), [mask]"r"(mask), [out]"r"(native)
          : "x0", "p0", "p7", "z20", "z31", "memory");
        %s(base, data, mask, got);
        if (memcmp(got, native, sizeof(got)) || memcmp(got, scalar, sizeof(got)) ||
            memcmp(memory, reference, sizeof(memory)) || memcmp(memory, scalar_memory, sizeof(memory))) {
          fprintf(stderr, "%s: vl=%%u phase=%%u\n", vl, phase);
          return %d;
        }
      }
    }
`, elementBytes, memoryBytes, isLoad, isSigned, offset, assembly, name, name, len(sigs))
				}
			}
		}
	}
	requireARM64SVEGoAssemblerResult(t, source.String(), true)
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
	compileAndRunRuntimeTestWithCompiler(t, llc, []string{"aarch64-linux-gnu-gcc", "-march=armv8.2-a+sve"}, "ordinary_memory_offset", triple, ir, main,
		[]string{"qemu-aarch64", "-cpu", "max,sve-max-vq=16", "-L", "/usr/aarch64-linux-gnu"})
}
