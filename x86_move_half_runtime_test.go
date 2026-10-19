package plan9asm

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func TestX86MoveHalfCompleteBytesFlagsAndUpperAliases(t *testing.T) {
	if runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64" {
		t.Skip("portable whole-byte runtime needs its native 64-bit host counterpart")
	}
	llc, clang, ok := findLlcAndClang(t)
	if !ok {
		t.Fatal("LLVM22 llc and clang required")
	}
	triple := testTargetTriple(runtime.GOOS, runtime.GOARCH)
	var source, declarations, prototypes, metadata strings.Builder
	for index, row := range x86MoveHalfOracle {
		for _, alias := range []bool{false, true} {
			name, destination := fmt.Sprintf("Half%d_%t", index, alias), 2
			if alias {
				destination = 1
			}
			args := fmt.Sprintf("(SI),X%d", destination)
			if row.vector {
				args = fmt.Sprintf("(SI),X1,X%d", destination)
			}
			fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-32\nMOVQ out+0(FP),AX\nMOVQ first+8(FP),BX\n"+
				"MOVQ second+16(FP),CX\nMOVQ half+24(FP),SI\nVMOVDQU64 (BX),Z1\nVMOVDQU64 (CX),Z2\n"+
				"MOVL $127,DX\nADDB $1,DL\nSTC\nSTD\n%s %s\nSETCS 64(AX)\nPUSHFQ\nPOPQ DX\nANDL $0xcd5,DX\nCLD\n"+
				"VMOVDQU64 Z%d,(AX)\nMOVQ DX,72(AX)\n%s X%d,88(AX)\nRET\n", name, row.op, args, destination, row.op, destination)
			fmt.Fprintf(&declarations, "func %s(out *[112]byte,first,second *[64]byte,half *uint64)\n", name)
			fmt.Fprintf(&prototypes, "extern void %s(uint8_t *,const uint8_t *,const uint8_t *,const uint64_t *);\n", name)
			flag := func(value bool) int {
				if value {
					return 1
				}
				return 0
			}
			fmt.Fprintf(&metadata, "{%s,%d,%d,%d},\n", name, flag(row.vector), flag(row.high), flag(alias))
		}
	}
	goAssemblerOracleOutput(t, "amd64", source.String(), "", true)
	ir, err := x86ScalarFPBitViewAPI(t, source.String(), declarations.String(), "amd64", triple, "Go")
	if err != nil {
		t.Fatal("complete original initialized half-move value source rejected", err)
	}
	driver := fmt.Sprintf(x86MoveHalfWholeBytesC, prototypes.String(), metadata.String())
	x86ConflictFunnelPortableRuntime(t, llc, clang, "move_half_whole_bytes", triple, ir, driver)
}

const x86MoveHalfWholeBytesC = `
#include <stdint.h>
#include <stdio.h>
#include <string.h>
%s
struct row {void (*run)(uint8_t *,const uint8_t *,const uint8_t *,const uint64_t *); int vector,high,alias;};
static const struct row rows[]={%s};
int main(void) {
  const uint64_t patterns[]={0,UINT64_MAX,UINT64_C(0x8000000000000000),UINT64_C(0x7ff0000000000001),
    UINT64_C(0xfff123456789abcd),UINT64_C(0x7ff8123456789abc),UINT64_C(0xaaaaaaaa55555555),UINT64_C(0x0123456789abcdef)};
  unsigned outcomes=0;
  for(unsigned r=0;r<sizeof(rows)/sizeof(rows[0]);r++) for(unsigned sample=0;sample<32;sample++) {
    uint8_t first[64],second[64],out[112],want[112];
    for(unsigned b=0;b<64;b++) {first[b]=(uint8_t)(sample*13+b*17);second[b]=(uint8_t)(sample*23+b*31+7);}
    uint64_t half=patterns[sample%%8]^(sample>=8?((uint64_t)sample<<21):0),flags=UINT64_C(0xc91);
    memset(out,0xa5,sizeof(out)); memset(want,0xa5,sizeof(want));
    const uint8_t *base=rows[r].vector||rows[r].alias?first:second;
    memcpy(want,base,64);
    memcpy(want+rows[r].high*8,&half,8);
    if(rows[r].vector) memset(want+16,0,48);
    want[64]=1;
    memcpy(want+72,&flags,8); memcpy(want+88,&half,8);
    rows[r].run(out,first,second,&half);
    if(memcmp(out,want,sizeof(out))) {fprintf(stderr,"half row=%%u sample=%%u\n",r,sample);return 1;}
    outcomes++;
  }
  return outcomes==512?0:2;
}
`
