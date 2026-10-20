package plan9asm

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/xgo-dev/plan9asm/internal/gotoolchain"
)

// Diagnose this original owner's actual execution route. These observations
// cannot authorize another runner or replace the original numerical verdict.
func x86OriginalFloatingRuntimeRequireCPU(t *testing.T, profile, clang, triple string, prefix []string) {
	t.Helper()
	var required int
	switch profile {
	case "scalar-precision":
		required = 1
	case "fma-caller":
		required = 2
	default:
		t.Fatal("unknown original floating execution profile", profile)
	}
	execution := x86FloatCompareExecution{
		compiler: []string{clang, "-target", triple},
		triple:   triple,
		prefix:   prefix,
	}
	directory := t.TempDir()
	source := filepath.Join(directory, "original-floating-preflight.c")
	executable := filepath.Join(directory, gotoolchain.HostExecutable("original-floating-preflight"))
	program := fmt.Sprintf("#define REQUIRED_PROFILE %d\n", required) +
		x86OriginalFloatingRuntimeStateC + x86OriginalFloatingRuntimeOpcodeC
	if err := os.WriteFile(source, []byte(program), 0600); err != nil {
		t.Fatal(err)
	}
	execution.compileC(t, source, executable)
	output := execution.run(t, executable)
	t.Logf("actual original floating allocation profile=%s triple=%s prefix=%q: %s", profile, triple, prefix, output)
}

// Scalar EVEX conversions require F, not packed narrow VL, DQ or BW. The
// complete FMA owner additionally executes VEX FMA and narrow packed EVEX.
// Keep this predicate portable so independent negative rows never execute ISA.
const x86OriginalFloatingRuntimeStateC = `#include <stdint.h>
static int original_floating_required_state(unsigned profile,unsigned sse2,unsigned avx,
                                            unsigned osxsave,unsigned f,unsigned vl,
                                            unsigned fma,uint64_t xcr0) {
    if (!sse2 || !avx || !osxsave || !f || (xcr0&0xe6)!=0xe6) return 0;
    return profile==1 || (profile==2 && vl && fma);
}
`

// Only these independent probe functions have native target attributes. The
// baseline entry prints real CPUID/XCR0 before any unsupported opcode can run.
const x86OriginalFloatingRuntimeOpcodeC = `
#include <stdio.h>
#include <cpuid.h>
#if !defined(_WIN32)
#include <sys/resource.h>
#endif

__attribute__((target("avx512f")))
static int original_scalar_precision(void) {
    uint32_t narrow=1;
    uint64_t wide=1;
    unsigned mask=1;
    __asm__ volatile("vxorps %%xmm0,%%xmm0,%%xmm0; kmovw %2,%%k1; "
                     "vcvtsd2ss %{rn-sae%},%%xmm0,%%xmm0,%%xmm1%{%%k1%}; vmovss %%xmm1,%0; "
                     "vcvtss2sd %{sae%},%%xmm0,%%xmm0,%%xmm1%{%%k1%}%{z%}; vmovsd %%xmm1,%1"
                     : "=m"(narrow),"=m"(wide) : "r"(mask) : "xmm0","xmm1","k1","memory");
    return narrow==0 && wide==0;
}

__attribute__((target("avx,fma,avx512f,avx512vl")))
static int original_fma_caller(void) {
    uint32_t scalar=1,packed[8];
    uint64_t wide[8];
    unsigned mask=0xffff;
    __asm__ volatile("vxorps %%xmm0,%%xmm0,%%xmm0; "
                     "vfmadd132ss %%xmm0,%%xmm0,%%xmm0; vmovss %%xmm0,%0; "
                     "vxorps %%ymm0,%%ymm0,%%ymm0; kmovw %3,%%k1; "
                     "vfmadd132ps %%ymm0,%%ymm0,%%ymm0%{%%k1%}; vmovups %%ymm0,%1; "
                     "vpxord %%zmm0,%%zmm0,%%zmm0; "
                     "vfmadd132pd %%zmm0,%%zmm0,%%zmm0; vmovupd %%zmm0,%2"
                     : "=m"(scalar),"=m"(packed),"=m"(wide) : "r"(mask) : "zmm0","k1","memory");
    if (scalar) return 0;
    for (unsigned i=0;i<8;i++) if (packed[i] || wide[i]) return 0;
    return 1;
}

int main(void) {
#if !defined(_WIN32)
    struct rlimit no_core={0,0};
    if (setrlimit(RLIMIT_CORE,&no_core)) return 77;
#endif
    unsigned a,b,c,d,sse2=0,avx=0,osxsave=0,f=0,vl=0,fma=0;
    uint64_t xcr0=0;
    if (__get_cpuid(1,&a,&b,&c,&d)) {
        sse2=(d>>26)&1; avx=(c>>28)&1; osxsave=(c>>27)&1; fma=(c>>12)&1;
        if (osxsave) {
            unsigned low,high;
            __asm__ volatile("xgetbv" : "=a"(low),"=d"(high) : "c"(0));
            xcr0=((uint64_t)high<<32)|low;
        }
    }
    if (__get_cpuid_count(7,0,&a,&b,&c,&d)) {
        f=(b>>16)&1; vl=(b>>31)&1;
    }
    unsigned profile=REQUIRED_PROFILE;
    printf("original-floating profile=%u SSE2=%u AVX=%u OSXSAVE=%u AVX512F=%u AVX512VL=%u FMA=%u XCR0=%llx\n",
           profile,sse2,avx,osxsave,f,vl,fma,(unsigned long long)xcr0);
    fflush(stdout);
    if (!original_floating_required_state(profile,sse2,avx,osxsave,f,vl,fma,xcr0)) return 78;
    if (profile==1 && !original_scalar_precision()) return 79;
    if (profile==2 && !original_fma_caller()) return 80;
    puts("ORIGINAL_FLOATING_RUNTIME_CPU_OPCODES_OK");
    return 0;
}
`
