package plan9asm

import (
	"strings"
	"testing"
)

func TestTranslateARM64SVEPredicateSelectCompleteGo127Family(t *testing.T) {
	const source = "TEXT svepredicateselectforms(SB),$0-0\n\tPSEL P15.B, P14.B, P13, P12.B\n\tRET\n"
	requireARM64SVEGoAssemblerResult(t, source, true)
	file, err := Parse(ArchARM64, source)
	if err != nil {
		t.Fatal(err)
	}
	for _, triple := range []string{"aarch64-apple-darwin", "aarch64-unknown-linux-gnu", "aarch64-pc-windows-msvc"} {
		t.Run(triple, func(t *testing.T) {
			ll, err := Translate(file, Options{TargetTriple: triple, Goarch: "arm64", Sigs: map[string]FuncSig{"svepredicateselectforms": {Name: "svepredicateselectforms", Ret: Void}}})
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{`"target-features"="+sve"`, "select <vscale x 16 x i1>"} {
				if !strings.Contains(ll, want) {
					t.Fatalf("%s SVE predicate select lowering omitted %q:\n%s", triple, want, ll)
				}
			}
			llc := findLLVM22Tool("llc")
			if llc == "" {
				t.Fatal("LLVM 22 llc not found")
			}
			compileLLVMToObject(t, llc, triple, "arm64-sve-predicate-select.ll", "arm64-sve-predicate-select.o", ll)
		})
	}
}

func TestTranslateARM64SVEPredicateSelectRejectsFormsOutsideGo127Table(t *testing.T) {
	for _, instruction := range []string{
		"PSEL P1.H, P2.H, P3, P4.H",
		"PSEL P1.B, P2.B, P3.Z, P4.B",
		"PSEL P1.B, P2.B, P16, P4.B",
		"PSEL P1.B, P2.B, P3, P4.H",
		"PSEL P1.B, P2.B, P3",
	} {
		t.Run(strings.NewReplacer(" ", "_", ".", "_").Replace(instruction), func(t *testing.T) {
			source := "TEXT badsvepredicateselect(SB),$0-0\n\t" + instruction + "\n\tRET\n"
			requireARM64SVEGoAssemblerResult(t, source, false)
			file, err := Parse(ArchARM64, source)
			if err != nil {
				return
			}
			if _, err := Translate(file, Options{TargetTriple: "aarch64-unknown-linux-gnu", Goarch: "arm64", Sigs: map[string]FuncSig{"badsvepredicateselect": {Name: "badsvepredicateselect", Ret: Void}}}); err == nil {
				t.Fatalf("Translate accepted %q outside Go 1.27's SVE predicate select forms", instruction)
			}
		})
	}
}
