package plan9asm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func arm64NativeTargetPoolSource(t *testing.T, payload []uint32) string {
	t.Helper()
	// Actual code loads eight data bytes, kills its pool address and returns.
	// The data deliberately resembles every preflight family and invalid code.
	words := assembleARM64LLVMWords(t, []string{
		"adr x9,#16", "ldr x0,[x9]", "mov x9,xzr", "ret",
	}, "")
	var source strings.Builder
	source.WriteString("TEXT comparison(SB),516,$0-0\n")
	for _, word := range append(words, payload...) {
		fmt.Fprintf(&source, "WORD $%#08x\n", word)
	}
	source.WriteString("RET\n")
	return source.String()
}

func TestARM64NativeTargetPreflightPreservesProvenPoolData(t *testing.T) {
	payload := []uint32{
		0x9e620000, 0x1e602000, 0x1e600400, 0x1e602800, 0x1e604800,
		0xd53b4400, 0xd53b2400, 0x1eaf9499, 0xffffffff, 0x14000000,
	}
	source := arm64NativeTargetPoolSource(t, payload)
	requireARM64GoAssemblerResult(t, source, true)
	file, err := Parse(ArchARM64, source)
	if err != nil {
		t.Fatal(err)
	}
	original, err := Parse(ArchARM64, source)
	if err != nil {
		t.Fatal(err)
	}
	_, blobs, err := prepareARM64RawPCRelative(file.Funcs[0])
	if err != nil || len(blobs) != 1 || len(blobs[0].bytes) != 4*len(payload) {
		t.Fatalf("independent existing pool proof lost the fixture: blobs=%v err=%v", blobs, err)
	}
	for _, target := range []string{"", "x86_64-unknown-linux-gnu", "aarch64-unknown-linux-gnu"} {
		if err := validateARM64NativeTargets(file, Options{Goarch: "arm64", TargetTriple: target}); err != nil {
			t.Errorf("data acquired an instruction/target requirement for %q: %v", target, err)
		}
	}
	if !reflect.DeepEqual(file, original) {
		t.Fatal("preflight changed the original source consumed by later proofs")
	}
	for _, target := range []string{"aarch64-unknown-linux-gnu", "aarch64-apple-darwin", "aarch64-pc-windows-msvc"} {
		for _, api := range []string{"text", "module", "Go", "public-module"} {
			present, err := arm64FPControlTargetAPI(t, api, source, target)
			if !present || err != nil {
				t.Errorf("%s %s rejected proven pool data: %v", target, api, err)
			}
		}
	}
}

func TestARM64NativeTargetPreflightCannotHideExecutablePoolWords(t *testing.T) {
	// A branch into the apparent pool makes it code. A dead component after
	// RET followed by another RET is likewise not a proven literal island.
	for _, prefix := range []string{"", "RET\n", "B code\ncode:\n"} {
		for _, word := range []uint32{0x1e602000, 0x1eaf9499} {
			source := fmt.Sprintf("TEXT comparison(SB),516,$0-0\n%sWORD $%#08x\nRET\n", prefix, word)
			for _, api := range []string{"text", "module", "Go", "public-module"} {
				present, err := arm64FPControlTargetAPI(t, api, source, "")
				if present || err == nil || errors.Is(err, ErrProbeNeedsContext) != (word == 0x1e602000) {
					t.Errorf("executable WORD lost its grammar/target gate: %s %s: %v", api, source, err)
				}
			}
		}
	}
}
