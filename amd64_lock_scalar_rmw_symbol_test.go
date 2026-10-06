package plan9asm

import (
	"strings"
	"testing"
)

func TestX86LockScalarRMWAddressLiteralIsNotMemory(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		for _, row := range x86LockScalarOracleForms(arch) {
			body := strings.Replace(row.text, "1(SI)", "$·cell(SB)", 1)
			t.Run(arch+"/"+body, func(t *testing.T) {
				source := "GLOBL ·cell(SB),8,$16\n" +
					x86NamedLockOracleSource(arch, "CLC\nLOCK\n"+body)
				// An emitted address literal is Go TYPE_ADDR, not writable
				// memory. This is actual assembler rejection, not execution.
				goAssemblerOracleOutput(t, arch, source, "", false)
				file, err := Parse(ArchAMD64, source)
				if err != nil {
					t.Fatal(err)
				}
				ins := file.Funcs[0].Instrs[len(file.Funcs[0].Instrs)-2]
				form, recognized, err := x86LockScalarRMWProperties(ins.Op, ins, arch)
				if !recognized || err == nil {
					t.Error("typed memory grammar accepted an address literal", form, err)
				}
				for _, api := range []string{"text", "module", "Go"} {
					ir, err := x86OrderingHintTranslateAPI(t, api, source,
						"package ordering\nfunc locked(p *[16]byte)\n", arch, "")
					if err == nil || ir != "" {
						t.Error("address literal became a locked memory object", api, err)
					}
				}
			})
		}
	}
}
