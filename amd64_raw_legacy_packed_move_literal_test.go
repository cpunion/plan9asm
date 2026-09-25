package plan9asm

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

var x86RawLegacyPackedMoveForms = []struct {
	name   Op
	prefix byte
	opcode byte
}{
	{"MOVUPS", 0, 0x10},
	{"MOVAPS", 0, 0x28},
	{"MOVUPD", 0x66, 0x10},
	{"MOVAPD", 0x66, 0x28},
	{"MOVO", 0x66, 0x6f},
	{"MOVOU", 0xf3, 0x6f},
}

func x86RawLegacyPackedMoveConstant(prefix, opcode byte) []byte {
	var code []byte
	if prefix != 0 {
		code = append(code, prefix)
	}
	code = append(code, 0x0f, opcode, 0x05, 1, 0, 0, 0, 0xc3)
	for i := 0; i < 16; i++ {
		code = append(code, byte(i*31+7))
	}
	return code
}

func TestDecodeX86RawLegacyPackedMoveConstantCompleteFamily(t *testing.T) {
	for _, form := range x86RawLegacyPackedMoveForms {
		t.Run(string(form.name), func(t *testing.T) {
			code := x86RawLegacyPackedMoveConstant(form.prefix, form.opcode)
			decoded, err := decodeX86RawDirectiveGroup(code, 64, 0, string(form.name), map[string]bool{})
			if err != nil {
				t.Fatal(err)
			}
			if len(decoded) != 2 || decoded[0].Op != form.name ||
				decoded[0].Args[0].Kind != OpSym || !decoded[0].x86RIPLiteral ||
				decoded[1].Op != OpRET {
				t.Fatalf("decoded %x as %#v, want %s source-local data", code, decoded, form.name)
			}
		})
	}
}

func TestDecodeX86RawLegacyPackedMoveConstantRejectsUnsafeSources(t *testing.T) {
	valid := x86RawLegacyPackedMoveConstant(0, 0x10)
	for name, mutate := range map[string]func([]byte) []byte{
		"truncated": func(code []byte) []byte { return code[:len(code)-1] },
		"overlap": func(code []byte) []byte {
			code[3] = 0
			return code
		},
		"outside": func(code []byte) []byte {
			code[3] = 127
			return code
		},
		"segment": func(code []byte) []byte { return append([]byte{0x64}, code...) },
		"address": func(code []byte) []byte { return append([]byte{0x67}, code...) },
	} {
		t.Run(name, func(t *testing.T) {
			code := mutate(append([]byte(nil), valid...))
			if _, err := decodeX86RawDirectiveGroup(code, 64, 0, name, map[string]bool{}); err == nil {
				t.Fatalf("accepted unsafe RIP source %x", code)
			}
		})
	}
}

func TestTranslateX86RawLegacyPackedMoveConstantObjects(t *testing.T) {
	var source strings.Builder
	sigs := make(map[string]FuncSig)
	var pools [][]byte
	for index, form := range x86RawLegacyPackedMoveForms {
		code := x86RawLegacyPackedMoveConstant(form.prefix, form.opcode)
		name := fmt.Sprintf("legacyPackedMoveConstant%d", index)
		fmt.Fprintf(&source, "TEXT %s(SB),$0-0\n", name)
		for _, value := range code {
			fmt.Fprintf(&source, "\tBYTE $%#02x\n", value)
		}
		sigs[name] = FuncSig{Name: name, Ret: Void}
		pools = append(pools, code[len(code)-16:])
	}
	requireX86GoAssemblerResult(t, "amd64", source.String(), true)
	file, err := Parse(ArchAMD64, source.String())
	if err != nil {
		t.Fatal(err)
	}
	normalized, err := normalizeX86RawFile(file, "amd64")
	if err != nil {
		t.Fatal(err)
	}
	if len(normalized.Data) != len(pools) {
		t.Fatalf("materialized %d constants, want %d", len(normalized.Data), len(pools))
	}
	for index, datum := range normalized.Data {
		if !bytes.Equal(datum.Payload, pools[index]) {
			t.Fatalf("constant %d = %x, want %x", index, datum.Payload, pools[index])
		}
	}
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc not found")
	}
	for _, triple := range []string{
		"x86_64-apple-darwin", "x86_64-unknown-linux-gnu", "x86_64-pc-windows-msvc",
	} {
		t.Run(triple, func(t *testing.T) {
			ir, err := Translate(file, Options{Goarch: "amd64", TargetTriple: triple, Sigs: sigs})
			if err != nil {
				t.Fatal(err)
			}
			compileLLVMToObject(t, llc, triple, "legacy-packed-move-literal.ll", "legacy-packed-move-literal.o", ir)
		})
	}
}
