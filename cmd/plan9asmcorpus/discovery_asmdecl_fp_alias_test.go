package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDiscoveryAsmDeclFPAliasActualNamedOffsets(t *testing.T) {
	for _, fixture := range []struct {
		name, declaration, body, args string
	}{
		{"byte", "func Value(x, y uint8) uint8", "MOVB x+1(FP),AX\nMOVB AX,y+8(FP)", "9"},
		{"word", "func Value(x, y uint16) uint16", "MOVW x+2(FP),AX\nMOVW AX,y+8(FP)", "10"},
		{"long", "func Value(x, y uint32) uint32", "MOVL x+4(FP),AX\nMOVL AX,y+8(FP)", "12"},
		{"quad", "func Value(x, y uint64) uint64", "MOVQ x+8(FP),AX\nMOVQ AX,y+16(FP)", "24"},
		{"float", "func Value(x, y float32) float32", "MOVSS x+4(FP),X0\nMOVSS X0,y+8(FP)", "12"},
		{"double", "func Value(x, y float64) float64", "MOVSD x+8(FP),X0\nMOVSD X0,y+16(FP)", "24"},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			assembly := "TEXT ·Value(SB),4,$0-" + fixture.args + "\n" + fixture.body + "\nRET\n"
			root, env := discoveryAsmDeclResultAliasFixture(t, fixture.declaration, assembly)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			plan, err := captureDiscoveryAsmDeclDirect(ctx, root, env, nil, []string{"."}, "")
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := plan.close(); err != nil {
					t.Error(err)
				}
			}()
			if len(plan.Units) != 1 {
				t.Fatalf("actual units: %d", len(plan.Units))
			}
			unit := plan.Units[0]
			_, original := runCapturedCommandOutput(ctx, plan.Dir, plan.Env, plan.GoBinary, "tool", "vet", "-asmdecl", unit.Filename)
			if _, observed := discoveryAsmDeclObservedFailure(original); !observed ||
				!strings.Contains(discoveryCommandDiagnostic(original), "invalid offset x+") ||
				!strings.Contains(discoveryCommandDiagnostic(original), "RET without writing") {
				t.Fatalf("actual original offset/name diagnostics missing: %v", original)
			}
			t.Logf("ORIGINAL_FP_NAME_DIAGNOSTIC: %v", original)
			if err := plan.filterResultAliasWarnings(unit, original); err != nil {
				t.Fatalf("exact owned FP bytes with complete typed result were rejected: %v", err)
			}
			if err := runDiscoveryAsmDecl(ctx, root, env, "linux/amd64", nil, []string{"."}); err != nil {
				t.Fatalf("full actual producer still rejected semantic FP aliases: %v", err)
			}
		})
	}
}

func TestDiscoveryAsmDeclFPAliasActualObjectAndCaller(t *testing.T) {
	declaration := "func Value(x, y uint64) uint64"
	alias := "TEXT ·Value(SB),4,$0-24\nMOVQ x+8(FP),AX\nMOVQ AX,y+16(FP)\nRET\n"
	canonical := "TEXT ·Value(SB),4,$0-24\nMOVQ y+8(FP),AX\nMOVQ AX,ret+16(FP)\nRET\n"
	root, env := discoveryAsmDeclResultAliasFixture(t, declaration, alias)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	writeTestFile(t, filepath.Join(root, "canonical.s"), canonical)
	var listings []string
	for _, name := range []string{"decl_amd64.s", "canonical.s"} {
		output, err := runCapturedCommandOutput(ctx, root, env, "go", "tool", "asm", "-S", "-p", "example.invalid/result-alias", "-o", filepath.Join(root, name+".o"), name)
		if err != nil {
			t.Fatal(err)
		}
		var bytes []string
		for _, line := range strings.Split(string(output), "\n") {
			if strings.HasPrefix(line, "\t0x") && !strings.Contains(line, "(") {
				bytes = append(bytes, line)
			}
		}
		if len(bytes) == 0 {
			t.Fatalf("actual original Go assembler has no object bytes: %s", output)
		}
		listings = append(listings, strings.Join(bytes, "\n"))
		t.Logf("ORIGINAL_GO_OBJECT %s:\n%s", name, output)
	}
	if listings[0] != listings[1] {
		t.Fatalf("names affected actual Go machine bytes:\nalias %s\ncanonical %s", listings[0], listings[1])
	}
	// Keep only the original alias assembly selected for the actual typed Go
	// caller. The canonical source is an object control, not a replacement.
	writeTestFile(t, filepath.Join(root, "canonical.s"), "//go:build ignore\n\n"+canonical)
	writeTestFile(t, filepath.Join(root, "caller_test.go"), `package alias
import "testing"
func TestOriginalFPAddress(t *testing.T) {
	values := []uint64{0, 1, 0x8000000000000000, 0xffffffffffffffff}
	for _, x := range values {
		for _, y := range values {
			left, right := uint64(0x0123456789abcdef), uint64(0xfedcba9876543210)
			got := Value(x, y)
			if got != y || left != 0x0123456789abcdef || right != 0xfedcba9876543210 {
				t.Fatalf("x=%x y=%x result=%x", x, y, got)
			}
		}
	}
}
`)
	callerEnv := replaceEnv(env, map[string]string{"GOOS": "darwin", "GOARCH": "amd64"})
	output, err := runCapturedCommandOutput(ctx, root, callerEnv, "go", "test", "-vet=off", "-count=1", "-run", "^TestOriginalFPAddress$", ".")
	if err != nil {
		t.Fatalf("original unchanged assembly typed caller failed: %v", err)
	}
	t.Logf("ORIGINAL_TYPED_GO_CALLER (Darwin AMD64 execution): %s", output)
}
