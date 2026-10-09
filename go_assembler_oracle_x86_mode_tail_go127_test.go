//go:build go1.27

package plan9asm

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xgo-dev/plan9asm/internal/goasmcmd"
	"github.com/xgo-dev/plan9asm/internal/goasmoracle"
)

func TestGoAssemblerOracleActualX86Mode32ClassTails(t *testing.T) {
	for _, test := range []struct{ name, instruction string }{
		{"alias-high-destination", "PADDD X0,X8"},
		{"alias-high-source", "PADDD X8,X0"},
		{"alias-high-address", "PADDD 8(R9),X0"},
		{"scalar-single", "MOVSS $0.0,X8"},
		{"scalar-double", "MOVSD $0.0,X8"},
		{"vector-single", "VMOVSS $0.0,X8"},
		{"vector-double", "VMOVSD $0.0,X8"},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := x86DiagnosticFixture(t, test.instruction)
			object := filepath.Join(filepath.Dir(source), "forms.o")
			stage, cancelStage := runtimeOracleStageContext(t)
			defer cancelStage()
			ctx, cancel := context.WithTimeout(stage, time.Minute)
			defer cancel()
			observed, err := goasmcmd.Observe(ctx, filepath.Dir(source),
				append(os.Environ(), "GOTOOLCHAIN=local", "GOOS=linux", "GOARCH=386"),
				"go", "tool", "asm", "-o", object, source)
			encoded, encodeErr := json.Marshal(observed)
			if encodeErr != nil {
				t.Fatal(encodeErr)
			}
			t.Logf("ORACLE_CHILD %s", encoded)
			if !observed.Waited || observed.ExitCode != 1 || ctx.Err() != nil ||
				!strings.Contains(string(observed.Stderr), "asmins: illegal in mode 32:") {
				t.Fatalf("actual mode-32 source refusal absent: %+v %v", observed, err)
			}
			additional := func(source, output string) bool {
				return goAssemblerOracleX86SourceDiagnostic(source, "386", output)
			}
			if failure := goasmoracle.CheckOutcome(source, object, observed.Stdout, observed.Stderr,
				err, ctx.Err(), false, 2, additional); failure != nil {
				t.Fatalf("actual original class-tail refusal lost: %v\n%s", failure, observed.Stderr)
			}
		})
	}
}
