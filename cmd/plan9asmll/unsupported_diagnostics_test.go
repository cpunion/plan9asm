package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Exercise the real report path. Its hints must describe the same expanded,
// target-selected instructions that failed, not a second raw-source scan.
func TestUnsupportedDiagnosticsUseActualCompilationInput(t *testing.T) {
	repo, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOFLAGS", "")
	t.Setenv("GOWORK", "off")

	cases := []struct {
		name, source, failure string
		ops                   []string
		hits                  int
	}{
		{
			name:    "supported-multiline-macro-context",
			source:  "#define SLEEPCALL \\\nMOVD $101,R8; \\\nSVC; \\\nRET\nTEXT ·F(SB),NOSPLIT,$0-0\nSLEEPCALL\n",
			failure: "instruction probe needs function context",
		},
		{
			name:    "inactive-target-instruction",
			source:  "TEXT ·F(SB),NOSPLIT,$0-0\n#ifdef GOOS_windows\nHIDDENOP R0,R1\n#else\nSVC\n#endif\nRET\n",
			failure: "instruction probe needs function context",
		},
		{
			name:    "actual-expanded-unknown-instruction",
			source:  "#define BROKEN \\\nNOTIMPLEMENTED R0,R1; \\\nRET\nTEXT ·F(SB),NOSPLIT,$0-0\nBROKEN\n",
			failure: "NOTIMPLEMENTED", ops: []string{"NOTIMPLEMENTED"}, hits: 1,
		},
		{
			name:    "parse-failure-has-no-instruction-inventory",
			source:  "TEXT ·F(SB),NOSPLIT,$invalid-0\n#ifdef GOOS_windows\nHIDDENOP R0,R1\n#endif\nRET\n",
			failure: "parse asm:",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, source := range map[string]string{
				"go.mod":        "module example.invalid/opdiagnostic\n\ngo 1.24\n",
				"value.go":      "package opdiagnostic\nfunc F()\n",
				"value_arm64.s": "#include \"textflag.h\"\n" + test.source,
			} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0600); err != nil {
					t.Fatal(err)
				}
			}
			t.Chdir(dir)
			report, _, err := runOneTarget(targetSpec{Goos: "linux", Goarch: "arm64"},
				[]string{"."}, nil, nil, "", t.TempDir(), false, 0, true, false, true, repo, compileConfig{})
			if err != nil || report.TotalAsm != 1 || report.Success != 0 || report.Failed != 1 ||
				report.NotApplicable != 0 || len(report.Fails) != 1 {
				t.Fatalf("real failed compilation was lost: report=%+v err=%v", report, err)
			}
			failure := report.Fails[0]
			if !strings.Contains(failure.Err, test.failure) {
				t.Fatalf("wrong actual failure: %s", failure.Err)
			}
			if !reflect.DeepEqual(failure.Unsupported, test.ops) || len(failure.UnsupportedHits) != test.hits {
				t.Fatalf("hints differ from actual selected instructions: ops=%v hits=%+v", failure.Unsupported, failure.UnsupportedHits)
			}
			for _, hit := range failure.UnsupportedHits {
				if hit.Op != "NOTIMPLEMENTED" || !strings.Contains(hit.Source, "NOTIMPLEMENTED R0,R1") {
					t.Fatalf("macro text or noninstruction became an opcode hint: %+v", hit)
				}
			}
		})
	}
}
