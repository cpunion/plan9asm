package plan9asm

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestWASMPackedGoRuntimeHelperUsesPortableNodeCommand(t *testing.T) {
	for _, directory := range []string{"lib", "misc"} {
		t.Run(directory, func(t *testing.T) {
			root := t.TempDir()
			dir := filepath.Join(root, directory, "wasm")
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"wasm_exec_node.js", "wasm_exec.js"} {
				// JavaScript files do not require POSIX execute permission,
				// including in a Windows GOROOT using node.exe.
				if err := os.WriteFile(filepath.Join(dir, name), []byte("// helper fixture\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			node := filepath.Join(root, "Node Program Files", "node.exe")
			wasm := filepath.Join(root, "Go Program Files", "oracle.wasm")
			cmd, err := wasmPackedGoRuntimeCommand(root, node, wasm)
			if err != nil {
				t.Fatal(err)
			}
			want := []string{node, "--stack-size=8192", filepath.Join(dir, "wasm_exec_node.js"), wasm}
			if cmd.Path != node || !reflect.DeepEqual(cmd.Args, want) {
				t.Fatalf("must invoke Node directly with separate arguments, got path=%q args=%q", cmd.Path, cmd.Args)
			}
		})
	}
}

func TestWASMPackedGoRuntimeHelperRequiresCompleteRegularFiles(t *testing.T) {
	for _, invalid := range []string{"missing", "node_directory", "missing_companion", "companion_directory"} {
		t.Run(invalid, func(t *testing.T) {
			root := t.TempDir()
			dir := filepath.Join(root, "lib", "wasm")
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
			if invalid == "node_directory" {
				if err := os.Mkdir(filepath.Join(dir, "wasm_exec_node.js"), 0700); err != nil {
					t.Fatal(err)
				}
			} else if invalid != "missing" {
				if err := os.WriteFile(filepath.Join(dir, "wasm_exec_node.js"), nil, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if invalid == "companion_directory" {
				if err := os.Mkdir(filepath.Join(dir, "wasm_exec.js"), 0700); err != nil {
					t.Fatal(err)
				}
			}
			if cmd, err := wasmPackedGoRuntimeCommand(root, "node.exe", "oracle.wasm"); err == nil || cmd != nil {
				t.Fatalf("incomplete runtime must fail without a fake command: cmd=%v err=%v", cmd, err)
			}
		})
	}
}
