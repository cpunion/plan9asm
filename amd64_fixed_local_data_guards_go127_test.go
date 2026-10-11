//go:build go1.27

package plan9asm

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoAMD64FixedLocalReadonlyDataConsumerIdentity(t *testing.T) {
	t.Setenv("GOCACHE", t.TempDir())
	pkg, request, file, options, assembly := fixedLocalDataFixture(t,
		"MOVOU literal<>(SB),X0\nPADDQ 16(SP),X0", fixedLocalLiteralData)
	fixedLocalDataAPIs(t, pkg, file, assembly, options, fixedLocalRequireContext)
	source := requireGoAMD64FrameCapture(t, pkg, request)
	ctx, cancel := runtimeOracleStageContext(t)
	defer cancel()
	bound, err := source.BindFixedLocal(ctx, pkg, file, options)
	if err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []string{"resolver", "JSON", "byte", "extent", "flags", "missing DATA", "new DATA"} {
		t.Run(mutation, func(t *testing.T) {
			changed, input := bound, assembly
			switch mutation {
			case "resolver":
				changed.ResolveSym = func(symbol string) string {
					if symbol == "·literal<>" {
						return "foreign.literal<>"
					}
					return options.ResolveSym(symbol)
				}
			case "JSON":
				changed.Sigs, err = cloneFixedLocalSigs(bound.Sigs)
				if err != nil {
					t.Fatal(err)
				}
			case "byte":
				input = strings.Replace(input, "0x0102030405060708", "0x0102030405060709", 1)
			case "extent":
				input = strings.Replace(input, ",8,$32", ",8,$64", 1)
			case "flags":
				input = strings.Replace(input, ",8,$32", ",16,$32", 1)
			case "missing DATA":
				input = strings.TrimPrefix(input, fixedLocalLiteralData)
			case "new DATA":
				input = "GLOBL extra<>(SB),8,$8\n" + input
			}
			parsed, err := Parse(ArchAMD64, input)
			if err != nil {
				t.Fatal(err)
			}
			fixedLocalDataAPIs(t, pkg, parsed, input, changed, fixedLocalRequireContext)
		})
	}
	// A callback's later answer cannot replace the already verified symbol.
	t.Run("stateful", func(t *testing.T) {
		calls := 0
		changed := bound
		changed.ResolveSym = func(symbol string) string {
			if symbol == "·literal<>" {
				calls++
				if calls > 1 {
					return "foreign.literal<>"
				}
			}
			return options.ResolveSym(symbol)
		}
		// Each API gets a fresh callback counter.
		fixedLocalDataAPIs(t, pkg, file, assembly, changed, func(t *testing.T, ir string, err error) {
			if err != nil || strings.Contains(ir, "foreign.literal") || calls != 1 {
				t.Fatal("mutable resolver escaped sealed literal identity", calls, err)
			}
			calls = 0
		})
	})
	path := filepath.Join(request.Directory, request.AssemblyFile)
	changed := strings.Replace(assembly, "0x0102030405060708", "0x0102030405060709", 1)
	if err := os.WriteFile(path, []byte(changed), 0600); err != nil {
		t.Fatal(err)
	}
	fixedLocalDataAPIs(t, pkg, file, assembly, bound, fixedLocalRequireContext)
	if err := os.WriteFile(path, []byte(assembly), 0600); err != nil {
		t.Fatal(err)
	}
	fixedLocalDataAPIs(t, pkg, file, assembly, bound, fixedLocalRequireContext)
}

func TestGoAMD64FixedLocalReadonlyDataRejectsCoalescedDefinitions(t *testing.T) {
	t.Setenv("GOCACHE", t.TempDir())
	// Go's bare and qualified static names are distinct definitions. The
	// ordinary LLVM resolver cannot fold them into one global in a binding.
	data := fixedLocalLiteralData + "DATA ·literal<>+0(SB)/8,$99\nGLOBL ·literal<>(SB),8,$32\n"
	pkg, request, file, options, assembly := fixedLocalDataFixture(t,
		"MOVOU literal<>(SB),X0\nPADDQ 16(SP),X0", data)
	x86LEAActualTargetGoObject(t, "amd64", options.TargetTriple, assembly)
	source := requireGoAMD64FrameCapture(t, pkg, request)
	ctx, cancel := runtimeOracleStageContext(t)
	defer cancel()
	if _, err := source.BindFixedLocal(ctx, pkg, file, options); !errors.Is(err, ErrProbeNeedsContext) {
		t.Fatal("distinct original globals acquired one LLVM identity", err)
	}
}
