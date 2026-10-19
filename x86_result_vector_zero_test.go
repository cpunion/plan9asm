package plan9asm

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

// The complete original eight named/VEX2/VEX3/WIG source programs retain every
// X/Y/Z bank and K/FLAGS canary. Actual native MXCSR checks remain separate.
func TestX86VectorZeroOriginalWholeBankPortableExecution(t *testing.T) {
	llc, clang, ok := findLlcAndClang(t)
	if !ok {
		t.Fatal("LLVM22 llc/clang required")
	}
	source, _, prototypes, checks := x86VectorZeroRuntimeFixture(t)
	var declarations strings.Builder
	for index := 0; index < 8; index++ {
		fmt.Fprintf(&declarations, "func vectorzero%d(out,in *byte,mask uint64)\n", index)
	}
	x86LEAActualTargetGoObject(t, "amd64", "x86_64-unknown-linux-gnu", source)
	main := "#include <stdint.h>\n#include <stdio.h>\n#include <string.h>\n" + prototypes + `
static int checkVectorZero(void (*program)(uint8_t*,const uint8_t*,uint64_t),
                           const char *name,int clearLow) {
    for (unsigned round=0; round<16; round++) {
        uint8_t input[2050], original[2050], output[3596];
        for (unsigned i=0; i<sizeof(input); i++) input[i]=(uint8_t)(i*73+round*19+1);
        memcpy(original,input,sizeof(input));
        memset(output,0xc3,sizeof(output));
        uint64_t mask=UINT64_C(0x8123456789abcdef)^round;
        program(output+1,input+1,mask);
        uint64_t gotMask;
        memcpy(&gotMask,output+3587,8);
        if (output[0]!=0xc3 || output[3595]!=0xc3 || output[3585]!=1 ||
            output[3586]!=0 || gotMask!=mask || memcmp(input,original,sizeof(input))) {
            fprintf(stderr,"%s round=%u FLAGS/K/input/canary failure\n",name,round);
            return 1;
        }
        const unsigned widths[]={16,32,64}, offsets[]={0,16,48};
        for (unsigned reg=0; reg<32; reg++) {
            for (unsigned view=0; view<3; view++) {
                for (unsigned byte=0; byte<widths[view]; byte++) {
                    uint8_t want=reg<16 && (clearLow || byte>=16) ? 0 : input[1+reg*64+byte];
                    uint8_t got=output[1+reg*112+offsets[view]+byte];
                    if (got!=want) {
                        fprintf(stderr,"%s round=%u reg=%u view=%u byte=%u\n",
                                name,round,reg,view,byte);
                        return 1;
                    }
                }
            }
        }
    }
    return 0;
}
int main(void) {
` + checks + "return 0;\n}\n"
	triple := testTargetTriple(runtime.GOOS, runtime.GOARCH)
	for _, api := range []string{"text", "module", "Go"} {
		t.Run(api, func(t *testing.T) {
			ir, err := x86ScalarFPBitViewAPI(t, source, declarations.String(), "amd64", triple, api)
			if err != nil {
				t.Fatal("complete original vector-zero source lost its native-read continuation", err)
			}
			for index, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
				for _, level := range []string{"-O0", "-O2"} {
					compileAndRunRuntimeTestForTarget(t, llc, clang,
						fmt.Sprintf("vector-zero-%s-ir%d-%s", api, index, level[1:]), triple, text, main, nil)
				}
			}
		})
	}
}
