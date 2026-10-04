package plan9asm

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestWASMPackedGoRuntimeHelperPreservesNodeGlobalCompatibility(t *testing.T) {
	node := wasmPackedNode(t)
	for _, tc := range []struct {
		name, setup, initial, beforeBootstrap string
		refused                               bool
	}{
		{
			name: "configurable_getter", initial: "original",
			setup: `Object.defineProperty(globalThis, "crypto", {
  get() { return source; }, enumerable: true, configurable: true
});`,
		},
		{
			name: "configurable_readonly_data", initial: "original",
			setup: `Object.defineProperty(globalThis, "crypto", {
  value: source, writable: false, enumerable: true, configurable: true
});`,
		},
		{
			name: "writable_provider", initial: "original",
			setup: `Object.defineProperty(globalThis, "crypto", {
  value: source, writable: true, enumerable: true, configurable: true
});`,
		},
		{
			name: "absent_provider",
			setup: `delete globalThis.crypto;
globalThis.__plan9asmOriginalCrypto = undefined;`,
			// Node installs eval-context builtin getters after --require runs.
			// Establish this fixture's genuinely absent binding immediately
			// before the unchanged bootstrap, not merely before eval startup.
			beforeBootstrap: "delete globalThis.crypto;\n",
		},
		{
			name: "nonconfigurable_getter", refused: true,
			setup: `Object.defineProperty(globalThis, "crypto", {
  get() { return source; }, enumerable: true, configurable: false
});`,
		},
		{
			name: "nonconfigurable_readonly_data", refused: true,
			setup: `Object.defineProperty(globalThis, "crypto", {
  value: source, writable: false, enumerable: true, configurable: false
});`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			dir := filepath.Join(root, "misc", "wasm")
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
			setup := filepath.Join(root, "node setup.cjs")
			files := map[string]string{
				setup: `const source = { provider: "original" };
globalThis.__plan9asmOriginalCrypto = source;
` + tc.setup,
				filepath.Join(dir, "wasm_exec.js"): `module.exports = { fixture: "same-sdk" };`,
				filepath.Join(dir, "wasm_exec_node.js"): `"use strict";
const original = globalThis.crypto;
globalThis.crypto = { provider: "go-runtime" };
console.log(JSON.stringify({
  args: process.argv.slice(2), initial: original ? original.provider : "",
  preserved: original === globalThis.__plan9asmOriginalCrypto,
  provider: globalThis.crypto.provider,
  companion: require("./wasm_exec").fixture
}));`,
			}
			for name, data := range files {
				if err := os.WriteFile(name, []byte(data), 0600); err != nil {
					t.Fatal(err)
				}
			}
			wasm := filepath.Join(root, "Go Program Files", "oracle.wasm")
			ctx, cancel := runtimeOracleStageContext(t)
			defer cancel()
			command, err := wasmPackedGoRuntimeCommand(ctx, root, node, wasm)
			if err != nil {
				t.Fatal(err)
			}
			args := append([]string{"--require", setup}, command.Args[1:]...)
			if tc.beforeBootstrap != "" {
				found := false
				for i := 0; i+1 < len(args); i++ {
					if args[i] == "-e" && args[i+1] == wasmPackedNodeBootstrap {
						args[i+1] = tc.beforeBootstrap + args[i+1]
						found = true
						break
					}
				}
				if !found {
					t.Fatal("fixture cannot bind the actual bootstrap environment")
				}
			}
			output, err := runtimeOracleCombinedOutput(ctx, exec.CommandContext(ctx, node, args...))
			if tc.refused {
				if err == nil || strings.Contains(string(output), `"companion":"same-sdk"`) {
					t.Fatalf("immutable binding must fail before a fake runtime pass: %v\n%s", err, output)
				}
				return
			}
			if err != nil {
				t.Fatalf("actual Node cannot load its selected Go runtime helper: %v\n%s", err, output)
			}
			var result struct {
				Args                         []string
				Initial, Provider, Companion string
				Preserved                    bool
			}
			if err := json.Unmarshal(output, &result); err != nil {
				t.Fatalf("runtime fixture did not execute: %v\n%s", err, output)
			}
			if !reflect.DeepEqual(result.Args, []string{wasm}) || result.Initial != tc.initial ||
				!result.Preserved || result.Provider != "go-runtime" || result.Companion != "same-sdk" {
				t.Fatalf("original provider, Go argv or same-SDK module resolution changed: %+v", result)
			}
		})
	}
}
