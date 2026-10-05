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

func TestCompactCPPInventoryDriverDefinesReachOriginalInvocations(t *testing.T) {
	cases := []struct {
		name    string
		source  string
		defines []string
		names   []string
	}{
		{
			name:    "driver only",
			source:  "TEXT ·Probe(SB),$0-0\nEMIT\nRET\n",
			defines: []string{"EMIT=GLOBL ·driverMarker(SB),$8"},
			names:   []string{"EMIT"},
		},
		{
			name:   "ordered overwrite and defaults",
			source: "TEXT ·Probe(SB),$0-0\n#ifdef ENABLED\nEMIT\n#endif\nRET\n",
			defines: []string{
				"ENABLED", "EMPTY=", "EMIT=GLOBL ·oldMarker(SB),$8",
				"EMIT=GLOBL ·driverMarker(SB),$8",
			},
			names: []string{"EMIT", "EMPTY", "ENABLED"},
		},
		{
			name:    "literal generated definition",
			source:  "MAKE\nNEW\nTEXT ·Probe(SB),$0-0\nRET\n",
			defines: []string{"MAKE=#define NEW GLOBL ·driverMarker(SB),$8\n"},
			names:   []string{"MAKE", "NEW"},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			for _, target := range []string{"linux/386", "linux/amd64", "linux/arm", "linux/arm64", "js/wasm"} {
				t.Run(target, func(t *testing.T) {
					defines := compactCPPDriverOriginalObject(t, target, test.source, test.defines)
					inventory, err := captureDiscoveryCompactCPPInventory(map[string][]byte{
						"module/original.s": []byte(test.source),
					})
					if err != nil {
						t.Fatal(err)
					}
					for _, name := range test.names {
						if !compactCPPDriverHasName(inventory.InvocationNames, name) {
							t.Fatalf("actual driver seed %q disappeared from raw name closure: names=%v defines=%q", name, inventory.InvocationNames, defines)
						}
					}
					if len(inventory.Sources["module/original.s"].Records) == 0 {
						t.Fatal("actual driver-only invocation disappeared from original source inventory")
					}
				})
			}
		})
	}
}

func compactCPPDriverHasName(names []string, wanted string) bool {
	for _, name := range names {
		if name == wanted {
			return true
		}
	}
	return false
}

// This compiles first-party original assembly only. Exact ordered -D inputs,
// source bytes and actual object bytes are retained independently of the query.
func compactCPPDriverOriginalObject(t *testing.T, target, source string, explicit []string) []string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	goBinary, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	observed, err := captureDiscoveryTargetFeatures(ctx, goBinary, root, targetFeatureTestEnv(), target, nil)
	if err != nil {
		t.Fatal(err)
	}
	profile := &discoveryAsmCommandProfile{
		Context: ctx, Observed: observed, GoBinary: goBinary, GoRoot: runtime.GOROOT(),
		PackagePath: "example.invalid/compact-driver", Environment: targetFeatureTestEnv(),
	}
	parts := strings.Split(target, "/")
	binary, args, env, _, err := discoveryAssemblyProfileCommand(profile, parts[0], parts[1], runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	for _, definition := range explicit {
		args = append(args, "-D", definition)
	}
	defines := []string{}
	for index := 0; index < len(args); index++ {
		if args[index] == "-D" {
			if index+1 == len(args) {
				t.Fatal("actual assembler profile has an incomplete -D argument")
			}
			index++
			defines = append(defines, args[index])
		}
	}
	file, object := filepath.Join(root, "original.s"), filepath.Join(root, "original.o")
	writeTestFile(t, file, source)
	args = append(args, "-o", object, file)
	_, _, err = runDiscoveryMachineCommand(ctx, root, env, binary, args...)
	if err != nil {
		t.Fatalf("original explicit -D source is not Go accepted: %v", err)
	}
	nm, _, err := runDiscoveryMachineCommand(ctx, root, env, binary, "tool", "nm", object)
	if err != nil || !strings.Contains(string(nm), "compact-driver.driverMarker") || strings.Contains(string(nm), "oldMarker") {
		t.Fatalf("original driver macro invocation/overwrite was not observed: %s %v", nm, err)
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
		Defines   []string
		Object    []byte
		ObjectSHA string
		Symbols   string
		Arguments []string
	}{observed.GoVersion, target, source, discoveryFeatureBytesSHA256([]byte(source)), defines,
		data, discoveryFeatureBytesSHA256(data), string(nm), args})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("COMPACT_DRIVER_ACTUAL_GO_RECEIPT %s", receipt)
	return defines
}
