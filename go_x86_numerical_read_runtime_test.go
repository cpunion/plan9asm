//go:build go1.27

package plan9asm

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// This carrier provides actual mutable linked numerical storage. It is not a
// Go native-entry/ABI0 permission: the required Linux oracle separately executes
// the original Go-linked declarations. A retargeted host run is diagnostic only.
func numericalRuntimeC(cases []numericalExecutionCase, arch string) string {
	word, seed := "uint64_t", "UINT64_C(0x8877665544332211)"
	if arch == "386" {
		word, seed = "uint32_t", "UINT64_C(0x44332211)"
	}
	var c strings.Builder
	c.WriteString(`#include <stdint.h>
#include <stdio.h>
#include <string.h>
#include <signal.h>
#include <sys/mman.h>
#include <sys/wait.h>
#include <unistd.h>
_Alignas(4096) unsigned char mutable[4096] __asm__("example.invalid/numericalfixture.Mutable");
_Alignas(4096) unsigned char cpu[4096] __asm__("internal/cpu.X86");
static uint64_t mask(unsigned n) {
 if(n==8) return UINT64_MAX;
 return (UINT64_C(1)<<(n*8))-1;
}
static uint64_t load(unsigned n) {
 uint64_t value=0;
 for(unsigned i=0;i<n;i++) value|=(uint64_t)mutable[3+i]<<(8*i);
 return value;
}
`)
	for index := range cases {
		fmt.Fprintf(&c, "extern %s numerical%d(void);\n", word, index)
	}
	c.WriteString("int main(void) {\n memset(mutable,0xa5,sizeof(mutable));\n memset(cpu,0x6b,sizeof(cpu));\n")
	c.WriteString(" for(unsigned round=0;round<3;round++) {\n  for(unsigned i=0;i<32;i++) mutable[i]=(unsigned char)(i*37+round*71);\n  cpu[64]=round&1;\n")
	dead := -1
	for index, test := range cases {
		fmt.Fprintf(&c, "  { uint64_t want=%s, value=load(%d);\n", seed, test.source)
		switch {
		case test.dead:
			dead = index
			c.WriteString("   want=7;\n")
		case test.cpu:
			c.WriteString("   want=(want&~UINT64_C(255))|cpu[64];\n")
		case test.compare:
			partner := seed
			if test.immediate {
				partner = "UINT64_C(4294967295)"
				if test.source == 8 {
					partner = "UINT64_MAX" // Actual CMPQ sign-extended imm32.
				}
			}
			fmt.Fprintf(&c, "   uint64_t a=value, b=%s&mask(%d);\n", partner, test.source)
			if test.reversed {
				c.WriteString("   uint64_t swap=a; a=b; b=swap;\n")
			}
			switch test.predicate {
			case "LT":
				fmt.Fprintf(&c, "   uint64_t sign=UINT64_C(1)<<%d; want=(a^sign)<(b^sign);\n", test.source*8-1)
			case "CS":
				c.WriteString("   want=a<b;\n")
			case "EQ":
				c.WriteString("   want=a==b;\n")
			}
		default:
			if test.signed {
				fmt.Fprintf(&c, "   if(value&(UINT64_C(1)<<%d)) value|=~mask(%d);\n", test.source*8-1, test.source)
			}
			if test.high {
				c.WriteString("   want=(want&~UINT64_C(0xff00))|(value<<8);\n")
			} else if test.gp == 4 || test.gp == 8 {
				fmt.Fprintf(&c, "   want=value&mask(%d);\n", test.gp)
			} else {
				fmt.Fprintf(&c, "   want=(want&~mask(%d))|(value&mask(%d));\n", test.gp, test.gp)
			}
		}
		fmt.Fprintf(&c, "   uint64_t got=numerical%d();\n   if(got!=want) { fprintf(stderr,\"round/row=%%u/%d got=%%llx want=%%llx\\n\",round,(unsigned long long)got,(unsigned long long)want); return 1; }\n   printf(\"%%016llx\\n\",(unsigned long long)got);\n  }\n", index, index)
	}
	c.WriteString("  for(unsigned i=32;i<sizeof(mutable);i++) if(mutable[i]!=0xa5) return 2;\n  for(unsigned i=0;i<sizeof(cpu);i++) if(i!=64&&cpu[i]!=0x6b) return 3;\n }\n")
	// A volatile read must still happen after both IR and backend optimization,
	// even when its result is overwritten. The source is inside the valid extent;
	// PROT_NONE, not an out-of-bounds source, creates this independent fault axis.
	if dead >= 0 {
		fmt.Fprintf(&c, " pid_t child=fork();\n if(child<0) return 4;\n if(!child) {\n  if(mprotect(mutable,sizeof(mutable),PROT_NONE)) _exit(5);\n  (void)numerical%d(); _exit(6);\n }\n int status=0;\n if(waitpid(child,&status,0)!=child) return 7;\n if(!WIFSIGNALED(status)||(WTERMSIG(status)!=SIGSEGV&&WTERMSIG(status)!=SIGBUS)) return 8;\n", dead)
	}
	c.WriteString(" return 0;\n}\n")
	return c.String()
}

func numericalRuntimeIR(ir string, pkg GoPackage, cases []numericalExecutionCase, arch string) string {
	word := "i64"
	if arch == "386" {
		word = "i32"
	}
	var wrappers strings.Builder
	wrappers.WriteString(ir)
	for index, test := range cases {
		fmt.Fprintf(&wrappers, "define %s @numerical%d() {\n %%value=call %s %s()\n ret %s %%value\n}\n",
			word, index, word, llvmGlobal(pkg.Path+"."+test.name), word)
	}
	return wrappers.String()
}

func numericalRuntimeExecute(t *testing.T, arch string, execution x86FloatCompareExecution, original bool) {
	t.Helper()
	pkg, file, options, cases := numericalExecutionFixture(t, arch)
	ir := numericalFamilyTranslate(t, "context", pkg, file, options)
	ir = numericalRuntimeIR(ir, pkg, cases, arch)
	main := numericalRuntimeC(cases, arch)
	var expected []byte
	if original {
		expected = numericalOriginalGoRuntime(t, options, cases, execution)
	}
	numericalLLVM22Axes(t, ir, execution.triple, func(t *testing.T, object, _ string) {
		dir := t.TempDir()
		source, executable := filepath.Join(dir, "main.c"), filepath.Join(dir, "mutable")
		if err := os.WriteFile(source, []byte(main), 0600); err != nil {
			t.Fatal(err)
		}
		compiler := append([]string(nil), execution.compiler...)
		compiler = append(compiler, "-O2", object, source, "-o", executable)
		x86FloatCompareCaptureCommand(t, exec.Command(compiler[0], compiler[1:]...))
		output := execution.run(t, executable)
		if original && !bytes.Equal(output, expected) {
			t.Fatalf("actual original Go/LLVM mutable results differ:\nGo=%s\nLLVM=%s", expected, output)
		}
	})
}

func numericalOriginalGoRuntime(t *testing.T, options Options, cases []numericalExecutionCase, execution x86FloatCompareExecution) []byte {
	t.Helper()
	root := filepath.Dir(options.NumericalReads.source.request.Directory)
	dir := t.TempDir()
	var main strings.Builder
	main.WriteString("package main\nimport (\"fmt\"; data \"example.invalid/numericalfixture\"; asm \"example.invalid/numericalfixture/consumer\")\nfunc cpuGet() byte\nfunc cpuSet(value byte)\nfunc main() {\n old:=cpuGet(); defer cpuSet(old)\n")
	fmt.Fprintf(&main, " values:=make([]uint64,%d)\n for round:=0;round<3;round++ {\n  for i:=range data.Mutable { data.Mutable[i]=byte(i*37+round*71) }\n  cpuSet(byte(round&1))\n", len(cases))
	for index, test := range cases {
		fmt.Fprintf(&main, "  values[%d]=uint64(asm.%s())\n", index, test.name)
	}
	main.WriteString("  cpuSet(old)\n  for _,value:=range values { fmt.Printf(\"%016x\\n\",value) }\n  if asm.Sentinel!=0x5a { panic(\"read changed source sentinel\") }\n }\n}\n")
	assembly := "#include \"textflag.h\"\nTEXT ·cpuGet(SB),NOSPLIT,$0-1\nMOVB internal∕cpu·X86+64(SB),AL\nMOVB AL,ret+0(FP)\nRET\nTEXT ·cpuSet(SB),NOSPLIT,$0-1\nMOVB value+0(FP),AL\nMOVB AL,internal∕cpu·X86+64(SB)\nRET\n"
	for name, data := range map[string]string{
		"go.mod":  "module numericaloracle\ngo 1.27\nrequire example.invalid/numericalfixture v0.0.0\nreplace example.invalid/numericalfixture => " + filepath.ToSlash(root) + "\n",
		"main.go": main.String(), "cpu_" + options.Goarch + ".s": assembly,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	request := options.NumericalReads.source.request
	executable := filepath.Join(dir, "original")
	command := exec.Command(request.GoBinary, "build", "-p=1", "-mod=mod", "-o", executable, ".")
	command.Dir, command.Env = dir, request.Environment
	x86FloatCompareCaptureCommand(t, command)
	return execution.run(t, executable)
}
