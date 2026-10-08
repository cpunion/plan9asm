//go:build go1.27

package gostaticstorage

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func indexedForeignArchiveFixture(t *testing.T, arch string) ([]byte, []byte, map[string]SDKReferencedArchive) {
	t.Helper()
	directory, compiler, _ := fixture(t)
	consumer := filepath.Join(directory, "auxfixture")
	if err := os.Mkdir(consumer, 0700); err != nil {
		t.Fatal(err)
	}
	goSource := "package auxfixture\nimport \"sync/atomic\"\nvar Flag atomic.Bool\nfunc Entry()\n"
	assembly := "#include \"textflag.h\"\nTEXT ·Entry(SB),NOSPLIT,$0-0\nRET\n"
	for name, source := range map[string]string{"value.go": goSource, "entry.s": assembly} {
		if err := os.WriteFile(filepath.Join(consumer, name), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	raw := run(t, compiler, consumer, fixtureEnv(t, "linux/"+arch),
		"list", "-deps", "-json", "-compiled", "-export", "-mod=readonly", "-gcflags=-std", "-asmflags=-std", ".")
	decoder := json.NewDecoder(bytes.NewReader(raw))
	referenced := make(map[string]SDKReferencedArchive)
	var archive []byte
	for {
		var selected struct {
			ImportPath, Export string
			SFiles             []string
			Standard           bool
		}
		if err := decoder.Decode(&selected); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		if selected.Export == "" {
			continue
		}
		data, err := os.ReadFile(selected.Export)
		if err != nil {
			t.Fatal(err)
		}
		if selected.ImportPath == fixtureModule+"/auxfixture" {
			if selected.Standard || !reflect.DeepEqual(selected.SFiles, []string{"entry.s"}) {
				t.Fatal("first-party binary format fixture acquired SDK source identity")
			}
			archive = data
		} else {
			if !selected.Standard {
				t.Fatal("actual metadata dependency is not this SDK", selected.ImportPath)
			}
			referenced[selected.ImportPath] = SDKReferencedArchive{selected.ImportPath, selected.SFiles, data}
		}
	}
	if archive == nil || referenced["sync/atomic"].Archive == nil {
		t.Fatal("actual SDK dependency/export route is incomplete")
	}
	builtin, err := os.ReadFile(filepath.Join(filepath.Dir(filepath.Dir(compiler)), "src", "cmd", "internal", "goobj", "builtinlist.go"))
	if err != nil {
		t.Fatal(err)
	}
	retainTextArtifact(t, "linux/"+arch, "indexed-aux", map[string][]byte{
		"original.a": archive, "atomic-original.a": referenced["sync/atomic"].Archive,
		"graph.json": raw, "source-value.go": []byte(goSource), "source-entry.s": []byte(assembly),
	})
	return archive, builtin, referenced
}

func TestGoSDKAssemblyActualAuxOnlyIndexedReference(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		t.Run(arch, func(t *testing.T) {
			archive, builtin, referenced := indexedForeignArchiveFixture(t, arch)
			observed, err := ObserveX86SDKAssemblyArchiveReferences(archive, builtin, "linux/"+arch, []string{"entry.s"}, referenced)
			if err != nil || len(observed) != 3 {
				t.Fatal("actual Aux-only indexed type reference was rejected despite its complete SDK export object", err)
			}
			if _, err := ObserveX86SDKAssemblyArchive(archive, builtin, "linux/"+arch, []string{"entry.s"}); err == nil {
				t.Fatal("SDK observer guessed an unresolved indexed type without actual referenced bytes")
			}
			if _, err := ObserveX86AssemblyArchive(archive, builtin, "linux/"+arch, []string{"entry.s"}); err == nil {
				t.Fatal("default owned observer inherited SDK referenced-object permissions")
			}
		})
	}
}
