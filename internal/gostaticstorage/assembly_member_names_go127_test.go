//go:build go1.27

package gostaticstorage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestGoSDKAssemblyActualTruncatedMemberNames(t *testing.T) {
	directory, compiler, _ := fixture(t)
	oracle := officialReaderForSource(t, compiler, sdkNamespaceReaderSource)
	consumer := filepath.Join(directory, "longmember")
	if err := os.Mkdir(consumer, 0700); err != nil {
		t.Fatal(err)
	}
	name := "long_original_assembly_member.s"
	goSource := "package longmember\nfunc Entry()\n"
	assembly := "#include \"textflag.h\"\nTEXT ·Entry(SB),NOSPLIT,$0-0\nRET\n"
	for file, data := range map[string]string{"value.go": goSource, name: assembly} {
		if err := os.WriteFile(filepath.Join(consumer, file), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	builtin, err := os.ReadFile(filepath.Join(filepath.Dir(filepath.Dir(compiler)), "src", "cmd", "internal", "goobj", "builtinlist.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, arch := range []string{"386", "amd64"} {
		t.Run(arch, func(t *testing.T) {
			target := "linux/" + arch
			env := fixtureEnv(t, target)
			raw := run(t, compiler, consumer, env, "list", "-json", "-compiled", "-export", "-mod=readonly",
				"-gcflags=-std", "-asmflags=-std", ".")
			var selected struct {
				Export   string
				SFiles   []string
				Standard bool
			}
			if err := json.Unmarshal(raw, &selected); err != nil || selected.Standard ||
				!reflect.DeepEqual(selected.SFiles, []string{name}) {
				t.Fatal("actual first-party source role differs", err)
			}
			archive, err := os.ReadFile(selected.Export)
			if err != nil {
				t.Fatal(err)
			}
			originalRaw := run(t, oracle, consumer, env, selected.Export)
			var original []namespaceOracleMember
			if err := json.Unmarshal(originalRaw, &original); err != nil || len(original) != 3 {
				t.Fatal("official original archive inventory differs", err)
			}
			wanted := (strings.TrimSuffix(name, ".s") + ".o")[:16]
			if original[2].Name != wanted || original[2].Flags != 20 {
				t.Fatal("official original byte-truncated member/flags differ", original[2])
			}
			retainTextArtifact(t, target, "indexed-long-member", map[string][]byte{
				"original.a": archive, "namespace-sdk.json": originalRaw, "selection.json": raw,
				"source-value.go": []byte(goSource), "source-assembly.s": []byte(assembly),
			})
			observed, err := ObserveX86SDKAssemblyArchive(archive, builtin, target, selected.SFiles)
			if err != nil || len(observed) != 3 || observed[2].Name != wanted || len(observed[2].Texts) != 1 {
				t.Fatal("exact actual SDK-format long source/member route was rejected", err)
			}
		})
	}
}
