package plan9asm

import "testing"

func TestX86TypedCodeEntryRequiresBoundTargetProfile(t *testing.T) {
	// This role returns an actual code pointer in a source-native uintptr,
	// not a portable numerical constant. Unknown/default layouts and a 32/64
	// pointer-width mismatch must not silently truncate or retype that entry.
	for _, arch := range []string{"amd64", "386"} {
		triples := []string{
			"", "unknown-unknown-unknown", "x86_64-unknown-linux-gnux32",
			"wasm32-unknown-unknown", "wasm64-unknown-unknown", "aarch64-unknown-linux-gnu",
		}
		if arch == "amd64" {
			triples = append(triples, "i386-unknown-linux-gnu", "i686-pc-windows-msvc")
		} else {
			triples = append(triples, "x86_64-unknown-linux-gnu", "x86_64-pc-windows-msvc")
		}
		for _, triple := range triples {
			t.Run(arch+"/"+triple, func(t *testing.T) {
				pkg, source, options := x86TypedCodeEntryFixture(t, arch, "MOV", 0)
				options.TargetTriple = triple
				module, err := TranslateGoModule(pkg, []byte(source), options)
				if module != nil {
					module.Module.Dispose()
				}
				if err == nil {
					t.Fatal("unbound target profile acquired a source-word code pointer")
				}
			})
		}
	}
}
