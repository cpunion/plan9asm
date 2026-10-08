//go:build go1.27

package gostaticstorage

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func indexedNamespaceObjectFixture(t *testing.T, arch string) ([]byte, []TextReference) {
	t.Helper()
	directory, compiler, _ := fixture(t)
	goSource := `package staticstorage
import "sync"
var Value byte
var First = sync.OnceFunc(func() { Value++ })
var Second = sync.OnceFunc(func() { Value += 2 })
func Read() byte { return Value }
`
	if err := os.WriteFile(filepath.Join(directory, "storage.go"), []byte(goSource), 0600); err != nil {
		t.Fatal(err)
	}
	raw := run(t, compiler, directory, fixtureEnv(t, "linux/"+arch),
		"list", "-json", "-compiled", "-export", "-mod=readonly", "-gcflags=-std", ".")
	var selected struct{ Export string }
	if err := json.Unmarshal(raw, &selected); err != nil || selected.Export == "" {
		t.Fatal("actual namespace fixture export differs", err)
	}
	archive, err := os.ReadFile(selected.Export)
	if err != nil {
		t.Fatal(err)
	}
	members, err := archiveMembers(archive)
	if err != nil || len(members) != 2 {
		t.Fatal("actual namespace fixture members differ", err)
	}
	body, err := objectBody(members[1].data, "linux/"+arch)
	if err != nil {
		t.Fatal(err)
	}
	builtin, err := os.ReadFile(filepath.Join(filepath.Dir(filepath.Dir(compiler)), "src", "cmd", "internal", "goobj", "builtinlist.go"))
	if err != nil {
		t.Fatal(err)
	}
	builtins, err := textBuiltins(builtin)
	if err != nil {
		t.Fatal(err)
	}
	return body, builtins
}

func TestGoSDKAssemblyNamespaceBinaryBoundaries(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		t.Run(arch, func(t *testing.T) {
			original, builtins := indexedNamespaceObjectFixture(t, arch)
			view, err := checkedObject(original, 16, false)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := readTextObject(original, "linux/"+arch, 16, false, builtins); err != nil {
				t.Fatal("actual original namespace object failed precondition", err)
			}
			blockOffset := func(block int) int { return int(binary.LittleEndian.Uint32(original[20+4*block:])) }
			counts := [5]int{}
			for namespace := range counts {
				counts[namespace] = len(view.block(namespace+3)) / 21
				if counts[namespace] == 0 {
					t.Fatal("actual fixture did not populate namespace", namespace)
				}
			}
			for block := 3; block <= 7; block++ {
				t.Run("truncated-symbol-block-"+string(rune('0'+block)), func(t *testing.T) {
					changed := append([]byte(nil), original...)
					binary.LittleEndian.PutUint32(changed[20+4*(block+1):], uint32(blockOffset(block+1)-1))
					if _, err := readTextObject(changed, "linux/"+arch, 16, false, builtins); !errors.Is(err, ErrContext) {
						t.Fatal("truncated original namespace acquired observation", err)
					}
				})
			}
			for block := 4; block <= 7; block++ {
				t.Run("namespace-crossover-"+string(rune('0'+block)), func(t *testing.T) {
					changed := append([]byte(nil), original...)
					binary.LittleEndian.PutUint32(changed[20+4*block:], uint32(blockOffset(block)+21))
					if _, err := readTextObject(changed, "linux/"+arch, 16, false, builtins); !errors.Is(err, ErrContext) {
						t.Fatal("changed namespace boundary acquired another index identity", err)
					}
				})
			}
			for _, block := range []int{9, 10} {
				t.Run("hash-inventory-"+string(rune('0'+block)), func(t *testing.T) {
					changed := append([]byte(nil), original...)
					width := 8
					if block == 10 {
						width = 16
					}
					binary.LittleEndian.PutUint32(changed[20+4*(block+1):], uint32(blockOffset(block+1)-width))
					if _, err := readTextObject(changed, "linux/"+arch, 16, false, builtins); !errors.Is(err, ErrContext) {
						t.Fatal("missing hash entry acquired an indexed definition", err)
					}
				})
			}
			if len(view.block(14)) < 23 {
				t.Fatal("actual original lacks a relocation boundary witness")
			}
			for _, reference := range []struct {
				name  string
				pkg   uint32
				bound uint32
			}{
				{"self", 2147483643, uint32(counts[0])},
				{"hashed64", 2147483646, uint32(counts[1])},
				{"hashed", 2147483645, uint32(counts[2])},
				{"nonpkg-def-ref", 2147483647, uint32(counts[3] + counts[4])},
				{"builtin", 2147483644, uint32(len(builtins))},
				{"foreign-package", uint32(len(view.block(1)) / 8), 0},
				{"nil-nonzero", 0, 1},
				{"overflow", 2147483645, math.MaxUint32},
			} {
				t.Run("reference-"+reference.name, func(t *testing.T) {
					changed := append([]byte(nil), original...)
					offset := blockOffset(14) + 15
					binary.LittleEndian.PutUint32(changed[offset:], reference.pkg)
					binary.LittleEndian.PutUint32(changed[offset+4:], reference.bound)
					if _, err := readTextObject(changed, "linux/"+arch, 16, false, builtins); !errors.Is(err, ErrContext) {
						t.Fatal("out-of-bound namespace reference acquired a different symbol", err)
					}
				})
			}
		})
	}
}

func TestGoSDKAssemblyNamedCollisionPermissions(t *testing.T) {
	text := indexedObjectSymbol{Namespace: TextSymbolNonpkgDefinition, Name: "fixture.Entry", ABI: 0, Kind: 1}
	for _, control := range []struct {
		name    string
		first   indexedObjectSymbol
		second  indexedObjectSymbol
		allowed bool
	}{
		{"ordinary named duplicate", text, text, false},
		{"different ABI TEXT", text, func() indexedObjectSymbol { v := text; v.ABI = 1; return v }(), true},
		{"DUPok observations", func() indexedObjectSymbol { v := text; v.Flags = 1; return v }(), text, true},
		{"storage collision", text, func() indexedObjectSymbol { v := text; v.Kind = 4; return v }(), false},
		{"hashed same name", text, func() indexedObjectSymbol { v := text; v.Namespace = TextSymbolHashed; return v }(), true},
		{"static same name", text, func() indexedObjectSymbol { v := text; v.ABI = 65535; return v }(), true},
	} {
		t.Run(control.name, func(t *testing.T) {
			definitions := make(sdkAssemblyNamedDefinitions)
			if err := definitions.observe(control.first, false); err != nil {
				t.Fatal(err)
			}
			err := definitions.observe(control.second, false)
			if (err == nil) != control.allowed {
				t.Fatal("named collision rule acquired or lost observation permission", err)
			}
			if control.allowed && control.second.Namespace == TextSymbolNonpkgDefinition && control.second.ABI != 65535 {
				if !reflect.DeepEqual(definitions[text.Name], []indexedObjectSymbol{control.first, control.second}) {
					t.Fatal("named observations selected or discarded a body winner")
				}
			}
		})
	}
	definitions := make(sdkAssemblyNamedDefinitions)
	data := text
	data.Namespace, data.Kind = TextSymbolHashed, 4
	if !errors.Is(definitions.observe(data, true), ErrContext) {
		t.Fatal("a hashed assembly DATA definition bypassed its independent role")
	}
}
