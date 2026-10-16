package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

func TestX86RawStaticRegisterReturnCompleteMoveWidths(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		for _, row := range []struct {
			op, register string
			typ          LLVMType
		}{{"MOVB", "AL", I8}, {"MOVW", "AX", I16}, {"MOVL", "AX", I32}, {"MOVQ", "AX", I64}} {
			if row.typ == I64 && target.arch == "386" {
				continue // This direct scalar role has no physical mode32 qword GP transport.
			}
			for _, encoding := range []string{"BYTE $0xc3", "BYTE $0xc2\nWORD $0"} {
				t.Run(target.triple+"/"+row.op+"/"+encoding, func(t *testing.T) {
					source := fmt.Sprintf("GLOBL saved<>(SB),8,$32\nTEXT Raw(SB),4,$0-0\n%s $17,%s\n%s %s,saved<>+4(SB)\n%s saved<>+4(SB),%s\n%s\nRET\n",
						row.op, row.register, row.op, row.register, row.op, row.register, encoding)
					requireX86GoAssemblerResult(t, target.arch, source, true)
					options := Options{Goarch: target.arch, TargetTriple: target.triple,
						Sigs: map[string]FuncSig{"Raw": {Name: "Raw", Ret: row.typ}}}
					x86RawReturnPointerAPIs(t, source, options, false)
				})
			}
		}
	}
}

func TestX86RawStaticRegisterReturnDoesNotInventSourceBytes(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		for _, row := range []struct {
			op, register string
			typ          LLVMType
		}{{"MOVB", "AL", I8}, {"MOVW", "AX", I16}, {"MOVL", "AX", I32}, {"MOVQ", "AX", I64}} {
			if row.typ == I64 && target.arch == "386" {
				continue
			}
			body := fmt.Sprintf("%s $17,%s\n%s %s,saved<>+4(SB)\n%s saved<>+4(SB),%s\n",
				row.op, row.register, row.op, row.register, row.op, row.register)
			for _, mutation := range []struct{ name, body string }{
				{"unknown-store-input", strings.Replace(body, row.op+" $17,"+row.register, "NOP", 1)},
				{"unbound-static-read", strings.Replace(body, "saved<>+4(SB),"+row.register, "external+4(SB),"+row.register, 1)},
				{"outside-static-extent", strings.Replace(body, "saved<>+4(SB),"+row.register, "saved<>+32(SB),"+row.register, 1)},
				{"unknown-final-result", body + "MOVL DX,AX\n"},
			} {
				t.Run(target.triple+"/"+row.op+"/"+mutation.name, func(t *testing.T) {
					source := "GLOBL saved<>(SB),8,$32\nTEXT Raw(SB),4,$0-0\n" + mutation.body + "BYTE $0xc3\nRET\n"
					requireX86GoAssemblerResult(t, target.arch, source, true)
					options := Options{Goarch: target.arch, TargetTriple: target.triple,
						Sigs: map[string]FuncSig{"Raw": {Name: "Raw", Ret: row.typ}}}
					x86RawReturnPointerAPIs(t, source, options, true)
				})
			}
		}
	}
}
