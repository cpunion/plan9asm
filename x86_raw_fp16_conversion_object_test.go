package plan9asm

import (
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

type x86FP16ConversionObjectEncoding struct {
	member int
	code   []byte
}

// Compare actual LLVM object fields with independent original physical bytes.
// Register allocation may change register numbers, but not opcode, source role,
// vector length, mask, zeroing, broadcast or ER/SAE controls. Scalar LLIG and
// widening register-SAE LL are intentionally ignored by the ISA.
func x86FP16ConversionCheckObject(t *testing.T, llc, triple, ir, level string, want map[string]x86FP16ConversionObjectEncoding) {
	t.Helper()
	disassembler := findLLVM22Tool("llvm-objdump")
	if disassembler == "" {
		t.Fatal("LLVM22 disassembler required")
	}
	dir := t.TempDir()
	input, object := filepath.Join(dir, "conversions.ll"), filepath.Join(dir, "conversions.o")
	if err := os.WriteFile(input, []byte(ir), 0600); err != nil {
		t.Fatal(err)
	}
	x86FloatCompareCaptureCommand(t, exec.Command(llc, level, "-mtriple="+triple, "-filetype=obj", input, "-o", object))
	listing := x86FloatCompareCaptureCommand(t, exec.Command(disassembler, "--disassemble", object))
	sections := regexp.MustCompile(`(?m)^[0-9a-f]+ <_?(conversion[0-9]+)>:\n`).FindAllSubmatchIndex(listing, -1)
	seen := make(map[string]bool)
	for index, section := range sections {
		name := string(listing[section[2]:section[3]])
		expected, found := want[name]
		if !found || seen[name] {
			t.Fatal("unexpected or duplicated conversion object symbol", name)
		}
		end := len(listing)
		if index+1 < len(sections) {
			end = sections[index+1][0]
		}
		conversions := 0
		for _, line := range strings.Split(string(listing[section[1]:end]), "\n") {
			fields := strings.Fields(line)
			if len(fields) < 7 || !strings.HasSuffix(fields[0], ":") {
				continue
			}
			var code []byte
			for _, field := range fields[1:] {
				if len(field) != 2 {
					break
				}
				value, err := hex.DecodeString(field)
				if err != nil {
					break
				}
				code = append(code, value[0])
			}
			if len(code) < 6 || code[0] != 0x62 || code[1]&7 < 5 {
				continue
			}
			conversions++
			original := expected.code
			register := code[5]&0xc0 == 0xc0
			if code[1]&7 != original[1]&7 || code[2]&0x87 != original[2]&0x87 ||
				code[3]&0x97 != original[3]&0x97 || code[4] != original[4] ||
				register != (original[5]&0xc0 == 0xc0) {
				t.Fatalf("%s changed native FP16 opcode, mask, zeroing or source/control role: %x versus %x", name, code, original)
			}
			lengthIgnored := expected.member >= 2 && (original[3]&0x10 == 0 || expected.member%2 == 0) ||
				expected.member == 0 && register && original[3]&0x10 != 0
			if !lengthIgnored && code[3]&0x60 != original[3]&0x60 {
				t.Fatalf("%s changed native source width or rounding: %x versus %x", name, code, original)
			}
		}
		if conversions != 1 {
			t.Fatal("conversion object did not retain exactly one original FP16 operation", name, conversions)
		}
		seen[name] = true
	}
	if len(seen) != len(want) {
		t.Fatal("incomplete native conversion object inventory", len(seen), len(want))
	}
}
