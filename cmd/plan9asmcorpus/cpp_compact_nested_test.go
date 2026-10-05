package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestCompactCPPInventoryGeneratedFunctionNamesRequireContext(t *testing.T) {
	source := strings.Join([]string{
		"#define MAKE " + strings.Repeat("\\", 1),
		"#define USE(NAME) " + strings.Repeat("\\", 3),
		"#define NAME GLOBL ·generatedMarker(SB),$8 " + strings.Repeat("\\", 1),
		"",
		"MAKE",
		"USE(NEW)",
		"NEW",
		"TEXT ·Probe(SB),$0-0",
		"RET",
		"",
	}, "\n")
	goBinary, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"linux/386", "linux/amd64", "linux/arm", "linux/arm64", "js/wasm"} {
		t.Run(target, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			observed, err := captureDiscoveryTargetFeatures(ctx, goBinary, t.TempDir(), targetFeatureTestEnv(), target, nil)
			if err != nil {
				t.Fatal(err)
			}
			profile := &discoveryAsmCommandProfile{
				Context: ctx, Observed: observed, GoBinary: goBinary, GoRoot: runtime.GOROOT(),
				PackagePath: "example.invalid/compact-nested", Environment: targetFeatureTestEnv(),
			}
			parts := strings.Split(target, "/")
			binary, args, env, _, err := discoveryAssemblyProfileCommand(profile, parts[0], parts[1], runtime.GOROOT())
			if err != nil {
				t.Fatal(err)
			}
			root := t.TempDir()
			file, object := filepath.Join(root, "original.s"), filepath.Join(root, "original.o")
			writeTestFile(t, file, source)
			args = append(args, "-o", object, file)
			_, _, err = runDiscoveryMachineCommand(ctx, root, env, binary, args...)
			if err != nil {
				t.Fatalf("original nested generator is not Go accepted: %v", err)
			}
			nm, _, err := runDiscoveryMachineCommand(ctx, root, env, binary, "tool", "nm", object)
			if err != nil || !strings.Contains(string(nm), "generatedMarker") {
				t.Fatalf("actual Go did not invoke the generated macro: %s %v", nm, err)
			}
			data, err := os.ReadFile(object)
			if err != nil || len(data) == 0 {
				t.Fatalf("missing actual original object: %v", err)
			}
			receipt, err := json.Marshal(struct {
				Version   string
				Target    string
				Source    string
				SourceSHA string
				Object    []byte
				ObjectSHA string
				Symbols   string
				Arguments []string
			}{observed.GoVersion, target, source, discoveryFeatureBytesSHA256([]byte(source)), data,
				discoveryFeatureBytesSHA256(data), string(nm), args})
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("COMPACT_NESTED_ACTUAL_GO_RECEIPT %s", receipt)
			if inventory, err := captureDiscoveryCompactCPPInventory(map[string][]byte{"module/original.s": []byte(source)}); err == nil {
				t.Fatalf("generated function formal silently omitted later NEW invocation: %+v", inventory)
			}
		})
	}
}
