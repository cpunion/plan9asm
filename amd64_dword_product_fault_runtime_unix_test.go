//go:build unix

package plan9asm

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

// Independent fault envelopes include all eight bytes of each potentially
// active qword, even though only its low dword affects the numerical product.
// Original SSE instructions are the bytes emitted by Go, not mnemonic aliases
// assembled by the C compiler. No original AVX512 execution is claimed here.
func TestCrossLinuxRuntimeMatrixX86DwordProductDeadResultFaultEnvelopes(t *testing.T) {
	execution := x86FloatCompareExecutionTools(t)
	var assembly, declarations, prototypes, originals, cases strings.Builder
	declarations.WriteString("package faultoracle\n")
	count := 0
	for _, name := range x86DwordProductOracleNames {
		legacy := !strings.HasPrefix(name, "V")
		widths := []int{16}
		if !legacy {
			widths = append(widths, 32, 64)
		}
		for _, width := range widths {
			for broadcast := 0; broadcast < 2; broadcast++ {
				if legacy && broadcast != 0 {
					continue
				}
				for masking := 0; masking < 3; masking++ {
					if legacy && masking != 0 {
						continue
					}
					fn := fmt.Sprintf("F%d", count)
					count++
					bank := map[int]string{16: "X", 32: "Y", 64: "Z"}[width]
					op, args := name, "(BX),"+bank+"1"
					if broadcast != 0 {
						op += ".BCST"
					}
					if masking == 2 {
						op += ".Z"
					}
					if !legacy {
						if masking != 0 {
							args += ",K1"
						}
						args += "," + bank + "2"
					}
					fmt.Fprintf(&declarations, "func %s(p *[64]byte,mask uint64)\n", fn)
					fmt.Fprintf(&assembly, "TEXT ·%s(SB),4,$0-16\nMOVQ p+0(FP),BX\nMOVQ mask+8(FP),AX\nVPXORD Z1,Z1,Z1\nVPXORD Z2,Z2,Z2\nKMOVQ AX,K1\n%s %s\nRET\n", fn, op, args)
					fmt.Fprintf(&prototypes, "extern void %s(const uint8_t*,uint64_t);\n", fn)
					original := "NULL"
					if legacy {
						original = "original_" + fn
						actual := assembleX87ControlBytes(t, "amd64", "TEXT probe(SB),4,$0-0\nPXOR X1,X1\n"+name+" (DI),X1\nRET\n")
						var bytes []string
						for _, value := range actual {
							bytes = append(bytes, fmt.Sprintf("0x%02x", value))
						}
						symbol := original
						if strings.Contains(execution.triple, "apple") {
							symbol = "_" + symbol
						}
						fmt.Fprintf(&originals, "extern void %s(const uint8_t*,uint64_t);\n__asm__(%q);\n", original, ".text\n.globl "+symbol+"\n"+symbol+":\n.byte "+strings.Join(bytes, ",")+"\n")
					}
					fmt.Fprintf(&cases, "{%s,%s,%d,%d,%d,%t},\n", fn, original, width, broadcast, masking, legacy)
				}
			}
		}
	}
	goAssemblerOracleOutput(t, "amd64", assembly.String(), "", true)
	ir := x86NarrowOracleIR(t, assembly.String(), declarations.String(), execution.triple)
	strict, deviations := 1, 0
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		strict, deviations = 0, 8
	}
	reference := fmt.Sprintf(x86DwordProductFaultC, prototypes.String(), originals.String(), cases.String(), strict)
	for _, level := range []string{"-O0", "-O2"} {
		t.Run(level, func(t *testing.T) {
			output := execution.compileAndRunLLVM(t, "dword-product-fault", ir, reference, level)
			want := fmt.Sprintf("dword-product fault outcomes %d original SSE outcomes 80 alignment deviations %d failures 0\n", count*10*4, deviations)
			if string(output) != want {
				t.Fatalf("native width/alignment, dead result or qword masked fault suppression: %s", output)
			}
			t.Log(strings.TrimSpace(string(output)))
		})
	}
}

const x86DwordProductFaultC = `
#include <stdint.h>
#include <stdio.h>
#include <string.h>
#include <signal.h>
#include <unistd.h>
#include <sys/mman.h>
#include <sys/resource.h>
#include <sys/wait.h>
#include <stdbool.h>
%s
%s
typedef struct {void(*translated)(const uint8_t*,uint64_t);void(*original)(const uint8_t*,uint64_t);unsigned width,broadcast,masking,legacy;} Kernel;
static Kernel kernels[]={%s};
static const uint64_t masks[]={0,1,0x80,UINT64_MAX};
static unsigned need(Kernel k,uint64_t mask) {
 if(!k.masking) return k.broadcast?8:k.width;
 unsigned high=0;
 for(unsigned i=0;i<k.width/8;i++) if((mask>>i)&1) high=k.broadcast?8:(i+1)*8;
 return high;
}
static int observe(Kernel k,unsigned address,uint64_t mask,int original) {
 pid_t child=fork();
 if(child<0) return -1;
 if(child==0) {
  struct rlimit no_core={0,0};
  if(setrlimit(RLIMIT_CORE,&no_core)) _exit(120);
  _Alignas(64) uint8_t data[128];memset(data,0xa7,sizeof(data));
  const uint8_t *p=data;
  if(address==1) p++;
  if(address==2) p=NULL;
  if(address>=3) {
   long page=sysconf(_SC_PAGESIZE);if(page<=0) _exit(121);
   uint8_t *mapping=mmap(NULL,2*(size_t)page,PROT_READ|PROT_WRITE,MAP_PRIVATE|MAP_ANONYMOUS,-1,0);
   if(mapping==MAP_FAILED) _exit(122);
   memset(mapping,0xa7,(size_t)page);
   if(mprotect(mapping+page,(size_t)page,PROT_NONE)) _exit(123);
   unsigned available[]={0,4,7,8,k.width-1,k.width,64};
   p=mapping+page-available[address-3];
  }
  if(original) k.original(p,mask);else k.translated(p,mask);
  _exit(0);
 }
 int status=0;if(waitpid(child,&status,0)!=child) return -2;
 if(WIFEXITED(status)) return WEXITSTATUS(status)?-3:0;
 if(WIFSIGNALED(status)&&(WTERMSIG(status)==SIGSEGV||WTERMSIG(status)==SIGBUS)) return 1;
 return -4;
}
static int expected(Kernel k,unsigned address,uint64_t mask) {
 unsigned bytes=need(k,mask),available[]={0,4,7,8,k.width-1,k.width,64};
 if(k.legacy&&(address==1||(address>=3&&(available[address-3]%%16)!=0))) return 1;
 return (address==2&&bytes)||(address>=3&&bytes>available[address-3]);
}
int main(void) {
 unsigned count=0,native=0,deviations=0,failures=0;const int strict=%d;
 for(unsigned i=0;i<sizeof(kernels)/sizeof(kernels[0]);i++) for(unsigned a=0;a<10;a++) for(unsigned m=0;m<4;m++) {
  Kernel k=kernels[i];int want=expected(k,a,masks[m]),got=observe(k,a,masks[m],0);count++;
  if(k.original) {
   int actual=observe(k,a,masks[m],1);native++;
   // This exact readable alignment-only Rosetta deviation is reported,
   // never substituted into the strict native Linux architectural gate.
   if(!strict&&a==1&&actual==0&&want==1) {deviations++;want=actual;}
   if(actual!=want) {fprintf(stderr,"original SSE kernel%%u address%%u got%%d want%%d\n",i,a,actual,want);failures++;}
  }
  if(got!=want) {fprintf(stderr,"kernel%%u address%%u mask%%llx got%%d want%%d\n",i,a,(unsigned long long)masks[m],got,want);failures++;}
 }
 printf("dword-product fault outcomes %%u original SSE outcomes %%u alignment deviations %%u failures %%u\n",count,native,deviations,failures);
 return failures?1:0;
}
`
