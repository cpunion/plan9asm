package plan9asm

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func x86LegacyLogicalOriginalBytes(t *testing.T, arch, name string) []byte {
	t.Helper()
	listing := x86InsertExtractGoListing(t, arch, "TEXT original(SB),4,$0-0\n"+name+" (AX),X0\nRET\n")
	var code []byte
	for _, line := range strings.Split(string(listing), "\n") {
		code = append(code, x86InsertExtractListedBytes(line)...)
	}
	// Independent Intel encoding, not obtained from the translator's spec.
	opcode := map[string]byte{
		"ANDPS": 0x54, "ANDPD": 0x54, "ANDNPS": 0x55, "ANDNPD": 0x55,
		"ORPS": 0x56, "ORPD": 0x56, "XORPS": 0x57, "XORPD": 0x57,
		"PAND": 0xdb, "PANDN": 0xdf, "POR": 0xeb, "PXOR": 0xef,
	}[name]
	want := []byte{0x0f, opcode, 0x00, 0xc3}
	if strings.HasSuffix(name, "PD") || strings.HasPrefix(name, "P") {
		want = append([]byte{0x66}, want...)
	}
	if !bytes.Equal(code, want) {
		t.Fatalf("%s/%s actual Go memory opcode=%x independent encoding=%x", arch, name, code, want)
	}
	t.Logf("%s/%s original complete opcode+RET=%x", arch, name, code)
	return append([]byte(nil), code[:len(code)-1]...)
}

func TestX86LegacyLogicalMemoryOriginalGoBytes(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		for _, name := range x86LegacyLogicalMemoryNames {
			t.Run(arch+"/"+name, func(t *testing.T) { x86LegacyLogicalOriginalBytes(t, arch, name) })
		}
	}
}

// Independent legacy namespace: these twelve Go names all have an m128
// source whose original SSE instruction requires aligned native storage.
var x86LegacyLogicalMemoryNames = []string{
	"ANDPS", "ANDPD", "ANDNPS", "ANDNPD", "ORPS", "ORPD",
	"XORPS", "XORPD", "PAND", "PANDN", "POR", "PXOR",
}

var x86LegacyLogicalMemoryTargets = []struct{ arch, triple string }{
	{"amd64", "x86_64-apple-darwin"},
	{"amd64", "x86_64-unknown-linux-gnu"},
	{"amd64", "x86_64-pc-windows-msvc"},
	{"386", "i386-unknown-linux-gnu"},
	{"386", "i686-pc-windows-msvc"},
}

func TestX86LegacyLogicalMemoryFPNeedsNativeBacking(t *testing.T) {
	for _, target := range x86LegacyLogicalMemoryTargets {
		for _, name := range x86LegacyLogicalMemoryNames {
			source := "TEXT ·view(SB),4,$0-16\nMOVOU a+0(FP),X0\n" + name + " a+0(FP),X0\nRET\n"
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(target.triple+"/"+name+"/"+api, func(t *testing.T) {
					_, err := x86ScalarFPBitViewAPI(t, source, "func view(a,b uint64)", target.arch, target.triple, api)
					if !errors.Is(err, ErrProbeNeedsContext) || !strings.Contains(err.Error(), "original native storage") {
						t.Fatalf("independent FP bytes fabricated native m128 backing/alignment: %v", err)
					}
				})
			}
		}
	}
}

func TestX86LegacyLogicalMemoryVirtualStackNeedsCertificate(t *testing.T) {
	for _, target := range x86LegacyLogicalMemoryTargets {
		for _, name := range x86LegacyLogicalMemoryNames {
			for _, address := range []string{"16(SP)", "16(BP)", "16(AX)(BP*1)"} {
				source := "TEXT ·view(SB),4,$0-0\nPXOR X0,X0\nMOVL $0,AX\n" + name + " " + address + ",X0\nRET\n"
				x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
				for _, api := range []string{"text", "module", "Go"} {
					t.Run(target.triple+"/"+name+"/"+address+"/"+api, func(t *testing.T) {
						_, err := x86ScalarFPBitViewAPI(t, source, "func view()", target.arch, target.triple, api)
						if !errors.Is(err, ErrProbeNeedsContext) {
							t.Fatalf("uncertified virtual stack became original m128 storage: %v", err)
						}
					})
				}
			}
		}
	}
}

func TestX86LegacyLogicalMemoryNativeOpcodeObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, target := range x86LegacyLogicalMemoryTargets {
		var source, declaration strings.Builder
		for _, name := range x86LegacyLogicalMemoryNames {
			move, size := "MOVQ", 8
			if target.arch == "386" {
				move, size = "MOVL", 4
			}
			fn := strings.ToLower(name)
			fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-%d\n%s p+0(FP),AX\nPXOR X0,X0\n%s (AX),X0\nRET\n", fn, size, move, name)
			fmt.Fprintf(&declaration, "func %s(p *byte)\n", fn)
		}
		x86LEAActualTargetGoObject(t, target.arch, target.triple, source.String())
		for _, api := range []string{"text", "module", "Go"} {
			t.Run(target.triple+"/"+api, func(t *testing.T) {
				ir, err := x86ScalarFPBitViewAPI(t, source.String(), declaration.String(), target.arch, target.triple, api)
				if err != nil {
					t.Fatal(err)
				}
				for _, name := range x86LegacyLogicalMemoryNames {
					want := "asm sideeffect \"" + strings.ToLower(name) + " ($2), $0\", \"=x,0,r,~{memory}\""
					if !strings.Contains(ir, want) {
						t.Fatalf("dead result lost original single native %s memory effect", name)
					}
				}
				compileLLVMToObject(t, llc, target.triple, "logical-native.ll", "logical-native.o", ir)
			})
		}
	}
}
