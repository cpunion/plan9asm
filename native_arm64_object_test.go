package plan9asm

import (
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func nativeObjectFixture(t *testing.T) []byte {
	t.Helper()
	dir := t.TempDir()
	src := filepath.Join(dir, "native.s")
	out := filepath.Join(dir, "native.o")
	code := `#include "textflag.h"
TEXT callback<>(SB), NOSPLIT|NOFRAME, $0
 SUB $16, RSP
 MOVD R30, (RSP)
 BL imported_strlen(SB)
 MOVD (RSP), R30
 ADD $16, RSP
 RET
GLOBL ·entry(SB), RODATA, $8
DATA ·entry(SB)/8, $callback<>(SB)
TEXT mixedtramp<>(SB), NOSPLIT, $0-0
 FMOVD R1, F0
 JMP imported_mixed(SB)
GLOBL ·mixedEntry(SB), RODATA, $8
DATA ·mixedEntry(SB)/8, $mixedtramp<>(SB)
`
	if err := os.WriteFile(src, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "tool", "asm", "-p", "probe", "-I", filepath.Join(runtime.GOROOT(), "pkg", "include"), "-o", out, src)
	cmd.Env = append(os.Environ(), "GOOS=darwin", "GOARCH=arm64")
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("asm: %v\n%s", err, b)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestNativeARM64Object(t *testing.T) {
	obj := nativeObjectFixture(t)
	funcs := map[string]bool{"callback": true, "mixedtramp": true}
	imports := map[string]string{"imported_strlen": "strlen", "imported_mixed": "mixed"}
	asm, data, err := TranslateNativeARM64Object(obj, funcs, imports, "probe")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 2 || data[0].Name != "probe.entry" || data[0].Size != 8 {
		t.Fatalf("data = %v", data)
	}
	if !strings.Contains(asm, `bl "_strlen"`) || !strings.Contains(asm, `.quad "Lllgo_native_`) {
		t.Fatalf("missing relocations:\n%s", asm)
	}
	if _, _, err := TranslateNativeARM64Object(obj, funcs, nil, "probe"); err == nil {
		t.Fatal("accepted undeclared foreign call")
	}
	if _, _, err := TranslateNativeARM64Object(obj, map[string]bool{"missing": true}, imports, "probe"); err == nil {
		t.Fatal("accepted missing function")
	}
	for n := 0; n < len(obj); n++ {
		if _, _, err := TranslateNativeARM64Object(obj[:n], funcs, imports, "probe"); err == nil {
			t.Fatalf("accepted truncated object at %d", n)
		}
	}
	t.Run("reject unsupported relocation", func(t *testing.T) {
		damaged := append([]byte(nil), obj...)
		r, err := readNativeObject(damaged)
		if err != nil {
			t.Fatal(err)
		}
		for i, sym := range r.syms[:r.ndef] {
			if sym.name != "callback" {
				continue
			}
			first := binary.LittleEndian.Uint32(r.block(11)[i*4:])
			rel := r.block(14)[first*23:]
			binary.LittleEndian.PutUint16(rel[5:], 0xffff)
			break
		}
		if _, _, err := TranslateNativeARM64Object(damaged, funcs, imports, "probe"); err == nil {
			t.Fatal("accepted unknown relocation")
		}
	})
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		return
	}
	dir := t.TempDir()
	s := filepath.Join(dir, "native.s")
	c := filepath.Join(dir, "main.c")
	exe := filepath.Join(dir, "probe")
	os.WriteFile(s, []byte(asm), 0600)
	os.WriteFile(c, []byte(`extern void *entry __asm("_probe.entry");
extern void *mixedEntry __asm("_probe.mixedEntry");
unsigned long mixed(unsigned long x, double y) { return x + (unsigned long)y; }
int main(void) {
 if (((unsigned long (*)(const char *))entry)("native ABI") != 10) return 1;
 return ((unsigned long (*)(unsigned long, unsigned long))mixedEntry)(5, 0x4000000000000000UL) != 7;
}
`), 0600)
	cmd := exec.Command("clang", s, c, "-o", exe)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("link: %v\n%s", err, b)
	}
	if b, err := exec.Command(exe).CombinedOutput(); err != nil {
		t.Fatalf("callback: %v\n%s", err, b)
	}
}

func TestForeignARM64Selection(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want bool
	}{
		{"TEXT raw<>(SB), NOSPLIT, $0-0\n JMP imported(SB)\n", true},
		{"TEXT raw<>(SB), NOSPLIT|NOFRAME, $0\n BL imported(SB)\n RET\n", true},
		{"TEXT ·declared(SB), NOSPLIT, $0-0\n RET\n", false},
		{"TEXT raw<>(SB), 0, $0-0\n RET\n", false},
		{"TEXT raw<>(SB), NOSPLIT, $8-0\n RET\n", false},
		{"TEXT raw<>(SB), NOSPLIT, $0-0\n BL imported(SB)\n RET\n", false},
		{"TEXT raw<>(SB), NOSPLIT, $0-0\n RET\nTEXT ·goFunc(SB),NOSPLIT,$0\nRET\n", false},
	} {
		if got := len(ForeignARM64Functions([]byte(tc.src))) != 0; got != tc.want {
			t.Errorf("selection(%q) = %v", tc.src, got)
		}
	}
}
