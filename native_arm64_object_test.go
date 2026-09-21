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
	if err := os.WriteFile(s, []byte(asm), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(c, []byte(`extern void *entry __asm("_probe.entry");
extern void *mixedEntry __asm("_probe.mixedEntry");
unsigned long mixed(unsigned long x, double y) { return x + (unsigned long)y; }
int main(void) {
 if (((unsigned long (*)(const char *))entry)("native ABI") != 10) return 1;
 return ((unsigned long (*)(unsigned long, unsigned long))mixedEntry)(5, 0x4000000000000000UL) != 7;
}
`), 0600); err != nil {
		t.Fatal(err)
	}
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
		{"TEXT raw<>(SB), NOSPLIT, $0-0\r\n JMP imported(SB)\r\n", true},
		{"TEXT raw<>(SB), NOSPLIT, $0-0\n RET\nTEXT raw2<>(SB),NOSPLIT,$0\nRET\n", true},
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

// Exercise the parser's rejection contract using valid assembler output with
// one damaged field at a time, so each diagnostic identifies the broken field.
func TestNativeARM64MalformedFields(t *testing.T) {
	obj := nativeObjectFixture(t)
	funcs := map[string]bool{"callback": true, "mixedtramp": true}
	imports := map[string]string{"imported_strlen": "strlen", "imported_mixed": "mixed"}
	cases := []struct {
		name, want string
		damage     func(*nativeObject)
	}{
		{"header", "invalid native object header", func(r *nativeObject) { binary.LittleEndian.PutUint32(r.b[20:], nativeHeaderSize-1) }},
		{"symbol table", "invalid native symbol table", func(r *nativeObject) {
			binary.LittleEndian.PutUint32(r.b[20+nativeBlkNonpkgref*4:], r.blocks[nativeBlkNonpkgref]-1)
		}},
		{"symbol name", "invalid native symbol name", func(r *nativeObject) { binary.LittleEndian.PutUint32(r.block(nativeBlkSymdef), ^uint32(0)) }},
		{"symbol size", "native symbol too large", func(r *nativeObject) {
			binary.LittleEndian.PutUint32(r.block(nativeBlkSymdef)[nativeSymSizeOffset:], 1<<30)
		}},
		{"symbol payload", "oversized native symbol", func(r *nativeObject) {
			binary.LittleEndian.PutUint32(r.block(nativeBlkSymdef)[nativeSymSizeOffset:], 0)
		}},
		{"alignment", "invalid native alignment", func(r *nativeObject) {
			binary.LittleEndian.PutUint32(nativeFixtureSym(r, "probe.entry")[nativeSymAlign:], 3)
		}},
		{"indices length", "invalid native object indices", func(r *nativeObject) {
			binary.LittleEndian.PutUint32(r.b[20+nativeBlkRelocIdx*4:], r.blocks[nativeBlkRelocIdx]+1)
		}},
		{"index range", "invalid native object index", func(r *nativeObject) { binary.LittleEndian.PutUint32(r.block(nativeBlkDataIdx), ^uint32(0)) }},
		{"split function", "not NOSPLIT", func(r *nativeObject) { nativeFixtureSym(r, "callback")[nativeSymFlags] &^= nativeNoSplit }},
		{"relocation bounds", "invalid native relocation", func(r *nativeObject) { binary.LittleEndian.PutUint32(nativeCallbackReloc(r), ^uint32(0)) }},
		{"Go ABI reference", "Go package index", func(r *nativeObject) { binary.LittleEndian.PutUint32(nativeCallbackReloc(r)[nativeRelocPkg:], 1) }},
		{"reference range", "invalid native symbol reference", func(r *nativeObject) {
			binary.LittleEndian.PutUint32(nativeCallbackReloc(r)[nativeRelocSym:], ^uint32(0))
		}},
		{"branch addend", "unsupported native branch addend", func(r *nativeObject) { binary.LittleEndian.PutUint64(nativeCallbackReloc(r)[nativeRelocAddend:], 8) }},
		{"branch opcode", "invalid native branch", func(r *nativeObject) {
			for i, s := range r.syms {
				if s.name == "callback" {
					binary.LittleEndian.PutUint32(r.payload(i)[binary.LittleEndian.Uint32(nativeCallbackReloc(r)):], 0)
				}
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			damaged := append([]byte(nil), obj...)
			r, err := readNativeObject(damaged)
			if err != nil {
				t.Fatal(err)
			}
			tc.damage(r)
			_, _, err = TranslateNativeARM64Object(damaged, funcs, imports, "probe")
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want %q", err, tc.want)
			}
		})
	}
}

func nativeFixtureSym(r *nativeObject, name string) []byte {
	for i, s := range r.syms {
		if s.name == name {
			for block, count := range r.counts {
				if i < count {
					return r.block(nativeBlkSymdef + block)[i*nativeSymSize:]
				}
				i -= count
			}
		}
	}
	panic("fixture missing symbol " + name)
}

func nativeCallbackReloc(r *nativeObject) []byte {
	for i, s := range r.syms {
		if s.name == "callback" {
			first := binary.LittleEndian.Uint32(r.block(nativeBlkRelocIdx)[i*4:])
			return r.block(nativeBlkReloc)[first*nativeRelocSize:]
		}
	}
	panic("fixture missing callback")
}

func TestNativeARM64SymbolReferences(t *testing.T) {
	r := &nativeObject{counts: [5]int{2, 3, 4, 5, 6}}
	for _, tc := range []struct {
		pkg  uint32
		sym  uint32
		want int
	}{
		{nativePkgSelf, 1, 1}, {nativePkgHashed64, 2, 4}, {nativePkgHashed, 3, 8}, {nativePkgNone, 10, 19},
	} {
		got, err := r.target(tc.pkg, tc.sym)
		if err != nil || got != tc.want {
			t.Fatalf("target(%x,%d) = %d, %v; want %d", tc.pkg, tc.sym, got, err, tc.want)
		}
	}
}
