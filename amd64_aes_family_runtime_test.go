package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

// FIPS-197 byte arithmetic, independent of LLVM intrinsics, the source grammar
// and the production defined-byte model. Compute the S-box from GF(2^8), not
// from a table shared with any emitter.
const x86AESIndependentC = `
#include <stdint.h>
#include <stdio.h>
#include <string.h>
#include <xmmintrin.h>
static uint8_t mul(uint8_t a,uint8_t b){uint8_t v=0;for(int i=0;i<8;i++){if(b&1)v^=a; a=(uint8_t)((a<<1)^((a&128)?0x1b:0));b>>=1;}return v;}
static uint8_t rol(uint8_t a,int n){return (uint8_t)((a<<n)|(a>>(8-n)));}
static uint8_t box(uint8_t a){uint8_t p=1,b=a;int n=254;if(!a)p=0;else while(n){if(n&1)p=mul(p,b);b=mul(b,b);n>>=1;}return p^rol(p,1)^rol(p,2)^rol(p,3)^rol(p,4)^0x63;}
static uint8_t inverse(uint8_t a){for(int i=0;i<256;i++)if(box((uint8_t)i)==a)return (uint8_t)i;return 0;}
static void mix(uint8_t *out,const uint8_t *in,int inv){for(int c=0;c<4;c++)for(int r=0;r<4;r++){
 const uint8_t forward[4]={2,3,1,1},backward[4]={14,11,13,9};out[c*4+r]=0;
 for(int k=0;k<4;k++)out[c*4+r]^=mul(in[c*4+k],inv?backward[(k-r+4)%4]:forward[(k-r+4)%4]);}}
static void aes(uint8_t *out,const uint8_t *state,const uint8_t *key,int op,int imm){
 uint8_t shifted[16],temp[16];
 if(op==4){mix(out,key,1);return;}
 if(op==5){for(int h=0;h<2;h++)for(int b=0;b<4;b++){out[h*8+b]=box(key[h*8+4+b]);out[h*8+4+b]=box(key[h*8+4+(b+1)%4])^(b==0?imm:0);}return;}
 for(int c=0;c<4;c++)for(int r=0;r<4;r++){int source=((c+(op>=2?4-r:r))%4)*4+r;shifted[c*4+r]=op>=2?inverse(state[source]):box(state[source]);}
 if(op==0||op==2)mix(temp,shifted,op==2);else memcpy(temp,shifted,16);
 for(int i=0;i<16;i++)out[i]=temp[i]^key[i];
}
typedef void(*Program)(uint8_t*,const uint8_t*,const uint8_t*,const uint8_t*,uint64_t,uint32_t);
struct Aligned {_Alignas(64) uint8_t before[64],value[80],after[64];};
static int check(Program p,const char *name,int width,int op,int legacy,int memory,int alias,int imm){
 for(int sample=0;sample<37;sample++)for(int profile=0;profile<2;profile++){
  struct Aligned a,b,old,out;memset(&a,0xa7,sizeof(a));memset(&b,0xb9,sizeof(b));memset(&old,0x6d,sizeof(old));memset(&out,0xc3,sizeof(out));
  for(int i=0;i<64;i++){a.value[i]=(uint8_t)(sample*19+i*37);b.value[i]=(uint8_t)(sample*71+i*11);old.value[i]=(uint8_t)(sample*97+i*3);}
  struct Aligned oa=a,ob=b,oo=old;uint64_t seed=profile?0x8d5:0;uint32_t csr=0x1f80|((sample%4)<<13)|(sample%64),saved=_mm_getcsr();_mm_setcsr(csr);
  p(out.value,a.value,b.value,old.value,seed,csr);uint32_t after=_mm_getcsr();_mm_setcsr(saved);
  uint64_t flags;uint32_t reported;memcpy(&flags,out.value+64,8);memcpy(&reported,out.value+72,4);
  if((flags&0x8d5)!=seed||reported!=csr||after!=csr||memcmp(&a,&oa,sizeof(a))||memcmp(&b,&ob,sizeof(b))||memcmp(&old,&oo,sizeof(old)))return 1;
  uint8_t expected[64];const uint8_t *retained=alias==0?a.value:alias==1?b.value:old.value;
  memcpy(expected,legacy?retained:(const uint8_t[64]){0},64);
  for(int lane=0;lane<width;lane+=16)aes(expected+lane,b.value+lane,a.value+lane,op,imm);
  for(int i=0;i<64;i++)if(out.value[i]!=expected[i]){fprintf(stderr,"%s sample%d byte%d got%x want%x\n",name,sample,i,out.value[i],expected[i]);return 2;}
  for(int i=0;i<64;i++)if(out.before[i]!=0xc3||out.after[i]!=0xc3)return 3;
  for(int i=76;i<80;i++)if(out.value[i]!=0xc3)return 4;
 }return 0;
}
`

func TestCrossLinuxRuntimeMatrixX86AESFamilyWholeByteAliases(t *testing.T) {
	execution := x86FloatCompareExecutionTools(t)
	var source, declarations, reference, checks strings.Builder
	declarations.WriteString("package oracle\n")
	reference.WriteString(x86AESIndependentC)
	count := 0
	for index, name := range x86AESOracleNames {
		op, legacy := index%6, !strings.HasPrefix(name, "V")
		widths := []int{16}
		if !legacy && op < 4 {
			widths = append(widths, 32, 64)
		}
		for _, width := range widths {
			bank := map[int]string{16: "X", 32: "Y", 64: "Z"}[width]
			for _, memory := range []bool{false, true} {
				for alias := 0; alias < 3; alias++ {
					if legacy && op < 4 && alias != 1 {
						continue
					}
					key := bank + "0"
					if memory {
						key = "(CX)"
					}
					row, destination := x86AESOracleRow(name, key, bank)
					if legacy && op < 4 {
						destination = 1
					} else {
						row = strings.TrimSuffix(row, bank+"2") + fmt.Sprintf("%s%d", bank, alias)
						destination = alias
					}
					fn := fmt.Sprintf("A%d", count)
					count++
					fmt.Fprintf(&declarations, "func %s(out *[80]byte,a,b,old *[64]byte,flags uint64,csr uint32)\n", fn)
					fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-44\nMOVQ out+0(FP),AX\nMOVQ a+8(FP),CX\nMOVQ b+16(FP),DX\nMOVQ old+24(FP),R8\nVMOVUPS (CX),Z0\nVMOVUPS (DX),Z1\nVMOVUPS (R8),Z2\nLDMXCSR csr+40(FP)\nPUSHQ flags+32(FP)\nPOPFQ\n%s\nVMOVUPS Z%d,(AX)\nPUSHFQ\nPOPQ R9\nMOVQ R9,64(AX)\nSTMXCSR 72(AX)\nRET\n", fn, row, destination)
					fmt.Fprintf(&reference, "extern void %s(uint8_t*,const uint8_t*,const uint8_t*,const uint8_t*,uint64_t,uint32_t);\n", fn)
					legacyInt, memoryInt := 0, 0
					if legacy {
						legacyInt = 1
					}
					if memory {
						memoryInt = 1
					}
					fmt.Fprintf(&checks, "if(check(%s,%q,%d,%d,%d,%d,%d,27))return 1;\n", fn, row, width, op, legacyInt, memoryInt, destination)
				}
			}
		}
	}
	reference.WriteString("int main(void){\n" + checks.String() + "return 0;}\n")
	goAssemblerOracleOutput(t, "amd64", source.String(), "", true)
	ir := x86NarrowOracleIR(t, source.String(), declarations.String(), execution.triple)
	for _, level := range []string{"-O0", "-O2"} {
		t.Run(level, func(t *testing.T) { execution.compileAndRunLLVM(t, "aes-byte-alias", ir, reference.String(), level) })
	}
	t.Run("IR_O2/backend_O2", func(t *testing.T) {
		execution.compileAndRunLLVM(t, "aes-byte-alias-ir-o2", x86DwordProductIROptimize(t, ir), reference.String(), "-O2")
	})
	t.Logf("%d AES-family kernels ×37 samples ×2 flag/MXCSR seeds on three explicit optimization axes; all64 vector bytes, aliases, exact native memory, inputs and output canaries; portable wide VAES is not original AVX512 execution", count)
}

// Execute genuine Go-assembler bytes, not bytes obtained from our grammar or
// LLVM. The host's legacy AES X128 axis is distinct from portable wide VAES;
// no original AVX512 execution or native-Linux claim is made here.
func TestCrossLinuxRuntimeMatrixX86AESFamilyOriginalGoFullX128Bytes(t *testing.T) {
	execution := x86FloatCompareExecutionTools(t)
	var original, checks strings.Builder
	count := 0
	for op, name := range x86AESOracleNames[:6] {
		for _, memory := range []bool{false, true} {
			for alias := 0; alias < 3; alias++ {
				if op < 4 && alias != 1 {
					continue
				}
				input := "X0"
				if memory {
					input = "(SI)"
				}
				row, destination := x86AESOracleRow(name, input, "X")
				if op >= 4 {
					row = strings.TrimSuffix(row, "X2") + fmt.Sprintf("X%d", alias)
					destination = alias
				}
				fn := fmt.Sprintf("NativeAES%d", count)
				count++
				// SysV DI/SI/DX/CX/R8/R9 provide out/a/b/old/flags/csr.
				body := fmt.Sprintf("TEXT actual(SB),4,$0-0\nMOVOU (SI),X0\nMOVOU (DX),X1\nMOVOU (CX),X2\nMOVL R9,76(DI)\nLDMXCSR 76(DI)\nPUSHQ R8\nPOPFQ\n%s\nMOVOU X%d,(DI)\nPUSHFQ\nPOPQ AX\nMOVQ AX,64(DI)\nSTMXCSR 72(DI)\nRET\n", row, destination)
				code := x86OrderingHintGoBytes(t, "amd64", body)
				var encoded []string
				for _, b := range code {
					encoded = append(encoded, fmt.Sprintf("0x%02x", b))
				}
				symbol := fn
				if strings.Contains(execution.triple, "apple") {
					symbol = "_" + symbol
				}
				fmt.Fprintf(&original, "extern void %s(uint8_t*,const uint8_t*,const uint8_t*,const uint8_t*,uint64_t,uint32_t);\n__asm__(%q);\n", fn, ".text\n.globl "+symbol+"\n"+symbol+":\n.byte "+strings.Join(encoded, ",")+"\n")
				fmt.Fprintf(&checks, "if(nativecheck(%s,%q,%d))return 1;\n", fn, row, op)
			}
		}
	}
	reference := x86AESIndependentC + original.String() + x86AESNativeCheckC + "int main(void){\n" + checks.String() + fmt.Sprintf("printf(\"original Go AES X128 kernels %d samples %d full-byte failures 0\\n\");return 0;}\n", count, count*37*2)
	output := execution.compileAndRunLLVM(t, "aes-native-go-bytes", "target triple = \""+execution.triple+"\"\n", reference, "-O2")
	want := fmt.Sprintf("original Go AES X128 kernels %d samples %d full-byte failures 0\n", count, count*37*2)
	if string(output) != want {
		t.Fatal("actual Go full-X128 source-byte oracle", string(output))
	}
	t.Log(strings.TrimSpace(string(output)))
}

const x86AESNativeCheckC = `
static int nativecheck(Program p,const char *name,int op){
 for(int sample=0;sample<37;sample++)for(int profile=0;profile<2;profile++){
  struct Aligned a,b,old,out;memset(&a,0xa7,sizeof(a));memset(&b,0xb9,sizeof(b));memset(&old,0x6d,sizeof(old));memset(&out,0xc3,sizeof(out));
  for(int i=0;i<64;i++){a.value[i]=(uint8_t)(sample*19+i*37);b.value[i]=(uint8_t)(sample*71+i*11);old.value[i]=(uint8_t)(sample*97+i*3);}
  struct Aligned oa=a,ob=b,oo=old;uint64_t seed=profile?0x8d5:0;uint32_t csr=0x1f80|((sample%4)<<13)|(sample%64),saved=_mm_getcsr();_mm_setcsr(csr);
  p(out.value,a.value,b.value,old.value,seed,csr);uint32_t after=_mm_getcsr();_mm_setcsr(saved);
  uint64_t flags;uint32_t reported,scratch;memcpy(&flags,out.value+64,8);memcpy(&reported,out.value+72,4);memcpy(&scratch,out.value+76,4);
  uint8_t expected[16];aes(expected,b.value,a.value,op,27);
  if(memcmp(out.value,expected,16)||(flags&0x8d5)!=seed||reported!=csr||scratch!=csr||after!=csr||memcmp(&a,&oa,sizeof(a))||memcmp(&b,&ob,sizeof(b))||memcmp(&old,&oo,sizeof(old))){fprintf(stderr,"original %s sample%d full bytes/flags/CSR/input mismatch\n",name,sample);return 1;}
  for(int i=16;i<64;i++)if(out.value[i]!=0xc3)return 2;
  for(int i=0;i<64;i++)if(out.before[i]!=0xc3||out.after[i]!=0xc3)return 3;
 }return 0;
}
`
