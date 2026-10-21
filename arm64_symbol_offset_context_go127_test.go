//go:build go1.27

package plan9asm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

const arm64SymbolOffsetDefinitions = "#define OFFSET 4\n" +
	"#define const_offsetARM64HasATOMICS 8\n#define machTimebaseInfo_denom 4\n"

// Independent asm7 C_ADDR rows 64/65 cover these seven spellings. Row 68
// additionally admits MOVD/MOVWU address values, not the five narrow aliases.
func arm64SymbolOffsetContextRows() []string {
	var rows []string
	for _, op := range []string{"MOVB", "MOVBU", "MOVH", "MOVHU", "MOVW", "MOVWU", "MOVD"} {
		rows = append(rows,
			op+" ·memory+OFFSET(SB),R2", op+" R2,·memory-OFFSET(SB)",
			op+" ZR,·memory+OFFSET(SB)", op+" $0,·memory+OFFSET(SB)",
			op+" ·memory+OFFSET(SB),ZR", op+" ·memory+OFFSET(SB),R27")
	}
	for _, expression := range []string{
		"+(OFFSET+4)", "+(OFFSET*2)", "+(OFFSET/2)", "+(OFFSET%3)",
		"+(OFFSET<<1)", "+(OFFSET>>1)", "+(OFFSET&3)", "+(OFFSET|3)",
		"+(OFFSET^3)", "+(~OFFSET)", "+(OFFSET+'x')",
		"+(OFFSET/(OFFSET+1))", "+(1/(OFFSET-1))", "+(OFFSET<<OFFSET)",
	} {
		rows = append(rows, "MOVD ·memory"+expression+"(SB),R2")
	}
	rows = append(rows,
		"MOVD $·memory+OFFSET(SB),R2", "MOVWU $·memory-OFFSET(SB),R2",
		"MOVD ·memory<>+OFFSET(SB),R2", "MOVD R2,·memory<>-OFFSET(SB)",
		"MOVBU internal∕cpu·ARM64+const_offsetARM64HasATOMICS(SB),R4",
		"MOVD $timebase<>+machTimebaseInfo_denom(SB),R21",
		"MOVD $timebase<>+machTimebaseInfo_denom(SB),R22")
	return rows
}

func TestARM64SymbolOffsetNeedsGeneratedConstantContext(t *testing.T) {
	rows := arm64SymbolOffsetContextRows()
	if len(rows) != 63 {
		t.Fatal("incomplete independent seven-member memory/address inventory", len(rows))
	}
	// Only this original-Go prior receives generated definitions. The public
	// probes and translations retain the exact unresolved source below.
	requireARM64GoAssemblerResult(t, arm64SymbolOffsetDefinitions+
		"TEXT symbolic(SB),$0-0\n"+strings.Join(rows, "\n")+"\nRET\n", true)
	for _, row := range rows {
		t.Run(row, func(t *testing.T) {
			source := "TEXT symbolic(SB),$0-0\n" + row + "\nRET\n"
			file, err := Parse(ArchARM64, source)
			if err != nil {
				t.Fatal(err)
			}
			instruction := file.Funcs[0].Instrs[1]
			original := append([]Operand(nil), instruction.Args...)
			for _, err := range []error{
				ProbeInstruction(ArchARM64, "arm64", instruction),
				ProbeInstructionSequence(ArchARM64, "arm64", []Instr{instruction}),
			} {
				if !errors.Is(err, ErrProbeNeedsContext) {
					t.Errorf("valid unresolved GP symbol addend lost its source context: %v", err)
				}
			}
			options := Options{Goarch: "arm64", TargetTriple: "aarch64-unknown-linux-gnu",
				Sigs: map[string]FuncSig{"symbolic": {Name: "symbolic", Ret: Void}}}
			ir, err := Translate(file, options)
			if !errors.Is(err, ErrProbeNeedsContext) || ir != "" {
				t.Errorf("unresolved GP symbol acquired textual output or lost Context: IR=%q error=%v", ir, err)
			}
			module, err := TranslateModule(file, options)
			empty := module.IsNil()
			if !empty {
				module.Dispose()
			}
			if !errors.Is(err, ErrProbeNeedsContext) || !empty {
				t.Errorf("unresolved GP symbol acquired a module or lost Context: empty=%v error=%v", empty, err)
			}
			if !reflect.DeepEqual(file.Funcs[0].Instrs[1].Args, original) ||
				file.Funcs[0].Instrs[1].Raw != row {
				t.Fatal("probe/translation changed the original unresolved operand")
			}
		})
	}
}

func TestARM64SymbolOffsetContextDoesNotHideInvalidForms(t *testing.T) {
	var rows []string
	for _, op := range []string{"MOVB", "MOVBU", "MOVH", "MOVHU", "MOVW", "MOVWU", "MOVD"} {
		rows = append(rows,
			op+".P ·memory+OFFSET(SB),R2", op+".W R2,·memory+OFFSET(SB)",
			op+" ·memory+OFFSET(SB),SP", op+" RSP,·memory+OFFSET(SB)",
			op+" ·memory+OFFSET(SB),W2", op+" R27,·memory+OFFSET(SB)",
			op+" ·memory+OFFSET(SB),·other(SB)", op+" $0.0,·memory+OFFSET(SB)",
			op+" ·memory+OFFSET(SB)", op+" ·memory+OFFSET(SB),R2,R3")
		if op != "MOVD" && op != "MOVWU" {
			rows = append(rows, op+" $·memory+OFFSET(SB),R2")
		}
	}
	for _, op := range []string{"MOVD", "MOVWU"} {
		rows = append(rows,
			op+".P $·memory+OFFSET(SB),R2", op+".W $·memory+OFFSET(SB),R2",
			op+" $·memory+OFFSET(SB),SP", op+" $·memory+OFFSET(SB),W2")
	}
	for _, expression := range []string{
		"+(OFFSET+)", "+(OFFSET/0)", "+(OFFSET%0)", "+(OFFSET<<-1)",
		"+(OFFSET<<1*2)", "+(OFFSET<<1<<2)", "+((-1)/OFFSET)",
		"+((1<<63)>>OFFSET)", "+1.5", "+(OFFSET==4)",
	} {
		rows = append(rows, "MOVD ·memory"+expression+"(SB),R2")
	}
	rows = append(rows,
		"MOVD ·memory+OFFSET(XX),R2", "MOVD ·memory+OFFSET(SB),F2",
		"MOVD F2,·memory+OFFSET(SB)", "MOVD $1,·memory+OFFSET(SB)")
	if len(rows) != 97 {
		t.Fatal("incomplete independent invalid GP symbol inventory", len(rows))
	}
	for _, row := range rows {
		t.Run(row, func(t *testing.T) {
			source := "TEXT invalidsymbol(SB),$0-0\n" + row + "\nRET\n"
			requireARM64GoAssemblerResult(t, arm64SymbolOffsetDefinitions+source, false)
			file, err := Parse(ArchARM64, source)
			if err != nil {
				return
			}
			instruction := file.Funcs[0].Instrs[1]
			for _, err := range []error{
				ProbeInstruction(ArchARM64, "arm64", instruction),
				ProbeInstructionSequence(ArchARM64, "arm64", []Instr{instruction}),
			} {
				if err == nil || errors.Is(err, ErrProbeNeedsContext) {
					t.Errorf("invalid complete GP grammar was hidden as Context: %v", err)
				}
			}
		})
	}
}

func TestARM64SymbolOffsetConcreteFormsStayConcrete(t *testing.T) {
	var rows []string
	for _, op := range []string{"MOVB", "MOVBU", "MOVH", "MOVHU", "MOVW", "MOVWU", "MOVD"} {
		rows = append(rows, op+" ·memory+4(SB),R2", op+" R2,·memory-4(SB)")
	}
	rows = append(rows, "MOVD $·memory+4(SB),R2", "MOVWU $·memory-4(SB),R2")
	requireARM64GoAssemblerResult(t, "TEXT concrete(SB),$0-0\n"+
		strings.Join(rows, "\n")+"\nRET\n", true)
	for _, row := range rows {
		t.Run(row, func(t *testing.T) {
			file, err := Parse(ArchARM64, fmt.Sprintf("TEXT concrete(SB),$0-0\n%s\nRET\n", row))
			if err != nil {
				t.Fatal(err)
			}
			if err := ProbeInstruction(ArchARM64, "arm64", file.Funcs[0].Instrs[1]); err != nil {
				t.Fatalf("a concrete GP symbol displacement became unresolved: %v", err)
			}
		})
	}
}
