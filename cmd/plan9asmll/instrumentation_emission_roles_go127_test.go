//go:build go1.27

package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm/internal/gotoolprofile"
	"golang.org/x/tools/go/packages"
)

func TestInstrumentationAssemblyEmissionRolesActualObjects(t *testing.T) {
	for _, test := range []struct {
		name, source, emission string
	}{
		{"comment-only", "// Original nonzero selected source without symbols.\n", "symbol_free"},
		{"macro-only", "#define UNUSED(X) MOVD X, R0\n", "symbol_free"},
		{"inactive-TEXT", "#ifdef NEVER_DEFINED\nTEXT ·Probe(SB),$0-8\nMOVD $41,R0\nMOVD R0,ret+0(FP)\nRET\n#endif\n", "symbol_free"},
		{"zero-byte", "", "symbol_free"},
		{"DATA-only", "#include \"textflag.h\"\nDATA ·Global+0(SB)/8,$41\nGLOBL ·Global(SB),RODATA|NOPTR,$8\n", "data_only"},
		{"TEXT-and-DATA", "#include \"textflag.h\"\nTEXT ·Probe(SB),NOSPLIT,$0-8\nMOVD $41,R0\nMOVD R0,ret+0(FP)\nRET\nDATA ·Global+0(SB)/8,$41\nGLOBL ·Global(SB),RODATA|NOPTR,$8\n", "assembly"},
	} {
		t.Run(test.name, func(t *testing.T) {
			input, ctx := instrumentationMetadataFixture(t, "darwin/arm64", gotoolprofile.InstrumentationRace, "1")
			input.Original.SignatureSourceProtocol = gotoolprofile.SignatureSourceProtocol
			const name = "probe_arm64.s"
			if err := os.WriteFile(filepath.Join(input.Original.SourceRoot, name), []byte(test.source), 0600); err != nil {
				t.Fatal(err)
			}
			input.Original.Sources[name] = featureBytesSHA256([]byte(test.source))
			input.Original.Headers[name] = test.source
			toolName := "pkg/include/textflag.h"
			toolBytes, err := os.ReadFile(filepath.Join(runtime.GOROOT(), toolName))
			if err != nil {
				t.Fatal(err)
			}
			input.Original.ToolSources = map[string]string{toolName: featureBytesSHA256(toolBytes)}
			replanInstrumentationMetadataFixture(t, input)
			producer, err := captureInstrumentationMetadata(ctx, input, input.Original.SourceRoot, os.Environ(), t.TempDir(), packages.Load)
			if err != nil {
				t.Fatal(err)
			}
			// Independent original stock-Go acceptance is not LLVM proof. In
			// particular, no-TEXT parsing cannot stand in for this real object.
			binary, err := exec.LookPath("go")
			if err != nil {
				t.Fatal(err)
			}
			object := filepath.Join(t.TempDir(), "original.o")
			args := []string{"tool", "asm", "-S", "-p", input.Plan.PackagePath,
				"-I", input.Original.SourceRoot, "-I", filepath.Join(producer.Source.Root, "pkg", "include"), "-o", object}
			for _, define := range producer.Metadata.Selection.Packages[0].Macros.Defines {
				args = append(args, "-D", define)
			}
			args = append(args, name)
			listing, stderr, err := gotoolprofile.RunBounded(ctx, input.Original.SourceRoot, producer.Source.Env, binary, args...)
			if err != nil || len(stderr) != 0 {
				t.Fatalf("original same-mode assembler: %v; %s", err, stderr)
			}
			if stat, err := os.Lstat(object); err != nil || !stat.Mode().IsRegular() || stat.Size() == 0 {
				t.Fatalf("original assembler produced no regular nonempty object: %v", err)
			}
			if test.emission == "symbol_free" && len(listing) != 0 {
				t.Fatalf("symbol-free original has a native listing: %s", listing)
			}
			if test.emission == "data_only" {
				if _, err := gotoolprofile.DataOnlySymbolsFromListing(listing); err != nil {
					t.Fatal(err)
				}
			}
			t.Logf("actual original same-mode source=%s listingSHA256=%s object=%s", name, featureBytesSHA256(listing), object)
			output := instrumentationObjectFixtureOutput(t, input)
			proof, tools, err := consumeInstrumentationAssembly(ctx, input, producer.Metadata, input.Original.SourceRoot, os.Environ(), output)
			if err != nil {
				t.Fatalf("actual accepted %s role must independently compile an LLVM object: %v", test.emission, err)
			}
			if len(proof.Files) != 1 || proof.Files[0].CPP.Emission != test.emission || proof.Files[0].Output.Object == "" {
				t.Fatalf("emission role/object lost: %#v", proof)
			}
			recordInstrumentationObjectFixture(t, output, input, producer.Metadata, tools, proof)
			data, err := json.Marshal(proof)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := gotoolprofile.DecodeInstrumentationConsumption(strings.NewReader(string(data)), input, producer.Metadata, tools, ".", nil); err != nil {
				t.Fatal(err)
			}
			for name, mutate := range map[string]func(*gotoolprofile.InstrumentationConsumptionProof){
				"old protocol": func(p *gotoolprofile.InstrumentationConsumptionProof) {
					p.Protocol = "actual_go_instrumentation_llvm_objects_v2"
				},
				"deleted object": func(p *gotoolprofile.InstrumentationConsumptionProof) { p.Files[0].Output.Object = "" },
				"paired role deletion": func(p *gotoolprofile.InstrumentationConsumptionProof) {
					p.Files[0].CPP.Emission = "assembly"
					p.Files[0].CPP.EmptyAssembly = nil
					p.Files[0].CPP.DataOnlyAssembly = nil
				},
				"contradictory role": func(p *gotoolprofile.InstrumentationConsumptionProof) { p.Files[0].CPP.Emission = "unknown" },
			} {
				if name == "paired role deletion" && test.emission == "assembly" {
					continue // This unchanged normal role is the legitimate control.
				}
				t.Run(name, func(t *testing.T) {
					var changed gotoolprofile.InstrumentationConsumptionProof
					if err := json.Unmarshal(data, &changed); err != nil {
						t.Fatal(err)
					}
					mutate(&changed)
					if err := gotoolprofile.ValidateInstrumentationConsumption(input, producer.Metadata, tools, ".", nil, &changed); err == nil {
						t.Fatal("changed native emission role/object or old consumer was accepted")
					}
				})
			}
		})
	}
}
