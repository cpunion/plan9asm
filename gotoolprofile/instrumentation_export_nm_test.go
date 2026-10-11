package gotoolprofile

import (
	"strings"
	"testing"
)

func TestInstrumentationExportNMUndefinedSymbolGrammarIsComplete(t *testing.T) {
	defined := "@export(mode_race.o):\t 42 T " + instrumentationExportModule + ".marker\n"
	for _, undefined := range []string{"U", "U unresolved.symbol"} {
		output := "@export(mode_race.o):\t " + undefined + "\n" + defined
		if err := validateInstrumentationExportNM(output, InstrumentationRace); err != nil {
			t.Fatalf("rejected actual Go nm undefined grammar %q: %v", undefined, err)
		}
	}
	for _, body := range []string{
		"U missing unexpected warning", "U missing error", "U two symbols", "U missing\twarning",
		"U missing -toolchain-error", "missing U", "warning", "", "42 U missing",
	} {
		t.Run(strings.ReplaceAll(body, "\t", "_"), func(t *testing.T) {
			output := "@export(mode_race.o):\t " + body + "\n" + defined
			if err := validateInstrumentationExportNM(output, InstrumentationRace); err == nil {
				t.Fatalf("accepted unknown nm undefined-symbol text %q", body)
			}
		})
	}
}
